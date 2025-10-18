package model

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBillingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}
	tables := []interface{}{
		&Plan{},
		&PlanAssignment{},
		&UsageCounter{},
		&VoucherBatch{},
		&VoucherRedemption{},
		&RequestFlag{},
		&RequestLog{},
		&RequestAggregate{},
	}
	if err := db.AutoMigrate(tables...); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
	oldDB := DB
	DB = db
	t.Cleanup(func() {
		DB = oldDB
	})
	common.UsingSQLite = true
	common.UsingPostgreSQL = false
	common.UsingMySQL = false
	return db
}

func TestGetActivePlanAssignments(t *testing.T) {
	db := setupBillingTestDB(t)
	now := time.Now().UTC()
	plan := Plan{
		Code:        "basic",
		Name:        "Basic",
		CycleType:   common.PlanCycleMonthly,
		QuotaMetric: common.PlanQuotaMetricRequests,
		QuotaAmount: 1000,
		IsActive:    true,
	}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	active := PlanAssignment{
		SubjectType: common.AssignmentSubjectTypeUser,
		SubjectId:   100,
		PlanId:      plan.Id,
		ActivatedAt: now.Add(-time.Hour),
	}
	expiredAt := now.Add(-5 * time.Minute)
	expired := PlanAssignment{
		SubjectType:   common.AssignmentSubjectTypeUser,
		SubjectId:     100,
		PlanId:        plan.Id,
		ActivatedAt:   now.Add(-2 * time.Hour),
		DeactivatedAt: &expiredAt,
	}
	future := PlanAssignment{
		SubjectType: common.AssignmentSubjectTypeUser,
		SubjectId:   100,
		PlanId:      plan.Id,
		ActivatedAt: now.Add(time.Hour),
	}
	otherSubject := PlanAssignment{
		SubjectType: common.AssignmentSubjectTypeToken,
		SubjectId:   200,
		PlanId:      plan.Id,
		ActivatedAt: now.Add(-time.Hour),
	}
	if err := db.Create(&[]PlanAssignment{active, expired, future, otherSubject}).Error; err != nil {
		t.Fatalf("failed to create assignments: %v", err)
	}
	assignments, err := GetActivePlanAssignments(common.AssignmentSubjectTypeUser, 100, now)
	if err != nil {
		t.Fatalf("failed to fetch assignments: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("expected 1 assignment, got %d", len(assignments))
	}
	if assignments[0].PlanId != plan.Id {
		t.Fatalf("unexpected plan assignment returned")
	}
}

func TestIncrementUsageCounter(t *testing.T) {
	db := setupBillingTestDB(t)
	now := time.Now().UTC()
	plan := Plan{
		Code:        "metered",
		Name:        "Metered",
		CycleType:   common.PlanCycleMonthly,
		QuotaMetric: common.PlanQuotaMetricRequests,
		QuotaAmount: 2000,
		IsActive:    true,
	}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}
	assignment := PlanAssignment{
		SubjectType: common.AssignmentSubjectTypeUser,
		SubjectId:   300,
		PlanId:      plan.Id,
		ActivatedAt: now.Add(-time.Hour),
	}
	if err := db.Create(&assignment).Error; err != nil {
		t.Fatalf("failed to create assignment: %v", err)
	}
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	if err := IncrementUsageCounter(assignment.Id, common.PlanQuotaMetricRequests, 10, start, end); err != nil {
		t.Fatalf("failed to increment counter: %v", err)
	}
	if err := IncrementUsageCounter(assignment.Id, common.PlanQuotaMetricRequests, 5, start, end); err != nil {
		t.Fatalf("failed to increment counter: %v", err)
	}
	var counter UsageCounter
	if err := db.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", assignment.Id, common.PlanQuotaMetricRequests, start).First(&counter).Error; err != nil {
		t.Fatalf("failed to load counter: %v", err)
	}
	if counter.ConsumedAmount != 15 {
		t.Fatalf("expected consumed amount 15, got %d", counter.ConsumedAmount)
	}
	if !counter.CycleEnd.Equal(end) {
		t.Fatalf("expected cycle end %v, got %v", end, counter.CycleEnd)
	}
}

func TestInsertRequestAggregate(t *testing.T) {
	setupBillingTestDB(t)
	start := time.Now().UTC().Truncate(time.Hour)
	end := start.Add(time.Hour)
	aggregate := &RequestAggregate{
		ModelAlias:     "gpt-4o",
		Upstream:       "openai",
		SubjectType:    common.AssignmentSubjectTypeUser,
		WindowStart:    start,
		WindowEnd:      end,
		TotalRequests:  2,
		TotalTokens:    500,
		UniqueSubjects: 1,
	}
	if err := InsertRequestAggregate(aggregate); err != nil {
		t.Fatalf("failed to insert aggregate: %v", err)
	}
	aggregateIncrement := &RequestAggregate{
		ModelAlias:     "gpt-4o",
		Upstream:       "openai",
		SubjectType:    common.AssignmentSubjectTypeUser,
		WindowStart:    start,
		WindowEnd:      end,
		TotalRequests:  3,
		TotalTokens:    700,
		UniqueSubjects: 4,
	}
	if err := InsertRequestAggregate(aggregateIncrement); err != nil {
		t.Fatalf("failed to merge aggregate: %v", err)
	}
	var stored RequestAggregate
	if err := DB.Where("model_alias = ? AND upstream = ? AND subject_type = ? AND window_start = ?", "gpt-4o", "openai", common.AssignmentSubjectTypeUser, start).First(&stored).Error; err != nil {
		t.Fatalf("failed to fetch stored aggregate: %v", err)
	}
	if stored.TotalRequests != 5 {
		t.Fatalf("expected total requests 5, got %d", stored.TotalRequests)
	}
	if stored.TotalTokens != 1200 {
		t.Fatalf("expected total tokens 1200, got %d", stored.TotalTokens)
	}
	if stored.UniqueSubjects != 4 {
		t.Fatalf("expected unique subjects 4, got %d", stored.UniqueSubjects)
	}
}
