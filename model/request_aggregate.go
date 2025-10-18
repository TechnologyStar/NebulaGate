package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RequestAggregate struct {
	Id             int       `json:"id"`
	ModelAlias     string    `json:"model_alias" gorm:"size:128;not null;uniqueIndex:uk_request_aggregates_window,priority:1"`
	Upstream       string    `json:"upstream" gorm:"size:64;not null;uniqueIndex:uk_request_aggregates_window,priority:2"`
	SubjectType    string    `json:"subject_type" gorm:"size:16;not null;uniqueIndex:uk_request_aggregates_window,priority:3"`
	WindowStart    time.Time `json:"window_start" gorm:"not null;uniqueIndex:uk_request_aggregates_window,priority:4"`
	WindowEnd      time.Time `json:"window_end" gorm:"not null;uniqueIndex:uk_request_aggregates_window,priority:5;index"`
	TotalRequests  int64     `json:"total_requests" gorm:"type:bigint;not null;default:0"`
	TotalTokens    int64     `json:"total_tokens" gorm:"type:bigint;not null;default:0"`
	UniqueSubjects int64     `json:"unique_subjects" gorm:"type:bigint;not null;default:0"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func InsertRequestAggregate(aggregate *RequestAggregate) error {
	if aggregate == nil {
		return errors.New("aggregate required")
	}
	if aggregate.ModelAlias == "" || aggregate.Upstream == "" || aggregate.SubjectType == "" {
		return errors.New("aggregate key fields required")
	}
	if aggregate.WindowStart.IsZero() || aggregate.WindowEnd.IsZero() {
		return errors.New("aggregate window required")
	}
	aggregate.WindowStart = aggregate.WindowStart.UTC()
	aggregate.WindowEnd = aggregate.WindowEnd.UTC()
	if aggregate.WindowEnd.Before(aggregate.WindowStart) {
		return errors.New("aggregate window invalid")
	}
	if common.RequestAggregateCacheEnabled && common.RedisEnabled {
		invalidateRequestAggregateCache(aggregate)
	}
	return DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "model_alias"},
			{Name: "upstream"},
			{Name: "subject_type"},
			{Name: "window_start"},
			{Name: "window_end"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"total_requests": gorm.Expr("total_requests + ?", aggregate.TotalRequests),
			"total_tokens":   gorm.Expr("total_tokens + ?", aggregate.TotalTokens),
			"unique_subjects": gorm.Expr(
				"CASE WHEN unique_subjects < ? THEN ? ELSE unique_subjects END",
				aggregate.UniqueSubjects,
				aggregate.UniqueSubjects,
			),
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}),
	}).Create(aggregate).Error
}

func invalidateRequestAggregateCache(aggregate *RequestAggregate) {}
