package model

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RedemptionCode represents a code that can be redeemed for a package
type RedemptionCode struct {
	Id             int            `json:"id" gorm:"primaryKey"`
	PackageId      int            `json:"package_id" gorm:"index;not null"`
	Code           string         `json:"code" gorm:"type:varchar(64);uniqueIndex;not null"`
	Status         int            `json:"status" gorm:"type:int;not null;default:1"` // 1: unused, 2: used, 3: revoked
	UsedByUserId   int            `json:"used_by_user_id" gorm:"index"`
	UsedAt         *time.Time     `json:"used_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	UserPackageId  *int           `json:"user_package_id"`
	Package        Package        `json:"package" gorm:"foreignKey:PackageId"`
}

func (rc *RedemptionCode) TableName() string {
	return "redemption_codes"
}

// BeforeCreate ensures the code is uppercased
func (rc *RedemptionCode) BeforeCreate(tx *gorm.DB) error {
	rc.Code = strings.ToUpper(strings.TrimSpace(rc.Code))
	return nil
}

// GetRedemptionCodeByCode retrieves a redis code by code string
func GetRedemptionCodeByCode(code string) (*RedemptionCode, error) {
	if code == "" {
		return nil, errors.New("invalid code")
	}
	var redemption RedemptionCode
	err := DB.Where("code = ?", strings.ToUpper(code)).First(&redemption).Error
	if err != nil {
		return nil, err
	}
	return &redemption, nil
}

// GetRedemptionCodeById retrieves a redemption code by ID
func GetRedemptionCodeById(id int) (*RedemptionCode, error) {
	if id == 0 {
		return nil, errors.New("invalid redemption code id")
	}
	var redemption RedemptionCode
	err := DB.Where("id = ?", id).First(&redemption).Error
	if err != nil {
		return nil, err
	}
	return &redemption, nil
}

// ListRedemptionCodes lists redemption codes with pagination and filters
func ListRedemptionCodes(startIdx, num int, packageId, status int, code string) ([]*RedemptionCode, int64, error) {
	var codes []*RedemptionCode
	var total int64

	query := DB.Model(&RedemptionCode{})

	if packageId > 0 {
		query = query.Where("package_id = ?", packageId)
	}
	if status > 0 {
		query = query.Where("status = ?", status)
	}
	if code != "" {
		query = query.Where("code LIKE ?", "%"+strings.ToUpper(code)+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("id desc").Limit(num).Offset(startIdx).Preload("Package").Find(&codes).Error
	if err != nil {
		return nil, 0, err
	}

	return codes, total, nil
}

// Insert creates a new redemption code
func (rc *RedemptionCode) Insert() error {
	return DB.Create(rc).Error
}

// Update updates a redemption code
func (rc *RedemptionCode) Update(fields ...string) error {
	if len(fields) == 0 {
		return DB.Save(rc).Error
	}
	return DB.Model(rc).Select(fields).Updates(rc).Error
}

// Delete soft deletes a redemption code
func (rc *RedemptionCode) Delete() error {
	return DB.Delete(rc).Error
}
