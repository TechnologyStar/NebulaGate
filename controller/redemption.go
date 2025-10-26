package controller

import (
    "errors"
    "net/http"
    "strconv"
    "unicode/utf8"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"

    "github.com/gin-gonic/gin"
)

func GetAllRedemptions(c *gin.Context) {
    pageInfo := common.GetPageQuery(c)
    redemptions, total, err := model.GetAllRedemptions(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
    if err != nil {
        common.ApiError(c, err)
        return
    }
    pageInfo.SetTotal(int(total))
    pageInfo.SetItems(redemptions)
    common.ApiSuccess(c, pageInfo)
    return
}

func SearchRedemptions(c *gin.Context) {
    keyword := c.Query("keyword")
    pageInfo := common.GetPageQuery(c)
    redemptions, total, err := model.SearchRedemptions(keyword, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
    if err != nil {
        common.ApiError(c, err)
        return
    }
    pageInfo.SetTotal(int(total))
    pageInfo.SetItems(redemptions)
    common.ApiSuccess(c, pageInfo)
    return
}

func GetRedemption(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        common.ApiError(c, err)
        return
    }
    redemption, err := model.GetRedemptionById(id)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    redemption,
    })
    return
}

func AddRedemption(c *gin.Context) {
    redemption := model.Redemption{}
    err := c.ShouldBindJSON(&redemption)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    if utf8.RuneCountInString(redemption.Name) == 0 || utf8.RuneCountInString(redemption.Name) > 20 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "兑换码名称长度必须在1-20之间",
        })
        return
    }
    if redemption.Count <= 0 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "兑换码个数必须大于0",
        })
        return
    }
    if redemption.Count > 100 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "一次兑换码批量生成的个数不能大于 100",
        })
        return
    }
    if redemption.PlanId <= 0 && redemption.Quota <= 0 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "兑换码至少需要包含套餐或额度之一",
        })
        return
    }
    if redemption.MaxUses == 0 {
        redemption.MaxUses = 1
    }
    if redemption.MaxUses < 0 {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": "兑换码最大使用次数不能为负数",
        })
        return
    }
    if err := validateExpiredTime(redemption.ExpiredTime); err != nil {
        c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
        return
    }
    if redemption.PlanId > 0 {
        if _, err := model.GetPlanById(redemption.PlanId); err != nil {
            c.JSON(http.StatusOK, gin.H{
                "success": false,
                "message": "关联的套餐不存在",
            })
            return
        }
    }
    var keys []string
    for i := 0; i < redemption.Count; i++ {
        key := common.GetUUID()
        cleanRedemption := model.Redemption{
            UserId:      c.GetInt("id"),
            Name:        redemption.Name,
            Key:         key,
            CreatedTime: common.GetTimestamp(),
            Quota:       redemption.Quota,
            PlanId:      redemption.PlanId,
            ExpiredTime: redemption.ExpiredTime,
            MaxUses:     redemption.MaxUses,
        }
        err = cleanRedemption.Insert()
        if err != nil {
            c.JSON(http.StatusOK, gin.H{
                "success": false,
                "message": err.Error(),
                "data":    keys,
            })
            return
        }
        keys = append(keys, key)
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    keys,
    })
    return
}

func DeleteRedemption(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    err := model.DeleteRedemptionById(id)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
    })
    return
}

func UpdateRedemption(c *gin.Context) {
    statusOnly := c.Query("status_only")
    redemption := model.Redemption{}
    err := c.ShouldBindJSON(&redemption)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    cleanRedemption, err := model.GetRedemptionById(redemption.Id)
    if err != nil {
        common.ApiError(c, err)
        return
    }
    if statusOnly == "" {
        if err := validateExpiredTime(redemption.ExpiredTime); err != nil {
            c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
            return
        }
        // If you add more fields, please also update redemption.Update()
        cleanRedemption.Name = redemption.Name
        cleanRedemption.Quota = redemption.Quota
        cleanRedemption.PlanId = redemption.PlanId
        cleanRedemption.ExpiredTime = redemption.ExpiredTime
        cleanRedemption.MaxUses = redemption.MaxUses
    }
    if statusOnly != "" {
        cleanRedemption.Status = redemption.Status
    }
    err = cleanRedemption.Update()
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    cleanRedemption,
    })
    return
}

func DeleteInvalidRedemption(c *gin.Context) {
    rows, err := model.DeleteInvalidRedemptions()
    if err != nil {
        common.ApiError(c, err)
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "",
        "data":    rows,
    })
    return
}

func validateExpiredTime(expired int64) error {
    if expired != 0 && expired < common.GetTimestamp() {
        return errors.New("过期时间不能早于当前时间")
    }
    return nil
}
