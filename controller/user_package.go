package controller

import (
	"net/http"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func RedeemPackageCode(c *gin.Context) {
	var request struct {
		Code string `json:"code"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}

	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	userPackage, err := service.RedeemPackageCode(request.Code, userId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "兑换成功",
		"data":    userPackage,
	})
}

func GetUserPackages(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	packages, err := model.GetUserPackagesByUserId(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Update status based on current time and quota
	now := time.Now()
	for _, pkg := range packages {
		if pkg.Status == common.UserPackageStatusActive {
			if pkg.ExpireAt.Before(now) {
				pkg.Status = common.UserPackageStatusExpired
				_ = pkg.Update("status")
			} else if pkg.TokenQuota <= 0 {
				pkg.Status = common.UserPackageStatusExhausted
				_ = pkg.Update("status")
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    packages,
	})
}

func GetActiveUserPackages(c *gin.Context) {
	userId := c.GetInt("id")
	if userId == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "未授权",
		})
		return
	}

	packages, err := model.GetActiveUserPackages(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    packages,
	})
}
