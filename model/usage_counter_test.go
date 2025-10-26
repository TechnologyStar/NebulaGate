package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUsageCounterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
	if err := db.AutoMigrate(&Plan{}, &PlanAssignment{}, &UsageCounter{}); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
	old := DB
	DB = db
	t.Cleanup(func() { DB = old })
	return db
}

func TestUsageCounterIncrement(t *testing.T) {
	db := setupUsageCounterTestDB(t)
	now := time.Now().UTC()
	p := &Plan{Code: "p", Name: "P", CycleType: common.PlanCycleMonthly, QuotaMetric: common.PlanQuotaMetricRequests, QuotaAmount: 100, IsActive: true}
	if err := db.Create(p).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	a := &PlanAssignment{SubjectType: common.AssignmentSubjectTypeUser, SubjectId: 1, PlanId: p.Id, ActivatedAt: now.Add(-time.Hour)}
	if err := db.Create(a).Error; err != nil {
		t.Fatalf("create assignment: %v", err)
	}
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	if err := IncrementUsageCounter(a.Id, p.QuotaMetric, 5, start, end); err != nil {
		t.Fatalf("increment: %v", err)
	}
	if err := IncrementUsageCounter(a.Id, p.QuotaMetric, 7, start, end); err != nil {
		t.Fatalf("increment2: %v", err)
	}
	var c UsageCounter
	if err := db.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", a.Id, p.QuotaMetric, start).First(&c).Error; err != nil {
		t.Fatalf("load: %v", err)
	}
	if c.ConsumedAmount != 12 {
		t.Fatalf("expected 12, got %d", c.ConsumedAmount)
	}
}
