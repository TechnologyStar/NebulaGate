package middleware

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "runtime"
    "strings"
    "sync"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/logger"
    "github.com/QuantumNous/new-api/model"
    "github.com/gin-gonic/gin"
)

// TelemetryConfig holds configuration for Heimdall telemetry
type TelemetryConfig struct {
    Enabled               bool
    GeolocationEnabled    bool
    BufferSize            int
    WorkerCount           int
    RetryAttempts         int
    RetryDelay            time.Duration
    DiskQueueEnabled      bool
    DiskQueuePath         string
    FlushInterval         time.Duration
}

// DefaultTelemetryConfig returns default configuration
func DefaultTelemetryConfig() TelemetryConfig {
    return TelemetryConfig{
        Enabled:               common.GetEnvOrDefaultBool("HEIMDALL_TELEMETRY_ENABLED", true),
        GeolocationEnabled:    common.GetEnvOrDefaultBool("HEIMDALL_GEOLOCATION_ENABLED", false),
        BufferSize:            common.GetEnvOrDefaultInt("HEIMDALL_BUFFER_SIZE", 10000),
        WorkerCount:           common.GetEnvOrDefaultInt("HEIMDALL_WORKER_COUNT", 5),
        RetryAttempts:         common.GetEnvOrDefaultInt("HEIMDALL_RETRY_ATTEMPTS", 3),
        RetryDelay:            time.Duration(common.GetEnvOrDefaultInt("HEIMDALL_RETRY_DELAY_MS", 1000)) * time.Millisecond,
        DiskQueueEnabled:      common.GetEnvOrDefaultBool("HEIMDALL_DISK_QUEUE_ENABLED", true),
        DiskQueuePath:         common.GetEnvOrDefault("HEIMDALL_DISK_QUEUE_PATH", "/tmp/heimdall_queue"),
        FlushInterval:         time.Duration(common.GetEnvOrDefaultInt("HEIMDALL_FLUSH_INTERVAL_MS", 5000)) * time.Millisecond,
    }
}

// TelemetryEntry represents a telemetry log entry
type TelemetryEntry struct {
    Log          *model.HeimdallRequestLog
    RequestStart time.Time
    RequestEnd   time.Time
    RequestBody  []byte
    ResponseBody []byte
}

// TelemetryWorker handles async persistence of telemetry data
type TelemetryWorker struct {
    config     TelemetryConfig
    entryChan  chan *TelemetryEntry
    stopChan   chan struct{}
    wg         sync.WaitGroup
    diskQueue  *DiskQueue
    mu         sync.RWMutex
    running    bool
}

// HeimdallTelemetryMiddleware creates a new Heimdall telemetry middleware
func HeimdallTelemetryMiddleware(config TelemetryConfig) gin.HandlerFunc {
    worker := NewTelemetryWorker(config)
    worker.Start()
    
    return func(c *gin.Context) {
        if !config.Enabled {
            c.Next()
            return
        }
        
        start := time.Now()
        
        // Capture request body for analysis
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }
        
        // Create response writer wrapper to capture response
        responseWriter := &responseBodyWriter{
            ResponseWriter: c.Writer,
            body:          &bytes.Buffer{},
        }
        c.Writer = responseWriter
        
        // Process request
        c.Next()
        
        end := time.Now()
        
        // Create telemetry entry
        entry := &TelemetryEntry{
            RequestStart: start,
            RequestEnd:   end,
            RequestBody:  requestBody,
            ResponseBody: responseWriter.body.Bytes(),
        }
        
        // Build log entry
        entry.Log = buildTelemetryLog(c, start, end, requestBody, responseWriter.body.Bytes())
        
        // Send to worker for async processing
        select {
        case worker.entryChan <- entry:
            // Successfully queued
        default:
            // Buffer full, log warning and try to process synchronously
            logger.LogError(c, "Heimdall telemetry buffer full, dropping log entry")
        }
    }
}

// responseBodyWriter wraps gin.ResponseWriter to capture response body
type responseBodyWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (r *responseBodyWriter) Write(b []byte) (int, error) {
    r.body.Write(b)
    return r.ResponseWriter.Write(b)
}

// buildTelemetryLog constructs a HeimdallRequestLog from the gin context
func buildTelemetryLog(c *gin.Context, start, end time.Time, requestBody, responseBody []byte) *model.HeimdallRequestLog {
    latencyMs := end.Sub(start).Milliseconds()
    
    // Extract client metadata
    clientMetadata, _ := model.ExtractClientMetadata(c.Request.Header)
    
    // Get user and token info from context
    userId, _ := c.Get("id")
    tokenId, _ := c.Get("token_id")
    
    // Parse request parameters
    var params map[string]interface{}
    if len(requestBody) > 0 {
        json.Unmarshal(requestBody, &params)
    } else {
        // For GET requests, use query parameters
        params = make(map[string]interface{})
        for key, values := range c.Request.URL.Query() {
            if len(values) > 0 {
                params[key] = values[0]
            }
        }
    }
    
    // Create log entry
    log := &model.HeimdallRequestLog{
        RequestId:            c.GetString("request_id"),
        OccurredAt:           start.UTC(),
        AuthKeyFingerprint:   model.CreateAuthKeyFingerprint(c.GetHeader("Authorization")),
        NormalizedURL:        model.NormalizeURL(c.Request.URL.Path, c.Request.Method),
        HTTPMethod:           c.Request.Method,
        HTTPStatus:           c.Writer.Status(),
        LatencyMs:            latencyMs,
        ClientIP:             clientMetadata["client_ip"],
        ClientUserAgent:      clientMetadata["user_agent"],
        ClientDeviceId:       clientMetadata["x-device-id"],
        RequestSizeBytes:     int64(len(requestBody)),
        ResponseSizeBytes:    int64(len(responseBody)),
        ParamDigest:          model.CreateParamDigest(params),
        SanitizedCookies:     model.SanitizeCookies(c.GetHeader("Cookie")),
        ModelName:            c.GetString("model_name"),
        UpstreamProvider:     c.GetString("channel_name"),
    }
    
    // Set user and token IDs if available
    if uid, ok := userId.(int); ok {
        log.UserId = &uid
    }
    if tid, ok := tokenId.(int); ok {
        log.TokenId = &tid
    }
    
    // Add geolocation if enabled
    if DefaultTelemetryConfig().GeolocationEnabled {
        if countryCode, region, city := getGeolocation(log.ClientIP); countryCode != "" {
            log.CountryCode = countryCode
            log.Region = region
            log.City = city
        }
    }
    
    // Add error information if request failed
    if c.Writer.Status() >= 400 {
        log.ErrorMessage = c.Errors.String()
        log.ErrorType = categorizeError(c.Writer.Status())
    }
    
    return log
}

// getGeolocation returns geolocation data for an IP address
// This is a placeholder implementation that should be replaced with actual geolocation service
func getGeolocation(ip string) (countryCode, region, city string) {
    // TODO: Implement actual geolocation lookup
    // This could use MaxMind GeoIP2, IP-API, or similar service
    return "", "", ""
}

// categorizeError categorizes HTTP status codes into error types
func categorizeError(statusCode int) string {
    switch {
    case statusCode >= 400 && statusCode < 500:
        return "client_error"
    case statusCode >= 500:
        return "server_error"
    default:
        return ""
    }
}

// NewTelemetryWorker creates a new telemetry worker
func NewTelemetryWorker(config TelemetryConfig) *TelemetryWorker {
    worker := &TelemetryWorker{
        config:    config,
        entryChan: make(chan *TelemetryEntry, config.BufferSize),
        stopChan:  make(chan struct{}),
    }
    
    if config.DiskQueueEnabled {
        worker.diskQueue = NewDiskQueue(config.DiskQueuePath)
    }
    
    return worker
}

// Start starts the telemetry worker
func (w *TelemetryWorker) Start() {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    if w.running {
        return
    }
    
    w.running = true
    
    // Start worker goroutines
    for i := 0; i < w.config.WorkerCount; i++ {
        w.wg.Add(1)
        go w.worker(i)
    }
    
    // Start flush goroutine
    w.wg.Add(1)
    go w.flusher()
    
    logger.SysLog(fmt.Sprintf("Heimdall telemetry worker started with %d workers", w.config.WorkerCount))
}

// Stop stops the telemetry worker
func (w *TelemetryWorker) Stop() {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    if !w.running {
        return
    }
    
    close(w.stopChan)
    close(w.entryChan)
    w.wg.Wait()
    w.running = false
    
    logger.SysLog("Heimdall telemetry worker stopped")
}

// worker processes telemetry entries
func (w *TelemetryWorker) worker(id int) {
    defer w.wg.Done()
    
    for {
        select {
        case entry, ok := <-w.entryChan:
            if !ok {
                return
            }
            w.processEntry(entry)
            
        case <-w.stopChan:
            return
        }
    }
}

// processEntry processes a single telemetry entry
func (w *TelemetryWorker) processEntry(entry *TelemetryEntry) {
    var err error
    
    // Try to persist to database first
    for attempt := 0; attempt < w.config.RetryAttempts; attempt++ {
        err = w.persistToDatabase(entry.Log)
        if err == nil {
            break
        }
        
        if attempt < w.config.RetryAttempts-1 {
            time.Sleep(w.config.RetryDelay)
        }
    }
    
    // If database persistence failed and disk queue is enabled, queue to disk
    if err != nil && w.config.DiskQueueEnabled && w.diskQueue != nil {
        if diskErr := w.diskQueue.Enqueue(entry.Log); diskErr != nil {
            logger.SysLog(fmt.Sprintf("Failed to queue telemetry to disk: %v", diskErr))
        }
    }
    
    // Update frequency metrics
    w.updateFrequencyMetrics(entry.Log)
}

// persistToDatabase persists the log entry to the database
func (w *TelemetryWorker) persistToDatabase(log *model.HeimdallRequestLog) error {
    return model.LOG_DB.Create(log).Error
}

// updateFrequencyMetrics updates Redis frequency metrics
func (w *TelemetryWorker) updateFrequencyMetrics(log *model.HeimdallRequestLog) {
    // Update per-URL counters
    if log.NormalizedURL != "" {
        urlKey := fmt.Sprintf("heimdall:url:%s:count", log.NormalizedURL)
        common.RedisIncrByOne(urlKey)
        common.RedisExpire(context.Background(), urlKey, time.Hour)
    }
    
    // Update per-token counters
    if log.TokenId != nil {
        tokenKey := fmt.Sprintf("heimdall:token:%d:count", *log.TokenId)
        common.RedisIncrByOne(tokenKey)
        common.RedisExpire(context.Background(), tokenKey, time.Hour)
    }
    
    // Update per-user counters
    if log.UserId != nil {
        userKey := fmt.Sprintf("heimdall:user:%d:count", *log.UserId)
        common.RedisIncrByOne(userKey)
        common.RedisExpire(context.Background(), userKey, time.Hour)
    }
}

// flusher periodically flushes disk queue to database
func (w *TelemetryWorker) flusher() {
    defer w.wg.Done()
    
    ticker := time.NewTicker(w.config.FlushInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if w.diskQueue != nil {
                w.flushDiskQueue()
            }
            
        case <-w.stopChan:
            // Final flush before stopping
            if w.diskQueue != nil {
                w.flushDiskQueue()
            }
            return
        }
    }
}

// flushDiskQueue flushes entries from disk queue to database
func (w *TelemetryWorker) flushDiskQueue() {
    entries, err := w.diskQueue.DequeueBatch(100) // Process in batches
    if err != nil {
        logger.SysLog(fmt.Sprintf("Failed to dequeue from disk queue: %v", err))
        return
    }
    
    for _, entry := range entries {
        if err := w.persistToDatabase(entry); err != nil {
            // Re-queue if persistence failed
            if requeueErr := w.diskQueue.Enqueue(entry); requeueErr != nil {
                logger.SysLog(fmt.Sprintf("Failed to re-queue telemetry entry: %v", requeueErr))
            }
        }
    }
}

// GetTelemetryStats returns telemetry worker statistics
func (w *TelemetryWorker) GetTelemetryStats() map[string]interface{} {
    w.mu.RLock()
    defer w.mu.RUnlock()
    
    stats := map[string]interface{}{
        "running":        w.running,
        "buffer_length":  len(w.entryChan),
        "buffer_capacity": w.config.BufferSize,
        "worker_count":   w.config.WorkerCount,
        "goroutines":     runtime.NumGoroutine(),
    }
    
    if w.diskQueue != nil {
        stats["disk_queue_size"] = w.diskQueue.Size()
    }
    
    return stats
}

// Global telemetry worker instance
var globalTelemetryWorker *TelemetryWorker

// InitHeimdallTelemetry initializes the global telemetry worker
func InitHeimdallTelemetry() {
    config := DefaultTelemetryConfig()
    globalTelemetryWorker = NewTelemetryWorker(config)
    globalTelemetryWorker.Start()
}

// StopHeimdallTelemetry stops the global telemetry worker
func StopHeimdallTelemetry() {
    if globalTelemetryWorker != nil {
        globalTelemetryWorker.Stop()
    }
}

// GetHeimdallTelemetryStats returns global telemetry statistics
func GetHeimdallTelemetryStats() map[string]interface{} {
    if globalTelemetryWorker != nil {
        return globalTelemetryWorker.GetTelemetryStats()
    }
    return map[string]interface{}{"running": false}
}
