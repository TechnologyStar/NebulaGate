package migrations

import (
	"testing"

	_ "github.com/QuantumNous/new-api/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBillingAndGovernanceMigration(t *testing.T) {
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
