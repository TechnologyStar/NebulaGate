package migrations

import (
	"errors"
	"sort"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	Version string
	Name    string
	Up      func(tx *gorm.DB) error
	Down    func(tx *gorm.DB) error
}

type schemaMigration struct {
	Version   string    `gorm:"primaryKey;size:128"`
	Name      string    `gorm:"size:255;not null"`
	AppliedAt time.Time `gorm:"not null"`
}

func (schemaMigration) TableName() string {
	return "schema_migrations"
}

var (
	migrationMu    sync.RWMutex
	migrationIndex = make(map[string]*Migration)
	migrationList  []*Migration
)

func registerMigration(m Migration) {
	migrationMu.Lock()
	defer migrationMu.Unlock()
	migrationCopy := m
	migrationIndex[m.Version] = &migrationCopy
	migrationList = append(migrationList, &migrationCopy)
}

func sortedMigrations() []*Migration {
	migrationMu.RLock()
	defer migrationMu.RUnlock()
	result := make([]*Migration, len(migrationList))
	copy(result, migrationList)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})
	return result
}

func getMigration(version string) (*Migration, bool) {
	migrationMu.RLock()
	defer migrationMu.RUnlock()
	m, ok := migrationIndex[version]
	return m, ok
}

func ensureSchemaTable(db *gorm.DB) error {
	return db.AutoMigrate(&schemaMigration{})
}

func appliedVersions(db *gorm.DB) (map[string]schemaMigration, error) {
	var records []schemaMigration
	if err := db.Find(&records).Error; err != nil {
		return nil, err
	}
	result := make(map[string]schemaMigration, len(records))
	for _, record := range records {
		result[record.Version] = record
	}
	return result, nil
}

func Run(db *gorm.DB) error {
	if db == nil {
		return errors.New("nil database")
	}
	if err := ensureSchemaTable(db); err != nil {
		return err
	}
	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}
	for _, migration := range sortedMigrations() {
		if _, ok := applied[migration.Version]; ok {
			continue
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := migration.Up(tx); err != nil {
				return err
			}
			record := schemaMigration{
				Version:   migration.Version,
				Name:      migration.Name,
				AppliedAt: time.Now().UTC(),
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func Down(db *gorm.DB, version string) error {
	if db == nil {
		return errors.New("nil database")
	}
	if err := ensureSchemaTable(db); err != nil {
		return err
	}
	migration, ok := getMigration(version)
	if !ok {
		return errors.New("migration not registered")
	}
	var record schemaMigration
	if err := db.Where("version = ?", version).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migration.Down(tx); err != nil {
			return err
		}
		if err := tx.Delete(&schemaMigration{}, "version = ?", version).Error; err != nil {
			return err
		}
		return nil
	})
}
