package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// SecurityViolation records detected policy violations
type SecurityViolation struct {
	Id              int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserId          int       `json:"user_id" gorm:"index;not null"`
	TokenId         *int      `json:"token_id" gorm:"index"`
	ViolatedAt      time.Time `json:"violated_at" gorm:"index;not null"`
	ContentSnippet  string    `json:"content_snippet" gorm:"type:text"` // Sanitized snippet
	MatchedKeywords string    `json:"matched_keywords" gorm:"type:varchar(500)"`
	Model           string    `json:"model" gorm:"type:varchar(100);index"`
	IpAddress       string    `json:"ip_address" gorm:"type:varchar(45)"`
	ActionTaken     string    `json:"action_taken" gorm:"type:varchar(50)"` // redirect, block, log
	RequestId       string    `json:"request_id" gorm:"type:varchar(64);index"`
	Severity        string    `json:"severity" gorm:"type:varchar(20)"` // malicious, violation
}

// TableName specifies the table name
func (SecurityViolation) TableName() string {
	return "security_violations"
}

// CreateSecurityViolation creates a new violation record
func CreateSecurityViolation(violation *SecurityViolation) error {
	if violation.UserId == 0 {
		return errors.New("user_id is required")
	}
	if violation.ViolatedAt.IsZero() {
		violation.ViolatedAt = time.Now()
	}
	return DB.Create(violation).Error
}

// GetSecurityViolations returns paginated violations
func GetSecurityViolations(offset, limit int, userId int, startTime, endTime *time.Time, keyword string) ([]*SecurityViolation, int64, error) {
	var violations []*SecurityViolation
	var total int64

	query := DB.Model(&SecurityViolation{})

	if userId > 0 {
		query = query.Where("user_id = ?", userId)
	}
	if startTime != nil {
		query = query.Where("violated_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("violated_at <= ?", endTime)
	}
	if keyword != "" {
		query = query.Where("matched_keywords LIKE ? OR content_snippet LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("violated_at DESC").Offset(offset).Limit(limit).Find(&violations).Error
	return violations, total, err
}

// DeleteSecurityViolation deletes a violation record
func DeleteSecurityViolation(id int) error {
	return DB.Delete(&SecurityViolation{}, id).Error
}

// GetViolationStatsByDateRange returns aggregated statistics
func GetViolationStatsByDateRange(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total violations count
	var totalCount int64
	err := DB.Model(&SecurityViolation{}).
		Where("violated_at >= ? AND violated_at <= ?", startTime, endTime).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}
	stats["total_count"] = totalCount

	// Unique users count
	var uniqueUsers int64
	err = DB.Model(&SecurityViolation{}).
		Where("violated_at >= ? AND violated_at <= ?", startTime, endTime).
		Distinct("user_id").
		Count(&uniqueUsers).Error
	if err != nil {
		return nil, err
	}
	stats["unique_users"] = uniqueUsers

	// Top keywords
	type KeywordCount struct {
		Keyword string
		Count   int64
	}
	var topKeywords []KeywordCount
	err = DB.Model(&SecurityViolation{}).
		Select("matched_keywords as keyword, COUNT(*) as count").
		Where("violated_at >= ? AND violated_at <= ? AND matched_keywords != ''", startTime, endTime).
		Group("matched_keywords").
		Order("count DESC").
		Limit(10).
		Scan(&topKeywords).Error
	if err != nil {
		return nil, err
	}
	stats["top_keywords"] = topKeywords

	// Daily trend (last 7 days)
	type DailyTrend struct {
		Date  string
		Count int64
	}
	var dailyTrend []DailyTrend

	// Use database-specific date formatting
	dateFormat := "DATE_FORMAT(violated_at, '%Y-%m-%d')"
	if common.UsingPostgreSQL {
		dateFormat = "TO_CHAR(violated_at, 'YYYY-MM-DD')"
	} else if common.UsingSQLite {
		dateFormat = "DATE(violated_at)"
	}

	err = DB.Model(&SecurityViolation{}).
		Select(dateFormat+" as date, COUNT(*) as count").
		Where("violated_at >= ? AND violated_at <= ?", startTime, endTime).
		Group("date").
		Order("date DESC").
		Scan(&dailyTrend).Error
	if err != nil {
		return nil, err
	}
	stats["daily_trend"] = dailyTrend

	return stats, nil
}

// GetViolationsByUser returns violations for a specific user
func GetViolationsByUser(userId int, offset, limit int) ([]*SecurityViolation, int64, error) {
	var violations []*SecurityViolation
	var total int64

	query := DB.Model(&SecurityViolation{}).Where("user_id = ?", userId)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("violated_at DESC").Offset(offset).Limit(limit).Find(&violations).Error
	return violations, total, err
}

// GetTopViolatingUsers returns users with most violations
func GetTopViolatingUsers(limit int, startTime, endTime *time.Time) ([]map[string]interface{}, error) {
	type UserViolation struct {
		UserId         int
		ViolationCount int64
		LastViolation  time.Time
	}

	var results []UserViolation
	query := DB.Model(&SecurityViolation{}).
		Select("user_id, COUNT(*) as violation_count, MAX(violated_at) as last_violation")

	if startTime != nil {
		query = query.Where("violated_at >= ?", startTime)
	}
	if endTime != nil {
		query = query.Where("violated_at <= ?", endTime)
	}

	err := query.Group("user_id").
		Order("violation_count DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Enhance with user info
	var enrichedResults []map[string]interface{}
	for _, result := range results {
		user, err := GetUserById(result.UserId, false)
		if err != nil {
			continue
		}
		enrichedResults = append(enrichedResults, map[string]interface{}{
			"user_id":         result.UserId,
			"username":        user.Username,
			"display_name":    user.DisplayName,
			"violation_count": result.ViolationCount,
			"last_violation":  result.LastViolation,
		})
	}

	return enrichedResults, nil
}
