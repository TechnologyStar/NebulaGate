package service

import (
	"errors"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"gorm.io/gorm"
)

// RedeemPackageCode is now in redemption_code service (for compatibility)

// ConsumePackageQuota consumes quota from available user packages prioritizing soon-to-expire packages
func ConsumePackageQuota(userId int, modelName string, quota int64, specificPackageId *int) ([]*model.UserPackage, error) {
	if userId <= 0 {
		return nil, errors.New("invalid user id")
	}
	if quota <= 0 {
		return nil, errors.New("invalid quota amount")
	}

	var consumedPackages []*model.UserPackage
	error := model.DB.Transaction(func(tx *gorm.DB) error {
		query := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("user_id = ? AND status = ? AND token_quota > 0 AND expire_at > ?",
				userId,
				common.UserPackageStatusActive,
				time.Now())

		if specificPackageId != nil && *specificPackageId > 0 {
			query = query.Where("id = ?", *specificPackageId)
		}

		var packages []*model.UserPackage
		if err := query.Order("expire_at ASC").Find(&packages).Error; err != nil {
			return err
		}

		if len(packages) == 0 {
			return errors.New("没有可用的积分包")
		}

		// Filter packages by model permissions
		allowedPackages := make([]*model.UserPackage, 0)
		for _, pkg := range packages {
			if pkg.IsModelAllowed(modelName) {
				allowedPackages = append(allowedPackages, pkg)
			}
		}

		if len(allowedPackages) == 0 {
			return errors.New("没有包含所选模型权限的积分包")
		}

		remaining := quota
		for _, pkg := range allowedPackages {
			if remaining <= 0 {
				break
			}

			consume := pkg.TokenQuota
			if consume > remaining {
				consume = remaining
			}

			pkg.TokenQuota -= consume
			if pkg.TokenQuota <= 0 {
				pkg.Status = common.UserPackageStatusExhausted
			}

			if err := tx.Model(pkg).Select("token_quota", "status").Updates(pkg).Error; err != nil {
				return err
			}

			remaining -= consume
			pkgCopy := *pkg
			pkgCopy.TokenQuota = consume
			consumedPackages = append(consumedPackages, &pkgCopy)
		}

		if remaining > 0 {
			return errors.New("积分包额度不足")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort consumed packages by expiry ascending (already done but ensure stable)
	sort.SliceStable(consumedPackages, func(i, j int) bool {
		return consumedPackages[i].ExpireAt.Before(consumedPackages[j].ExpireAt)
	})

	return consumedPackages, nil
}

// RefreshUserPackageStatuses updates expired packages status
func RefreshUserPackageStatuses(userId int) error {
	if userId <= 0 {
		return errors.New("invalid user id")
	}
	rn := time.Now()
	return model.DB.Model(&model.UserPackage{}).
		Where("user_id = ? AND status = ? AND expire_at <= ?", userId, common.UserPackageStatusActive, rn).
		Update("status", common.UserPackageStatusExpired).Error
}

// GetUserActivePackages returns active packages with up-to-date status
func GetUserActivePackages(userId int) ([]*model.UserPackage, error) {
	if err := RefreshUserPackageStatuses(userId); err != nil {
		return nil, err
	}

	packages, err := model.GetActiveUserPackages(userId)
	if err != nil {
		return nil, err
	}

	return packages, nil
}
