package dto

// VoucherBatchCreateRequest describes a batch creation of vouchers.
type VoucherBatchCreateRequest struct {
	Count       int     `json:"count" validate:"required,min=1,max=1000"`
	Prefix      string  `json:"prefix,omitempty" validate:"omitempty,max=16"`
	GrantType   string  `json:"grant_type" validate:"required,oneof=credit plan"`
	CreditAmount int64  `json:"credit_amount,omitempty" validate:"omitempty,min=1"`
	PlanID      int     `json:"plan_id,omitempty" validate:"omitempty,min=1"`
	ExpireDays  int     `json:"expire_days,omitempty" validate:"omitempty,min=1,max=3650"`
	Note        string  `json:"note,omitempty" validate:"omitempty,max=128"`
}

// VoucherRedeemRequest describes the redemption request payload.
type VoucherRedeemRequest struct {
	Code string `json:"code" validate:"required"`
}

// VoucherRedeemResponse is a minimal response when redeeming a voucher.
type VoucherRedeemResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message,omitempty"`
	CreditAmount int64  `json:"credit_amount,omitempty"`
	PlanID       int    `json:"plan_id,omitempty"`
}
