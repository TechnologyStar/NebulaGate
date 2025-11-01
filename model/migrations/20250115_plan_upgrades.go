package migrations

import (
	"errors"

	"gorm.io/gorm"
)

const PlanUpgradesVersion = "20250115_plan_upgrades"

func init() {
	registerMigration(Migration{
		Version: PlanUpgradesVersion,
		Name:    "Plan upgrades: voucher codes, token limits, model restrictions, validity",
		Up:      planUpgradesUp,
		Down:    planUpgradesDown,
	})
}

func planUpgradesUp(tx *gorm.DB) error {
	tables, ok := schemaTables(PlanUpgradesVersion)
	if !ok {
		return errors.New("schema provider not registered for plan upgrades migration")
	}
	if len(tables) == 0 {
		return nil
	}
	return tx.AutoMigrate(tables...)
}

func planUpgradesDown(tx *gorm.DB) error {
	if err := tx.Migrator().DropTable("voucher_codes"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(&struct{ TableName string `gorm:"-"`; TokenLimit int64; AllowedModels interface{}; ValidityDays int }{TableName: "plans"}, "token_limit"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(&struct{ TableName string `gorm:"-"`; TokenLimit int64; AllowedModels interface{}; ValidityDays int }{TableName: "plans"}, "allowed_models"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(&struct{ TableName string `gorm:"-"`; TokenLimit int64; AllowedModels interface{}; ValidityDays int }{TableName: "plans"}, "validity_days"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(&struct{ TableName string `gorm:"-"`; ExpiresAt interface{}; EnforcementMetadata interface{} }{TableName: "plan_assignments"}, "expires_at"); err != nil {
		return err
	}
	if err := tx.Migrator().DropColumn(&struct{ TableName string `gorm:"-"`; ExpiresAt interface{}; EnforcementMetadata interface{} }{TableName: "plan_assignments"}, "enforcement_metadata"); err != nil {
		return err
	}
	return nil
}
