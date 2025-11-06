package model

import (
	"time"
)

type DeviceFingerprint struct {
	Id            int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId        int       `json:"user_id" gorm:"index;not null"`
	Fingerprint   string    `json:"fingerprint" gorm:"type:varchar(256);index;not null"`
	UserAgent     string    `json:"user_agent" gorm:"type:text"`
	IpAddress     string    `json:"ip_address" gorm:"type:varchar(45);index"`
	FirstSeenAt   time.Time `json:"first_seen_at" gorm:"not null"`
	LastSeenAt    time.Time `json:"last_seen_at" gorm:"not null"`
	RequestCount  int       `json:"request_count" gorm:"default:0"`
	IsBlocked     bool      `json:"is_blocked" gorm:"default:false;index"`
	RiskScore     int       `json:"risk_score" gorm:"default:0"`
}

func (DeviceFingerprint) TableName() string {
	return "device_fingerprints"
}

func CreateDeviceFingerprint(device *DeviceFingerprint) error {
	return DB.Create(device).Error
}

func GetDeviceFingerprint(fingerprint string, userId int) (*DeviceFingerprint, error) {
	var device DeviceFingerprint
	err := DB.Where("fingerprint = ? AND user_id = ?", fingerprint, userId).First(&device).Error
	return &device, err
}

func UpdateDeviceFingerprint(device *DeviceFingerprint) error {
	return DB.Save(device).Error
}

func GetDeviceClusters(offset, limit int, blocked *bool) ([]*DeviceFingerprint, int64, error) {
	var devices []*DeviceFingerprint
	var total int64

	query := DB.Model(&DeviceFingerprint{})
	if blocked != nil {
		query = query.Where("is_blocked = ?", *blocked)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("risk_score DESC, last_seen_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&devices).Error

	return devices, total, err
}

func GetTopSuspiciousDevices(limit int) ([]*DeviceFingerprint, error) {
	var devices []*DeviceFingerprint
	err := DB.Where("risk_score > ?", 50).
		Order("risk_score DESC").
		Limit(limit).
		Find(&devices).Error
	return devices, err
}
