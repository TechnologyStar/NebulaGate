package model

import (
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model/migrations"

	"gorm.io/gorm"
)

type legacyPlan struct {
	Id                     int            `gorm:"column:id"`
	Code                   string         `gorm:"column:code;size:64;not null;uniqueIndex:uk_plans_code"`
	Name                   string         `gorm:"column:name;size:128;not null"`
	Description            string         `gorm:"column:description;type:text"`
	CycleType              string         `gorm:"column:cycle_type;size:16;not null"`
	CycleDurationDays      int            `gorm:"column:cycle_duration_days;not null;default:0"`
	QuotaMetric            string         `gorm:"column:quota_metric;size:16;not null"`
	QuotaAmount            int64          `gorm:"column:quota_amount;type:bigint;not null"`
	AllowCarryOver         bool           `gorm:"column:allow_carry_over;not null;default:false"`
	CarryLimitPercent      int            `gorm:"column:carry_limit_percent;not null;default:0"`
	UpstreamAliasWhitelist JSONValue      `gorm:"column:upstream_alias_whitelist;type:json"`
	IsActive               bool           `gorm:"column:is_active;not null;default:true"`
	IsPublic               bool           `gorm:"column:is_public;not null;default:false"`
	IsSystem               bool           `gorm:"column:is_system;not null;default:false"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (legacyPlan) TableName() string { return "plans" }

type legacyPlanAssignment struct {
	Id                  int            `gorm:"column:id"`
	SubjectType         string         `gorm:"column:subject_type;size:16;not null;index:idx_plan_assignments_subject,priority:1"`
	SubjectId           int            `gorm:"column:subject_id;not null;index:idx_plan_assignments_subject,priority:2"`
	PlanId              int            `gorm:"column:plan_id;not null;index"`
	BillingMode         string         `gorm:"column:billing_mode;size:16;not null;default:'plan'"`
	ActivatedAt         time.Time      `gorm:"column:activated_at;not null;index:idx_plan_assignments_window,priority:1"`
	DeactivatedAt       *time.Time     `gorm:"column:deactivated_at;index:idx_plan_assignments_window,priority:2"`
	RolloverPolicy      string         `gorm:"column:rollover_policy;size:16;not null;default:'none'"`
	RolloverAmount      int64          `gorm:"column:rollover_amount;type:bigint;not null;default:0"`
	RolloverExpiresAt   *time.Time     `gorm:"column:rollover_expires_at"`
	AutoFallbackEnabled bool           `gorm:"column:auto_fallback_enabled;not null;default:false"`
	FallbackPlanId      *int           `gorm:"column:fallback_plan_id;index"`
	Metadata            JSONValue      `gorm:"column:metadata;type:json"`
	CreatedAt           time.Time      `gorm:"column:created_at"`
	UpdatedAt           time.Time      `gorm:"column:updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (legacyPlanAssignment) TableName() string { return "plan_assignments" }

type legacyUsageCounter UsageCounter

type legacyVoucherBatch VoucherBatch

type legacyVoucherRedemption VoucherRedemption

type legacyRequestFlag RequestFlag

type legacyRequestLog RequestLog

type legacyRequestAggregate RequestAggregate

func init() {
	migrations.RegisterSchemaProvider(migrations.BillingGovernanceVersion, func() []interface{} {
		return []interface{}{
			&legacyPlan{},
			&legacyPlanAssignment{},
			&UsageCounter{},
			&VoucherBatch{},
			&VoucherRedemption{},
			&RequestFlag{},
			&RequestLog{},
			&RequestAggregate{},
		}
	})
}
