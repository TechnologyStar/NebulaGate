package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

type VoucherBatch struct {
	Id                   int            `json:"id"`
	CodePrefix           string         `json:"code_prefix" gorm:"size:48;not null;uniqueIndex:uk_voucher_code_prefix"`
	BatchLabel           string         `json:"batch_label" gorm:"size:128;not null"`
	GrantType            string         `json:"grant_type" gorm:"size:16;not null;default:'credit'"`
	CreditAmount         int64          `json:"credit_amount" gorm:"type:bigint;not null;default:0"`
	PlanGrantId          *int           `json:"plan_grant_id" gorm:"index"`
	PlanGrantDuration    *int           `json:"plan_grant_duration"`
	IsStackable          bool           `json:"is_stackable" gorm:"not null;default:false"`
	MaxRedemptions       int            `json:"max_redemptions" gorm:"not null;default:0"`
	MaxPerSubject        int            `json:"max_per_subject" gorm:"not null;default:0"`
	ValidFrom            *time.Time     `json:"valid_from" gorm:"index"`
	ValidUntil           *time.Time     `json:"valid_until" gorm:"index"`
	Metadata             JSONValue      `json:"metadata" gorm:"type:json"`
	CreatedBy            string         `json:"created_by" gorm:"size:64"`
	Notes                string         `json:"notes" gorm:"type:text"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

type VoucherRedemption struct {
	Id                int        `json:"id"`
	VoucherBatchId    int        `json:"voucher_batch_id" gorm:"not null;index"`
	Code              string     `json:"code" gorm:"size:96;not null;uniqueIndex:uk_voucher_redemption_code"`
	SubjectType       string     `json:"subject_type" gorm:"size:16;not null;index:idx_voucher_subject,priority:1"`
	SubjectId         int        `json:"subject_id" gorm:"not null;index:idx_voucher_subject,priority:2"`
	PlanAssignmentId  *int       `json:"plan_assignment_id" gorm:"index"`
	RedeemedAt        time.Time  `json:"redeemed_at" gorm:"not null"`
	RedeemedBy        string     `json:"redeemed_by" gorm:"size:64"`
	CreditAmount      int64      `json:"credit_amount" gorm:"type:bigint;not null;default:0"`
	PlanGrantedId     *int       `json:"plan_granted_id" gorm:"index"`
	Metadata          JSONValue  `json:"metadata" gorm:"type:json"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (batch *VoucherBatch) BeforeCreate(tx *gorm.DB) error {
	if batch.GrantType == "" {
		batch.GrantType = common.VoucherGrantTypeCredit
	}
	if batch.CreditAmount < 0 {
		batch.CreditAmount = 0
	}
	return nil
}

func (redemption *VoucherRedemption) BeforeCreate(tx *gorm.DB) error {
	if redemption.SubjectType == "" {
		redemption.SubjectType = common.AssignmentSubjectTypeUser
	}
	if redemption.RedeemedAt.IsZero() {
		redemption.RedeemedAt = time.Now().UTC()
	}
	return nil
}
