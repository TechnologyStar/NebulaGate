package model

import (
	"time"
)

type IPCluster struct {
	Id              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	IpAddress       string    `json:"ip_address" gorm:"type:varchar(45);uniqueIndex;not null"`
	Country         string    `json:"country" gorm:"type:varchar(2)"`
	City            string    `json:"city" gorm:"type:varchar(100)"`
	UniqueUsers     int       `json:"unique_users" gorm:"default:0"`
	TotalRequests   int       `json:"total_requests" gorm:"default:0"`
	ViolationCount  int       `json:"violation_count" gorm:"default:0"`
	IsBlocked       bool      `json:"is_blocked" gorm:"default:false;index"`
	RiskScore       int       `json:"risk_score" gorm:"default:0"`
	FirstSeenAt     time.Time `json:"first_seen_at" gorm:"not null"`
	LastSeenAt      time.Time `json:"last_seen_at" gorm:"not null"`
}

func (IPCluster) TableName() string {
	return "ip_clusters"
}

func CreateIPCluster(cluster *IPCluster) error {
	return DB.Create(cluster).Error
}

func GetIPCluster(ipAddress string) (*IPCluster, error) {
	var cluster IPCluster
	err := DB.Where("ip_address = ?", ipAddress).First(&cluster).Error
	return &cluster, err
}

func UpdateIPCluster(cluster *IPCluster) error {
	return DB.Save(cluster).Error
}

func GetIPClusters(offset, limit int, blocked *bool, minRiskScore int) ([]*IPCluster, int64, error) {
	var clusters []*IPCluster
	var total int64

	query := DB.Model(&IPCluster{})
	if blocked != nil {
		query = query.Where("is_blocked = ?", *blocked)
	}
	if minRiskScore > 0 {
		query = query.Where("risk_score >= ?", minRiskScore)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("risk_score DESC, last_seen_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&clusters).Error

	return clusters, total, err
}

func GetTopSuspiciousIPs(limit int) ([]*IPCluster, error) {
	var clusters []*IPCluster
	err := DB.Where("risk_score > ?", 50).
		Order("risk_score DESC").
		Limit(limit).
		Find(&clusters).Error
	return clusters, err
}
