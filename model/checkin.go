package model

import (
    "errors"
    "fmt"
    "time"

    "gorm.io/gorm"
)

type CheckInRecord struct {
    Id              int       `json:"id"`
    UserId          int       `json:"user_id" gorm:"not null;index"`
    CheckInDate     string    `json:"check_in_date" gorm:"size:10;not null;index:idx_user_date"`
    QuotaAwarded    int       `json:"quota_awarded" gorm:"not null;default:0"`
    ConsecutiveDays int       `json:"consecutive_days" gorm:"not null;default:1"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

func (c *CheckInRecord) TableName() string {
    return "check_in_records"
}

// CheckIn performs daily check-in and awards quota
func CheckIn(userId int) (quotaAwarded int, consecutiveDays int, err error) {
    if userId == 0 {
        return 0, 0, errors.New("invalid user id")
    }

    today := time.Now().UTC().Format("2006-01-02")
    yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

    err = DB.Transaction(func(tx *gorm.DB) error {
        var existing CheckInRecord
        err := tx.Where("user_id = ? AND check_in_date = ?", userId, today).First(&existing).Error
        if err == nil {
            return errors.New("already checked in today")
        }
        if !errors.Is(err, gorm.ErrRecordNotFound) {
            return err
        }

        var yesterdayRecord CheckInRecord
        err = tx.Where("user_id = ? AND check_in_date = ?", userId, yesterday).First(&yesterdayRecord).Error
        consecutiveDays = 1
        if err == nil {
            consecutiveDays = yesterdayRecord.ConsecutiveDays + 1
        }

        quotaAwarded = calculateCheckInQuota(consecutiveDays)

        record := CheckInRecord{
            UserId:          userId,
            CheckInDate:     today,
            QuotaAwarded:    quotaAwarded,
            ConsecutiveDays: consecutiveDays,
        }
        if err := tx.Create(&record).Error; err != nil {
            return err
        }

        if err := tx.Model(&User{}).Where("id = ?", userId).Update("quota", gorm.Expr("quota + ?", quotaAwarded)).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        return 0, 0, err
    }

    RecordLog(userId, LogTypeTopup, fmt.Sprintf("签到奖励：连续签到%d天，获得 %d 额度", consecutiveDays, quotaAwarded))
    return quotaAwarded, consecutiveDays, nil
}

func calculateCheckInQuota(consecutiveDays int) int {
    baseQuota := 100000
    if consecutiveDays >= 30 {
        return baseQuota * 5
    } else if consecutiveDays >= 14 {
        return baseQuota * 3
    } else if consecutiveDays >= 7 {
        return baseQuota * 2
    }
    return baseQuota
}

func GetCheckInHistory(userId int, limit int, offset int) ([]CheckInRecord, int64, error) {
    var records []CheckInRecord
    var total int64

    query := DB.Where("user_id = ?", userId)

    if err := query.Model(&CheckInRecord{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if err := query.Order("check_in_date desc").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
        return nil, 0, err
    }

    return records, total, nil
}

func GetTodayCheckInStatus(userId int) (bool, *CheckInRecord, error) {
    today := time.Now().UTC().Format("2006-01-02")
    var record CheckInRecord
    err := DB.Where("user_id = ? AND check_in_date = ?", userId, today).First(&record).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return false, nil, nil
        }
        return false, nil, err
    }
    return true, &record, nil
}

func GetUserConsecutiveDays(userId int) (int, error) {
    today := time.Now().UTC().Format("2006-01-02")
    yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

    var record CheckInRecord
    err := DB.Where("user_id = ? AND check_in_date IN ?", userId, []string{today, yesterday}).
        Order("check_in_date desc").First(&record).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return 0, nil
        }
        return 0, err
    }

    return record.ConsecutiveDays, nil
}

// GetCheckInRewardConfig returns the check-in reward configuration
func GetCheckInRewardConfig() []map[string]interface{} {
    return []map[string]interface{}{
        {
            "day_range":      "1",
            "reward_quota":   100000,
            "description":    "第1天",
            "is_bonus":       false,
        },
        {
            "day_range":      "2-6",
            "reward_quota":   100000,
            "description":    "第2-6天",
            "is_bonus":       false,
        },
        {
            "day_range":      "7-13",
            "reward_quota":   200000,
            "description":    "第7-13天",
            "is_bonus":       false,
        },
        {
            "day_range":      "14-29",
            "reward_quota":   300000,
            "description":    "第14-29天",
            "is_bonus":       false,
        },
        {
            "day_range":      "30+",
            "reward_quota":   500000,
            "description":    "第30天及以上",
            "is_bonus":       true,
        },
    }
}
