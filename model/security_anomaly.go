package model

import (
	"time"
)

type SecurityAnomaly struct {
	Id              int        `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId          int        `json:"user_id" gorm:"index;not null"`
	TokenId         *int       `json:"token_id" gorm:"index"`
	DetectedAt      time.Time  `json:"detected_at" gorm:"index;not null"`
	AnomalyType     string     `json:"anomaly_type" gorm:"type:varchar(50);index"`
	Severity        string     `json:"severity" gorm:"type:varchar(20);index"`
	Description     string     `json:"description" gorm:"type:text"`
	Metadata        string     `json:"metadata" gorm:"type:text"`
	IpAddress       string     `json:"ip_address" gorm:"type:varchar(45);index"`
	DeviceId        string     `json:"device_id" gorm:"type:varchar(256);index"`
	RiskScore       int        `json:"risk_score" gorm:"default:0"`
	ActionTaken     string     `json:"action_taken" gorm:"type:varchar(50)"`
	ActionedAt      *time.Time `json:"actioned_at"`
	Status          string     `json:"status" gorm:"type:varchar(20);default:'pending';index"`
	ReviewedBy      *int       `json:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	ReviewDecision  string     `json:"review_decision" gorm:"type:varchar(20)"`
	ReviewRationale string     `json:"review_rationale" gorm:"type:text"`
}

func (SecurityAnomaly) TableName() string {
	return "security_anomalies"
}

func CreateSecurityAnomaly(anomaly *SecurityAnomaly) error {
	if anomaly.DetectedAt.IsZero() {
		anomaly.DetectedAt = time.Now()
	}
	if anomaly.Status == "" {
		anomaly.Status = "pending"
	}
	return DB.Create(anomaly).Error
}

func GetSecurityAnomaly(id int) (*SecurityAnomaly, error) {
	var anomaly SecurityAnomaly
	err := DB.Where("id = ?", id).First(&anomaly).Error
	return &anomaly, err
}

func UpdateSecurityAnomaly(anomaly *SecurityAnomaly) error {
	return DB.Save(anomaly).Error
}

func GetSecurityAnomalies(offset, limit int, filters map[string]interface{}) ([]*SecurityAnomaly, int64, error) {
	var anomalies []*SecurityAnomaly
	var total int64

	query := DB.Model(&SecurityAnomaly{})

	if userId, ok := filters["user_id"].(int); ok && userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	if severity, ok := filters["severity"].(string); ok && severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if anomalyType, ok := filters["anomaly_type"].(string); ok && anomalyType != "" {
		query = query.Where("anomaly_type = ?", anomalyType)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if startTime, ok := filters["start_time"].(*time.Time); ok && startTime != nil {
		query = query.Where("detected_at >= ?", startTime)
	}
	if endTime, ok := filters["end_time"].(*time.Time); ok && endTime != nil {
		query = query.Where("detected_at <= ?", endTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("detected_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&anomalies).Error

	return anomalies, total, err
}

func GetAnomalyCountsBySeverity(startTime, endTime time.Time) (map[string]int64, error) {
	type SeverityCount struct {
		Severity string
		Count    int64
	}

	var results []SeverityCount
	err := DB.Model(&SecurityAnomaly{}).
		Select("severity, COUNT(*) as count").
		Where("detected_at >= ? AND detected_at <= ?", startTime, endTime).
		Group("severity").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Severity] = r.Count
	}

	return counts, nil
}

func GetAnomalyTrends(startTime, endTime time.Time) ([]map[string]interface{}, error) {
	type DailyTrend struct {
		Date     string
		Severity string
		Count    int64
	}

	var results []DailyTrend

	dateFormat := "DATE_FORMAT(detected_at, '%Y-%m-%d')"
	if UsingPostgreSQL {
		dateFormat = "TO_CHAR(detected_at, 'YYYY-MM-DD')"
	} else if UsingSQLite {
		dateFormat = "DATE(detected_at)"
	}

	err := DB.Model(&SecurityAnomaly{}).
		Select(dateFormat+" as date, severity, COUNT(*) as count").
		Where("detected_at >= ? AND detected_at <= ?", startTime, endTime).
		Group("date, severity").
		Order("date DESC, severity").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	trends := make([]map[string]interface{}, len(results))
	for i, r := range results {
		trends[i] = map[string]interface{}{
			"date":     r.Date,
			"severity": r.Severity,
			"count":    r.Count,
		}
	}

	return trends, nil
}

func GetPendingHighSeverityAnomalies(limit int) ([]*SecurityAnomaly, error) {
	var anomalies []*SecurityAnomaly
	err := DB.Where("status = ? AND severity = ?", "pending", "malicious").
		Order("detected_at ASC").
		Limit(limit).
		Find(&anomalies).Error
	return anomalies, err
}
