package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// GetAnomalies retrieves anomalies for the current user
func GetAnomalies(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	ruleType := c.Query("rule_type")
	severity := c.Query("severity")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	anomalies, total, err := model.GetSecurityAnomalies(offset, limit, userId, ruleType, severity, nil, nil)
	if err != nil {
		logger.LogError(c, "failed to get anomalies: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve anomalies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": anomalies,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetAnomalyStatistics retrieves anomaly statistics
func GetAnomalyStatistics(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 7
	}

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(days*24) * time.Hour)

	stats, err := service.GetAnomalyStatistics(startTime, endTime)
	if err != nil {
		logger.LogError(c, "failed to get anomaly statistics: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve statistics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
		"period": gin.H{
			"start_time": startTime,
			"end_time":   endTime,
			"days":       days,
		},
	})
}

// ResolveAnomaly marks an anomaly as resolved
func ResolveAnomaly(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	anomalyIdStr := c.Param("id")
	anomalyId, err := strconv.Atoi(anomalyIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid anomaly id"})
		return
	}

	// Verify ownership
	var anomaly model.SecurityAnomaly
	if err := model.DB.Where("id = ? AND user_id = ?", anomalyId, userId).First(&anomaly).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "anomaly not found"})
		return
	}

	if err := service.ResolveAnomaly(anomalyId); err != nil {
		logger.LogError(c, "failed to resolve anomaly: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve anomaly"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "anomaly resolved"})
}

// AdminGetAllAnomalies retrieves all anomalies (admin only)
func AdminGetAllAnomalies(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	userIdStr := c.Query("user_id")
	ruleType := c.Query("rule_type")
	severity := c.Query("severity")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var userId int
	if userIdStr != "" {
		userId, _ = strconv.Atoi(userIdStr)
	}

	anomalies, total, err := model.GetSecurityAnomalies(offset, limit, userId, ruleType, severity, nil, nil)
	if err != nil {
		logger.LogError(c, "failed to get anomalies: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve anomalies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": anomalies,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// AdminProcessUserAnomalies triggers anomaly detection for a specific user (admin only)
func AdminProcessUserAnomalies(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	userIdStr := c.Param("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	anomalies, err := service.ProcessUserAnomalies(userId, 24*time.Hour)
	if err != nil {
		logger.LogError(c, "failed to process user anomalies: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process anomalies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":   userId,
		"anomalies": anomalies,
		"count":     len(anomalies),
	})
}

// AdminGetAnomalySettings retrieves anomaly detection settings (admin only)
func AdminGetAnomalySettings(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	settings := map[string]interface{}{
		"enabled":                  model.GetOptionValue("anomaly_detection_enabled") == "true",
		"interval_seconds":         model.GetOptionValue("anomaly_detection_interval_seconds"),
		"window_hours":             model.GetOptionValue("anomaly_detection_window_hours"),
		"quota_spike_percent":      model.GetOptionValue("anomaly_quota_spike_percent"),
		"login_ratio_threshold":    model.GetOptionValue("anomaly_login_ratio_threshold"),
		"request_ratio_threshold":  model.GetOptionValue("anomaly_request_ratio_threshold"),
		"new_device_requests":      model.GetOptionValue("anomaly_new_device_requests"),
		"ip_change_threshold":      model.GetOptionValue("anomaly_ip_change_threshold"),
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

// AdminUpdateAnomalySettings updates anomaly detection settings (admin only)
func AdminUpdateAnomalySettings(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	for key, value := range req {
		optionKey := "anomaly_" + key
		stringValue := ""

		switch v := value.(type) {
		case bool:
			stringValue = "false"
			if v {
				stringValue = "true"
			}
		case float64:
			stringValue = strconv.FormatFloat(v, 'f', -1, 64)
		case string:
			stringValue = v
		default:
			continue
		}

		if err := model.UpdateOption(optionKey, stringValue); err != nil {
			logger.LogError(c, "failed to update option "+optionKey+": "+err.Error())
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

// Helper function to check if user is admin
func isAdmin(c *gin.Context) bool {
	role := c.GetString("role")
	return role == common.RoleAdminUser || role == common.RoleRootUser
}
