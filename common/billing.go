package common

const (
    PlanCycleDaily   = "daily"
    PlanCycleMonthly = "monthly"
    PlanCycleCustom  = "custom"
)

const (
    PlanQuotaMetricRequests = "requests"
    PlanQuotaMetricTokens   = "tokens"
)

const (
    PlanStatusDraft    = "draft"
    PlanStatusActive   = "active"
    PlanStatusArchived = "archived"
)

const (
    AssignmentSubjectTypeUser  = "user"
    AssignmentSubjectTypeToken = "token"
)

const (
    BillingModePlan     = "plan"
    BillingModePrepaid  = "prepaid"
    BillingModeVoucher  = "voucher"
    BillingModeFallback = "fallback"
    // Additional placeholder mode for balance-based billing
    BillingModeBalance = "balance"
)

const (
    RolloverPolicyNone     = "none"
    RolloverPolicyCarryAll = "carry_all"
    RolloverPolicyCap      = "cap"
)

const (
    FlagReasonAbuse     = "abuse"
    FlagReasonViolation = "violation"
)

const (
    VoucherGrantTypeCredit = "credit"
    VoucherGrantTypePlan   = "plan"
)

var (
    PlanAssignmentsCacheEnabled  bool
    UsageCounterCacheEnabled     bool
    RequestAggregateCacheEnabled bool
)
