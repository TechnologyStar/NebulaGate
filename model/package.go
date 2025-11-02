package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Package represents a token package/plan that can be purchased or redeemed
type Package struct {
	Id           int            `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"type:varchar(128);not null;index"`
	Description  string         `json:"description" gorm:"type:text"`
	TokenQuota   int64          `json:"token_quota" gorm:"type:bigint;not null;default:0"`
	ModelScope   string         `json:"model_scope" gorm:"type:text"` // JSON array of allowed models
	ValidityDays int            `json:"validity_days" gorm:"not null;default:30"`
	Price        float64        `json:"price" gorm:"type:decimal(10,2);default:0"`
	Status       int            `json:"status" gorm:"type:int;not null;default:1"` // 1: enabled, 2: disabled
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (p *Package) TableName() string {
	return "packages"
}

// GetAllowedModels returns the list of allowed models for this package
func (p *Package) GetAllowedModels() []string {
	if p.ModelScope == "" {
		return []string{}
	}
	var models []string
	if err := json.Unmarshal([]byte(p.ModelScope), &models); err != nil {
		return []string{}
	}
	return models
}

// SetAllowedModels sets the list of allowed models for this package
func (p *Package) SetAllowedModels(models []string) error {
	if models == nil {
		p.ModelScope = "[]"
		return nil
	}
	data, err := json.Marshal(models)
	if err != nil {
		return err
	}
	p.ModelScope = string(data)
	return nil
}

// IsModelAllowed checks if a model is allowed in this package
func (p *Package) IsModelAllowed(modelName string) bool {
	models := p.GetAllowedModels()
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

// GetPackageById retrieves a package by its ID
func GetPackageById(id int) (*Package, error) {
	if id == 0 {
		return nil, errors.New("invalid package id")
	}
	var pkg Package
	err := DB.Where("id = ?", id).First(&pkg).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// GetAllPackages retrieves all packages with pagination
func GetAllPackages(startIdx int, num int, status int) ([]*Package, int64, error) {
	var packages []*Package
	var total int64

	query := DB.Model(&Package{})
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&packages).Error
	if err != nil {
		return nil, 0, err
	}

	return packages, total, nil
}

// Insert creates a new package
func (p *Package) Insert() error {
	return DB.Create(p).Error
}

// Update updates a package
func (p *Package) Update() error {
	return DB.Save(p).Error
}

// Delete soft deletes a package
func (p *Package) Delete() error {
	return DB.Delete(p).Error
}

// DeletePackageById deletes a package by its ID
func DeletePackageById(id int) error {
	if id == 0 {
		return errors.New("invalid package id")
	}
	pkg, err := GetPackageById(id)
	if err != nil {
		return err
	}
	return pkg.Delete()
}
