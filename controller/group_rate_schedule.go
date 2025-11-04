package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func CreateGroupRateSchedule(c *gin.Context) {
	var schedule model.GroupRateSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate time format
	if !isValidTimeFormat(schedule.TimeStart) || !isValidTimeFormat(schedule.TimeEnd) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid time format. Use HH:MM format (e.g., 09:30)",
		})
		return
	}

	if schedule.RateMultiplier <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Rate multiplier must be greater than 0",
		})
		return
	}

	if err := model.CreateGroupRateSchedule(&schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create schedule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Schedule created successfully",
		"data":    schedule,
	})
}

func GetGroupRateSchedules(c *gin.Context) {
	groupName := c.Query("group_name")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "group_name is required",
		})
		return
	}

	schedules, err := model.GetGroupRateSchedulesByGroup(groupName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get schedules: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    schedules,
	})
}

func UpdateGroupRateSchedule(c *gin.Context) {
	var schedule model.GroupRateSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	if schedule.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Schedule ID is required",
		})
		return
	}

	// Validate time format if provided
	if schedule.TimeStart != "" && !isValidTimeFormat(schedule.TimeStart) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid time_start format. Use HH:MM format",
		})
		return
	}

	if schedule.TimeEnd != "" && !isValidTimeFormat(schedule.TimeEnd) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid time_end format. Use HH:MM format",
		})
		return
	}

	if err := model.UpdateGroupRateSchedule(&schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update schedule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Schedule updated successfully",
		"data":    schedule,
	})
}

func DeleteGroupRateSchedule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid schedule ID",
		})
		return
	}

	if err := model.DeleteGroupRateSchedule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete schedule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Schedule deleted successfully",
	})
}

func GetCurrentGroupRate(c *gin.Context) {
	groupName := c.Query("group_name")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "group_name is required",
		})
		return
	}

	multiplier, err := service.GetCachedGroupRateMultiplier(groupName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get current rate: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"group_name":      groupName,
			"rate_multiplier": multiplier,
			"timestamp":       time.Now().Format(time.RFC3339),
		},
	})
}

func GetAllGroupRateSchedules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	schedules, total, err := model.GetAllGroupRateSchedules(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get schedules: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items":      schedules,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func isValidTimeFormat(timeStr string) bool {
	_, err := time.Parse("15:04", timeStr)
	return err == nil
}

func ForceUpdateGroupRates(c *gin.Context) {
	if err := service.UpdateGroupRateMultipliers(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("Failed to update group rates: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Group rates updated successfully",
	})
}
