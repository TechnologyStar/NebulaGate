package controller

import (
    "net/http"
    "strconv"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    "github.com/QuantumNous/new-api/service"
    "github.com/gin-gonic/gin"
)

// GetAnomalyDetections retrieves anomaly detections with optional filters
func GetAnomalyDetections(c *gin.Context) {
    userId := c.GetInt("id")
    userRole := c.GetInt("role")
    
    // Parse query parameters
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    status := c.Query("status")
    anomalyType := c.Query("anomaly_type")
    minRiskScoreStr := c.Query("min_risk_score")
    targetUserIdStr := c.Query("user_id")
    
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    
    minRiskScore := 0.0
    if minRiskScoreStr != "" {
        minRiskScore, _ = strconv.ParseFloat(minRiskScoreStr, 64)
    }
    
    startIdx := (page - 1) * pageSize
    
    // Non-admin users can only view their own anomalies
    targetUserId := userId
    if userRole >= common.RoleAdminUser && targetUserIdStr != "" {
        targetUserId, _ = strconv.Atoi(targetUserIdStr)
    }
    
    // Get anomalies
    var anomalies []*model.AnomalyDetection
    var total int64
    var err error
    
    if userRole >= common.RoleAdminUser && targetUserIdStr == "" {
        // Admin viewing all anomalies
        anomalies, total, err = model.SearchAnomalies(0, anomalyType, minRiskScore, status, startIdx, pageSize)
    } else {
        // User viewing their own anomalies or admin viewing specific user
        anomalies, total, err = model.SearchAnomalies(targetUserId, anomalyType, minRiskScore, status, startIdx, pageSize)
    }
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve anomaly detections",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "anomalies": anomalies,
            "total":     total,
            "page":      page,
            "page_size": pageSize,
        },
    })
}

// GetAnomalyDetectionById retrieves a specific anomaly detection
func GetAnomalyDetectionById(c *gin.Context) {
    userId := c.GetInt("id")
    userRole := c.GetInt("role")
    
    anomalyId, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid anomaly ID",
        })
        return
    }
    
    anomaly, err := model.GetAnomalyDetectionById(anomalyId)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "message": "Anomaly detection not found",
        })
        return
    }
    
    // Check permissions
    if userRole < common.RoleAdminUser && anomaly.UserId != userId {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "message": "Access denied",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    anomaly,
    })
}

// UpdateAnomalyDetectionStatus updates the status of an anomaly detection
func UpdateAnomalyDetectionStatus(c *gin.Context) {
    userId := c.GetInt("id")
    userRole := c.GetInt("role")
    
    // Only admins can update status
    if userRole < common.RoleAdminUser {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "message": "Admin access required",
        })
        return
    }
    
    anomalyId, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid anomaly ID",
        })
        return
    }
    
    var req struct {
        Status string `json:"status" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }
    
    // Validate status
    validStatuses := map[string]bool{
        "detected":       true,
        "reviewing":      true,
        "resolved":       true,
        "false_positive": true,
    }
    
    if !validStatuses[req.Status] {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid status",
        })
        return
    }
    
    err = model.UpdateAnomalyDetectionStatus(anomalyId, req.Status, userId)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to update anomaly status",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Anomaly status updated successfully",
    })
}

// UpdateAnomalyDetectionAction updates the action for an anomaly detection
func UpdateAnomalyDetectionAction(c *gin.Context) {
    userRole := c.GetInt("role")
    
    // Only admins can update actions
    if userRole < common.RoleAdminUser {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "message": "Admin access required",
        })
        return
    }
    
    anomalyId, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid anomaly ID",
        })
        return
    }
    
    var req struct {
        Action string `json:"action" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }
    
    // Validate action
    validActions := map[string]bool{
        "none":       true,
        "alert":      true,
        "rate_limit": true,
        "block":      true,
    }
    
    if !validActions[req.Action] {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid action",
        })
        return
    }
    
    err = model.UpdateAnomalyDetectionAction(anomalyId, req.Action)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to update anomaly action",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Anomaly action updated successfully",
    })
}

// GetAnomalyStatistics retrieves anomaly statistics for the user
func GetAnomalyStatistics(c *gin.Context) {
    userId := c.GetInt("id")
    userRole := c.GetInt("role")
    
    // Parse time range
    startTimeStr := c.DefaultQuery("start_time", "0")
    endTimeStr := c.DefaultQuery("end_time", strconv.FormatInt(common.GetTimestamp(), 10))
    targetUserIdStr := c.Query("user_id")
    
    startTime, _ := strconv.ParseInt(startTimeStr, 10, 64)
    endTime, _ := strconv.ParseInt(endTimeStr, 10, 64)
    
    // Default to last 24 hours if not specified
    if startTime == 0 {
        endTime = common.GetTimestamp()
        startTime = endTime - 86400
    }
    
    // Non-admin users can only view their own stats
    targetUserId := userId
    if userRole >= common.RoleAdminUser && targetUserIdStr != "" {
        targetUserId, _ = strconv.Atoi(targetUserIdStr)
    }
    
    stats, err := model.GetAnomalyStatsByUser(targetUserId, startTime, endTime)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve anomaly statistics",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    stats,
    })
}

// GetDeviceAggregation retrieves all records for a specific device fingerprint
func GetDeviceAggregation(c *gin.Context) {
    userRole := c.GetInt("role")
    
    // Only admins can view device aggregation
    if userRole < common.RoleAdminUser {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "message": "Admin access required",
        })
        return
    }
    
    deviceFingerprint := c.Query("device_fingerprint")
    if deviceFingerprint == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Device fingerprint is required",
        })
        return
    }
    
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    
    startIdx := (page - 1) * pageSize
    
    // Get logs for device
    logs, total, err := model.GetHeimdallLogsByDevice(deviceFingerprint, startIdx, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve device logs",
        })
        return
    }
    
    // Get anomalies for device
    anomalies, anomalyTotal, err := model.GetAnomalyDetectionsByDevice(deviceFingerprint, 0, 10)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve device anomalies",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "logs":           logs,
            "logs_total":     total,
            "anomalies":      anomalies,
            "anomaly_total":  anomalyTotal,
            "page":           page,
            "page_size":      pageSize,
        },
    })
}

// GetIPAggregation retrieves all records for a specific IP address
func GetIPAggregation(c *gin.Context) {
    userRole := c.GetInt("role")
    
    // Only admins can view IP aggregation
    if userRole < common.RoleAdminUser {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "message": "Admin access required",
        })
        return
    }
    
    ip := c.Query("ip")
    if ip == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "IP address is required",
        })
        return
    }
    
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    if page < 1 {
        page = 1
    }
    if pageSize < 1 || pageSize > 100 {
        pageSize = 20
    }
    
    startIdx := (page - 1) * pageSize
    
    // Get logs for IP
    logs, total, err := model.GetHeimdallLogsByIP(ip, startIdx, pageSize)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve IP logs",
        })
        return
    }
    
    // Get anomalies for IP
    anomalies, anomalyTotal, err := model.GetAnomalyDetectionsByIP(ip, 0, 10)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve IP anomalies",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": gin.H{
            "logs":          logs,
            "logs_total":    total,
            "anomalies":     anomalies,
            "anomaly_total": anomalyTotal,
            "page":          page,
            "page_size":     pageSize,
        },
    })
}

// TriggerAnomalyDetection manually triggers anomaly detection for a user
func TriggerAnomalyDetection(c *gin.Context) {
    userRole := c.GetInt("role")
    
    // Only admins can trigger anomaly detection
    if userRole < common.RoleAdminUser {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "message": "Admin access required",
        })
        return
    }
    
    targetUserIdStr := c.Param("user_id")
    targetUserId, err := strconv.Atoi(targetUserIdStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid user ID",
        })
        return
    }
    
    // Run anomaly detection
    detector := service.NewAnomalyDetectorService()
    err = detector.AnalyzeUserBehavior(targetUserId)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to run anomaly detection: " + err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Anomaly detection completed successfully",
    })
}
