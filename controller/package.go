package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

func GetAllPackages(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	statusStr := c.Query("status")
	status := 0
	if statusStr != "" {
		status, _ = strconv.Atoi(statusStr)
	}

	packages, total, err := model.GetAllPackages(pageInfo.GetStartIdx(), pageInfo.GetPageSize(), status)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(packages)
	common.ApiSuccess(c, pageInfo)
}

func GetPackage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pkg, err := model.GetPackageById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    pkg,
	})
}

func CreatePackage(c *gin.Context) {
	var pkg model.Package
	err := c.ShouldBindJSON(&pkg)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if pkg.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "套餐名称不能为空",
		})
		return
	}

	if pkg.TokenQuota <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "Token额度必须大于0",
		})
		return
	}

	if pkg.ValidityDays <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "有效期必须大于0",
		})
		return
	}

	// Set default status if not provided
	if pkg.Status == 0 {
		pkg.Status = common.PackageStatusActive
	}

	err = pkg.Insert()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    pkg,
	})
}

func UpdatePackage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var pkg model.Package
	err = c.ShouldBindJSON(&pkg)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Get existing package
	existingPkg, err := model.GetPackageById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Update fields
	existingPkg.Name = pkg.Name
	existingPkg.Description = pkg.Description
	existingPkg.TokenQuota = pkg.TokenQuota
	existingPkg.ModelScope = pkg.ModelScope
	existingPkg.ValidityDays = pkg.ValidityDays
	existingPkg.Price = pkg.Price
	existingPkg.Status = pkg.Status

	err = existingPkg.Update()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    existingPkg,
	})
}

func DeletePackage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Check if there are unused redemption codes for this package
	codes, _, err := model.ListRedemptionCodes(0, 1, id, common.RedemptionCodeStatusUnused, "")
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if len(codes) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该套餐还有未使用的兑换码，无法删除",
		})
		return
	}

	err = model.DeletePackageById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

func GetPackageModels(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pkg, err := model.GetPackageById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var models []string
	if pkg.ModelScope != "" {
		err = json.Unmarshal([]byte(pkg.ModelScope), &models)
		if err != nil {
			common.ApiError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    models,
	})
}
