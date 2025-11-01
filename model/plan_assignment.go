package model

import (
    "errors"
    "time"

    "github.com/QuantumNous/new-api/common"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type PlanAssignment struct {
    Id                  int            `json:"id"`
    SubjectType         string         `json:"subject_type" gorm:"size:16;not null;index:idx_plan_assignments_subject,priority:1"`
    SubjectId           int            `json:"subject_id" gorm:"not null;index:idx_plan_assignments_subject,priority:2"`
    PlanId              int            `json:"plan_id" gorm:"not null;index"`
    BillingMode         string         `json:"billing_mode" gorm:"size:16;not null;default:'plan'"`
    ActivatedAt         time.Time      `json:"activated_at" gorm:"not null;index:idx_plan_assignments_window,priority:1"`
    DeactivatedAt       *time.Time     `json:"deactivated_at" gorm:"index:idx_plan_assignments_window,priority:2"`
    ExpiresAt           *time.Time     `json:"expires_at" gorm:"index"`
    RolloverPolicy      string         `json:"rollover_policy" gorm:"size:16;not null;default:'none'"`
    RolloverAmount      int64          `json:"rollover_amount" gorm:"type:bigint;not null;default:0"`
    RolloverExpiresAt   *time.Time     `json:"rollover_expires_at"`
    AutoFallbackEnabled bool           `json:"auto_fallback_enabled" gorm:"not null;default:false"`
    FallbackPlanId      *int           `json:"fallback_plan_id" gorm:"index"`
    EnforcementMetadata JSONValue      `json:"enforcement_metadata" gorm:"type:json"`
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

// GetActivePlanAssignmentsTx loads active assignments using the provided transaction.
// If forUpdate is true, it will apply row-level locks where supported by the DB.
func GetActivePlanAssignmentsTx(tx *gorm.DB, subjectType string, subjectId int, at time.Time, forUpdate bool) ([]*PlanAssignment, error) {
    if tx == nil {
        return nil, errors.New("tx required")
    }
    if at.IsZero() {
        at = time.Now().UTC()
    }
    var assignments []*PlanAssignment
    q := tx.Where("subject_type = ? AND subject_id = ?", subjectType, subjectId).
        Where("activated_at <= ?", at).
        Where("deactivated_at IS NULL OR deactivated_at > ?", at).
        Order("activated_at DESC")
    if forUpdate {
        q = q.Clauses(clause.Locking{Strength: "UPDATE"})
    }
    if err := q.Find(&assignments).Error; err != nil {
        return nil, err
    }
    return assignments, nil
}

func GetPlanAssignmentById(id int) (*PlanAssignment, error) {
    if id == 0 {
        return nil, errors.New("id required")
    }
    var assignment PlanAssignment
    err := DB.Where("id = ?", id).First(&assignment).Error
    if err != nil {
        return nil, err
    }
    return &assignment, nil
}

func loadPlanAssignmentsFromCache(subjectType string, subjectId int, at time.Time) ([]*PlanAssignment, bool) {
    return nil, false
}

func storePlanAssignmentsCache(subjectType string, subjectId int, at time.Time, assignments []*PlanAssignment) {}
