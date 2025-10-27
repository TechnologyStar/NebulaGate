package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func DrawLottery(c *gin.Context) {
	userId := c.GetInt("id")

	record, err := model.DrawLottery(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "抽奖成功",
		"data":    record,
	})
}

func GetLotteryConfigs(c *gin.Context) {
	configs, err := model.GetLotteryConfigs()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, configs)
}

func CreateLotteryConfig(c *gin.Context) {
	var config model.LotteryConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		common.ApiError(c, err)
		return
	}
	if err := config.Create(); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, config)
}

func UpdateLotteryConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "无效的配置ID")
		return
	}

	config, err := model.GetLotteryConfigById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if err := c.ShouldBindJSON(config); err != nil {
		common.ApiError(c, err)
		return
	}

	if err := config.Update(); err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, config)
}

func DeleteLotteryConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "无效的配置ID")
		return
	}

	if err := model.DeleteLotteryConfig(id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

func GetLotteryRecords(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)

	records, total, err := model.GetLotteryRecords(userId, pageInfo.GetPageSize(), pageInfo.GetStartIdx())
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}

func GetAllLotteryRecords(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)

	records, total, err := model.GetAllLotteryRecords(pageInfo.GetPageSize(), pageInfo.GetStartIdx())
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(records)
	common.ApiSuccess(c, pageInfo)
}
