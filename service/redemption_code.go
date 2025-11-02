package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"gorm.io/gorm"
)

const codeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateCode(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = codeCharset[rand.Intn(len(codeCharset))]
	}
	return string(b)
}

// GenerateRedemptionCodes generates redemption codes for a package
func GenerateRedemptionCodes(packageId int, quantity int) ([]string, error) {
	if packageId <= 0 {
		return nil, errors.New("invalid package id")
	}

	if quantity <= 0 || quantity > 500 {
		return nil, errors.New("quantity must be between 1 and 500")
	}

	// Verify package exists
	pkg, err := model.GetPackageById(packageId)
	if err != nil {
		return nil, fmt.Errorf("package not found: %w", err)
	}

	if pkg.Status != common.PackageStatusActive {
		return nil, errors.New("package is not active")
	}

	codes := make([]string, 0, quantity)
	existingCodes := make(map[string]bool)

	return codes, model.DB.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < quantity; i++ {
			var code string
			maxAttempts := 10

			for attempt := 0; attempt < maxAttempts; attempt++ {
				code = generateCode(16)

				// Check if already generated in this batch
				if existingCodes[code] {
					continue
				}

				// Check if code exists in database
				var count int64
				if err := tx.Model(&model.RedemptionCode{}).Where("code = ?", code).Count(&count).Error; err != nil {
					return err
				}

				if count == 0 {
					existingCodes[code] = true
					break
				}
			}

			if !existingCodes[code] {
				return fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
			}

			redemptionCode := &model.RedemptionCode{
				PackageId: packageId,
				Code:      code,
				Status:    common.RedemptionCodeStatusUnused,
			}

			if err := tx.Create(redemptionCode).Error; err != nil {
				return err
			}

			codes = append(codes, code)
		}

		return nil
	})
}

// RevokeRedemptionCode revokes a redemption code
func RevokeRedemptionCode(id int) error {
	if id <= 0 {
		return errors.New("invalid redemption code id")
	}

	code, err := model.GetRedemptionCodeById(id)
	if err != nil {
		return err
	}

	if code.Status == common.RedemptionCodeStatusRedeemed {
		return errors.New("cannot revoke an already used code")
	}

	if code.Status == common.RedemptionCodeStatusRevoked {
		return errors.New("code is already revoked")
	}

	code.Status = common.RedemptionCodeStatusRevoked
	return code.Update("status")
}

// RedeemPackageCode redeems a package code for a user
func RedeemPackageCode(code string, userId int) (*model.UserPackage, error) {
	if code == "" {
		return nil, errors.New("兑换码不能为空")
	}

	if userId <= 0 {
		return nil, errors.New("无效的用户ID")
	}

	code = strings.ToUpper(strings.TrimSpace(code))

	var userPackage *model.UserPackage

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		// Lock redemption code for update
		var redemptionCode model.RedemptionCode
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("code = ?", code).
			First(&redemptionCode).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("兑换码不存在")
			}
			return err
		}

		// Check if code is unused
		if redemptionCode.Status != common.RedemptionCodeStatusUnused {
			switch redemptionCode.Status {
			case common.RedemptionCodeStatusRedeemed:
				return errors.New("兑换码已被使用")
			case common.RedemptionCodeStatusRevoked:
				return errors.New("兑换码已被作废")
			default:
				return errors.New("兑换码状态异常")
			}
		}

		// Get package info
		var pkg model.Package
		err = tx.Where("id = ?", redemptionCode.PackageId).First(&pkg).Error
		if err != nil {
			return errors.New("套餐不存在")
		}

		if pkg.Status != common.PackageStatusActive {
			return errors.New("套餐已被禁用")
		}

		// Create user package
		expireAt := time.Now().AddDate(0, 0, pkg.ValidityDays)
		userPackage = &model.UserPackage{
			UserId:       userId,
			PackageId:    pkg.Id,
			TokenQuota:   pkg.TokenQuota,
			ModelScope:   pkg.ModelScope,
			ExpireAt:     expireAt,
			Status:       common.UserPackageStatusActive,
			InitialQuota: pkg.TokenQuota,
		}

		if err := tx.Create(userPackage).Error; err != nil {
			return err
		}

		// Update redemption code
		now := time.Now()
		redemptionCode.Status = common.RedemptionCodeStatusRedeemed
		redemptionCode.UsedByUserId = userId
		redemptionCode.UsedAt = &now
		redemptionCode.UserPackageId = &userPackage.Id

		if err := tx.Save(&redemptionCode).Error; err != nil {
			return err
		}

		// Load package relation
		userPackage.Package = pkg

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Log the redemption
	logContent := fmt.Sprintf("兑换套餐 %s，获得 %d tokens，有效期 %d 天",
		userPackage.Package.Name,
		userPackage.TokenQuota,
		userPackage.Package.ValidityDays)
	model.RecordLog(userId, model.LogTypeTopup, logContent)

	return userPackage, nil
}
