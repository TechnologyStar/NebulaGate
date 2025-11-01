package controller

import (
    "net/http"
    "strconv"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/dto"
    "github.com/QuantumNous/new-api/model"
    "github.com/QuantumNous/new-api/service"

    "github.com/gin-gonic/gin"
)

// GenerateVouchers creates a batch of voucher codes (admin endpoint)
func GenerateVouchers(c *gin.Context) {
    var req dto.VoucherBatchCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }

    if req.Count <= 0 || req.Count > 1000 {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Count must be between 1 and 1000",
        })
        return
    }

    if req.GrantType != common.VoucherGrantTypeCredit && req.GrantType != common.VoucherGrantTypePlan {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid grant type, must be 'credit' or 'plan'",
        })
        return
    }

    if req.GrantType == common.VoucherGrantTypeCredit && req.CreditAmount <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Credit amount must be positive for credit vouchers",
        })
        return
    }

    if req.GrantType == common.VoucherGrantTypePlan && req.PlanID <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Valid plan ID required for plan vouchers",
        })
        return
    }

    username := c.GetString("username")
    if username == "" {
        username = "admin"
    }

    codes, err := service.GenerateVoucherBatch(
        req.Prefix,
        req.Count,
        req.GrantType,
        req.CreditAmount,
        req.PlanID,
        req.ExpireDays,
        username,
        req.Note,
    )

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "Failed to generate vouchers: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Vouchers generated successfully",
        "data": gin.H{
            "codes": codes,
            "count": len(codes),
        },
    })
}

// RedeemVoucher redeems a voucher code (user endpoint)
func RedeemVoucher(c *gin.Context) {
    var req dto.VoucherRedeemRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }

    if req.Code == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Voucher code is required",
        })
        return
    }

    userId := c.GetInt("id")
    if userId == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "Unauthorized",
        })
        return
    }

    username := c.GetString("username")
    if username == "" {
        username = "user_" + strconv.Itoa(userId)
    }

    result, err := service.RedeemVoucher(req.Code, userId, username)
    if err != nil {
        c.JSON(http.StatusOK, gin.H{
            "success": false,
            "message": result.Message,
        })
        return
    }

    response := dto.VoucherRedeemResponse{
        Success:      result.Success,
        Message:      result.Message,
        CreditAmount: result.CreditAmount,
        PlanID:       result.PlanID,
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": result.Message,
        "data":    response,
    })
}

// GetVoucherBatches retrieves all voucher batches (admin endpoint)
func GetVoucherBatches(c *gin.Context) {
    var batches []model.VoucherBatch
    err := model.DB.Where("deleted_at IS NULL").Order("created_at DESC").Find(&batches).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }
    common.ApiSuccess(c, batches)
}

// GetVoucherRedemptions retrieves all redemptions for a batch (admin endpoint)
func GetVoucherRedemptions(c *gin.Context) {
    batchId, err := strconv.Atoi(c.Param("batch_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid batch ID",
        })
        return
    }

    var redemptions []model.VoucherRedemption
    err = model.DB.Where("voucher_batch_id = ?", batchId).Order("redeemed_at DESC").Find(&redemptions).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, redemptions)
}

// GetVoucherCodes retrieves all voucher codes for a batch (admin endpoint)
func GetVoucherCodes(c *gin.Context) {
    batchId, err := strconv.Atoi(c.Param("batch_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid batch ID",
        })
        return
    }

    status := c.Query("status")
    codes, err := model.GetVoucherCodesByBatch(batchId, status)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, codes)
}

// IssueVoucherCode marks a voucher code as issued (admin endpoint)
func IssueVoucherCode(c *gin.Context) {
    codeId, err := strconv.Atoi(c.Param("code_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid code ID",
        })
        return
    }

    var req struct {
        AssignToUserId *int    `json:"assign_to_user_id,omitempty"`
        AssignToEmail  *string `json:"assign_to_email,omitempty"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }

    var voucherCode model.VoucherCode
    err = model.DB.Where("id = ? AND deleted_at IS NULL", codeId).First(&voucherCode).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    if voucherCode.Status != model.VoucherCodeStatusAvailable {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Voucher code is not available",
        })
        return
    }

    now := time.Now().UTC()
    updates := map[string]interface{}{
        "status":    model.VoucherCodeStatusIssued,
        "issued_at": now,
    }
    if req.AssignToUserId != nil {
        updates["assigned_to_user_id"] = *req.AssignToUserId
    }
    if req.AssignToEmail != nil {
        updates["assigned_to_email"] = *req.AssignToEmail
    }

    err = model.DB.Model(&voucherCode).Updates(updates).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Voucher code marked as issued",
    })
}

// RevokeVoucherCode marks a voucher code as expired (admin endpoint)
func RevokeVoucherCode(c *gin.Context) {
    codeId, err := strconv.Atoi(c.Param("code_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid code ID",
        })
        return
    }

    var voucherCode model.VoucherCode
    err = model.DB.Where("id = ? AND deleted_at IS NULL", codeId).First(&voucherCode).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    if voucherCode.Status == model.VoucherCodeStatusRedeemed {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Cannot revoke an already redeemed voucher code",
        })
        return
    }

    err = model.DB.Model(&voucherCode).Update("status", model.VoucherCodeStatusExpired).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Voucher code revoked",
    })
}
