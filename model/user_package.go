package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// UserPackage represents a user's redeemed package with remaining quota
type UserPackage struct {
	Id            int            `json:"id" gorm:"primaryKey"`
	UserId        int            `json:"user_id" gorm:"index;not null"`
	PackageId     int            `json:"package_id" gorm:"index;not null"`
	TokenQuota    int64          `json:"token_quota" gorm:"type:bigint;not null"` // Remaining quota
	ModelScope    string         `json:"model_scope" gorm:"type:text"`            // JSON array of allowed models
	ExpireAt      time.Time      `json:"expire_at"`
	Status        int            `json:"status" gorm:"type:int;not null;default:1"` // 1: active, 2: exhausted, 3: expired
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	Package       Package        `json:"package" gorm:"foreignKey:PackageId"`
	InitialQuota  int64          `json:"initial_quota" gorm:"type:bigint;not null"` // Original quota for reference
}

func (up *UserPackage) TableName() string {
	return "user_packages"
}

// GetAllowedModels returns the list of allowed models for this user package
func (up *UserPackage) GetAllowedModels() []string {
	if up.ModelScope == "" {
		return []string{}
	}
	var models []string
	if err := json.Unmarshal([]byte(up.ModelScope), &models); err != nil {
		return []string{}
	}
	return models
}

// SetAllowedModels sets the list of allowed models for this user package
func (up *UserPackage) SetAllowedModels(models []string) error {
	if models == nil {
		up.ModelScope = "[]"
		return nil
	}
	data, err := json.Marshal(models)
	if err != nil {
		return err
	}
	up.ModelScope = string(data)
	return nil
}

// IsModelAllowed checks if a model is allowed in this user package
func (up *UserPackage) IsModelAllowed(modelName string) bool {
	models := up.GetAllowedModels()
	if len(models) == 0 {
		// Empty list means all models are allowed
		return true
	}
	for _, m := range models {
		if m == modelName {
			return true
		}
	}
	return false
}

// IsExpired checks if the package has expired
func (up *UserPackage) IsExpired() bool {
	return time.Now().After(up.ExpireAt)
}

// IsExhausted checks if the package quota is exhausted
func (up *UserPackage) IsExhausted() bool {
	return up.TokenQuota <= 0
}

// GetUserPackageById retrieves a user package by ID
func GetUserPackageById(id int) (*UserPackage, error) {
	if id == 0 {
		return nil, errors.New("invalid user package id")
	}
	var userPackage UserPackage
	err := DB.Where("id = ?", id).Preload("Package").First(&userPackage).Error
	if err != nil {
		return nil, err
	}
	return &userPackage, nil
}

// GetUserPackagesByUserId retrieves all packages for a user
func GetUserPackagesByUserId(userId int) ([]*UserPackage, error) {
	if userId == 0 {
		return nil, errors.New("invalid user id")
	}
	var packages []*UserPackage
	err := DB.Where("user_id = ?", userId).Preload("Package").Order("expire_at asc").Find(&packages).Error
	if err != nil {
		return nil, err
	}
	return packages, nil
}

// GetActiveUserPackages retrieves active (not expired, not exhausted) packages for a user
func GetActiveUserPackages(userId int) ([]*UserPackage, error) {
	if userId == 0 {
		return nil, errors.New("invalid user id")
	}
	var packages []*UserPackage
	now := time.Now()
	err := DB.Where("user_id = ? AND status = ? AND token_quota > ? AND expire_at > ?",
		userId, 1, 0, now).Preload("Package").Order("expire_at asc").Find(&packages).Error
	if err != nil {
		return nil, err
	}
	return packages, nil
}

// Insert creates a new user package
func (up *UserPackage) Insert() error {
	return DB.Create(up).Error
}

// Update updates a user package
func (up *UserPackage) Update(fields ...string) error {
	if len(fields) == 0 {
		return DB.Save(up).Error
	}
	return DB.Model(up).Select(fields).Updates(up).Error
}

// ConsumeQuota consumes quota from the user package
func (up *UserPackage) ConsumeQuota(amount int64) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}
	if up.TokenQuota < amount {
		return errors.New("insufficient quota")
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		// Lock the row for update
		var pkg UserPackage
		err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id = ?", up.Id).First(&pkg).Error
		if err != nil {
			return err
		}

		if pkg.TokenQuota < amount {
			return errors.New("insufficient quota")
		}

		pkg.TokenQuota -= amount
		if pkg.TokenQuota <= 0 {
			pkg.Status = 2 // exhausted
		}

		err = tx.Save(&pkg).Error
		if err != nil {
			return err
		}

		up.TokenQuota = pkg.TokenQuota
		up.Status = pkg.Status
		return nil
	})
}

// UpdateExpiredPackages updates the status of expired packages
func UpdateExpiredPackages() (int64, error) {
	now := time.Now()
	result := DB.Model(&UserPackage{}).
		Where("status = ? AND expire_at <= ?", 1, now).
		Update("status", 3)
	return result.RowsAffected, result.Error
}
