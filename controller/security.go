package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// GetSecurityDashboard returns security dashboard statistics
func GetSecurityDashboard(c *gin.Context) {
	// Parse time range from query params
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid start_time format",
			})
			return
		}
	} else {
		// Default to last 7 days
		startTime = time.Now().AddDate(0, 0, -7)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid end_time format",
			})
			return
		}
	} else {
		endTime = time.Now()
	}

	stats, err := service.GetDashboardStats(startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSecurityViolations returns paginated violation records
func GetSecurityViolations(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userId, _ := strconv.Atoi(c.Query("user_id"))
	keyword := c.Query("keyword")
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	offset := (page - 1) * pageSize

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			endTime = &t
		}
	}

	violations, total, err := model.GetSecurityViolations(offset, pageSize, userId, startTime, endTime, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"violations": violations,
			"total":      total,
			"page":       page,
			"page_size":  pageSize,
		},
	})
}

// DeleteSecurityViolation deletes a violation record
func DeleteSecurityViolation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid violation ID",
		})
		return
	}

	err = model.DeleteSecurityViolation(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Violation record deleted successfully",
	})
}

// GetSecurityUsers returns users with violations
func GetSecurityUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	bannedOnly := c.Query("banned_only") == "true"

	offset := (page - 1) * pageSize

	userSecList, total, err := model.GetAllUserSecurity(offset, pageSize, bannedOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Enrich with user info
	var enrichedUsers []map[string]interface{}
	for _, userSec := range userSecList {
		user, err := model.GetUserById(userSec.UserId, false)
		if err != nil {
			continue
		}

		enrichedUsers = append(enrichedUsers, map[string]interface{}{
			"user_id":           userSec.UserId,
			"username":          user.Username,
			"display_name":      user.DisplayName,
			"is_banned":         userSec.IsBanned,
			"redirect_model":    userSec.RedirectModel,
			"violation_count":   userSec.ViolationCount,
			"last_violation_at": userSec.LastViolationAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users":     enrichedUsers,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// BanUser bans a user
func BanUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	err = service.BanUser(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User banned successfully",
	})
}

// UnbanUser unbans a user
func UnbanUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	err = service.UnbanUser(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User unbanned successfully",
	})
}

// SetUserRedirect sets redirect model for a user
func SetUserRedirect(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	var req struct {
		Model string `json:"model" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	err = service.SetUserRedirect(userId, req.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User redirect set successfully",
	})
}

// ClearUserRedirect removes redirect for a user
func ClearUserRedirect(c *gin.Context) {
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	err = service.ClearUserRedirect(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User redirect cleared successfully",
	})
}

// GetSecuritySettings returns security settings
func GetSecuritySettings(c *gin.Context) {
	settings := service.GetSecuritySettings()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateSecuritySettings updates security settings
func UpdateSecuritySettings(c *gin.Context) {
	var settings map[string]interface{}

	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	err := service.UpdateSecuritySettings(settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
	})
}
