package service

import (
    "fmt"
    "sync"
    "testing"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    relaycommon "github.com/QuantumNous/new-api/relay/common"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to open sqlite: %v", err)
    }
    if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
        t.Fatalf("failed to enable foreign keys: %v", err)
    }
    tables := []interface{}{
        &model.User{},
        &model.Token{},
        &model.Plan{},
        &model.PlanAssignment{},
        &model.UsageCounter{},
        &model.RequestLog{},
    }
    if err := db.AutoMigrate(tables...); err != nil {
        t.Fatalf("failed to migrate tables: %v", err)
    }
    oldDB := model.DB
    model.DB = db
    t.Cleanup(func() { model.DB = oldDB })
    common.UsingSQLite = true
    common.UsingPostgreSQL = false
    common.UsingMySQL = false
    return db
}

func newRelayInfo(userId, tokenId int, unlimited bool) *relaycommon.RelayInfo {
    return &relaycommon.RelayInfo{UserId: userId, TokenId: tokenId, TokenUnlimited: unlimited, StartTime: time.Now()}
}

func createUserAndToken(t *testing.T, db *gorm.DB, quota int, unlimited bool) (*model.User, *model.Token) {
    user := &model.User{Username: "u", Password: "password", Quota: quota, Status: common.UserStatusEnabled}
    if err := db.Create(user).Error; err != nil {
        t.Fatalf("create user: %v", err)
    }
    token := &model.Token{UserId: user.Id, Key: "k", Name: "t", RemainQuota: quota, UnlimitedQuota: unlimited, Status: common.TokenStatusEnabled}
    if err := db.Create(token).Error; err != nil {
        t.Fatalf("create token: %v", err)
    }
    return user, token
}

func createPlanAndAssignment(t *testing.T, db *gorm.DB, userId int, quotaAmount int64, billingMode string, fallback bool) (*model.Plan, *model.PlanAssignment) {
    plan := &model.Plan{Code: "basic", Name: "Basic", CycleType: common.PlanCycleMonthly, QuotaMetric: common.PlanQuotaMetricRequests, QuotaAmount: quotaAmount, IsActive: true}
    if err := db.Create(plan).Error; err != nil {
        t.Fatalf("create plan: %v", err)
    }
    assignment := &model.PlanAssignment{SubjectType: common.AssignmentSubjectTypeUser, SubjectId: userId, PlanId: plan.Id, ActivatedAt: time.Now().Add(-time.Hour)}
    if billingMode != "" { assignment.BillingMode = billingMode }
    assignment.AutoFallbackEnabled = fallback
    if err := db.Create(assignment).Error; err != nil {
        t.Fatalf("create assignment: %v", err)
    }
    return plan, assignment
}

func TestBillingEngine_PlanOnly(t *testing.T) {
    db := setupServiceTestDB(t)
    user, _ := createUserAndToken(t, db, 100, true)
    plan, assignment := createPlanAndAssignment(t, db, user.Id, 50, common.BillingModePlan, false)
    _ = plan; _ = assignment

    engine := NewBillingEngine(db)
    pc, err := engine.PrepareCharge(nil, "user:"+fmt.Sprintf("%d", user.Id), newRelayInfo(user.Id, 0, true))
    if err != nil { t.Fatalf("prepare: %v", err) }
    if pc.Mode != common.BillingModePlan { t.Fatalf("expected plan mode, got %s", pc.Mode) }
    if pc.Allowance != 50 { t.Fatalf("expected allowance 50, got %d", pc.Allowance) }

    reqId := "req-plan-1"
    _, err = engine.CommitCharge(nil, &CommitParams{Prepared: pc, Amount: 10, RequestId: reqId, RelayInfo: newRelayInfo(user.Id, 0, true)})
    if err != nil { t.Fatalf("commit: %v", err) }

    // Verify usage counter
    start, _ := getCycleWindow(plan.CycleType, time.Now().UTC())
    var counter model.UsageCounter
    if err := db.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", assignment.Id, plan.QuotaMetric, start).First(&counter).Error; err != nil {
        t.Fatalf("load counter: %v", err)
    }
    if counter.ConsumedAmount != 10 { t.Fatalf("expected consumed 10, got %d", counter.ConsumedAmount) }

    // Idempotent retry
    _, err = engine.CommitCharge(nil, &CommitParams{Prepared: pc, Amount: 10, RequestId: reqId, RelayInfo: newRelayInfo(user.Id, 0, true)})
    if err != nil { t.Fatalf("retry commit: %v", err) }
    var counter2 model.UsageCounter
    if err := db.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", assignment.Id, plan.QuotaMetric, start).First(&counter2).Error; err != nil {
        t.Fatalf("load counter: %v", err)
    }
    if counter2.ConsumedAmount != 10 { t.Fatalf("expected consumed still 10 after retry, got %d", counter2.ConsumedAmount) }
}

func TestBillingEngine_BalanceOnly(t *testing.T) {
    db := setupServiceTestDB(t)
    user, token := createUserAndToken(t, db, 100, true)
    engine := NewBillingEngine(db)
    pc, err := engine.PrepareCharge(nil, "user:"+fmt.Sprintf("%d", user.Id), newRelayInfo(user.Id, token.Id, true))
    if err != nil { t.Fatalf("prepare: %v", err) }
    if pc.Mode != common.BillingModeBalance { t.Fatalf("expected balance mode, got %s", pc.Mode) }
    reqId := "req-balance-1"
    _, err = engine.CommitCharge(nil, &CommitParams{Prepared: pc, Amount: 20, RequestId: reqId, RelayInfo: newRelayInfo(user.Id, token.Id, true)})
    if err != nil { t.Fatalf("commit: %v", err) }
    var u model.User
    if err := db.First(&u, user.Id).Error; err != nil { t.Fatalf("load user: %v", err) }
    if u.Quota != 80 { t.Fatalf("expected user quota 80, got %d", u.Quota) }
    // Retry idempotent
    _, err = engine.CommitCharge(nil, &CommitParams{Prepared: pc, Amount: 20, RequestId: reqId, RelayInfo: newRelayInfo(user.Id, token.Id, true)})
    if err != nil { t.Fatalf("retry commit: %v", err) }
    var u2 model.User
    if err := db.First(&u2, user.Id).Error; err != nil { t.Fatalf("load user: %v", err) }
    if u2.Quota != 80 { t.Fatalf("expected user quota remains 80, got %d", u2.Quota) }
}

func TestBillingEngine_AutoFallback(t *testing.T) {
    db := setupServiceTestDB(t)
    user, token := createUserAndToken(t, db, 100, true)
    plan, assignment := createPlanAndAssignment(t, db, user.Id, 5, common.BillingModePlan, true)
    _ = plan; _ = assignment
    engine := NewBillingEngine(db)
    pc, err := engine.PrepareCharge(nil, "user:"+fmt.Sprintf("%d", user.Id), newRelayInfo(user.Id, token.Id, true))
    if err != nil { t.Fatalf("prepare: %v", err) }
    if pc.Mode != common.BillingModePlan { t.Fatalf("expected plan mode at prepare, got %s", pc.Mode) }
    reqId := "req-fallback-1"
    _, err = engine.CommitCharge(nil, &CommitParams{Prepared: pc, Amount: 10, RequestId: reqId, RelayInfo: newRelayInfo(user.Id, token.Id, true)})
    if err != nil { t.Fatalf("commit: %v", err) }
    // Should fallback to balance, not increment usage counter
    start, _ := getCycleWindow(plan.CycleType, time.Now().UTC())
    var cnt int64
    db.Model(&model.UsageCounter{}).Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", assignment.Id, plan.QuotaMetric, start).Select("ifnull(sum(consumed_amount),0)").Scan(&cnt)
    if cnt != 0 { t.Fatalf("expected plan usage 0 due to fallback, got %d", cnt) }
    var u model.User
    if err := db.First(&u, user.Id).Error; err != nil { t.Fatalf("load user: %v", err) }
    if u.Quota != 90 { t.Fatalf("expected user quota 90 after fallback, got %d", u.Quota) }
}

func TestBillingEngine_ConcurrencyGuard(t *testing.T) {
    db := setupServiceTestDB(t)
    user, token := createUserAndToken(t, db, 100, true)
    engine := NewBillingEngine(db)
    pc, err := engine.PrepareCharge(nil, "user:"+fmt.Sprintf("%d", user.Id), newRelayInfo(user.Id, token.Id, true))
    if err != nil { t.Fatalf("prepare: %v", err) }
    if pc.Mode != common.BillingModeBalance { t.Fatalf("expected balance mode, got %s", pc.Mode) }

    reqId := "req-concurrent-1"
    wg := sync.WaitGroup{}
    wg.Add(2)
    go func() {
        defer wg.Done()
        _, _ = engine.CommitCharge(nil, &CommitParams{Prepared: pc, Amount: 30, RequestId: reqId, RelayInfo: newRelayInfo(user.Id, token.Id, true)})
    }()
    go func() {
        defer wg.Done()
        _, _ = engine.CommitCharge(nil, &CommitParams{Prepared: pc, Amount: 30, RequestId: reqId, RelayInfo: newRelayInfo(user.Id, token.Id, true)})
    }()
    wg.Wait()
    var u model.User
    if err := db.First(&u, user.Id).Error; err != nil { t.Fatalf("load user: %v", err) }
    if u.Quota != 70 { t.Fatalf("expected single deduction to 70, got %d", u.Quota) }
}
