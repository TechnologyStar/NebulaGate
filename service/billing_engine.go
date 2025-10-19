package service

import (
    "context"
    "errors"
    "fmt"
    "strconv"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    relaycommon "github.com/QuantumNous/new-api/relay/common"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

// BillingEngine provides plan vs balance decision and atomic charge commit with audit log.
type BillingEngine struct {
    DB *gorm.DB
}

func NewBillingEngine(db *gorm.DB) *BillingEngine {
    if db == nil {
        db = model.DB
    }
    return &BillingEngine{DB: db}
}

// PreparedCharge describes the decision result from PrepareCharge.
type PreparedCharge struct {
    SubjectType string
    SubjectId   int
    // Mode decided to use for primary attempt (plan|balance|auto[fallback])
    Mode string
    // Selected assignment when Mode is plan or auto and plan exists
    Assignment *model.PlanAssignment
    Plan       *model.Plan
    CycleStart time.Time
    CycleEnd   time.Time
    // Remaining allowance in the selected plan (0 if none)
    Allowance int64
}

// CommitParams defines input for CommitCharge.
type CommitParams struct {
    Prepared         *PreparedCharge
    Amount           int64
    RequestId        string
    ModelAlias       string
    Upstream         string
    UsageMetric      string
    PromptTokens     int64
    CompletionTokens int64
    LatencyMs        int64
    RelayInfo        *relaycommon.RelayInfo
}

// PrepareCharge resolves active plan assignment and allowance. subjectKey format: "user:123" or "token:456".
func (be *BillingEngine) PrepareCharge(ctx context.Context, subjectKey string, relayInfo *relaycommon.RelayInfo) (*PreparedCharge, error) {
    subjectType, subjectId, err := parseSubjectKey(subjectKey, relayInfo)
    if err != nil {
        return nil, err
    }
    assignments, err := model.GetActivePlanAssignments(subjectType, subjectId, time.Now().UTC())
    if err != nil {
        return nil, err
    }
    // Determine desired mode: subject override -> global default
    desiredMode := strings.ToLower(common.BillingDefaultMode)
    if relayInfo != nil && relayInfo.BillingMode != "" {
        desiredMode = strings.ToLower(relayInfo.BillingMode)
    }
    // If globally configured to auto fallback, upgrade plan mode to fallback
    if desiredMode == common.BillingModePlan && common.BillingAutoFallbackEnabled {
        desiredMode = common.BillingModeFallback
    }
    pc := &PreparedCharge{
        SubjectType: subjectType,
        SubjectId:   subjectId,
        Mode:        desiredMode,
    }
    if len(assignments) == 0 {
        // No plan to use; force balance
        pc.Mode = common.BillingModeBalance
        return pc, nil
    }
    // Pick the most recent assignment
    assignment := assignments[0]
    pc.Assignment = assignment
    // If explicit override chooses non-plan charging, keep it; otherwise inherit assignment mode
    if desiredMode == common.BillingModePlan || desiredMode == common.BillingModeFallback {
        pc.Mode = desiredMode
    } else {
        pc.Mode = assignment.BillingMode
        if pc.Mode == common.BillingModePlan && common.BillingAutoFallbackEnabled {
            pc.Mode = common.BillingModeFallback
        }
    }
    // Load plan
    var plan model.Plan
    if err := be.DB.First(&plan, "id = ?", assignment.PlanId).Error; err != nil {
        return nil, err
    }
    pc.Plan = &plan
    start, end := getCycleWindow(plan.CycleType, time.Now().UTC())
    pc.CycleStart, pc.CycleEnd = start, end
    // Load current usage for the cycle
    var counter model.UsageCounter
    err = be.DB.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?",
        assignment.Id, plan.QuotaMetric, start).First(&counter).Error
    var consumed int64
    if err == nil {
        consumed = counter.ConsumedAmount
    } else if errors.Is(err, gorm.ErrRecordNotFound) {
        consumed = 0
    } else {
        return nil, err
    }
    allowance := plan.QuotaAmount + assignment.RolloverAmount - consumed
    if allowance < 0 {
        allowance = 0
    }
    pc.Allowance = allowance
    // When the assignment requests fallback or plan exhausted, we will decide on commit
    return pc, nil
}

// CommitCharge atomically applies the charge according to the prepared decision and writes audit log.
func (be *BillingEngine) CommitCharge(ctx context.Context, in *CommitParams) (*model.RequestLog, error) {
    if in == nil || in.Prepared == nil {
        return nil, fmt.Errorf("invalid commit input")
    }
    if in.RequestId == "" {
        return nil, fmt.Errorf("request id required")
    }
    pc := in.Prepared
    mode := strings.ToLower(pc.Mode)
    if mode == "auto" || mode == common.BillingModeFallback {
        // Decide based on allowance
        if pc.Allowance > 0 {
            mode = common.BillingModePlan
        } else {
            mode = common.BillingModeBalance
        }
    }
    // Defaults
    if in.UsageMetric == "" {
        if pc.Plan != nil {
            in.UsageMetric = pc.Plan.QuotaMetric
        } else {
            in.UsageMetric = common.PlanQuotaMetricRequests
        }
    }
    var resultLog *model.RequestLog
    err := be.DB.Transaction(func(tx *gorm.DB) error {
        // Idempotency barrier: try to create RequestLog first.
        log := BuildRequestLog(in.RequestId, pc.SubjectType, pc.SubjectId, in.ModelAlias, in.Upstream,
            pc.Assignment, in.UsageMetric, in.PromptTokens, in.CompletionTokens, in.PromptTokens+in.CompletionTokens, in.LatencyMs, in.Amount, mode,
            AuditMetadata{Mode: mode, Amount: in.Amount, SubjectKey: fmt.Sprintf("%s:%d", pc.SubjectType, pc.SubjectId), Engine: "billing_engine", Idempotent: false})
        if err := tx.Create(log).Error; err != nil {
            // If duplicate, treat as idempotent and return existing
            if isUniqueConstraintError(err) {
                var existing model.RequestLog
                if err2 := tx.Where("request_id = ?", in.RequestId).First(&existing).Error; err2 == nil {
                    resultLog = &existing
                    return nil
                }
            }
            return err
        }
        // apply the charge based on mode
        switch mode {
        case common.BillingModePlan:
            if pc.Assignment == nil || pc.Plan == nil {
                return fmt.Errorf("plan info missing")
            }
            // Lock counter row for update if exists
            counter, err := model.GetUsageCounterTx(tx, pc.Assignment.Id, pc.Plan.QuotaMetric, pc.CycleStart)
            if err != nil {
                return err
            }
            var consumed int64 = 0
            if counter != nil {
                consumed = counter.ConsumedAmount
            }
            remaining := pc.Plan.QuotaAmount + pc.Assignment.RolloverAmount - consumed
            if remaining < in.Amount {
                // Not enough plan allowance
                // If fallback enabled, switch to balance
                if pc.Assignment.AutoFallbackEnabled || pc.Mode == common.BillingModeFallback {
                    mode = common.BillingModeBalance
                    // Update log metadata to reflect fallback
                    // Delete created log and recreate at end with correct mode inside same tx.
                    if err := tx.Delete(&model.RequestLog{}, "id = ?", log.Id).Error; err != nil {
                        return err
                    }
                    // Recreate idempotency record with same request id and updated mode
                    log = BuildRequestLog(in.RequestId, pc.SubjectType, pc.SubjectId, in.ModelAlias, in.Upstream,
                        pc.Assignment, in.UsageMetric, in.PromptTokens, in.CompletionTokens, in.PromptTokens+in.CompletionTokens, in.LatencyMs, in.Amount, mode,
                        AuditMetadata{Mode: mode, Amount: in.Amount, SubjectKey: fmt.Sprintf("%s:%d", pc.SubjectType, pc.SubjectId), Engine: "billing_engine", Idempotent: false})
                    if err := tx.Create(log).Error; err != nil {
                        if isUniqueConstraintError(err) {
                            var existing model.RequestLog
                            if err2 := tx.Where("request_id = ?", in.RequestId).First(&existing).Error; err2 == nil {
                                resultLog = &existing
                                return nil
                            }
                        }
                        return err
                    }
                    // then go to balance branch below after switch
                } else {
                    return &ErrPlanExhausted{AssignmentId: pc.Assignment.Id, Metric: pc.Plan.QuotaMetric, Remaining: remaining, Needed: in.Amount}
                }
            } else {
                // Increment usage inside tx
                if err := model.IncrementUsageCounterTx(tx, pc.Assignment.Id, pc.Plan.QuotaMetric, in.Amount, pc.CycleStart, pc.CycleEnd); err != nil {
                    return err
                }
            }
        case common.BillingModeBalance:
            // Deduct from user (and token if not unlimited) atomically with conditions
            // Lock user row
            var user model.User
            if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", pc.SubjectId).Error; err != nil {
                return err
            }
            if user.Quota < int(in.Amount) {
                return &ErrBalanceInsufficient{UserId: user.Id, TokenId: in.RelayInfo.TokenId, Remaining: user.Quota, Needed: in.Amount}
            }
            if err := tx.Model(&model.User{}).Where("id = ? AND quota >= ?", user.Id, int(in.Amount)).Updates(map[string]any{
                "quota":      gorm.Expr("quota - ?", int(in.Amount)),
                "used_quota": gorm.Expr("used_quota + ?", int(in.Amount)),
            }).Error; err != nil {
                return err
            }
            // Token deduction if not unlimited and not playground
            if in.RelayInfo != nil && !in.RelayInfo.IsPlayground {
                var token model.Token
                if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&token, "id = ?", in.RelayInfo.TokenId).Error; err != nil {
                    return err
                }
                if !token.UnlimitedQuota {
                    if token.RemainQuota < int(in.Amount) {
                        return &ErrBalanceInsufficient{UserId: user.Id, TokenId: token.Id, Remaining: token.RemainQuota, Needed: in.Amount}
                    }
                    if err := tx.Model(&model.Token{}).Where("id = ? AND remain_quota >= ?", token.Id, int(in.Amount)).Updates(map[string]any{
                        "remain_quota": gorm.Expr("remain_quota - ?", int(in.Amount)),
                        "used_quota":   gorm.Expr("used_quota + ?", int(in.Amount)),
                        "accessed_time": common.GetTimestamp(),
                    }).Error; err != nil {
                        return err
                    }
                }
            }
        default:
            return fmt.Errorf("unsupported billing mode: %s", mode)
        }
        resultLog = log
        return nil
    })
    if err != nil {
        return nil, err
    }
    return resultLog, nil
}

// RollbackCharge is a best-effort reversal based on request id. It is idempotent and safe to call multiple times.
func (be *BillingEngine) RollbackCharge(ctx context.Context, requestId string) error {
    if requestId == "" {
        return fmt.Errorf("request id required")
    }
    return be.DB.Transaction(func(tx *gorm.DB) error {
        var log model.RequestLog
        if err := tx.Where("request_id = ?", requestId).First(&log).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return nil
            }
            return err
        }
        // For now we only delete the log entry to allow retry; reversing quotas precisely requires richer metadata.
        if err := tx.Delete(&model.RequestLog{}, log.Id).Error; err != nil {
            return err
        }
        return nil
    })
}

func parseSubjectKey(subjectKey string, relayInfo *relaycommon.RelayInfo) (string, int, error) {
    if subjectKey != "" {
        parts := strings.Split(subjectKey, ":")
        if len(parts) != 2 {
            return "", 0, fmt.Errorf("invalid subject key")
        }
        typ := parts[0]
        id64, err := strconv.Atoi(parts[1])
        if err != nil {
            return "", 0, err
        }
        return typ, id64, nil
    }
    // Default to user
    if relayInfo != nil && relayInfo.UserId != 0 {
        return common.AssignmentSubjectTypeUser, relayInfo.UserId, nil
    }
    return "", 0, fmt.Errorf("subject key not provided")
}

func getCycleWindow(cycleType string, now time.Time) (time.Time, time.Time) {
    now = now.UTC()
    switch cycleType {
    case common.PlanCycleDaily:
        start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
        return start, start.Add(24 * time.Hour)
    case common.PlanCycleMonthly:
        start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
        return start, start.AddDate(0, 1, 0)
    default:
        // custom or unknown, use monthly as a safe default
        start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
        return start, start.AddDate(0, 1, 0)
    }
}

func isUniqueConstraintError(err error) bool {
    if err == nil {
        return false
    }
    msg := strings.ToLower(err.Error())
    return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") || strings.Contains(msg, "constraint failed")
}
