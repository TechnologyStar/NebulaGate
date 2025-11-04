package model

import (
	"time"

	"gorm.io/gorm"
)

// GroupRateSchedule represents a time-based rate multiplier rule for a group
type GroupRateSchedule struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	GroupName      string    `json:"group_name" gorm:"size:64;not null;index"`
	TimeStart      string    `json:"time_start" gorm:"size:5;not null"` // HH:MM format
	TimeEnd        string    `json:"time_end" gorm:"size:5;not null"`   // HH:MM format
	RateMultiplier float64   `json:"rate_multiplier" gorm:"not null;default:1.0"`
	Enabled        bool      `json:"enabled" gorm:"not null;default:true;index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (g *GroupRateSchedule) TableName() string {
	return "group_rate_schedules"
}

// CreateGroupRateSchedule creates a new rate schedule rule
func CreateGroupRateSchedule(schedule *GroupRateSchedule) error {
	return DB.Create(schedule).Error
}

// GetGroupRateSchedule gets a schedule by ID
func GetGroupRateSchedule(id int) (*GroupRateSchedule, error) {
	var schedule GroupRateSchedule
	err := DB.Where("id = ?", id).First(&schedule).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// GetGroupRateSchedulesByGroup gets all schedules for a group
func GetGroupRateSchedulesByGroup(groupName string) ([]GroupRateSchedule, error) {
	var schedules []GroupRateSchedule
	err := DB.Where("group_name = ?", groupName).Order("time_start ASC").Find(&schedules).Error
	return schedules, err
}

// GetEnabledGroupRateSchedules gets all enabled schedules for a group
func GetEnabledGroupRateSchedules(groupName string) ([]GroupRateSchedule, error) {
	var schedules []GroupRateSchedule
	err := DB.Where("group_name = ? AND enabled = ?", groupName, true).
		Order("time_start ASC").Find(&schedules).Error
	return schedules, err
}

// UpdateGroupRateSchedule updates a schedule
func UpdateGroupRateSchedule(schedule *GroupRateSchedule) error {
	return DB.Model(&GroupRateSchedule{}).Where("id = ?", schedule.ID).Updates(schedule).Error
}

// DeleteGroupRateSchedule deletes a schedule
func DeleteGroupRateSchedule(id int) error {
	return DB.Where("id = ?", id).Delete(&GroupRateSchedule{}).Error
}

// GetAllGroupRateSchedules gets all schedules with pagination
func GetAllGroupRateSchedules(page int, pageSize int) ([]GroupRateSchedule, int64, error) {
	var schedules []GroupRateSchedule
	var total int64

	db := DB.Model(&GroupRateSchedule{})
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = db.Order("group_name ASC, time_start ASC").Limit(pageSize).Offset(offset).Find(&schedules).Error
	return schedules, total, err
}

// GetCurrentRateMultiplier gets the current effective rate multiplier for a group
func GetCurrentRateMultiplier(groupName string, currentTime time.Time) (float64, error) {
	schedules, err := GetEnabledGroupRateSchedules(groupName)
	if err != nil {
		return 1.0, err
	}

	if len(schedules) == 0 {
		return 1.0, nil
	}

	currentHHMM := currentTime.Format("15:04")

	for _, schedule := range schedules {
		if isTimeInRange(currentHHMM, schedule.TimeStart, schedule.TimeEnd) {
			return schedule.RateMultiplier, nil
		}
	}

	return 1.0, nil
}

// isTimeInRange checks if currentTime is within the range [startTime, endTime)
// Supports cross-midnight ranges (e.g., 22:00 - 02:00)
func isTimeInRange(current, start, end string) bool {
	if start <= end {
		// Normal range: 08:00 - 18:00
		return current >= start && current < end
	}
	// Cross-midnight range: 22:00 - 02:00
	return current >= start || current < end
}
