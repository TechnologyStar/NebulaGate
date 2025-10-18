package migrations

import (
	"errors"

	"gorm.io/gorm"
)

const BillingGovernanceVersion = "20241018_billing_and_governance"

func init() {
	registerMigration(Migration{
		Version: BillingGovernanceVersion,
		Name:    "Billing and governance data layer",
		Up:      billingAndGovernanceUp,
		Down:    billingAndGovernanceDown,
	})
}

func billingAndGovernanceUp(tx *gorm.DB) error {
	tables, ok := schemaTables(BillingGovernanceVersion)
	if !ok {
		return errors.New("schema provider not registered for billing governance migration")
	}
	if len(tables) == 0 {
		return nil
	}
	return tx.AutoMigrate(tables...)
}

func billingAndGovernanceDown(tx *gorm.DB) error {
	tables, ok := schemaTables(BillingGovernanceVersion)
	if !ok {
		return errors.New("schema provider not registered for billing governance migration")
	}
	for i := len(tables) - 1; i >= 0; i-- {
		if err := tx.Migrator().DropTable(tables[i]); err != nil {
			return err
		}
	}
	return nil
}
