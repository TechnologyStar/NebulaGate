package service

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// UpdateGroupRateMultipliers updates all group rate multipliers based on current time
func UpdateGroupRateMultipliers() error {
	if !common.RedisEnabled {
		common.SysLog("Redis not enabled, skipping group rate multiplier update")
		return nil
	}

	currentTime := time.Now()
	
	// Get all unique group names that have schedules
	var groupNames []string
	if err := model.DB.Model(&model.GroupRateSchedule{}).
		Distinct("group_name").
		Where("enabled = ?", true).
		Pluck("group_name", &groupNames).Error; err != nil {
		return fmt.Errorf("failed to get group names: %w", err)
	}

	for _, groupName := range groupNames {
		multiplier, err := model.GetCurrentRateMultiplier(groupName, currentTime)
		if err != nil {
			common.SysLog(fmt.Sprintf("Failed to get rate multiplier for group %s: %v", groupName, err))
			continue
		}

		// Store in Redis cache
		cacheKey := fmt.Sprintf("group:rate:current:%s", groupName)
		err = common.RedisSet(cacheKey, fmt.Sprintf("%.2f", multiplier), 5*time.Minute)
		if err != nil {
			common.SysLog(fmt.Sprintf("Failed to cache rate multiplier for group %s: %v", groupName, err))
		}
	}

	return nil
}

// GetCachedGroupRateMultiplier gets the cached rate multiplier for a group
func GetCachedGroupRateMultiplier(groupName string) (float64, error) {
	if !common.RedisEnabled {
		return 1.0, nil
	}

	cacheKey := fmt.Sprintf("group:rate:current:%s", groupName)
	value, err := common.RedisGet(cacheKey)
	if err != nil || value == "" {
		// Cache miss, calculate and cache
		multiplier, err := model.GetCurrentRateMultiplier(groupName, time.Now())
		if err != nil {
			return 1.0, err
		}
		
		// Cache for 5 minutes
		_ = common.RedisSet(cacheKey, fmt.Sprintf("%.2f", multiplier), 5*time.Minute)
		return multiplier, nil
	}

	var multiplier float64
	_, err = fmt.Sscanf(value, "%f", &multiplier)
	if err != nil {
		return 1.0, err
	}

	return multiplier, nil
}

// StartGroupRateScheduler starts the background scheduler for group rate multipliers
func StartGroupRateScheduler() {
	if !common.IsMasterNode {
		return
	}

	common.SysLog("Starting group rate scheduler")

	// Initial update
	if err := UpdateGroupRateMultipliers(); err != nil {
		common.SysLog(fmt.Sprintf("Initial group rate multiplier update failed: %v", err))
	}

	// Update every minute
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			if err := UpdateGroupRateMultipliers(); err != nil {
				common.SysLog(fmt.Sprintf("Group rate multiplier update failed: %v", err))
			}
		}
	}()
}
