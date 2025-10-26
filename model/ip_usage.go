package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TokenIPUsage struct {
	ID           int       `json:"id"`
	TokenId      int       `json:"token_id" gorm:"not null;index:idx_token_ip,priority:1"`
	IP           string    `json:"ip" gorm:"size:64;not null;index:idx_token_ip,priority:2"`
	FirstSeenAt  time.Time `json:"first_seen_at" gorm:"not null"`
	LastSeenAt   time.Time `json:"last_seen_at" gorm:"not null"`
	RequestCount int64     `json:"request_count" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserIPUsage struct {
	ID           int       `json:"id"`
	UserId       int       `json:"user_id" gorm:"not null;index:idx_user_ip,priority:1"`
	IP           string    `json:"ip" gorm:"size:64;not null;index:idx_user_ip,priority:2"`
	FirstSeenAt  time.Time `json:"first_seen_at" gorm:"not null"`
	LastSeenAt   time.Time `json:"last_seen_at" gorm:"not null"`
	RequestCount int64     `json:"request_count" gorm:"not null;default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func sanitizeIP(ip string) string {
	candidate := strings.TrimSpace(ip)
	if candidate == "" {
		return ""
	}
	if len(candidate) > 63 {
		candidate = candidate[:63]
	}
	return candidate
}

func upsertIPUsage(db *gorm.DB, table string, columnA string, columnB string, valueA any, valueB any) error {
	if db == nil {
		return gorm.ErrInvalidDB
	}
	now := time.Now().UTC()
	return db.Table(table).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: columnA}, {Name: columnB}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"last_seen_at":  now,
			"request_count": gorm.Expr("request_count + ?", 1),
		}),
	}).Create(map[string]interface{}{
		columnA:         valueA,
		columnB:         valueB,
		"first_seen_at": now,
		"last_seen_at":  now,
		"request_count": 1,
	}).Error
}

// RecordIPUsage tracks IP usage statistics for both the token and user scope.
func RecordIPUsage(tokenId int, userId int, rawIP string) {
	ip := sanitizeIP(rawIP)
	if ip == "" {
		return
	}
	db := LOG_DB
	if db == nil {
		db = DB
	}
	if tokenId > 0 {
		_ = upsertIPUsage(db, "token_ip_usages", "token_id", "ip", tokenId, ip)
	}
	if userId > 0 {
		_ = upsertIPUsage(db, "user_ip_usages", "user_id", "ip", userId, ip)
	}
}

func GetTokenIPUsage(tokenId int, since time.Time) ([]TokenIPUsage, int64, error) {
	db := LOG_DB
	if db == nil {
		db = DB
	}
	query := db.Model(&TokenIPUsage{}).Where("token_id = ?", tokenId)
	if !since.IsZero() {
		query = query.Where("last_seen_at >= ?", since)
	}
	var usages []TokenIPUsage
	if err := query.Order("last_seen_at desc").Find(&usages).Error; err != nil {
		return nil, 0, err
	}
	var totalRequests int64
	if err := query.Select("COALESCE(SUM(request_count),0)").Scan(&totalRequests).Error; err != nil {
		return nil, 0, err
	}
	return usages, totalRequests, nil
}

func GetUserIPUsage(userId int, since time.Time) ([]UserIPUsage, int64, error) {
	db := LOG_DB
	if db == nil {
		db = DB
	}
	query := db.Model(&UserIPUsage{}).Where("user_id = ?", userId)
	if !since.IsZero() {
		query = query.Where("last_seen_at >= ?", since)
	}
	var usages []UserIPUsage
	if err := query.Order("last_seen_at desc").Find(&usages).Error; err != nil {
		return nil, 0, err
	}
	var totalRequests int64
	if err := query.Select("COALESCE(SUM(request_count),0)").Scan(&totalRequests).Error; err != nil {
		return nil, 0, err
	}
	return usages, totalRequests, nil
}
