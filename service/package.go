package service

import (
	"errors"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// CreatePackage handles creation logic for packages
func CreatePackage(pkg *model.Package) error {
	if pkg == nil {
		return errors.New("invalid package payload")
	}

	if pkg.Name == "" {
		return errors.New("套餐名称不能为空")
	}

	if pkg.TokenQuota <= 0 {
		return errors.New("Token额度必须大于0")
	}

	if pkg.ValidityDays <= 0 {
		return errors.New("有效期必须大于0")
	}

	if pkg.Status == 0 {
		pkg.Status = common.PackageStatusActive
	}

	return pkg.Insert()
}

// UpdatePackage updates package details
func UpdatePackage(id int, updates *model.Package) (*model.Package, error) {
	if id <= 0 {
		return nil, errors.New("invalid package id")
	}

	existing, err := model.GetPackageById(id)
	if err != nil {
		return nil, err
	}

	existing.Name = updates.Name
	existing.Description = updates.Description
	existing.TokenQuota = updates.TokenQuota
	existing.ModelScope = updates.ModelScope
	existing.ValidityDays = updates.ValidityDays
	existing.Price = updates.Price
	if updates.Status != 0 {
		existing.Status = updates.Status
	}

	if err := existing.Update(); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeletePackage deletes a package after verifying no unused codes exist
func DeletePackage(id int) error {
	if id <= 0 {
		return errors.New("invalid package id")
	}

	codes, _, err := model.ListRedemptionCodes(0, 1, id, common.RedemptionCodeStatusUnused, "")
	if err != nil {
		return err
	}
	if len(codes) > 0 {
		return errors.New("该套餐还有未使用的兑换码，无法删除")
	}

	return model.DeletePackageById(id)
}
