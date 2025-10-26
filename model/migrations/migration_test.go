package migrations

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type (
	testPlan              struct{}
	testPlanAssignment    struct{}
	testUsageCounter      struct{}
	testVoucherBatch      struct{}
	testVoucherRedemption struct{}
	testRequestFlag       struct{}
	testRequestLog        struct{}
	testRequestAggregate  struct{}
)

func (testPlan) TableName() string              { return "plans" }
func (testPlanAssignment) TableName() string    { return "plan_assignments" }
func (testUsageCounter) TableName() string      { return "usage_counters" }
func (testVoucherBatch) TableName() string      { return "voucher_batches" }
func (testVoucherRedemption) TableName() string { return "voucher_redemptions" }
func (testRequestFlag) TableName() string       { return "request_flags" }
func (testRequestLog) TableName() string        { return "request_logs" }
func (testRequestAggregate) TableName() string  { return "request_aggregates" }

func TestBillingAndGovernanceMigration(t *testing.T) {
	RegisterSchemaProvider(BillingGovernanceVersion, func() []interface{} {
		return []interface{}{
			&testPlan{},
			&testPlanAssignment{},
			&testUsageCounter{},
			&testVoucherBatch{},
			&testVoucherRedemption{},
			&testRequestFlag{},
			&testRequestLog{},
			&testRequestAggregate{},
		}
	})
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := Run(db); err != nil {
		t.Fatalf("failed to run migration: %v", err)
	}
	requiredTables := []string{
		"plans",
		"plan_assignments",
		"usage_counters",
		"voucher_batches",
		"voucher_redemptions",
		"request_flags",
		"request_logs",
		"request_aggregates",
	}
	for _, table := range requiredTables {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("expected table %s to exist", table)
		}
	}
	if err := Down(db, BillingGovernanceVersion); err != nil {
		t.Fatalf("failed to rollback migration: %v", err)
	}
	for _, table := range requiredTables {
		if db.Migrator().HasTable(table) {
			t.Fatalf("expected table %s to be dropped", table)
		}
	}
}
