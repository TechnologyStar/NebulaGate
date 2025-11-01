package dto

// PlanCreateRequest represents the payload to create a billing plan.
// Validation tags are placeholders for future use.
type PlanCreateRequest struct {
    Name             string   `json:"name" validate:"required,min=1,max=64"`
    Description      string   `json:"description,omitempty" validate:"max=512"`
    Cycle            string   `json:"cycle" validate:"required,oneof=daily monthly custom"`
    CycleLengthDays  int      `json:"cycle_length_days,omitempty" validate:"omitempty,min=1,max=365"`
    Quota            int64    `json:"quota,omitempty" validate:"omitempty,min=0"`
    QuotaMetric      string   `json:"quota_metric,omitempty" validate:"omitempty,oneof=requests tokens"`
    RolloverPolicy   string   `json:"rollover_policy,omitempty" validate:"omitempty,oneof=none carry_all cap"`
    TokenLimit       int64    `json:"token_limit,omitempty" validate:"omitempty,min=0"`
    AllowedModels    []string `json:"allowed_models,omitempty"`
    ValidityDays     int      `json:"validity_days,omitempty" validate:"omitempty,min=0"`
    Price            float64  `json:"price,omitempty" validate:"omitempty,min=0"`
    Status           string   `json:"status,omitempty" validate:"omitempty,oneof=draft active archived"`
}

// PlanUpdateRequest represents the payload to update a billing plan.
type PlanUpdateRequest struct {
    Name             *string   `json:"name,omitempty" validate:"omitempty,min=1,max=64"`
    Description      *string   `json:"description,omitempty" validate:"omitempty,max=512"`
    Cycle            *string   `json:"cycle,omitempty" validate:"omitempty,oneof=daily monthly custom"`
    CycleLengthDays  *int      `json:"cycle_length_days,omitempty" validate:"omitempty,min=1,max=365"`
    Quota            *int64    `json:"quota,omitempty" validate:"omitempty,min=0"`
    QuotaMetric      *string   `json:"quota_metric,omitempty" validate:"omitempty,oneof=requests tokens"`
    RolloverPolicy   *string   `json:"rollover_policy,omitempty" validate:"omitempty,oneof=none carry_all cap"`
    TokenLimit       *int64    `json:"token_limit,omitempty" validate:"omitempty,min=0"`
    AllowedModels    *[]string `json:"allowed_models,omitempty"`
    ValidityDays     *int      `json:"validity_days,omitempty" validate:"omitempty,min=0"`
    Price            *float64  `json:"price,omitempty" validate:"omitempty,min=0"`
    Status           *string   `json:"status,omitempty" validate:"omitempty,oneof=draft active archived"`
}

// PlanView is a lightweight response model for listing and viewing plan data.
type PlanView struct {
    ID              int      `json:"id"`
    Name            string   `json:"name"`
    Description     string   `json:"description,omitempty"`
    Cycle           string   `json:"cycle"`
    CycleLengthDays int      `json:"cycle_length_days,omitempty"`
    Quota           int64    `json:"quota,omitempty"`
    QuotaMetric     string   `json:"quota_metric,omitempty"`
    RolloverPolicy  string   `json:"rollover_policy,omitempty"`
    TokenLimit      int64    `json:"token_limit,omitempty"`
    AllowedModels   []string `json:"allowed_models,omitempty"`
    ValidityDays    int      `json:"validity_days,omitempty"`
    Price           float64  `json:"price,omitempty"`
    Status          string   `json:"status"`
}

// AssignmentTarget describes where a plan is assigned (user or token).
type AssignmentTarget struct {
    SubjectType string `json:"subject_type"` // user | token
    SubjectID   int    `json:"subject_id"`
}

// PlanAssignmentRequest is used to assign a plan to a subject.
type PlanAssignmentRequest struct {
    PlanID   int               `json:"plan_id" validate:"required,min=1"`
    Targets  []AssignmentTarget `json:"targets" validate:"required,dive"`
    StartsAt int64             `json:"starts_at,omitempty"`
}
