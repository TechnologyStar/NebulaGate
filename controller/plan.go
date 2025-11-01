package controller

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/dto"
    "github.com/QuantumNous/new-api/model"

    "github.com/gin-gonic/gin"
)

// GetAllPlans retrieves all active plans (admin endpoint)
func GetAllPlans(c *gin.Context) {
    var plans []model.Plan
    err := model.DB.Where("deleted_at IS NULL").Order("created_at DESC").Find(&plans).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }
    common.ApiSuccess(c, plans)
}

// GetPlan retrieves a single plan by ID (admin endpoint)
func GetPlan(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid plan ID",
        })
        return
    }

    var plan model.Plan
    err = model.DB.Where("id = ? AND deleted_at IS NULL", id).First(&plan).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }
    common.ApiSuccess(c, plan)
}

// CreatePlan creates a new billing plan (admin endpoint)
func CreatePlan(c *gin.Context) {
    var req dto.PlanCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }

    plan := model.Plan{
        Code:              generatePlanCode(req.Name),
        Name:              req.Name,
        Description:       req.Description,
        CycleType:         req.Cycle,
        CycleDurationDays: req.CycleLengthDays,
        QuotaMetric:       req.QuotaMetric,
        QuotaAmount:       req.Quota,
        TokenLimit:        req.TokenLimit,
        ValidityDays:      req.ValidityDays,
        IsActive:          true,
        IsPublic:          false,
    }

    if req.RolloverPolicy != "" {
        plan.AllowCarryOver = req.RolloverPolicy != common.RolloverPolicyNone
        if req.RolloverPolicy == common.RolloverPolicyCap {
            plan.CarryLimitPercent = 100
        }
    }

    if len(req.AllowedModels) > 0 {
        allowedModelsJSON, err := marshalJSONValue(req.AllowedModels)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "message": "Invalid allowed_models format: " + err.Error(),
            })
            return
        }
        plan.AllowedModels = allowedModelsJSON
    }

    err := model.DB.Create(&plan).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, plan)
}

// UpdatePlan updates an existing billing plan (admin endpoint)
func UpdatePlan(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid plan ID",
        })
        return
    }

    var req dto.PlanUpdateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }

    var plan model.Plan
    err = model.DB.Where("id = ?", id).First(&plan).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    updates := make(map[string]interface{})
    if req.Name != nil {
        updates["name"] = *req.Name
    }
    if req.Description != nil {
        updates["description"] = *req.Description
    }
    if req.Cycle != nil {
        updates["cycle_type"] = *req.Cycle
    }
    if req.CycleLengthDays != nil {
        updates["cycle_duration_days"] = *req.CycleLengthDays
    }
    if req.Quota != nil {
        updates["quota_amount"] = *req.Quota
    }
    if req.QuotaMetric != nil {
        updates["quota_metric"] = *req.QuotaMetric
    }
    if req.RolloverPolicy != nil {
        updates["allow_carry_over"] = *req.RolloverPolicy != common.RolloverPolicyNone
        if *req.RolloverPolicy == common.RolloverPolicyCap {
            updates["carry_limit_percent"] = 100
        }
    }
    if req.TokenLimit != nil {
        updates["token_limit"] = *req.TokenLimit
    }
    if req.AllowedModels != nil {
        allowedModelsJSON, err := marshalJSONValue(*req.AllowedModels)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "success": false,
                "message": "Invalid allowed_models format: " + err.Error(),
            })
            return
        }
        updates["allowed_models"] = allowedModelsJSON
    }
    if req.ValidityDays != nil {
        updates["validity_days"] = *req.ValidityDays
    }

    err = model.DB.Model(&plan).Updates(updates).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    err = model.DB.Where("id = ?", id).First(&plan).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, plan)
}

// DeletePlan soft deletes a billing plan (admin endpoint)
func DeletePlan(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid plan ID",
        })
        return
    }

    var plan model.Plan
    err = model.DB.Where("id = ?", id).First(&plan).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    err = model.DB.Delete(&plan).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Plan deleted successfully",
    })
}

// AssignPlan assigns a plan to one or more subjects (admin endpoint)
func AssignPlan(c *gin.Context) {
    var req dto.PlanAssignmentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid request: " + err.Error(),
        })
        return
    }

    if req.PlanID <= 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Valid plan_id is required",
        })
        return
    }

    if len(req.Targets) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "At least one target is required",
        })
        return
    }

    var plan model.Plan
    err := model.DB.Where("id = ?", req.PlanID).First(&plan).Error
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Plan not found",
        })
        return
    }

    activatedAt := time.Now().UTC()
    if req.StartsAt > 0 {
        activatedAt = time.Unix(req.StartsAt, 0).UTC()
    }

    assignments := make([]*model.PlanAssignment, 0, len(req.Targets))
    for _, target := range req.Targets {
        if target.SubjectType != common.AssignmentSubjectTypeUser && target.SubjectType != common.AssignmentSubjectTypeToken {
            continue
        }
        if target.SubjectID <= 0 {
            continue
        }

        assignment := &model.PlanAssignment{
            SubjectType:   target.SubjectType,
            SubjectId:     target.SubjectID,
            PlanId:        req.PlanID,
            BillingMode:   common.BillingModePlan,
            ActivatedAt:   activatedAt,
            RolloverPolicy: common.RolloverPolicyNone,
        }

        err := model.DB.Create(assignment).Error
        if err != nil {
            common.SysLog("failed to create assignment: " + err.Error())
            continue
        }
        assignments = append(assignments, assignment)
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Plan assigned successfully",
        "data": gin.H{
            "assigned_count": len(assignments),
            "assignments":    assignments,
        },
    })
}

// DetachPlanAssignment removes a plan assignment (admin endpoint)
func DetachPlanAssignment(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid assignment ID",
        })
        return
    }

    var assignment model.PlanAssignment
    err = model.DB.Where("id = ?", id).First(&assignment).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    now := time.Now().UTC()
    assignment.DeactivatedAt = &now

    err = model.DB.Save(&assignment).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Plan assignment deactivated successfully",
    })
}

// GetUserPlans retrieves active plan assignments for the authenticated user
func GetUserPlans(c *gin.Context) {
    userId := c.GetInt("id")
    if userId == 0 {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "Unauthorized",
        })
        return
    }

    assignments, err := model.GetActivePlanAssignments(common.AssignmentSubjectTypeUser, userId, time.Now().UTC())
    if err != nil {
        common.ApiError(c, err)
        return
    }

    planIds := make([]int, 0, len(assignments))
    for _, a := range assignments {
        planIds = append(planIds, a.PlanId)
    }

    var plans []model.Plan
    if len(planIds) > 0 {
        err = model.DB.Where("id IN ?", planIds).Find(&plans).Error
        if err != nil {
            common.ApiError(c, err)
            return
        }
    }

    type AssignmentWithPlan struct {
        Assignment model.PlanAssignment `json:"assignment"`
        Plan       *model.Plan          `json:"plan"`
    }

    result := make([]AssignmentWithPlan, 0, len(assignments))
    planMap := make(map[int]*model.Plan)
    for i := range plans {
        planMap[plans[i].Id] = &plans[i]
    }

    for _, a := range assignments {
        result = append(result, AssignmentWithPlan{
            Assignment: *a,
            Plan:       planMap[a.PlanId],
        })
    }

    common.ApiSuccess(c, result)
}

// GetPlanUsage retrieves usage counters for a plan assignment (admin endpoint)
func GetPlanUsage(c *gin.Context) {
    assignmentId, err := strconv.Atoi(c.Param("assignment_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "Invalid assignment ID",
        })
        return
    }

    var counters []model.UsageCounter
    err = model.DB.Where("plan_assignment_id = ?", assignmentId).Order("cycle_start DESC").Find(&counters).Error
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, counters)
}

func generatePlanCode(name string) string {
    return "plan_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

func marshalJSONValue(v interface{}) (model.JSONValue, error) {
    data, err := json.Marshal(v)
    if err != nil {
        return nil, err
    }
    return model.JSONValue(data), nil
}
