package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func CheckIn(c *gin.Context) {
	userId := c.GetInt("id")

	quotaAwarded, consecutiveDays, err := model.CheckIn(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "签到成功",
		"data": gin.H{
			"quota_awarded":    quotaAwarded,
			"consecutive_days": consecutiveDays,
		},
	})
}

func GetCheckInStatus(c *gin.Context) {
	userId := c.GetInt("id")

	hasCheckedIn, record, err := model.GetTodayCheckInStatus(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	consecutiveDays, err := model.GetUserConsecutiveDays(userId)
	if err != nil {
		consecutiveDays = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"has_checked_in":   hasCheckedIn,
			"today_record":     record,
			"consecutive_days": consecutiveDays,
		},
	})
}

func GetCheckInHistory(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)

	records, total, err := model.GetCheckInHistory(userId, pageInfo.GetPageSize(), pageInfo.GetStartIdx())
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}
