package scheduler

import (
    "context"
    "encoding/hex"
    "errors"
    "fmt"
    "strconv"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/dto"
    "github.com/QuantumNous/new-api/model"
    "github.com/QuantumNous/new-api/service"
    cfg "github.com/QuantumNous/new-api/setting/config"

    "gorm.io/gorm"
)

// Start begins background scheduler jobs according to feature flags and configuration.
// It returns a cancel function to stop all jobs.
func Start() context.CancelFunc {
    ctx, cancel := context.WithCancel(context.Background())

    // Plan cycle reset job (billing)
    if common.BillingFeatureEnabled {
        // Run every hour to catch up missed windows; the job itself is idempotent.
        go startTicker(ctx, time.Hour, func() { _ = RunPlanCycleResetOnce(context.Background()) })
    }

    // TTL cleanup job (governance + public logs)
    if common.GovernanceFeatureEnabled || common.PublicLogsFeatureEnabled {
        // Run every hour
        go startTicker(ctx, time.Hour, func() { _ = RunTTLCleanupOnce(context.Background()) })
    }

    // Anomaly detection job
    // Run every hour to analyze user behavior patterns
    go startTicker(ctx, time.Hour, func() { _ = RunAnomalyDetectionOnce() })

    return cancel
}

func startTicker(ctx context.Context, interval time.Duration, fn func()) {
    // Fire once on start to ensure progress after boot
    fn()
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            fn()
        }
    }
}

// RunPlanCycleResetOnce executes one sweep that resets expired usage counters
// based on the associated plan's cycle. Also handles plan expiry via expires_at.
// The operation is idempotent.
func RunPlanCycleResetOnce(ctx context.Context) error {
    now := time.Now().UTC()

    type pendingNotify struct {
        userId  int
        email   string
        setting dto.UserSetting
        data    dto.Notify
    }
    var notifications []pendingNotify

    err := model.DB.Transaction(func(tx *gorm.DB) error {
        // 1. Handle expired plan assignments
        var expiredAssignments []model.PlanAssignment
        if err := tx.Where("expires_at IS NOT NULL AND expires_at <= ? AND (deactivated_at IS NULL OR deactivated_at > ?)", now, now).Find(&expiredAssignments).Error; err != nil {
            return err
        }
        for _, a := range expiredAssignments {
            if err := tx.Model(&model.PlanAssignment{}).Where("id = ?", a.Id).Updates(map[string]any{
                "deactivated_at":    now,
                "rollover_amount":   0,
                "rollover_policy":   common.RolloverPolicyNone,
                "rollover_expires_at": gorm.Expr("NULL"),
                "updated_at":        gorm.Expr("CURRENT_TIMESTAMP"),
            }).Error; err != nil {
                return err
            }
            common.SysLog(fmt.Sprintf("plan assignment expired: id=%d, subject=%s:%d", a.Id, a.SubjectType, a.SubjectId))
            // Zero out any active usage counters
            if err := tx.Model(&model.UsageCounter{}).Where("plan_assignment_id = ?", a.Id).Updates(map[string]any{
                "consumed_amount": 0,
                "updated_at":      gorm.Expr("CURRENT_TIMESTAMP"),
            }).Error; err != nil {
                return err
            }
        }

        // 2. Handle cycle-based counter reset
        var expired []model.UsageCounter
        if err := tx.Where("cycle_end <= ?", now).Find(&expired).Error; err != nil {
            return err
        }
        for _, c := range expired {
            // Load assignment and plan to decide cycle type
            var a model.PlanAssignment
            if err := tx.First(&a, "id = ?", c.PlanAssignmentId).Error; err != nil {
                // Assignment missing; remove stale counter
                if errors.Is(err, gorm.ErrRecordNotFound) {
                    if err := tx.Delete(&model.UsageCounter{}, c.Id).Error; err != nil {
                        return err
                    }
                    continue
                }
                return err
            }
            var p model.Plan
            if err := tx.First(&p, "id = ?", a.PlanId).Error; err != nil {
                if errors.Is(err, gorm.ErrRecordNotFound) {
                    // Plan missing; delete stale counter
                    if err := tx.Delete(&model.UsageCounter{}, c.Id).Error; err != nil {
                        return err
                    }
                    continue
                }
                return err
            }
            start, end := getCycleWindow(p.CycleType, now)

            // If a counter for the new window already exists, delete the expired record and continue
            var existing model.UsageCounter
            qerr := tx.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", c.PlanAssignmentId, c.Metric, start).First(&existing).Error
            if qerr == nil {
                if err := tx.Delete(&model.UsageCounter{}, c.Id).Error; err != nil {
                    return err
                }
                continue
            }
            if !errors.Is(qerr, gorm.ErrRecordNotFound) {
                return qerr
            }

            // Compute carry-over according to plan settings
            leftover := p.QuotaAmount - c.ConsumedAmount
            if leftover < 0 {
                leftover = 0
            }
            carried := int64(0)
            eventType := "reset"
            if p.AllowCarryOver {
                carried = leftover
                if p.CarryLimitPercent > 0 {
                    capAmt := (p.QuotaAmount * int64(p.CarryLimitPercent)) / 100
                    if carried > capAmt {
                        carried = capAmt
                    }
                }
                if carried > 0 {
                    eventType = "carry_over"
                }
                policy := common.RolloverPolicyCarryAll
                if p.CarryLimitPercent > 0 {
                    policy = common.RolloverPolicyCap
                }
                // Set rollover on assignment for the upcoming window
                if err := tx.Model(&model.PlanAssignment{}).Where("id = ?", a.Id).Updates(map[string]any{
                    "rollover_policy":      policy,
                    "rollover_amount":      carried,
                    "rollover_expires_at":  end,
                    "updated_at":           gorm.Expr("CURRENT_TIMESTAMP"),
                }).Error; err != nil {
                    return err
                }
            } else {
                // No carry-over: clear rollover fields
                if err := tx.Model(&model.PlanAssignment{}).Where("id = ?", a.Id).Updates(map[string]any{
                    "rollover_policy":     common.RolloverPolicyNone,
                    "rollover_amount":     0,
                    "rollover_expires_at": gorm.Expr("NULL"),
                    "updated_at":          gorm.Expr("CURRENT_TIMESTAMP"),
                }).Error; err != nil {
                    return err
                }
            }

            // Move counter to the new cycle window and reset consumed
            if err := tx.Model(&model.UsageCounter{}).Where("id = ?", c.Id).Updates(map[string]any{
                "consumed_amount": 0,
                "cycle_start":     start,
                "cycle_end":       end,
                "updated_at":      gorm.Expr("CURRENT_TIMESTAMP"),
            }).Error; err != nil {
                return err
            }

            // Record an audit RequestLog entry for reset/carry_over (idempotent)
            subjHash := anonymizeSubject(a.SubjectType, a.SubjectId)
            metaPayload := map[string]any{
                "event":          eventType,
                "carried_amount": carried,
                "leftover":       leftover,
                "prev_consumed":  c.ConsumedAmount,
                "plan_quota":     p.QuotaAmount,
                "engine":         "scheduler",
            }
            b, _ := common.Marshal(metaPayload)
            reqId := fmt.Sprintf("%s:%d:%s:%d", eventType, a.Id, c.Metric, start.Unix())
            rl := &model.RequestLog{
                RequestId:             reqId,
                OccurredAt:            now,
                ModelAlias:            "",
                UpstreamProvider:      "",
                SubjectType:           a.SubjectType,
                AnonymizedSubjectHash: subjHash,
                PlanId:                &a.PlanId,
                PlanAssignmentId:      &a.Id,
                UsageMetric:           c.Metric,
                PromptTokens:          0,
                CompletionTokens:      0,
                TotalTokens:           0,
                LatencyMs:             0,
                Metadata:              model.JSONValue(b),
            }
            if err := tx.Create(rl).Error; err != nil {
                if !isUniqueConstraintError(err) {
                    return err
                }
            }

            common.SysLog("plan cycle reset: assignment=" + strconv.Itoa(a.Id) + " metric=" + c.Metric)

            // Queue an optional user notification if enabled
            if bc := cfg.GetBillingConfig(); bc != nil && bc.Enabled && bc.NotifyOnReset && a.SubjectType == common.AssignmentSubjectTypeUser {
                var user model.User
                if err := tx.First(&user, "id = ?", a.SubjectId).Error; err == nil {
                    title := "Plan cycle reset"
                    if p.CycleType == common.PlanCycleDaily {
                        title = "Daily quota reset"
                    }
                    content := "Your plan cycle has been reset."
                    if eventType == "carry_over" {
                        content = fmt.Sprintf("Your plan has reset. Carried over %d to the new cycle. New allowance base: %d.", carried, p.QuotaAmount)
                    }
                    notifications = append(notifications, pendingNotify{
                        userId:  user.Id,
                        email:   user.Email,
                        setting: user.GetSetting(),
                        data:    dto.NewNotify("plan_reset", title, content, nil),
                    })
                }
            }
        }
        return nil
    })
    if err != nil {
        return err
    }

    // Send notifications outside the transaction
    for _, n := range notifications {
        _ = service.NotifyUser(n.userId, n.email, n.setting, n.data)
    }
    return nil
}

// RunTTLCleanupOnce performs a TTL cleanup sweep for governance flags, public logs and vouchers.
func RunTTLCleanupOnce(ctx context.Context) error {
    now := time.Now().UTC()
    // Governance: request flags cleanup
    if common.GovernanceFeatureEnabled {
        ttlHours := 24
        if g := cfg.GetGovernanceConfig(); g != nil {
            // Optional new config field; default to 24 if zero
            if g.FlagTTLHours > 0 {
                ttlHours = g.FlagTTLHours
            }
        }
        cutoff := now.Add(-time.Duration(ttlHours) * time.Hour)
        // TTLAt wins; otherwise rely on CreatedAt cutoff
        if err := model.DB.Where("ttl_at IS NOT NULL AND ttl_at <= ?", now).Delete(&model.RequestFlag{}).Error; err != nil {
            return err
        }
        if err := model.DB.Where("ttl_at IS NULL AND created_at < ?", cutoff).Delete(&model.RequestFlag{}).Error; err != nil {
            return err
        }
    }

    // Public logs retention cleanup
    if common.PublicLogsFeatureEnabled {
        days := common.PublicLogsRetentionDays
        if plc := cfg.GetPublicLogsConfig(); plc != nil && plc.RetentionDays > 0 {
            days = plc.RetentionDays
        }
        if days > 0 {
            cutoff := now.AddDate(0, 0, -days)
            if err := model.DB.Where("occurred_at < ?", cutoff).Delete(&model.RequestLog{}).Error; err != nil {
                return err
            }
        }
    }

    // Expired vouchers cleanup (soft delete batches past valid_until)
    if err := model.DB.Where("valid_until IS NOT NULL AND valid_until <= ?", now).Delete(&model.VoucherBatch{}).Error; err != nil {
        return err
    }

    return nil
}

// getCycleWindow is a local copy of the helper from BillingEngine to avoid import cycles
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

func anonymizeSubject(subjectType string, subjectId int) string {
    plain := []byte(subjectType + ":" + fmt.Sprintf("%d", subjectId))
    sum := common.HmacSha256Raw(plain, []byte(common.SessionSecret))
    return hex.EncodeToString(sum)
}

func isUniqueConstraintError(err error) bool {
    if err == nil {
        return false
    }
    msg := strings.ToLower(err.Error())
    return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate") || strings.Contains(msg, "constraint failed")
}

// RunAnomalyDetectionOnce runs anomaly detection analysis for all active users
func RunAnomalyDetectionOnce() error {
    common.SysLog("Starting scheduled anomaly detection analysis")
    detector := service.NewAnomalyDetectorService()
    detector.RunPeriodicAnalysis()
    common.SysLog("Completed scheduled anomaly detection analysis")
    return nil
}
