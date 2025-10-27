package controller

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

func GetUserLeaderboard(c *gin.Context) {
	window := c.DefaultQuery("window", "24h")
	limit := parseQueryInt(c, "limit", 100)

	entries, err := service.GetUserLeaderboard(window, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, entries)
}

func GetUserStats(c *gin.Context) {
	userId := c.GetInt("id")
	window := c.DefaultQuery("window", "24h")

	stats, err := service.GetUserStats(userId, window)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, stats)
}
