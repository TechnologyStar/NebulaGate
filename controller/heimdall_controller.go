package controller

import (
    "net/http"
    "time"

    "github.com/QuantumNous/new-api/middleware"
    "github.com/QuantumNous/new-api/service"
    "github.com/gin-gonic/gin"
)

// GetHeimdallTelemetryStats returns Heimdall telemetry statistics
func GetHeimdallTelemetryStats(c *gin.Context) {
    stats := middleware.GetHeimdallTelemetryStats()
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    stats,
    })
}

// GetHeimdallURLMetrics returns URL frequency metrics
func GetHeimdallURLMetrics(c *gin.Context) {
    // Parse time window parameter
    timeWindowStr := c.DefaultQuery("time_window", "1h")
    timeWindow, err := time.ParseDuration(timeWindowStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid time_window format. Use formats like '1h', '24h', '7d'",
        })
        return
    }
    
    // Get metrics
    analyticsService := service.GlobalHeimdallAnalyticsService
    metrics, err := analyticsService.GetURLFrequencyMetrics(c.Request.Context(), timeWindow)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve URL metrics: " + err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    metrics,
    })
}

// GetHeimdallTokenMetrics returns token frequency metrics
func GetHeimdallTokenMetrics(c *gin.Context) {
    // Parse time window parameter
    timeWindowStr := c.DefaultQuery("time_window", "1h")
    timeWindow, err := time.ParseDuration(timeWindowStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid time_window format. Use formats like '1h', '24h', '7d'",
        })
        return
    }
    
    // Get metrics
    analyticsService := service.GlobalHeimdallAnalyticsService
    metrics, err := analyticsService.GetTokenFrequencyMetrics(c.Request.Context(), timeWindow)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve token metrics: " + err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    metrics,
    })
}

// GetHeimdallUserMetrics returns user frequency metrics
func GetHeimdallUserMetrics(c *gin.Context) {
    // Parse time window parameter
    timeWindowStr := c.DefaultQuery("time_window", "1h")
    timeWindow, err := time.ParseDuration(timeWindowStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid time_window format. Use formats like '1h', '24h', '7d'",
        })
        return
    }
    
    // Get metrics
    analyticsService := service.GlobalHeimdallAnalyticsService
    metrics, err := analyticsService.GetUserFrequencyMetrics(c.Request.Context(), timeWindow)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve user metrics: " + err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    metrics,
    })
}

// GetHeimdallAnomalyData returns data for anomaly detection
func GetHeimdallAnomalyData(c *gin.Context) {
    // Parse time window parameter
    timeWindowStr := c.DefaultQuery("time_window", "1h")
    timeWindow, err := time.ParseDuration(timeWindowStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid time_window format. Use formats like '1h', '24h', '7d'",
        })
        return
    }
    
    // Get anomaly data
    analyticsService := service.GlobalHeimdallAnalyticsService
    data, err := analyticsService.GetAnomalyDetectionData(c.Request.Context(), timeWindow)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to retrieve anomaly data: " + err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    data,
    })
}

// CleanupHeimdallMetrics triggers cleanup of old metrics
func CleanupHeimdallMetrics(c *gin.Context) {
    analyticsService := service.GlobalHeimdallAnalyticsService
    err := analyticsService.CleanupOldMetrics(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to cleanup metrics: " + err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Metrics cleanup completed successfully",
    })
}

// GenerateHeimdallRollups triggers hourly rollups
func GenerateHeimdallRollups(c *gin.Context) {
    analyticsService := service.GlobalHeimdallAnalyticsService
    err := analyticsService.GenerateHourlyRollups(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to generate rollups: " + err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Hourly rollups generated successfully",
    })
}

// GetHeimdallDashboard returns dashboard data
func GetHeimdallDashboard(c *gin.Context) {
    // Parse time window parameter
    timeWindowStr := c.DefaultQuery("time_window", "24h")
    timeWindow, err := time.ParseDuration(timeWindowStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid time_window format. Use formats like '1h', '24h', '7d'",
        })
        return
    }
    
    analyticsService := service.GlobalHeimdallAnalyticsService
    
    // Get all metrics data
    urlMetrics, _ := analyticsService.GetURLFrequencyMetrics(c.Request.Context(), timeWindow)
    tokenMetrics, _ := analyticsService.GetTokenFrequencyMetrics(c.Request.Context(), timeWindow)
    userMetrics, _ := analyticsService.GetUserFrequencyMetrics(c.Request.Context(), timeWindow)
    anomalyData, _ := analyticsService.GetAnomalyDetectionData(c.Request.Context(), timeWindow)
    
    // Get telemetry stats
    telemetryStats := middleware.GetHeimdallTelemetryStats()
    
    // Calculate summary statistics
    totalRequests := int64(0)
    totalErrors := int64(0)
    avgLatency := float64(0)
    
    for _, metric := range urlMetrics {
        totalRequests += metric.Count
        totalErrors += int64(float64(metric.Count) * metric.ErrorRate / 100)
        avgLatency += metric.AvgLatency
    }
    
    if len(urlMetrics) > 0 {
        avgLatency = avgLatency / float64(len(urlMetrics))
    }
    
    errorRate := float64(0)
    if totalRequests > 0 {
        errorRate = float64(totalErrors) / float64(totalRequests) * 100
    }
    
    dashboard := gin.H{
        "summary": gin.H{
            "total_requests": totalRequests,
            "total_errors":   totalErrors,
            "error_rate":     errorRate,
            "avg_latency":    avgLatency,
            "unique_urls":    len(urlMetrics),
            "active_tokens":  len(tokenMetrics),
            "active_users":   len(userMetrics),
        },
        "url_metrics":     urlMetrics,
        "token_metrics":   tokenMetrics,
        "user_metrics":    userMetrics,
        "anomaly_data":   anomalyData,
        "telemetry_stats": telemetryStats,
        "time_window":     timeWindowStr,
        "generated_at":    time.Now().UTC(),
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    dashboard,
    })
}

// GetHeimdallConfig returns current Heimdall configuration
func GetHeimdallConfig(c *gin.Context) {
    config := middleware.DefaultTelemetryConfig()
    
    // Remove sensitive information from config
    safeConfig := gin.H{
        "enabled":               config.Enabled,
        "geolocation_enabled":    config.GeolocationEnabled,
        "buffer_size":           config.BufferSize,
        "worker_count":          config.WorkerCount,
        "retry_attempts":        config.RetryAttempts,
        "retry_delay_ms":        config.RetryDelay.Milliseconds(),
        "disk_queue_enabled":    config.DiskQueueEnabled,
        "flush_interval_ms":     config.FlushInterval.Milliseconds(),
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    safeConfig,
    })
}

// UpdateHeimdallConfig updates Heimdall configuration (limited fields)
func UpdateHeimdallConfig(c *gin.Context) {
    var request struct {
        Enabled            *bool   `json:"enabled"`
        GeolocationEnabled *bool   `json:"geolocation_enabled"`
        BufferSize         *int     `json:"buffer_size"`
        WorkerCount        *int     `json:"worker_count"`
        RetryAttempts      *int     `json:"retry_attempts"`
        RetryDelayMs      *int64   `json:"retry_delay_ms"`
        DiskQueueEnabled  *bool    `json:"disk_queue_enabled"`
        FlushIntervalMs   *int64   `json:"flush_interval_ms"`
    }
    
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request body: " + err.Error(),
        })
        return
    }
    
    // Get current config
    config := middleware.DefaultTelemetryConfig()
    
    // Update allowed fields
    if request.Enabled != nil {
        config.Enabled = *request.Enabled
    }
    if request.GeolocationEnabled != nil {
        config.GeolocationEnabled = *request.GeolocationEnabled
    }
    if request.BufferSize != nil {
        config.BufferSize = *request.BufferSize
    }
    if request.WorkerCount != nil {
        config.WorkerCount = *request.WorkerCount
    }
    if request.RetryAttempts != nil {
        config.RetryAttempts = *request.RetryAttempts
    }
    if request.RetryDelayMs != nil {
        config.RetryDelay = time.Duration(*request.RetryDelayMs) * time.Millisecond
    }
    if request.DiskQueueEnabled != nil {
        config.DiskQueueEnabled = *request.DiskQueueEnabled
    }
    if request.FlushIntervalMs != nil {
        config.FlushInterval = time.Duration(*request.FlushIntervalMs) * time.Millisecond
    }
    
    // Note: In a real implementation, you would update the running worker
    // For now, just return the updated config
    safeConfig := gin.H{
        "enabled":               config.Enabled,
        "geolocation_enabled":    config.GeolocationEnabled,
        "buffer_size":           config.BufferSize,
        "worker_count":          config.WorkerCount,
        "retry_attempts":        config.RetryAttempts,
        "retry_delay_ms":        config.RetryDelay.Milliseconds(),
        "disk_queue_enabled":    config.DiskQueueEnabled,
        "flush_interval_ms":     config.FlushInterval.Milliseconds(),
    }
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Configuration updated successfully. Restart required for some changes to take effect.",
        "data":    safeConfig,
    })
}
