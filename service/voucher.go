package service

import (
    "errors"
    "fmt"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type VoucherRedeemResult struct {
    Success          bool
    Message          string
    CreditAmount     int64
    PlanID           int
    PlanAssignmentID int
}

func GenerateVoucherBatch(
    prefix string,
    count int,
    grantType string,
    creditAmount int64,
    planID int,
    validDays int,
    createdBy string,
    notes string,
) ([]string, error) {
    if count <= 0 || count > 1000 {
        return nil, errors.New("voucher count must be between 1 and 1000")
    }

    if grantType != common.VoucherGrantTypeCredit && grantType != common.VoucherGrantTypePlan {
        return nil, errors.New("invalid grant type")
    }

    if grantType == common.VoucherGrantTypeCredit && creditAmount <= 0 {
        return nil, errors.New("credit amount must be positive for credit vouchers")
    }

    if grantType == common.VoucherGrantTypePlan && planID <= 0 {
        return nil, errors.New("valid plan ID required for plan vouchers")
    }

    prefix = strings.TrimSpace(prefix)
    if prefix == "" {
        prefix = strings.ToUpper(uuid.New().String()[:6])
    } else {
        prefix = strings.ToUpper(prefix)
    }

    originalPrefix := prefix
    finalPrefix := ""
    for attempts := 0; attempts < 5; attempts++ {
        candidate := originalPrefix
        if attempts > 0 {
            candidate = fmt.Sprintf("%s%s", originalPrefix, strings.ToUpper(uuid.New().String()[:4]))
        }
        var existing int64
        err := model.DB.Model(&model.VoucherBatch{}).Where("code_prefix = ?", candidate).Count(&existing).Error
        if err != nil {
            return nil, fmt.Errorf("failed to verify prefix uniqueness: %w", err)
        }
        if existing == 0 {
            finalPrefix = candidate
            break
        }
    }

    if finalPrefix == "" {
        return nil, errors.New("unable to allocate unique voucher prefix")
    }

    prefix = finalPrefix

    var validFrom, validUntil *time.Time
    now := time.Now().UTC()
    validFrom = &now
    if validDays > 0 {
        until := now.Add(time.Duration(validDays) * 24 * time.Hour)
        validUntil = &until
    }

    batchLabel := fmt.Sprintf("%s_%s_%d", prefix, time.Now().Format("20060102"), count)

    batch := &model.VoucherBatch{
        CodePrefix:     prefix,
        BatchLabel:     batchLabel,
        GrantType:      grantType,
        CreditAmount:   creditAmount,
        IsStackable:    false,
        MaxRedemptions: count,
        MaxPerSubject:  1,
        ValidFrom:      validFrom,
        ValidUntil:     validUntil,
        CreatedBy:      createdBy,
        Notes:          notes,
    }

    if grantType == common.VoucherGrantTypePlan {
        batch.PlanGrantId = &planID
    }

    err := model.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(batch).Error; err != nil {
            return fmt.Errorf("failed to create voucher batch: %w", err)
        }

        voucherCodes := make([]*model.VoucherCode, count)
        for i := 0; i < count; i++ {
            code := generateVoucherCode(prefix)
            voucherCode := &model.VoucherCode{
                VoucherBatchId: batch.Id,
                Code:           code,
                Status:         model.VoucherCodeStatusAvailable,
            }
            if grantType == common.VoucherGrantTypePlan && planID > 0 {
                voucherCode.PlanId = &planID
            }
            voucherCodes[i] = voucherCode
        }

        if err := tx.CreateInBatches(voucherCodes, 100).Error; err != nil {
            return fmt.Errorf("failed to create voucher codes: %w", err)
        }

        return nil
    })
    if err != nil {
        return nil, err
    }

    codes := make([]string, count)
    voucherCodes, err := model.GetVoucherCodesByBatch(batch.Id, "")
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve generated codes: %w", err)
    }
    for i, vc := range voucherCodes {
        if i < count {
            codes[i] = vc.Code
        }
    }

    return codes, nil
}

func RedeemVoucher(code string, userId int, username string) (*VoucherRedeemResult, error) {
    if code == "" {
        return nil, errors.New("voucher code is required")
    }
    if userId <= 0 {
        return nil, errors.New("invalid user ID")
    }

    result := &VoucherRedeemResult{
        Success: false,
    }

    err := model.DB.Transaction(func(tx *gorm.DB) error {
        var voucherCode model.VoucherCode
        err := tx.Where("code = ? AND deleted_at IS NULL", code).First(&voucherCode).Error
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return errors.New("invalid voucher code")
            }
            return fmt.Errorf("error loading voucher code: %w", err)
        }

        if voucherCode.Status == model.VoucherCodeStatusRedeemed {
            return errors.New("voucher code already redeemed")
        }
        if voucherCode.Status == model.VoucherCodeStatusExpired {
            return errors.New("voucher code has expired")
        }

        var existingRedemption model.VoucherRedemption
        err = tx.Where("code = ?", code).First(&existingRedemption).Error
        if err == nil {
            return errors.New("voucher code already redeemed")
        }
        if !errors.Is(err, gorm.ErrRecordNotFound) {
            return fmt.Errorf("error checking redemption: %w", err)
        }

        var batch model.VoucherBatch
        err = tx.Where("id = ? AND deleted_at IS NULL", voucherCode.VoucherBatchId).First(&batch).Error
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return errors.New("voucher batch not found")
            }
            return fmt.Errorf("error loading voucher batch: %w", err)
        }

        now := time.Now().UTC()
        if batch.ValidFrom != nil && now.Before(*batch.ValidFrom) {
            return errors.New("voucher is not yet valid")
        }
        if batch.ValidUntil != nil && now.After(*batch.ValidUntil) {
            return errors.New("voucher has expired")
        }

        var redemptionCount int64
        err = tx.Model(&model.VoucherRedemption{}).Where("voucher_batch_id = ?", batch.Id).Count(&redemptionCount).Error
        if err != nil {
            return fmt.Errorf("error counting redemptions: %w", err)
        }
        if batch.MaxRedemptions > 0 && redemptionCount >= int64(batch.MaxRedemptions) {
            return errors.New("voucher batch has reached maximum redemptions")
        }

        if batch.MaxPerSubject > 0 {
            var userRedemptionCount int64
            err = tx.Model(&model.VoucherRedemption{}).
                Where("voucher_batch_id = ? AND subject_type = ? AND subject_id = ?",
                    batch.Id, common.AssignmentSubjectTypeUser, userId).
                Count(&userRedemptionCount).Error
            if err != nil {
                return fmt.Errorf("error counting user redemptions: %w", err)
            }
            if userRedemptionCount >= int64(batch.MaxPerSubject) {
                return errors.New("you have already redeemed this voucher")
            }
        }

        redemption := &model.VoucherRedemption{
            VoucherBatchId: batch.Id,
            Code:           code,
            SubjectType:    common.AssignmentSubjectTypeUser,
            SubjectId:      userId,
            RedeemedAt:     now,
            RedeemedBy:     username,
            CreditAmount:   batch.CreditAmount,
        }

        if batch.GrantType == common.VoucherGrantTypeCredit {
            err = model.IncreaseUserQuota(userId, int(batch.CreditAmount), false)
            if err != nil {
                return fmt.Errorf("failed to increase user quota: %w", err)
            }
            result.CreditAmount = batch.CreditAmount
            result.Message = fmt.Sprintf("Successfully redeemed %d credits", batch.CreditAmount)
        } else if batch.GrantType == common.VoucherGrantTypePlan {
            targetPlanId := 0
            if voucherCode.PlanId != nil && *voucherCode.PlanId > 0 {
                targetPlanId = *voucherCode.PlanId
            } else if batch.PlanGrantId != nil && *batch.PlanGrantId > 0 {
                targetPlanId = *batch.PlanGrantId
            }
            if targetPlanId <= 0 {
                return errors.New("voucher does not reference a valid plan")
            }

            var plan model.Plan
            err = tx.Where("id = ?", targetPlanId).First(&plan).Error
            if err != nil {
                return fmt.Errorf("plan not found: %w", err)
            }

            assignment := &model.PlanAssignment{
                SubjectType:    common.AssignmentSubjectTypeUser,
                SubjectId:      userId,
                PlanId:         targetPlanId,
                BillingMode:    common.BillingModePlan,
                ActivatedAt:    now,
                RolloverPolicy: common.RolloverPolicyNone,
            }

            if plan.ValidityDays > 0 {
                expiresAt := now.Add(time.Duration(plan.ValidityDays) * 24 * time.Hour)
                assignment.ExpiresAt = &expiresAt
            }

            err = tx.Create(assignment).Error
            if err != nil {
                return fmt.Errorf("failed to create plan assignment: %w", err)
            }

            redemption.PlanAssignmentId = &assignment.Id
            redemption.PlanGrantedId = &targetPlanId

            result.PlanID = targetPlanId
            result.PlanAssignmentID = assignment.Id
            result.Message = fmt.Sprintf("Successfully redeemed plan: %s", plan.Name)
        }

        if err := tx.Create(redemption).Error; err != nil {
            return fmt.Errorf("failed to record redemption: %w", err)
        }

        update := map[string]interface{}{
            "status":              model.VoucherCodeStatusRedeemed,
            "redeemed_at":         now,
            "redeemed_by_user_id": userId,
        }
        if err := tx.Model(&model.VoucherCode{}).
            Where("id = ? AND status IN ?", voucherCode.Id, []string{model.VoucherCodeStatusAvailable, model.VoucherCodeStatusIssued}).
            Updates(update).Error; err != nil {
            return fmt.Errorf("failed to update voucher code status: %w", err)
        }

        voucherCode.Status = model.VoucherCodeStatusRedeemed
        voucherCode.RedeemedAt = &now
        voucherCode.RedeemedByUserId = &userId

        result.Success = true
        return nil
    })

    if err != nil {
        result.Message = err.Error()
        return result, err
    }

    return result, nil
}

func generateVoucherCode(prefix string) string {
    uid := uuid.New().String()
    uid = uid[:8]
    return fmt.Sprintf("%s-%s", prefix, uid)
}

func extractPrefix(code string) string {
    for i, c := range code {
        if c == '-' {
            return code[:i]
        }
    }
    return code
}
