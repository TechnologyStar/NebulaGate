package model

import "github.com/QuantumNous/new-api/model/migrations"

func init() {
	migrations.RegisterSchemaProvider(migrations.PackageSystemVersion, func() []interface{} {
		return []interface{}{
			&Package{},
			&RedemptionCode{},
			&UserPackage{},
		}
	})
}
