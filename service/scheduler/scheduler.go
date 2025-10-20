package scheduler

import (
    "context"
    "errors"
    "strconv"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
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
// based on the associated plan's cycle. The operation is idempotent.
func RunPlanCycleResetOnce(ctx context.Context) error {
    now := time.Now().UTC()

    return model.DB.Transaction(func(tx *gorm.DB) error {
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

            // If a counter for the new window already exists, delete the expired record
            var existing model.UsageCounter
            err := tx.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", c.PlanAssignmentId, c.Metric, start).First(&existing).Error
            if err == nil {
                if err := tx.Delete(&model.UsageCounter{}, c.Id).Error; err != nil {
                    return err
                }
                continue
            }
            if !errors.Is(err, gorm.ErrRecordNotFound) {
                return err
            }

            // Update in place to keep a single row per assignment+metric
            if err := tx.Model(&model.UsageCounter{}).Where("id = ?", c.Id).Updates(map[string]any{
                "consumed_amount": 0,
                "cycle_start":     start,
                "cycle_end":       end,
                "updated_at":      gorm.Expr("CURRENT_TIMESTAMP"),
            }).Error; err != nil {
                return err
            }

            // Optional audit log
            common.SysLog("plan cycle reset: assignment=" + strconv.Itoa(a.Id) + " metric=" + c.Metric)
        }
        return nil
    })
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

// helper removed; use strconv.Itoa directly
