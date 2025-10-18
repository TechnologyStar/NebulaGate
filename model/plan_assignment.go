package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

type PlanAssignment struct {
	Id                  int            `json:"id"`
	SubjectType         string         `json:"subject_type" gorm:"size:16;not null;index:idx_plan_assignments_subject,priority:1"`
	SubjectId           int            `json:"subject_id" gorm:"not null;index:idx_plan_assignments_subject,priority:2"`
	PlanId              int            `json:"plan_id" gorm:"not null;index"`
	BillingMode         string         `json:"billing_mode" gorm:"size:16;not null;default:'plan'"`
	ActivatedAt         time.Time      `json:"activated_at" gorm:"not null;index:idx_plan_assignments_window,priority:1"`
	DeactivatedAt       *time.Time     `json:"deactivated_at" gorm:"index:idx_plan_assignments_window,priority:2"`
	RolloverPolicy      string         `json:"rollover_policy" gorm:"size:16;not null;default:'none'"`
	RolloverAmount      int64          `json:"rollover_amount" gorm:"type:bigint;not null;default:0"`
	RolloverExpiresAt   *time.Time     `json:"rollover_expires_at"`
	AutoFallbackEnabled bool           `json:"auto_fallback_enabled" gorm:"not null;default:false"`
	FallbackPlanId      *int           `json:"fallback_plan_id" gorm:"index"`
	Metadata            JSONValue      `json:"metadata" gorm:"type:json"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}

func (assignment *PlanAssignment) BeforeCreate(tx *gorm.DB) error {
	if assignment.SubjectType == "" {
		assignment.SubjectType = common.AssignmentSubjectTypeUser
	}
	if assignment.BillingMode == "" {
		assignment.BillingMode = common.BillingModePlan
	}
	if assignment.RolloverPolicy == "" {
		assignment.RolloverPolicy = common.RolloverPolicyNone
	}
	if assignment.ActivatedAt.IsZero() {
		assignment.ActivatedAt = time.Now().UTC()
	}
	return nil
}

func GetActivePlanAssignments(subjectType string, subjectId int, at time.Time) ([]*PlanAssignment, error) {
	if subjectType == "" {
		return nil, errors.New("subject type required")
	}
	if subjectId == 0 {
		return nil, errors.New("subject id required")
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	if common.PlanAssignmentsCacheEnabled && common.RedisEnabled {
		if assignments, ok := loadPlanAssignmentsFromCache(subjectType, subjectId, at); ok {
			return assignments, nil
		}
	}
	var assignments []*PlanAssignment
	err := DB.Where("subject_type = ? AND subject_id = ?", subjectType, subjectId).
		Where("activated_at <= ?", at).
		Where("deactivated_at IS NULL OR deactivated_at > ?", at).
		Order("activated_at DESC").
		Find(&assignments).Error
	if err != nil {
		return nil, err
	}
	if common.PlanAssignmentsCacheEnabled && common.RedisEnabled {
		storePlanAssignmentsCache(subjectType, subjectId, at, assignments)
	}
	return assignments, nil
}

func loadPlanAssignmentsFromCache(subjectType string, subjectId int, at time.Time) ([]*PlanAssignment, bool) {
	return nil, false
}

func storePlanAssignmentsCache(subjectType string, subjectId int, at time.Time, assignments []*PlanAssignment) {}
