package scheduler

import (
    "testing"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupSchedulerTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite: %v", err)
    }
    if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
        t.Fatalf("failed to enable foreign keys: %v", err)
    }
    // Minimal set of tables
    if err := db.AutoMigrate(&model.Plan{}, &model.PlanAssignment{}, &model.UsageCounter{}, &model.RequestFlag{}, &model.RequestLog{}, &model.VoucherBatch{}); err != nil {
        t.Fatalf("failed to migrate tables: %v", err)
    }
    old := model.DB
    model.DB = db
    t.Cleanup(func() { model.DB = old })
    return db
}

func TestPlanResetJob(t *testing.T) {
    db := setupSchedulerTestDB(t)
    now := time.Now().UTC()

    // Create plan with daily cycle
    p := &model.Plan{Code: "daily", Name: "Daily", CycleType: common.PlanCycleDaily, QuotaMetric: common.PlanQuotaMetricRequests, QuotaAmount: 100, IsActive: true}
    if err := db.Create(p).Error; err != nil { t.Fatalf("create plan: %v", err) }
    a := &model.PlanAssignment{SubjectType: common.AssignmentSubjectTypeUser, SubjectId: 1, PlanId: p.Id, ActivatedAt: now.Add(-48 * time.Hour)}
    if err := db.Create(a).Error; err != nil { t.Fatalf("create assignment: %v", err) }

    // Insert an expired counter for previous day window
    prevStart := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)
    prevEnd := prevStart.Add(24 * time.Hour)
    c := &model.UsageCounter{PlanAssignmentId: a.Id, Metric: p.QuotaMetric, CycleStart: prevStart, CycleEnd: prevEnd, ConsumedAmount: 42}
    if err := db.Create(c).Error; err != nil { t.Fatalf("create counter: %v", err) }

    // Run reset sweep
    if err := RunPlanCycleResetOnce(nil); err != nil { t.Fatalf("reset once: %v", err) }

    // Expect counter moved to current window with consumed 0
    currStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
    var after model.UsageCounter
    if err := db.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", a.Id, p.QuotaMetric, currStart).First(&after).Error; err != nil {
        t.Fatalf("load after: %v", err)
    }
    if after.ConsumedAmount != 0 {
        t.Fatalf("expected consumed 0 after reset, got %d", after.ConsumedAmount)
    }
    if !after.CycleStart.Equal(currStart) || !after.CycleEnd.Equal(currStart.Add(24*time.Hour)) {
        t.Fatalf("unexpected cycle anchors after reset")
    }
}

func TestTTLCleanupFlags(t *testing.T) {
    db := setupSchedulerTestDB(t)
    _ = db
    common.GovernanceFeatureEnabled = true

    // Insert a flag older than 25 hours with no ttl_at
    old := time.Now().UTC().Add(-25 * time.Hour)
    f := &model.RequestFlag{RequestId: "r1", SubjectType: common.AssignmentSubjectTypeUser, SubjectId: 1, Reason: common.FlagReasonViolation, CreatedAt: old}
    if err := model.DB.Create(f).Error; err != nil { t.Fatalf("create flag: %v", err) }

    if err := RunTTLCleanupOnce(nil); err != nil { t.Fatalf("cleanup once: %v", err) }

    var cnt int64
    if err := model.DB.Model(&model.RequestFlag{}).Where("request_id = ?", "r1").Count(&cnt).Error; err != nil {
        t.Fatalf("count flags: %v", err)
    }
    if cnt != 0 {
        t.Fatalf("expected old flag to be deleted, still %d left", cnt)
    }
}
