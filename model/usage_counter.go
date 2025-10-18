package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UsageCounter struct {
	Id               int       `json:"id"`
	PlanAssignmentId int       `json:"plan_assignment_id" gorm:"not null;index;uniqueIndex:uk_usage_assignment_metric_cycle,priority:1"`
	Metric           string    `json:"metric" gorm:"size:16;not null;uniqueIndex:uk_usage_assignment_metric_cycle,priority:2"`
	CycleStart       time.Time `json:"cycle_start" gorm:"not null;uniqueIndex:uk_usage_assignment_metric_cycle,priority:3"`
	CycleEnd         time.Time `json:"cycle_end" gorm:"not null;index"`
	ConsumedAmount   int64     `json:"consumed_amount" gorm:"type:bigint;not null;default:0"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func IncrementUsageCounter(assignmentId int, metric string, amount int64, cycleStart time.Time, cycleEnd time.Time) error {
	if assignmentId == 0 {
		return errors.New("assignment id required")
	}
	if metric == "" {
		return errors.New("metric required")
	}
	if amount == 0 {
		return nil
	}
	if cycleEnd.Before(cycleStart) {
		return errors.New("cycle end before start")
	}
	cycleStart = cycleStart.UTC()
	cycleEnd = cycleEnd.UTC()
	counter := UsageCounter{
		PlanAssignmentId: assignmentId,
		Metric:           metric,
		CycleStart:       cycleStart,
		CycleEnd:         cycleEnd,
		ConsumedAmount:   amount,
	}
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "plan_assignment_id"}, {Name: "metric"}, {Name: "cycle_start"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"consumed_amount": gorm.Expr("consumed_amount + ?", amount),
			"cycle_end":       cycleEnd,
			"updated_at":      gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(&counter).Error
}

func ResetUsageCounter(assignmentId int, metric string, cycleStart time.Time) error {
	if assignmentId == 0 || metric == "" {
		return errors.New("assignment id and metric required")
	}
	return DB.Where("plan_assignment_id = ? AND metric = ? AND cycle_start = ?", assignmentId, metric, cycleStart.UTC()).Delete(&UsageCounter{}).Error
}
