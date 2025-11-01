package model

import (
    "time"

    "gorm.io/gorm"
)

const (
    VoucherCodeStatusAvailable = "available"
    VoucherCodeStatusIssued    = "issued"
    VoucherCodeStatusRedeemed  = "redeemed"
    VoucherCodeStatusExpired   = "expired"
)

type VoucherCode struct {
    Id              int            `json:"id"`
    VoucherBatchId  int            `json:"voucher_batch_id" gorm:"not null;index"`
    PlanId          *int           `json:"plan_id" gorm:"index"`
    Code            string         `json:"code" gorm:"size:96;not null;uniqueIndex:uk_voucher_code"`
    Status          string         `json:"status" gorm:"size:16;not null;default:'available';index"`
    AssignedToUserId *int          `json:"assigned_to_user_id" gorm:"index"`
    AssignedToEmail *string        `json:"assigned_to_email" gorm:"size:128"`
    IssuedAt        *time.Time     `json:"issued_at"`
    RedeemedAt      *time.Time     `json:"redeemed_at"`
    RedeemedByUserId *int          `json:"redeemed_by_user_id" gorm:"index"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (code *VoucherCode) BeforeCreate(tx *gorm.DB) error {
    if code.Status == "" {
        code.Status = VoucherCodeStatusAvailable
    }
    return nil
}

func GetVoucherCodeByCode(code string) (*VoucherCode, error) {
    var voucherCode VoucherCode
    err := DB.Where("code = ? AND deleted_at IS NULL", code).First(&voucherCode).Error
    if err != nil {
        return nil, err
    }
    return &voucherCode, nil
}

func GetVoucherCodesByBatch(batchId int, status string) ([]*VoucherCode, error) {
    var codes []*VoucherCode
    query := DB.Where("voucher_batch_id = ? AND deleted_at IS NULL", batchId)
    if status != "" {
        query = query.Where("status = ?", status)
    }
    err := query.Order("created_at ASC").Find(&codes).Error
    if err != nil {
        return nil, err
    }
    return codes, nil
}
