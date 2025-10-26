package controller

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func parseQueryInt(c *gin.Context, key string, def int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return def
	}
	if v, err := strconv.Atoi(valueStr); err == nil {
		return v
	}
	return def
}

func GetPublicLogs(c *gin.Context) {
	if !common.PublicLogsFeatureEnabled {
		common.ApiErrorMsg(c, "公开日志未启用")
		return
	}
	query := service.PublicLogQuery{
		Window:   c.DefaultQuery("window", "24h"),
		Model:    c.Query("model"),
		Search:   c.Query("search"),
		Page:     parseQueryInt(c, "page", 1),
		PageSize: parseQueryInt(c, "page_size", 20),
	}
	result, err := service.GetPublicLogs(query)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, result)
}

func GetPublicLogModels(c *gin.Context) {
	if !common.PublicLogsFeatureEnabled {
		common.ApiErrorMsg(c, "公开日志未启用")
		return
	}
	models, err := service.GetPublicLogModels(c.DefaultQuery("window", "24h"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, models)
}

func ExportPublicLogs(c *gin.Context) {
	if !common.PublicLogsFeatureEnabled {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "公开日志未启用"})
		return
	}
	query := service.PublicLogQuery{
		Window: c.DefaultQuery("window", "24h"),
		Model:  c.Query("model"),
		Search: c.Query("search"),
	}
	rows, err := service.ExportPublicLogs(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	filename := "public-logs-" + time.Now().Format("20060102-150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	writer := csv.NewWriter(c.Writer)
	if err := service.WriteCSV(writer, rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
}

func GetPublicLeaderboard(c *gin.Context) {
	window := c.DefaultQuery("window", "24h")
	limit := parseQueryInt(c, "limit", 50)
	entries, err := service.GetModelLeaderboard(window, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, entries)
}

func GetAdminLeaderboard(c *gin.Context) {
	window := c.DefaultQuery("window", "24h")
	limit := parseQueryInt(c, "limit", 100)
	entries, err := service.GetModelLeaderboard(window, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, entries)
}

func ExportAdminLeaderboard(c *gin.Context) {
	window := c.DefaultQuery("window", "24h")
	limit := parseQueryInt(c, "limit", 100)
	entries, err := service.GetModelLeaderboard(window, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	rows := [][]string{{"Model", "Requests", "Tokens", "Unique Users", "Unique Tokens"}}
	for _, entry := range entries {
		rows = append(rows, []string{
			entry.Model,
			strconv.FormatInt(entry.RequestCount, 10),
			strconv.FormatInt(entry.TokenCount, 10),
			strconv.FormatInt(entry.UniqueUsers, 10),
			strconv.FormatInt(entry.UniqueTokens, 10),
		})
	}
	filename := "model-leaderboard-" + time.Now().Format("20060102-150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	writer := csv.NewWriter(c.Writer)
	if err := service.WriteCSV(writer, rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
}

func GetTokenIPUsage(c *gin.Context) {
	id := parseQueryInt(c, "id", 0)
	if param := c.Param("id"); param != "" {
		if v, err := strconv.Atoi(param); err == nil {
			id = v
		}
	}
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的令牌ID")
		return
	}
	window := c.DefaultQuery("window", "30d")
	summary, err := service.GetTokenIPUsageSummary(id, window)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, summary)
}

func GetUserIPUsage(c *gin.Context) {
	id := parseQueryInt(c, "id", 0)
	if param := c.Param("id"); param != "" {
		if v, err := strconv.Atoi(param); err == nil {
			id = v
		}
	}
	if id <= 0 {
		common.ApiErrorMsg(c, "无效的用户ID")
		return
	}
	window := c.DefaultQuery("window", "30d")
	summary, err := service.GetUserIPUsageSummary(id, window)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, summary)
}
