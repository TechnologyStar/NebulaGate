package service

import (
	"time"

	"github.com/QuantumNous/new-api/model"
)

type UserLeaderboardEntry struct {
	UserId        int    `json:"user_id"`
	Username      string `json:"username"`
	RequestCount  int64  `json:"request_count"`
	TokenCount    int64  `json:"token_count"`
	QuotaConsumed int64  `json:"quota_consumed"`
	UniqueModels  int64  `json:"unique_models"`
}

func GetUserLeaderboard(window string, limit int) ([]UserLeaderboardEntry, error) {
	start, allTime, err := getWindowStart(window)
	if err != nil {
		return nil, err
	}

	db := getLogDB()
	query := db.Table("logs").
		Select(`
			user_id,
			username,
			COUNT(*) as request_count,
			SUM(prompt_tokens + completion_tokens) as token_count,
			SUM(quota) as quota_consumed,
			COUNT(DISTINCT model_name) as unique_models
		`).
		Where("type = ?", model.LogTypeConsume).
		Where("user_id > 0")

	if !allTime {
		query = query.Where("created_at >= ?", start.Unix())
	}

	query = query.Group("user_id, username").
		Order("request_count DESC").
		Limit(limit)

	var entries []UserLeaderboardEntry
	if err := query.Scan(&entries).Error; err != nil {
		return nil, err
	}

	return entries, nil
}

func GetUserStats(userId int, window string) (*UserLeaderboardEntry, error) {
	start, allTime, err := getWindowStart(window)
	if err != nil {
		return nil, err
	}

	db := getLogDB()
	query := db.Table("logs").
		Select(`
			user_id,
			username,
			COUNT(*) as request_count,
			SUM(prompt_tokens + completion_tokens) as token_count,
			SUM(quota) as quota_consumed,
			COUNT(DISTINCT model_name) as unique_models
		`).
		Where("type = ?", model.LogTypeConsume).
		Where("user_id = ?", userId)

	if !allTime {
		query = query.Where("created_at >= ?", start.Unix())
	}

	query = query.Group("user_id, username")

	var entry UserLeaderboardEntry
	if err := query.Scan(&entry).Error; err != nil {
		return nil, err
	}

	return &entry, nil
}

func GetTokenIPStats(tokenId int, window string) (map[string]interface{}, error) {
	start, allTime, err := getWindowStart(window)
	if err != nil {
		return nil, err
	}

	var since time.Time
	if !allTime {
		since = start
	}

	usages, totalRequests, err := model.GetTokenIPUsage(tokenId, since)
	if err != nil {
		return nil, err
	}

	token, err := model.GetTokenById(tokenId)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"token_id":         tokenId,
		"token_name":       token.Name,
		"unique_ip_count":  len(usages),
		"total_requests":   totalRequests,
		"window":           window,
		"ip_list":          usages,
	}

	return result, nil
}
