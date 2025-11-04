package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// UserSecurity tracks user security status
type UserSecurity struct {
	UserId          int        `json:"user_id" gorm:"primaryKey;not null"`
	IsBanned        bool       `json:"is_banned" gorm:"default:false;index"`
	RedirectModel   string     `json:"redirect_model" gorm:"type:varchar(100)"` // User-level redirect target
	ViolationCount  int        `json:"violation_count" gorm:"default:0"`
	LastViolationAt *time.Time `json:"last_violation_at"`
	CreatedAt       time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (UserSecurity) TableName() string {
	return "user_security"
}

// GetUserSecurity retrieves security status for a user
func GetUserSecurity(userId int) (*UserSecurity, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("user_security:%d", userId)
	if cachedData, found := common.MemoryCacheGet(cacheKey); found {
		if userSec, ok := cachedData.(*UserSecurity); ok {
			return userSec, nil
		}
	}

	var userSec UserSecurity
	err := DB.Where("user_id = ?", userId).First(&userSec).Error
	if err != nil {
		if errors.Is(err, errors.New("record not found")) {
			// Create default security record
			userSec = UserSecurity{
				UserId:         userId,
				IsBanned:       false,
				ViolationCount: 0,
			}
			if err := DB.Create(&userSec).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Cache for 5 minutes
	common.MemoryCacheSet(cacheKey, &userSec, 5*time.Minute)
	return &userSec, nil
}

// UpdateUserSecurity updates user security status
func UpdateUserSecurity(userSec *UserSecurity) error {
	if userSec.UserId == 0 {
		return errors.New("user_id is required")
	}

	err := DB.Save(userSec).Error
	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user_security:%d", userSec.UserId)
	common.MemoryCacheDelete(cacheKey)

	// Also sync to Redis
	syncUserSecurityToRedis(userSec)
	return nil
}

// IncrementViolationCount increments the violation count for a user
func IncrementViolationCount(userId int) error {
	now := time.Now()
	err := DB.Model(&UserSecurity{}).
		Where("user_id = ?", userId).
		Updates(map[string]interface{}{
			"violation_count":   DB.Raw("violation_count + 1"),
			"last_violation_at": now,
		}).Error

	if err != nil {
		// If record doesn't exist, create it
		userSec := UserSecurity{
			UserId:          userId,
			ViolationCount:  1,
			LastViolationAt: &now,
		}
		err = DB.Create(&userSec).Error
		if err != nil {
			return err
		}
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("user_security:%d", userId)
	common.MemoryCacheDelete(cacheKey)

	// Reload and sync to Redis
	userSec, _ := GetUserSecurity(userId)
	if userSec != nil {
		syncUserSecurityToRedis(userSec)
	}

	return nil
}

// BanUser bans a user
func BanUser(userId int) error {
	userSec, err := GetUserSecurity(userId)
	if err != nil {
		return err
	}

	userSec.IsBanned = true
	return UpdateUserSecurity(userSec)
}

// UnbanUser unbans a user
func UnbanUser(userId int) error {
	userSec, err := GetUserSecurity(userId)
	if err != nil {
		return err
	}

	userSec.IsBanned = false
	return UpdateUserSecurity(userSec)
}

// SetUserRedirect sets redirect model for a user
func SetUserRedirect(userId int, model string) error {
	userSec, err := GetUserSecurity(userId)
	if err != nil {
		return err
	}

	userSec.RedirectModel = model
	return UpdateUserSecurity(userSec)
}

// ClearUserRedirect removes redirect configuration
func ClearUserRedirect(userId int) error {
	userSec, err := GetUserSecurity(userId)
	if err != nil {
		return err
	}

	userSec.RedirectModel = ""
	return UpdateUserSecurity(userSec)
}

// GetAllUserSecurity retrieves security status with pagination
func GetAllUserSecurity(offset, limit int, bannedOnly bool) ([]*UserSecurity, int64, error) {
	var userSecList []*UserSecurity
	var total int64

	query := DB.Model(&UserSecurity{})
	if bannedOnly {
		query = query.Where("is_banned = ?", true)
	}

	// Only include users with violations or banned status
	query = query.Where("violation_count > 0 OR is_banned = ? OR redirect_model != ''", true)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("violation_count DESC, last_violation_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&userSecList).Error

	return userSecList, total, err
}

// syncUserSecurityToRedis syncs user security status to Redis for fast access
func syncUserSecurityToRedis(userSec *UserSecurity) {
	if !common.RedisEnabled {
		return
	}

	key := fmt.Sprintf("user_security:%d", userSec.UserId)
	data, err := json.Marshal(userSec)
	if err != nil {
		return
	}

	// Store in Redis with 1 hour expiry
	common.RedisSet(key, string(data), 3600)
}

// GetUserSecurityFromRedis retrieves user security from Redis
func GetUserSecurityFromRedis(userId int) (*UserSecurity, bool) {
	if !common.RedisEnabled {
		return nil, false
	}

	key := fmt.Sprintf("user_security:%d", userId)
	data, err := common.RedisGet(key)
	if err != nil || data == "" {
		return nil, false
	}

	var userSec UserSecurity
	if err := json.Unmarshal([]byte(data), &userSec); err != nil {
		return nil, false
	}

	return &userSec, true
}
