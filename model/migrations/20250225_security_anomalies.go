package migrations

import (
	"errors"

	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

const SecurityAnomaliesVersion = "20250225_security_anomalies"

func init() {
	registerMigration(Migration{
		Version: SecurityAnomaliesVersion,
		Name:    "Security anomalies and baselines for behavioral analysis",
		Up:      securityAnomaliesUp,
		Down:    securityAnomaliesDown,
	})
	RegisterSchemaProvider(SecurityAnomaliesVersion, securityAnomaliesSchema)
}

func securityAnomaliesSchema() []interface{} {
	return []interface{}{
		&model.SecurityAnomaly{},
		&model.AnomalyBaseline{},
	}
}

func securityAnomaliesUp(tx *gorm.DB) error {
	tables, ok := schemaTables(SecurityAnomaliesVersion)
	if !ok {
		return errors.New("schema provider not registered for security anomalies migration")
	}
	if len(tables) == 0 {
		return nil
	}
	return tx.AutoMigrate(tables...)
}

func securityAnomaliesDown(tx *gorm.DB) error {
	tables, ok := schemaTables(SecurityAnomaliesVersion)
	if !ok {
		return errors.New("schema provider not registered for security anomalies migration")
	}
	for i := len(tables) - 1; i >= 0; i-- {
		if err := tx.Migrator().DropTable(tables[i]); err != nil {
			return err
		}
	}
	return nil
}
