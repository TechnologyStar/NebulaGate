package router

import (
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

// SetHeimdallRouter configures routes for Heimdall telemetry endpoints
func SetHeimdallRouter(router *gin.Engine) {
	// Heimdall API routes
	heimdallRouter := router.Group("/heimdall")
	heimdallRouter.Use(middleware.UserAuth()) // Require authentication
	{
		// Telemetry stats and configuration
		heimdallRouter.GET("/stats", controller.GetHeimdallTelemetryStats)
		heimdallRouter.GET("/config", controller.GetHeimdallConfig)
		heimdallRouter.PUT("/config", controller.UpdateHeimdallConfig)
		
		// Metrics endpoints
		heimdallRouter.GET("/metrics/urls", controller.GetHeimdallURLMetrics)
		heimdallRouter.GET("/metrics/tokens", controller.GetHeimdallTokenMetrics)
		heimdallRouter.GET("/metrics/users", controller.GetHeimdallUserMetrics)
		heimdallRouter.GET("/metrics/anomaly", controller.GetHeimdallAnomalyData)
		
		// Dashboard
		heimdallRouter.GET("/dashboard", controller.GetHeimdallDashboard)
		
		// Management endpoints (admin only)
		adminRouter := heimdallRouter.Group("/admin")
		adminRouter.Use(middleware.AdminAuth()) // Require admin authentication
		{
			adminRouter.POST("/cleanup", controller.CleanupHeimdallMetrics)
			adminRouter.POST("/rollups", controller.GenerateHeimdallRollups)
		}
	}
}
