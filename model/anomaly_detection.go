package model

import (
	"time"

	"gorm.io/gorm"
)

// AnomalyDetection stores anomaly detection results and risk scores
type AnomalyDetection struct {
	Id                   int            `json:"id" gorm:"primaryKey"`
	UserId               int            `json:"user_id" gorm:"index;not null"`
	DeviceFingerprint    string         `json:"device_fingerprint" gorm:"type:varchar(255);index"`
	IPAddress            string         `json:"ip_address" gorm:"type:varchar(64);index"`
	AnomalyType          string         `json:"anomaly_type" gorm:"type:varchar(64);index"` // e.g., "high_frequency", "suspicious_pattern", "data_spike"
	RiskScore            float64        `json:"risk_score" gorm:"index"`                    // 0-100
	LoginCount           int            `json:"login_count"`
	TotalAccessCount     int            `json:"total_access_count"`
	AccessRatio          float64        `json:"access_ratio"` // access_count / login_count
	APIRequestCount      int            `json:"api_request_count"`
	TimeWindowStart      int64          `json:"time_window_start"`
	TimeWindowEnd        int64          `json:"time_window_end"`
	AverageInterval      float64        `json:"average_interval"` // seconds between requests
	Status               string         `json:"status" gorm:"type:varchar(32);index;default:'detected'"` // detected, reviewing, resolved, false_positive
	Action               string         `json:"action" gorm:"type:varchar(64)"` // rate_limit, block, alert, none
	Description          string         `json:"description" gorm:"type:text"`
	Metadata             string         `json:"metadata" gorm:"type:text"` // JSON for additional context
	DetectedAt           int64          `json:"detected_at" gorm:"index;not null"`
	ReviewedAt           int64          `json:"reviewed_at"`
	ReviewedBy           int            `json:"reviewed_by"` // admin user ID
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (AnomalyDetection) TableName() string {
	return "anomaly_detections"
}

// CreateAnomalyDetection creates a new anomaly detection record
func CreateAnomalyDetection(anomaly *AnomalyDetection) error {
	anomaly.DetectedAt = time.Now().Unix()
	if anomaly.Status == "" {
		anomaly.Status = "detected"
	}
	return DB.Create(anomaly).Error
}

// GetAnomalyDetectionById retrieves an anomaly detection record by ID
func GetAnomalyDetectionById(id int) (*AnomalyDetection, error) {
	var anomaly AnomalyDetection
	err := DB.First(&anomaly, id).Error
	return &anomaly, err
}

// GetAnomalyDetectionsByUserId retrieves anomaly detections for a specific user
func GetAnomalyDetectionsByUserId(userId int, startIdx int, num int) ([]*AnomalyDetection, int64, error) {
	var anomalies []*AnomalyDetection
	var total int64

	err := DB.Model(&AnomalyDetection{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("user_id = ?", userId).
		Order("detected_at desc").
		Limit(num).
		Offset(startIdx).
		Find(&anomalies).Error

	return anomalies, total, err
}

// GetAnomalyDetectionsByStatus retrieves anomaly detections by status
func GetAnomalyDetectionsByStatus(status string, startIdx int, num int) ([]*AnomalyDetection, int64, error) {
	var anomalies []*AnomalyDetection
	var total int64

	err := DB.Model(&AnomalyDetection{}).Where("status = ?", status).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("status = ?", status).
		Order("detected_at desc").
		Limit(num).
		Offset(startIdx).
		Find(&anomalies).Error

	return anomalies, total, err
}

// GetHighRiskAnomalies retrieves anomalies with risk score above threshold
func GetHighRiskAnomalies(minRiskScore float64, startIdx int, num int) ([]*AnomalyDetection, int64, error) {
	var anomalies []*AnomalyDetection
	var total int64

	err := DB.Model(&AnomalyDetection{}).Where("risk_score >= ?", minRiskScore).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("risk_score >= ?", minRiskScore).
		Order("risk_score desc, detected_at desc").
		Limit(num).
		Offset(startIdx).
		Find(&anomalies).Error

	return anomalies, total, err
}

// GetAnomalyDetectionsByDevice retrieves anomaly detections for a specific device
func GetAnomalyDetectionsByDevice(deviceFingerprint string, startIdx int, num int) ([]*AnomalyDetection, int64, error) {
	var anomalies []*AnomalyDetection
	var total int64

	err := DB.Model(&AnomalyDetection{}).Where("device_fingerprint = ?", deviceFingerprint).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("device_fingerprint = ?", deviceFingerprint).
		Order("detected_at desc").
		Limit(num).
		Offset(startIdx).
		Find(&anomalies).Error

	return anomalies, total, err
}

// GetAnomalyDetectionsByIP retrieves anomaly detections for a specific IP
func GetAnomalyDetectionsByIP(ip string, startIdx int, num int) ([]*AnomalyDetection, int64, error) {
	var anomalies []*AnomalyDetection
	var total int64

	err := DB.Model(&AnomalyDetection{}).Where("ip_address = ?", ip).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("ip_address = ?", ip).
		Order("detected_at desc").
		Limit(num).
		Offset(startIdx).
		Find(&anomalies).Error

	return anomalies, total, err
}

// UpdateAnomalyDetectionStatus updates the status of an anomaly detection
func UpdateAnomalyDetectionStatus(id int, status string, reviewedBy int) error {
	updates := map[string]interface{}{
		"status":      status,
		"reviewed_at": time.Now().Unix(),
		"reviewed_by": reviewedBy,
	}
	return DB.Model(&AnomalyDetection{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateAnomalyDetectionAction updates the action taken for an anomaly
func UpdateAnomalyDetectionAction(id int, action string) error {
	return DB.Model(&AnomalyDetection{}).Where("id = ?", id).Update("action", action).Error
}

// DeleteAnomalyDetection soft deletes an anomaly detection record
func DeleteAnomalyDetection(id int) error {
	return DB.Delete(&AnomalyDetection{}, id).Error
}

// GetAnomalyStatsByUser calculates aggregated anomaly statistics for a user
func GetAnomalyStatsByUser(userId int, startTime int64, endTime int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalCount int64
	var highRiskCount int64
	var avgRiskScore float64

	// Total anomalies
	err := DB.Model(&AnomalyDetection{}).
		Where("user_id = ? AND detected_at >= ? AND detected_at <= ?", userId, startTime, endTime).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	// High risk anomalies (score >= 70)
	err = DB.Model(&AnomalyDetection{}).
		Where("user_id = ? AND detected_at >= ? AND detected_at <= ? AND risk_score >= ?", userId, startTime, endTime, 70.0).
		Count(&highRiskCount).Error
	if err != nil {
		return nil, err
	}

	// Average risk score
	err = DB.Model(&AnomalyDetection{}).
		Where("user_id = ? AND detected_at >= ? AND detected_at <= ?", userId, startTime, endTime).
		Select("COALESCE(AVG(risk_score), 0)").
		Scan(&avgRiskScore).Error
	if err != nil {
		return nil, err
	}

	stats["total_anomalies"] = totalCount
	stats["high_risk_count"] = highRiskCount
	stats["avg_risk_score"] = avgRiskScore

	return stats, nil
}

// SearchAnomalies searches anomaly detections with filters
func SearchAnomalies(userId int, anomalyType string, minRiskScore float64, status string, startIdx int, num int) ([]*AnomalyDetection, int64, error) {
	var anomalies []*AnomalyDetection
	var total int64

	query := DB.Model(&AnomalyDetection{})

	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	if anomalyType != "" {
		query = query.Where("anomaly_type = ?", anomalyType)
	}
	if minRiskScore > 0 {
		query = query.Where("risk_score >= ?", minRiskScore)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("detected_at desc").
		Limit(num).
		Offset(startIdx).
		Find(&anomalies).Error

	return anomalies, total, err
}
