package securityanalytics

import (
    "fmt"
    "sync"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    "gorm.io/datatypes"
)

// Engine orchestrates anomaly detection across users
type Engine struct {
    detectors         []AnomalyDetector
    processingTimeout time.Duration
    mutex             sync.RWMutex
    running           bool
}

// NewEngine creates a new anomaly detection engine with default detectors
func NewEngine() *Engine {
    quotaSpikeThreshold := GetConfiguredThreshold("quota_spike_percent", 150.0)
    loginRatioThreshold := GetConfiguredThreshold("login_ratio_threshold", 1000.0)
    requestRatioThreshold := GetConfiguredThreshold("request_ratio_threshold", 500.0)
    newDeviceThreshold := int(GetConfiguredThreshold("new_device_requests", 100.0))
    ipChangeThreshold := int(GetConfiguredThreshold("ip_change_threshold", 5.0))

    engine := &Engine{
        processingTimeout: 5 * time.Minute,
        running:           false,
        detectors: []AnomalyDetector{
            NewQuotaSpikeDetector(quotaSpikeThreshold),
            NewAbnormalLoginRatioDetector(loginRatioThreshold),
            NewHighRequestRatioDetector(requestRatioThreshold),
            NewUnusualDeviceActivityDetector(newDeviceThreshold, ipChangeThreshold),
        },
    }

    return engine
}

// ProcessUser analyzes a specific user for anomalies
func (e *Engine) ProcessUser(userId int, windowDuration time.Duration) (anomalies []*model.SecurityAnomaly, err error) {
    e.mutex.RLock()
    detectors := make([]AnomalyDetector, len(e.detectors))
    copy(detectors, e.detectors)
    e.mutex.RUnlock()

    // Prepare detection context
    endTime := time.Now()
    startTime := endTime.Add(-windowDuration)

    ctx, err := e.buildDetectionContext(userId, startTime, endTime)
    if err != nil {
        return nil, fmt.Errorf("failed to build detection context: %w", err)
    }

    // Run all detectors
    for _, detector := range detectors {
        result, err := detector.Detect(ctx)
        if err != nil {
            common.SysLog(fmt.Sprintf("detector %s failed: %v", detector.GetRuleType(), err))
            continue
        }

        if result.Detected {
            anomaly, err := e.createAnomalyRecord(userId, result, ctx)
            if err != nil {
                common.SysLog(fmt.Sprintf("failed to create anomaly record: %v", err))
                continue
            }

            // Check for duplicates
            if !e.isDuplicate(anomaly) {
                anomalies = append(anomalies, anomaly)
            }
        }
    }

    return anomalies, nil
}

// ProcessBatch analyzes multiple users for anomalies
func (e *Engine) ProcessBatch(userIds []int, windowDuration time.Duration, concurrency int) (map[int][]*model.SecurityAnomaly, error) {
    results := make(map[int][]*model.SecurityAnomaly)
    var mutex sync.Mutex

    // Use semaphore for concurrency control
    sem := make(chan struct{}, concurrency)
    var wg sync.WaitGroup

    for _, userId := range userIds {
        wg.Add(1)
        go func(uid int) {
            defer wg.Done()

            sem <- struct{}{}        // acquire
            defer func() { <-sem }() // release

            anomalies, err := e.ProcessUser(uid, windowDuration)
            if err != nil {
                common.SysLog(fmt.Sprintf("failed to process user %d: %v", uid, err))
                return
            }

            if len(anomalies) > 0 {
                mutex.Lock()
                results[uid] = anomalies
                mutex.Unlock()
            }
        }(userId)
    }

    wg.Wait()
    return results, nil
}

// buildDetectionContext gathers all necessary data for anomaly detection
func (e *Engine) buildDetectionContext(userId int, startTime, endTime time.Time) (*DetectionContext, error) {
    ctx := &DetectionContext{
        UserId:               userId,
        WindowStartTime:      startTime,
        WindowEndTime:        endTime,
        TimeWindow:           endTime.Sub(startTime),
        LatestBaseline:       make(map[string]*model.AnomalyBaseline),
        DeviceAggregations:   make([]*DeviceAggregationResult, 0),
        IPAggregations:       make([]*IPAggregationWindow, 0),
        ConversationLinkages: make([]*ConversationLinkageResult, 0),
    }

    // Get request logs
    var logs []*model.Log
    err := model.LOG_DB.
        Where("user_id = ? AND created_at >= ? AND created_at <= ?", userId, startTime.Unix(), endTime.Unix()).
        Find(&logs).Error
    if err != nil {
        return nil, fmt.Errorf("failed to fetch logs: %w", err)
    }

    // Count requests and quota
    for _, log := range logs {
        if log.Type == model.LogTypeConsume {
            ctx.RequestCount++
            ctx.QuotaUsed += int64(log.Quota)
        }
        // Count logins (could be marked in log type or other field)
        // This is simplified; adjust based on actual login logging
    }

    // Aggregate device activity
    deviceAggs, err := AggregateDeviceActivity(userId, startTime, endTime)
    if err == nil {
        ctx.DeviceAggregations = deviceAggs
        ctx.UniqueDevices = int64(len(deviceAggs))
    }

    // Aggregate IP activity
    ipAggs, err := AggregateIPActivity(userId, startTime, endTime)
    if err == nil {
        ctx.IPAggregations = ipAggs
        ctx.UniqueIPs = int64(len(ipAggs))
    }

    // Link conversations with requests
    convLinks, err := LinkConversationsWithRequests(userId, startTime, endTime)
    if err == nil {
        ctx.ConversationLinkages = convLinks
    }

    // Load baselines
    for _, metricType := range []string{"quota_usage", "login_ratio", "request_count"} {
        baseline, _ := model.GetAnomalyBaseline(userId, metricType)
        ctx.LatestBaseline[metricType] = baseline
    }

    return ctx, nil
}

// createAnomalyRecord converts a detection result to a database record
func (e *Engine) createAnomalyRecord(userId int, result *DetectionResult, ctx *DetectionContext) (*model.SecurityAnomaly, error) {
    evidenceData, err := common.Marshal(result.Evidence)
    if err != nil {
        evidenceData = []byte("{}")
    }

    anomaly := &model.SecurityAnomaly{
        UserId:     userId,
        RuleType:   result.RuleType,
        Severity:   result.Severity,
        Message:    result.Message,
        Evidence:   datatypes.JSON(evidenceData),
        DetectedAt: time.Now(),
        IpAddress:  "", // Could extract from context if needed
        TTLUntil:   getTTLForSeverity(result.Severity),
    }

    if err := model.CreateSecurityAnomaly(anomaly); err != nil {
        return nil, err
    }

    return anomaly, nil
}

// isDuplicate checks if an anomaly has already been recently reported
func (e *Engine) isDuplicate(anomaly *model.SecurityAnomaly) bool {
    // Build a deduplication key based on user, rule type, and time window
    dedupeWindow := 1 * time.Hour
    windowStart := time.Now().Add(-dedupeWindow)

    var count int64
    err := model.DB.Model(&model.SecurityAnomaly{}).
        Where("user_id = ? AND rule_type = ? AND detected_at >= ? AND is_resolved = ?",
            anomaly.UserId, anomaly.RuleType, windowStart, false).
        Count(&count).Error

    if err != nil {
        return false
    }

    return count > 0
}

// AddDetector adds a custom detector to the engine
func (e *Engine) AddDetector(detector AnomalyDetector) {
    e.mutex.Lock()
    defer e.mutex.Unlock()
    e.detectors = append(e.detectors, detector)
}

// Start begins background anomaly detection processing
func (e *Engine) Start(interval time.Duration, windowDuration time.Duration, batchSize int) chan struct{} {
    e.mutex.Lock()
    e.running = true
    e.mutex.Unlock()

    stopCh := make(chan struct{})

    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()

        for {
            select {
            case <-stopCh:
                e.mutex.Lock()
                e.running = false
                e.mutex.Unlock()
                return
            case <-ticker.C:
                e.processPendingUsers(windowDuration, batchSize)
            }
        }
    }()

    return stopCh
}

// processPendingUsers fetches active users and processes them
func (e *Engine) processPendingUsers(windowDuration time.Duration, batchSize int) {
    // Get active users (with recent activity)
    var users []struct {
        Id int
    }

    recentTime := time.Now().Add(-7 * 24 * time.Hour)
    err := model.DB.
        Table("users").
        Select("DISTINCT users.id").
        Joins("LEFT JOIN logs ON users.id = logs.user_id").
        Where("logs.created_at > ? AND users.status = ?", recentTime.Unix(), common.UserStatusEnabled).
        Limit(batchSize).
        Scan(&users).Error

    if err != nil {
        common.SysLog(fmt.Sprintf("failed to fetch pending users: %v", err))
        return
    }

    if len(users) == 0 {
        return
    }

    userIds := make([]int, 0, len(users))
    for _, u := range users {
        userIds = append(userIds, u.Id)
    }

    // Process in batches
    results, err := e.ProcessBatch(userIds, windowDuration, 5)
    if err != nil {
        common.SysLog(fmt.Sprintf("batch processing failed: %v", err))
        return
    }

    // Log results
    totalAnomalies := 0
    for uid, anomalies := range results {
        totalAnomalies += len(anomalies)
        common.SysLog(fmt.Sprintf("anomaly engine: detected %d anomalies for user %d", len(anomalies), uid))
    }

    if totalAnomalies > 0 {
        common.SysLog(fmt.Sprintf("anomaly engine: batch processing complete, detected %d total anomalies", totalAnomalies))
    }
}

// Stop stops the background processing
func (e *Engine) Stop(stopCh chan struct{}) {
    e.mutex.Lock()
    running := e.running
    e.mutex.Unlock()

    if running {
        close(stopCh)
        time.Sleep(100 * time.Millisecond) // Give goroutine time to exit
    }
}

// getTTLForSeverity returns the TTL duration for a given severity level
func getTTLForSeverity(severity string) *time.Time {
    var ttlDuration time.Duration

    switch severity {
    case "critical":
        ttlDuration = 24 * time.Hour
    case "high":
        ttlDuration = 7 * 24 * time.Hour
    case "medium":
        ttlDuration = 14 * 24 * time.Hour
    case "low":
        ttlDuration = 30 * 24 * time.Hour
    default:
        ttlDuration = 14 * 24 * time.Hour
    }

    ttl := time.Now().Add(ttlDuration)
    return &ttl
}

// UpdateBaselines updates the rolling baselines for a user
func UpdateUserBaselines(userId int) error {
    // Get recent activity (past 30 days)
    thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)

    // Update quota usage baseline
    var totalQuota int64
    err := model.LOG_DB.
        Where("user_id = ? AND created_at > ? AND type = ?", userId, thirtyDaysAgo.Unix(), model.LogTypeConsume).
        Select("SUM(quota)").
        Row().
        Scan(&totalQuota)

    if err == nil && totalQuota > 0 {
        baseline := &model.AnomalyBaseline{
            UserId:            userId,
            MetricType:        "quota_usage",
            BaselineValue:     float64(totalQuota),
            WindowSizeSeconds: 86400 * 30, // 30 days
            SampleSize:        1,
        }
        _ = model.UpdateAnomalyBaseline(baseline)
    }

    return nil
}
