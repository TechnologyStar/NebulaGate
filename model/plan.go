package model

import (
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model/migrations"

    "gorm.io/gorm"
)

var billingSchemaVersion = migrations.BillingGovernanceVersion

type Plan struct {
    Id                       int            `json:"id"`
    Code                     string         `json:"code" gorm:"size:64;not null;uniqueIndex:uk_plans_code"`
    Name                     string         `json:"name" gorm:"size:128;not null"`
    Description              string         `json:"description" gorm:"type:text"`
    CycleType                string         `json:"cycle_type" gorm:"size:16;not null"`
    CycleDurationDays        int            `json:"cycle_duration_days" gorm:"not null;default:0"`
    QuotaMetric              string         `json:"quota_metric" gorm:"size:16;not null"`
    QuotaAmount              int64          `json:"quota_amount" gorm:"type:bigint;not null"`
    // Carry-over configuration
    AllowCarryOver           bool           `json:"allow_carry_over" gorm:"not null;default:false"`
    CarryLimitPercent        int            `json:"carry_limit_percent" gorm:"not null;default:0"`
    UpstreamAliasWhitelist   JSONValue      `json:"upstream_alias_whitelist" gorm:"type:json"`
    IsActive                 bool           `json:"is_active" gorm:"not null;default:true"`
    IsPublic                 bool           `json:"is_public" gorm:"not null;default:false"`
    IsSystem                 bool           `json:"is_system" gorm:"not null;default:false"`
    CreatedAt                time.Time      `json:"created_at"`
    UpdatedAt                time.Time      `json:"updated_at"`
    DeletedAt                gorm.DeletedAt `json:"-" gorm:"index"`
}

func (plan *Plan) BeforeCreate(tx *gorm.DB) error {
    if plan.CycleType == "" {
        plan.CycleType = common.PlanCycleMonthly
    }
    if plan.QuotaMetric == "" {
        plan.QuotaMetric = common.PlanQuotaMetricRequests
    }
    return nil
}

func GetPlanById(id int) (*Plan, error) {
    if id == 0 {
        return nil, nil
    }
    var plan Plan
    err := DB.Where("id = ? AND deleted_at IS NULL", id).First(&plan).Error
    if err != nil {
        return nil, err
    }
    return &plan, nil
}

func init() {
    migrations.RegisterSchemaProvider(billingSchemaVersion, func() []interface{} {
        return []interface{}{
            &Plan{},
            &PlanAssignment{},
            &UsageCounter{},
            &VoucherBatch{},
            &VoucherRedemption{},
            &RequestFlag{},
            &RequestLog{},
            &RequestAggregate{},
        }
    })
}
