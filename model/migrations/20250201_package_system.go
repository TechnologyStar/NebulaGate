package migrations

import (
	"errors"

	"gorm.io/gorm"
)

const PackageSystemVersion = "20250201_package_system"

func init() {
	registerMigration(Migration{
		Version: PackageSystemVersion,
		Name:    "Package and redemption code system",
		Up:      packageSystemUp,
		Down:    packageSystemDown,
	})
}

func packageSystemUp(tx *gorm.DB) error {
	tables, ok := schemaTables(PackageSystemVersion)
	if !ok {
		return errors.New("schema provider not registered for package system migration")
	}
	if len(tables) == 0 {
		return nil
	}
	return tx.AutoMigrate(tables...)
}

func packageSystemDown(tx *gorm.DB) error {
	tables, ok := schemaTables(PackageSystemVersion)
	if !ok {
		return errors.New("schema provider not registered for package system migration")
	}
	for i := len(tables) - 1; i >= 0; i-- {
		if err := tx.Migrator().DropTable(tables[i]); err != nil {
			return err
		}
	}
	return nil
}
