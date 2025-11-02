package controller

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func GenerateRedemptionCodes(c *gin.Context) {
	var request struct {
		PackageId int `json:"package_id"`
		Quantity  int `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		common.ApiError(c, err)
		return
	}

	if request.PackageId <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "必须指定套餐",
		})
		return
	}

	if request.Quantity <= 0 || request.Quantity > 500 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "生成数量需在1-500之间",
		})
		return
	}

	codes, err := service.GenerateRedemptionCodes(request.PackageId, request.Quantity)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"codes": codes,
		},
	})
}

func GetRedemptionCodes(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	packageId, _ := strconv.Atoi(c.Query("package_id"))
	status, _ := strconv.Atoi(c.Query("status"))
	code := c.Query("code")

	codes, total, err := model.ListRedemptionCodes(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), packageId, status, code)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(codes)
	common.ApiSuccess(c, pageInfo)
}

func RevokeRedemptionCode(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	err = service.RevokeRedemptionCode(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func ExportRedemptionCodes(c *gin.Context) {
	packageId, _ := strconv.Atoi(c.Query("package_id"))
	status, _ := strconv.Atoi(c.Query("status"))
	code := c.Query("code")

	codes, _, err := model.ListRedemptionCodes(0, 1000, packageId, status, code)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=redemption_codes_%s.csv", time.Now().Format("20060102_150405")))
	c.Header("Cache-Control", "no-cache")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	_ = writer.Write([]string{"Code", "Package", "Status", "UsedBy", "UsedAt"})

	// Write data rows
	for _, rc := range codes {
		statusText := "未使用"
		switch rc.Status {
		case common.RedemptionCodeStatusRedeemed:
			statusText = "已使用"
		case common.RedemptionCodeStatusRevoked:
			statusText = "已作废"
		}

		usedBy := ""
		if rc.UsedByUserId > 0 {
			usedBy = strconv.Itoa(rc.UsedByUserId)
		}

		usedAt := ""
		if rc.UsedAt != nil {
			usedAt = rc.UsedAt.Format("2006-01-02 15:04:05")
		}

		pkgName := ""
		if rc.Package.Id > 0 {
			pkgName = rc.Package.Name
		}

		_ = writer.Write([]string{
			rc.Code,
			pkgName,
			statusText,
			usedBy,
			usedAt,
		})
	}
}
