package model

import (
	"time"

	"gorm.io/gorm"
)

// HeimdallLog stores request logs captured by Heimdall security gateway
type HeimdallLog struct {
	Id                int            `json:"id" gorm:"primaryKey"`
	UserId            int            `json:"user_id" gorm:"index"`
	TokenKey          string         `json:"token_key" gorm:"type:varchar(255);index"`
	RequestPath       string         `json:"request_path" gorm:"type:varchar(512)"`
	RequestMethod     string         `json:"request_method" gorm:"type:varchar(16)"`
	RealIP            string         `json:"real_ip" gorm:"type:varchar(64);index"`
	ForwardedFor      string         `json:"forwarded_for" gorm:"type:varchar(255)"`
	UserAgent         string         `json:"user_agent" gorm:"type:text"`
	RequestHeaders    string         `json:"request_headers" gorm:"type:text"`
	RequestBody       string         `json:"request_body" gorm:"type:text"`
	ContentFingerprint string        `json:"content_fingerprint" gorm:"type:varchar(255)"`
	DeviceFingerprint string         `json:"device_fingerprint" gorm:"type:varchar(255);index"`
	Cookies           string         `json:"cookies" gorm:"type:text"`
	ResponseStatus    int            `json:"response_status" gorm:"index"`
	ResponseTime      int            `json:"response_time"` // milliseconds
	Timestamp         int64          `json:"timestamp" gorm:"index;not null"`
	CreatedAt         time.Time      `json:"created_at"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (HeimdallLog) TableName() string {
	return "heimdall_logs"
}

// CreateHeimdallLog creates a new Heimdall log entry
func CreateHeimdallLog(log *HeimdallLog) error {
	log.Timestamp = time.Now().Unix()
	return DB.Create(log).Error
}

// GetHeimdallLogsByUserId retrieves Heimdall logs for a specific user
func GetHeimdallLogsByUserId(userId int, startIdx int, num int) ([]*HeimdallLog, int64, error) {
	var logs []*HeimdallLog
	var total int64

	err := DB.Model(&HeimdallLog{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("user_id = ?", userId).
		Order("timestamp desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error

	return logs, total, err
}

// GetHeimdallLogsByIP retrieves all logs from a specific IP
func GetHeimdallLogsByIP(ip string, startIdx int, num int) ([]*HeimdallLog, int64, error) {
	var logs []*HeimdallLog
	var total int64

	err := DB.Model(&HeimdallLog{}).Where("real_ip = ?", ip).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("real_ip = ?", ip).
		Order("timestamp desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error

	return logs, total, err
}

// GetHeimdallLogsByDevice retrieves all logs from a specific device fingerprint
func GetHeimdallLogsByDevice(deviceFingerprint string, startIdx int, num int) ([]*HeimdallLog, int64, error) {
	var logs []*HeimdallLog
	var total int64

	err := DB.Model(&HeimdallLog{}).Where("device_fingerprint = ?", deviceFingerprint).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = DB.Where("device_fingerprint = ?", deviceFingerprint).
		Order("timestamp desc").
		Limit(num).
		Offset(startIdx).
		Find(&logs).Error

	return logs, total, err
}

// GetRequestFrequencyByToken calculates request frequency for a token within a time window
func GetRequestFrequencyByToken(tokenKey string, startTime int64, endTime int64) (int64, error) {
	var count int64
	err := DB.Model(&HeimdallLog{}).
		Where("token_key = ? AND timestamp >= ? AND timestamp <= ?", tokenKey, startTime, endTime).
		Count(&count).Error
	return count, err
}

// GetRequestFrequencyByUser calculates request frequency for a user within a time window
func GetRequestFrequencyByUser(userId int, startTime int64, endTime int64) (int64, error) {
	var count int64
	err := DB.Model(&HeimdallLog{}).
		Where("user_id = ? AND timestamp >= ? AND timestamp <= ?", userId, startTime, endTime).
		Count(&count).Error
	return count, err
}
