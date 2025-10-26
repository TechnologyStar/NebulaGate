package model

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gorm.io/gorm"
)

type LotteryConfig struct {
	Id          int       `json:"id"`
	Name        string    `json:"name" gorm:"size:128;not null"`
	PrizeType   string    `json:"prize_type" gorm:"size:32;not null"`
	PrizeValue  int       `json:"prize_value" gorm:"not null"`
	Probability float64   `json:"probability" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"default:-1"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LotteryRecord struct {
	Id         int       `json:"id"`
	UserId     int       `json:"user_id" gorm:"not null;index"`
	ConfigId   int       `json:"config_id" gorm:"not null;index"`
	PrizeType  string    `json:"prize_type" gorm:"size:32;not null"`
	PrizeValue int       `json:"prize_value" gorm:"not null"`
	PrizeName  string    `json:"prize_name" gorm:"size:128"`
	CreatedAt  time.Time `json:"created_at"`
}

func DrawLottery(userId int) (*LotteryRecord, error) {
	if userId == 0 {
		return nil, errors.New("invalid user id")
	}

	var configs []LotteryConfig
	err := DB.Where("is_active = ?", true).Find(&configs).Error
	if err != nil {
		return nil, err
	}

	if len(configs) == 0 {
		return nil, errors.New("no active lottery prizes")
	}

	var totalProb float64
	for _, cfg := range configs {
		if cfg.Stock == 0 {
			continue
		}
		totalProb += cfg.Probability
	}

	if totalProb <= 0 {
		return nil, errors.New("no available prizes")
	}

	randMax := big.NewInt(1000000)
	randNum, err := rand.Int(rand.Reader, randMax)
	if err != nil {
		return nil, err
	}
	roll := float64(randNum.Int64()) / 1000000.0 * totalProb

	var selectedConfig *LotteryConfig
	var cumulative float64
	for i := range configs {
		if configs[i].Stock == 0 {
			continue
		}
		cumulative += configs[i].Probability
		if roll <= cumulative {
			selectedConfig = &configs[i]
			break
		}
	}

	if selectedConfig == nil {
		return nil, errors.New("lottery draw failed")
	}

	var record *LotteryRecord
	err = DB.Transaction(func(tx *gorm.DB) error {
		if selectedConfig.Stock > 0 {
			if err := tx.Model(&LotteryConfig{}).Where("id = ? AND stock > 0", selectedConfig.Id).
				Update("stock", gorm.Expr("stock - 1")).Error; err != nil {
				return err
			}
		}

		record = &LotteryRecord{
			UserId:     userId,
			ConfigId:   selectedConfig.Id,
			PrizeType:  selectedConfig.PrizeType,
			PrizeValue: selectedConfig.PrizeValue,
			PrizeName:  selectedConfig.Name,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		switch selectedConfig.PrizeType {
		case "quota":
			if err := tx.Model(&User{}).Where("id = ?", userId).
				Update("quota", gorm.Expr("quota + ?", selectedConfig.PrizeValue)).Error; err != nil {
				return err
			}
			RecordLog(userId, LogTypeTopup, fmt.Sprintf("抽奖奖励：%s，获得 %d 额度", selectedConfig.Name, selectedConfig.PrizeValue))

		case "plan":
			plan, err := GetPlanById(selectedConfig.PrizeValue)
			if err != nil {
				return errors.New("invalid plan id in lottery prize")
			}
			assignment := &PlanAssignment{
				SubjectType: "user",
				SubjectId:   userId,
				PlanId:      plan.Id,
			}
			if err := tx.Create(assignment).Error; err != nil {
				return err
			}
			RecordLog(userId, LogTypeTopup, fmt.Sprintf("抽奖奖励：%s，获得套餐 %s", selectedConfig.Name, plan.Name))
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return record, nil
}

func GetLotteryConfigs() ([]LotteryConfig, error) {
	var configs []LotteryConfig
	err := DB.Order("id desc").Find(&configs).Error
	return configs, err
}

func GetLotteryConfigById(id int) (*LotteryConfig, error) {
	var config LotteryConfig
	err := DB.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (config *LotteryConfig) Create() error {
	return DB.Create(config).Error
}

func (config *LotteryConfig) Update() error {
	return DB.Model(config).Updates(map[string]interface{}{
		"name":        config.Name,
		"prize_type":  config.PrizeType,
		"prize_value": config.PrizeValue,
		"probability": config.Probability,
		"stock":       config.Stock,
		"is_active":   config.IsActive,
	}).Error
}

func DeleteLotteryConfig(id int) error {
	return DB.Delete(&LotteryConfig{}, id).Error
}

func GetLotteryRecords(userId int, limit int, offset int) ([]LotteryRecord, int64, error) {
	var records []LotteryRecord
	var total int64

	query := DB.Where("user_id = ?", userId)

	if err := query.Model(&LotteryRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func GetAllLotteryRecords(limit int, offset int) ([]LotteryRecord, int64, error) {
	var records []LotteryRecord
	var total int64

	if err := DB.Model(&LotteryRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}
