package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

type RequestFlag struct {
	Id                 int        `json:"id"`
	RequestId          string     `json:"request_id" gorm:"size:64;not null;index"`
	SubjectType        string     `json:"subject_type" gorm:"size:16;not null;index:idx_request_flags_subject,priority:1"`
	SubjectId          int        `json:"subject_id" gorm:"not null;index:idx_request_flags_subject,priority:2"`
	UserId             *int       `json:"user_id" gorm:"index"`
	TokenId            *int       `json:"token_id" gorm:"index"`
	Reason             string     `json:"reason" gorm:"size:16;not null"`
	ReroutedModelAlias string     `json:"rerouted_model_alias" gorm:"size:128"`
	TtlAt              *time.Time `json:"ttl_at" gorm:"index"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (flag *RequestFlag) BeforeCreate(tx *gorm.DB) error {
	if flag.SubjectType == "" {
		flag.SubjectType = common.AssignmentSubjectTypeUser
	}
	if flag.Reason == "" {
		flag.Reason = common.FlagReasonViolation
	}
	return nil
}
