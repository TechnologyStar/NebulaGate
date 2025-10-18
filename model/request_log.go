package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

type RequestLog struct {
	Id                    int            `json:"id"`
	RequestId             string         `json:"request_id" gorm:"size:64;not null;uniqueIndex"`
	OccurredAt            time.Time      `json:"occurred_at" gorm:"not null;index"`
	ModelAlias            string         `json:"model_alias" gorm:"size:128;index:idx_request_logs_model_window,priority:1"`
	UpstreamProvider      string         `json:"upstream_provider" gorm:"size:64;index:idx_request_logs_model_window,priority:2"`
	SubjectType           string         `json:"subject_type" gorm:"size:16;not null;index:idx_request_logs_subject,priority:1"`
	AnonymizedSubjectHash string         `json:"anonymized_subject_hash" gorm:"size:128;not null;index:idx_request_logs_subject,priority:2"`
	PlanId                *int           `json:"plan_id" gorm:"index"`
	PlanAssignmentId      *int           `json:"plan_assignment_id" gorm:"index"`
	UsageMetric           string         `json:"usage_metric" gorm:"size:16;not null;default:'requests'"`
	PromptTokens          int64          `json:"prompt_tokens" gorm:"type:bigint;not null;default:0"`
	CompletionTokens      int64          `json:"completion_tokens" gorm:"type:bigint;not null;default:0"`
	TotalTokens           int64          `json:"total_tokens" gorm:"type:bigint;not null;default:0"`
	LatencyMs             int64          `json:"latency_ms" gorm:"not null;default:0"`
	FlagIds               JSONValue      `json:"flag_ids" gorm:"type:json"`
	Metadata              JSONValue      `json:"metadata" gorm:"type:json"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

func (log *RequestLog) BeforeCreate(tx *gorm.DB) error {
	if log.SubjectType == "" {
		log.SubjectType = common.AssignmentSubjectTypeUser
	}
	if log.OccurredAt.IsZero() {
		log.OccurredAt = time.Now().UTC()
	}
	if log.UsageMetric == "" {
		log.UsageMetric = common.PlanQuotaMetricRequests
	}
	if log.TotalTokens == 0 {
		log.TotalTokens = log.PromptTokens + log.CompletionTokens
	}
	return nil
}
