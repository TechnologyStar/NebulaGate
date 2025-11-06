package model

import (
    "errors"
    "time"

    "gorm.io/datatypes"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type SecurityAnomaly struct {
    Id        int            `json:"id" gorm:"primaryKey;autoIncrement"`
    UserId    int            `json:"user_id" gorm:"index;not null"`
    TokenId   *int           `json:"token_id" gorm:"index"`
    DeviceId  *string        `json:"device_id" gorm:"index;size:255"`
    IpAddress string         `json:"ip_address" gorm:"type:varchar(45);index"`
    RuleType  string         `json:"rule_type" gorm:"type:varchar(100);not null;index"` // quota_spike, abnormal_login_ratio, high_request_ratio, etc
    Severity  string         `json:"severity" gorm:"type:varchar(20)"`                   // low, medium, high, critical
    Evidence  datatypes.JSON `json:"evidence" gorm:"type:json"`                          // JSON evidence data
    Message   string         `json:"message" gorm:"type:text"`                           // Human-readable message
    DetectedAt time.Time     `json:"detected_at" gorm:"index;not null"`
    CreatedAt time.Time      `json:"created_at" gorm:"index;autoCreateTime"`
    UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
    TTLUntil  *time.Time     `json:"ttl_until" gorm:"index"`
    IsResolved bool          `json:"is_resolved" gorm:"default:false"`
    ResolvedAt *time.Time    `json:"resolved_at"`
}

func (SecurityAnomaly) TableName() string {
    return "security_anomalies"
}

type AnomalyEvidence struct {
    BaselineValue float64                 `json:"baseline_value"`
    ActualValue   float64                 `json:"actual_value"`
    Deviation     float64                 `json:"deviation_percent"`
    Details       map[string]interface{} `json:"details,omitempty"`
}

func CreateSecurityAnomaly(anomaly *SecurityAnomaly) error {
    if anomaly.UserId == 0 {
        return errors.New("user_id is required")
    }
    if anomaly.RuleType == "" {
        return errors.New("rule_type is required")
    }
    if anomaly.DetectedAt.IsZero() {
        anomaly.DetectedAt = time.Now()
    }
    if anomaly.Severity == "" {
        anomaly.Severity = "medium"
    }
    return DB.Create(anomaly).Error
}

func GetSecurityAnomalies(offset, limit int, userId int, ruleType string, severity string, startTime, endTime *time.Time) ([]*SecurityAnomaly, int64, error) {
    var anomalies []*SecurityAnomaly
    var total int64

    query := DB.Model(&SecurityAnomaly{}).Where("is_resolved = ?", false)

    if userId > 0 {
        query = query.Where("user_id = ?", userId)
    }
    if ruleType != "" {
        query = query.Where("rule_type = ?", ruleType)
    }
    if severity != "" {
        query = query.Where("severity = ?", severity)
    }
    if startTime != nil {
        query = query.Where("detected_at >= ?", startTime)
    }
    if endTime != nil {
        query = query.Where("detected_at <= ?", endTime)
    }

    err := query.Count(&total).Error
    if err != nil {
        return nil, 0, err
    }

    err = query.Order("detected_at DESC").Offset(offset).Limit(limit).Find(&anomalies).Error
    return anomalies, total, err
}

func GetAnomalyStatsByDateRange(startTime, endTime time.Time) (map[string]interface{}, error) {
    stats := make(map[string]interface{})

    var totalCount int64
    err := DB.Model(&SecurityAnomaly{}).
        Where("detected_at >= ? AND detected_at <= ?", startTime, endTime).
        Count(&totalCount).Error
    if err != nil {
        return nil, err
    }
    stats["total_count"] = totalCount

    var uniqueUsers int64
    err = DB.Model(&SecurityAnomaly{}).
        Where("detected_at >= ? AND detected_at <= ?", startTime, endTime).
        Distinct("user_id").
        Count(&uniqueUsers).Error
    if err != nil {
        return nil, err
    }
    stats["unique_users"] = uniqueUsers

    type AnomalyCount struct {
        RuleType string
        Count    int64
    }
    var anomalyCounts []AnomalyCount
    err = DB.Model(&SecurityAnomaly{}).
        Select("rule_type, COUNT(*) as count").
        Where("detected_at >= ? AND detected_at <= ?", startTime, endTime).
        Group("rule_type").
        Order("count DESC").
        Scan(&anomalyCounts).Error
    if err != nil {
        return nil, err
    }
    stats["by_rule_type"] = anomalyCounts

    type SeverityCount struct {
        Severity string
        Count    int64
    }
    var severityCounts []SeverityCount
    err = DB.Model(&SecurityAnomaly{}).
        Select("severity, COUNT(*) as count").
        Where("detected_at >= ? AND detected_at <= ?", startTime, endTime).
        Group("severity").
        Scan(&severityCounts).Error
    if err != nil {
        return nil, err
    }
    stats["by_severity"] = severityCounts

    return stats, nil
}

func ResolveSecurityAnomaly(id int) error {
    now := time.Now()
    return DB.Model(&SecurityAnomaly{}).
        Where("id = ?", id).
        Updates(map[string]interface{}{
            "is_resolved": true,
            "resolved_at": now,
        }).Error
}

func CleanupExpiredAnomalies(ctx ...interface{}) (int64, error) {
    result := DB.Where("ttl_until IS NOT NULL AND ttl_until <= ?", time.Now()).Delete(&SecurityAnomaly{})
    return result.RowsAffected, result.Error
}

// DeviceAggregation represents aggregated device activity
type DeviceAggregation struct {
    NormalizedDeviceId string
    UserId             int
    RequestCount       int64
    LastSeenAt         time.Time
    IPs                []string
    Models             []string
}

// IPAggregation represents aggregated IP activity with time windows
type IPAggregation struct {
    IP         string
    UserId     int
    TokenId    *int
    Window     string // "5m", "1h", "24h"
    Count      int64
    ASN        string
    Subnet     string
    LastSeenAt time.Time
}

// AnomalyBaseline stores baseline metrics for users
type AnomalyBaseline struct {
    Id                  int       `json:"id" gorm:"primaryKey;autoIncrement"`
    UserId              int       `json:"user_id" gorm:"uniqueIndex:uk_user_metric;not null"`
    MetricType          string    `json:"metric_type" gorm:"uniqueIndex:uk_user_metric;not null"` // request_count, login_count, quota_usage
    BaselineValue       float64   `json:"baseline_value"`
    StandardDeviation   float64   `json:"standard_deviation"`
    WindowSizeSeconds   int       `json:"window_size_seconds"`
    SampleSize          int       `json:"sample_size"`
    LastUpdatedAt       time.Time `json:"last_updated_at"`
    CreatedAt           time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (AnomalyBaseline) TableName() string {
    return "anomaly_baselines"
}

func UpdateAnomalyBaseline(baseline *AnomalyBaseline) error {
    if baseline.UserId == 0 || baseline.MetricType == "" {
        return errors.New("user_id and metric_type are required")
    }
    baseline.LastUpdatedAt = time.Now()
    return DB.Clauses(clause.OnConflict{
        UpdateAll: true,
    }).Create(baseline).Error
}

func GetAnomalyBaseline(userId int, metricType string) (*AnomalyBaseline, error) {
    var baseline AnomalyBaseline
    err := DB.Where("user_id = ? AND metric_type = ?", userId, metricType).First(&baseline).Error
    if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &baseline, err
}
