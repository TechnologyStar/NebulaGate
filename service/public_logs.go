package service

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/model"

	"gorm.io/gorm"
)

type channelMetadata struct {
	Name string
	Type string
}

type PublicLogQuery struct {
	Window   string
	Model    string
	Search   string
	Page     int
	PageSize int
}

type PublicLogItem struct {
	ID            string `json:"id"`
	CreatedAt     int64  `json:"created_at"`
	Model         string `json:"model"`
	SubjectLabel  string `json:"subject_label"`
	TokenLabel    string `json:"token_label"`
	TokenName     string `json:"token_name"`
	Tokens        int64  `json:"tokens"`
	PromptTokens  int64  `json:"prompt_tokens"`
	Completion    int64  `json:"completion_tokens"`
	Status        string `json:"status"`
	UpstreamAlias string `json:"upstream_alias"`
	Summary       string `json:"summary"`
}

type PublicLogMetrics struct {
	TotalRequests int64 `json:"total_requests"`
	TotalTokens   int64 `json:"total_tokens"`
	SuccessCount  int64 `json:"success_count"`
	FailedCount   int64 `json:"failed_count"`
}

type PublicLogsResult struct {
	Items   []PublicLogItem  `json:"items"`
	Total   int64            `json:"total"`
	Metrics PublicLogMetrics `json:"metrics"`
}

type ModelLeaderboardEntry struct {
	Model        string `json:"model"`
	RequestCount int64  `json:"request_count"`
	TokenCount   int64  `json:"token_count"`
	UniqueUsers  int64  `json:"unique_users"`
	UniqueTokens int64  `json:"unique_tokens"`
}

type IPUsageResponseItem struct {
	IP           string `json:"ip"`
	RequestCount int64  `json:"request_count"`
	FirstSeenAt  int64  `json:"first_seen_at"`
	LastSeenAt   int64  `json:"last_seen_at"`
}

type IPUsageResponse struct {
	SubjectID     int                   `json:"subject_id"`
	SubjectLabel  string                `json:"subject_label"`
	SubjectType   string                `json:"subject_type"`
	Window        string                `json:"window"`
	UniqueCount   int                   `json:"unique_count"`
	TotalRequests int64                 `json:"total_requests"`
	Items         []IPUsageResponseItem `json:"items"`
}

const (
	defaultPageSize   = 20
	maxPublicPageSize = 200
	exportLimit       = 2000
)

func getLogDB() *gorm.DB {
	if model.LOG_DB != nil {
		return model.LOG_DB
	}
	return model.DB
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPublicPageSize {
		pageSize = maxPublicPageSize
	}
	return page, pageSize
}

func parseSearchFilter(tx *gorm.DB, search string) *gorm.DB {
	normalized := strings.TrimSpace(search)
	if normalized == "" {
		return tx
	}
	lower := strings.ToLower(normalized)
	if strings.HasPrefix(lower, "token#") {
		digits := strings.TrimSpace(lower[len("token#"):])
		if id, err := strconv.Atoi(digits); err == nil {
			return tx.Where("token_id = ?", id)
		}
	}
	if id, err := strconv.Atoi(normalized); err == nil {
		return tx.Where("token_id = ?", id)
	}
	like := fmt.Sprintf("%%%s%%", normalized)
	return tx.Where("content LIKE ? OR token_name LIKE ? OR model_name LIKE ?", like, like, like)
}

func getWindowStart(window string) (time.Time, bool, error) {
	start, allTime, err := common.ParseRelativeWindow(window)
	if err != nil {
		return time.Time{}, false, err
	}
	if allTime {
		return time.Time{}, true, nil
	}
	return start, false, nil
}

func buildPublicLogQuery(query PublicLogQuery) (*gorm.DB, time.Time, bool, error) {
	start, allTime, err := getWindowStart(query.Window)
	if err != nil {
		return nil, time.Time{}, false, err
	}
	tx := getLogDB().Model(&model.Log{}).Where("type IN ?", []int{model.LogTypeConsume, model.LogTypeError})
	if !allTime {
		tx = tx.Where("created_at >= ?", start.Unix())
	}
	if query.Model != "" {
		tx = tx.Where("model_name = ?", query.Model)
	}
	if query.Search != "" {
		tx = parseSearchFilter(tx, query.Search)
	}
	return tx, start, allTime, nil
}

func mapChannelInfo(channelIDs []int) map[int]channelMetadata {
	if len(channelIDs) == 0 {
		return map[int]channelMetadata{}
	}
	var records []struct {
		ID   int
		Name string
		Type int
	}
	if err := model.DB.Table("channels").Select("id, name, type").Where("id IN ?", channelIDs).Find(&records).Error; err != nil {
		return map[int]channelMetadata{}
	}
	result := make(map[int]channelMetadata, len(records))
	for _, rec := range records {
		result[rec.ID] = channelMetadata{
			Name: rec.Name,
			Type: strings.ToLower(constant.GetChannelTypeName(rec.Type)),
		}
	}
	return result
}

func buildStatus(log *model.Log) string {
	if log.Type == model.LogTypeError {
		return "failed"
	}
	if log.PromptTokens == 0 && log.CompletionTokens == 0 && log.Quota == 0 {
		return "failed"
	}
	if other, err := common.StrToMap(log.Other); err == nil {
		if cached, ok := other["cache_tokens"].(float64); ok && cached > 0 {
			return "cached"
		}
		if cached, ok := other["cache_creation_tokens"].(float64); ok && cached > 0 {
			return "cached"
		}
	}
	return "success"
}

func summarizeLogs(logs []model.Log, channelMeta map[int]channelMetadata) ([]PublicLogItem, PublicLogMetrics) {
	items := make([]PublicLogItem, 0, len(logs))
	var metrics PublicLogMetrics
	for _, log := range logs {
		metrics.TotalRequests++
		if log.Type == model.LogTypeConsume {
			metrics.TotalTokens += int64(log.PromptTokens + log.CompletionTokens)
		}
		status := buildStatus(&log)
		if status == "failed" {
			metrics.FailedCount++
		} else {
			metrics.SuccessCount++
		}
		meta := channelMeta[log.ChannelId]
		tokens := int64(log.PromptTokens + log.CompletionTokens)
		subjectLabel := ""
		tokenLabel := ""
		if log.TokenId > 0 {
			subjectLabel = fmt.Sprintf("Token #%d", log.TokenId)
			tokenLabel = subjectLabel
		} else if log.UserId > 0 {
			subjectLabel = fmt.Sprintf("User #%d", log.UserId)
		} else {
			subjectLabel = "Anonymous"
		}
		items = append(items, PublicLogItem{
			ID:            fmt.Sprintf("%d", log.Id),
			CreatedAt:     log.CreatedAt,
			Model:         log.ModelName,
			SubjectLabel:  subjectLabel,
			TokenLabel:    tokenLabel,
			TokenName:     log.TokenName,
			Tokens:        tokens,
			PromptTokens:  int64(log.PromptTokens),
			Completion:    int64(log.CompletionTokens),
			Status:        status,
			UpstreamAlias: pickUpstream(meta),
			Summary:       log.Content,
		})
	}
	return items, metrics
}

func pickUpstream(meta channelMetadata) string {
	if meta.Name != "" {
		return meta.Name
	}
	return meta.Type
}

func GetPublicLogs(query PublicLogQuery) (*PublicLogsResult, error) {
	if !common.PublicLogsFeatureEnabled {
		return nil, fmt.Errorf("public logs are disabled")
	}
	page, pageSize := normalizePagination(query.Page, query.PageSize)
	tx, _, _, err := buildPublicLogQuery(query)
	if err != nil {
		return nil, err
	}
	var total int64
	if err := tx.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, err
	}
	var logs []model.Log
	if err := tx.Session(&gorm.Session{}).Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&logs).Error; err != nil {
		return nil, err
	}
	channelIDs := make([]int, 0)
	seen := make(map[int]struct{})
	for _, log := range logs {
		if log.ChannelId != 0 {
			if _, ok := seen[log.ChannelId]; !ok {
				seen[log.ChannelId] = struct{}{}
				channelIDs = append(channelIDs, log.ChannelId)
			}
		}
	}
	channelMeta := mapChannelInfo(channelIDs)
	items, metrics := summarizeLogs(logs, channelMeta)
	return &PublicLogsResult{Items: items, Total: total, Metrics: metrics}, nil
}

func GetPublicLogModels(window string) ([]string, error) {
	if !common.PublicLogsFeatureEnabled {
		return nil, fmt.Errorf("public logs are disabled")
	}
	tx, _, _, err := buildPublicLogQuery(PublicLogQuery{Window: window})
	if err != nil {
		return nil, err
	}
	var models []string
	if err := tx.Distinct("model_name").Pluck("model_name", &models).Error; err != nil {
		return nil, err
	}
	sort.Strings(models)
	return models, nil
}

func ExportPublicLogs(query PublicLogQuery) ([][]string, error) {
	page, pageSize := normalizePagination(1, exportLimit)
	query.Page = page
	query.PageSize = pageSize
	result, err := GetPublicLogs(query)
	if err != nil {
		return nil, err
	}
	rows := [][]string{{"ID", "Timestamp", "Model", "Tokens", "Status", "Token", "Upstream", "Summary"}}
	for _, item := range result.Items {
		rows = append(rows, []string{
			item.ID,
			time.Unix(item.CreatedAt, 0).UTC().Format(time.RFC3339),
			item.Model,
			fmt.Sprintf("%d", item.Tokens),
			item.Status,
			item.SubjectLabel,
			item.UpstreamAlias,
			item.Summary,
		})
	}
	return rows, nil
}

func WriteCSV(writer *csv.Writer, rows [][]string) error {
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func GetModelLeaderboard(window string, limit int) ([]ModelLeaderboardEntry, error) {
	if !common.PublicLogsFeatureEnabled {
		return nil, fmt.Errorf("public logs are disabled")
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	query := PublicLogQuery{Window: window}
	tx, _, _, err := buildPublicLogQuery(query)
	if err != nil {
		return nil, err
	}
	tx = tx.Where("type = ?", model.LogTypeConsume)
	rows := make([]struct {
		Model        string
		RequestCount int64
		TokenCount   int64
		UniqueUsers  int64
		UniqueTokens int64
	}, 0)
	if err := tx.Select("model_name as model, COUNT(*) as request_count, COALESCE(SUM(prompt_tokens + completion_tokens),0) as token_count, COUNT(DISTINCT user_id) as unique_users, COUNT(DISTINCT token_id) as unique_tokens").Group("model_name").Order("request_count desc").Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]ModelLeaderboardEntry, 0, len(rows))
	for _, row := range rows {
		result = append(result, ModelLeaderboardEntry{
			Model:        row.Model,
			RequestCount: row.RequestCount,
			TokenCount:   row.TokenCount,
			UniqueUsers:  row.UniqueUsers,
			UniqueTokens: row.UniqueTokens,
		})
	}
	return result, nil
}

func GetTokenIPUsageSummary(tokenId int, window string) (*IPUsageResponse, error) {
	start, _, err := getWindowStart(window)
	if err != nil {
		return nil, err
	}
	usages, totalRequests, err := model.GetTokenIPUsage(tokenId, start)
	if err != nil {
		return nil, err
	}
	items := make([]IPUsageResponseItem, 0, len(usages))
	for _, usage := range usages {
		items = append(items, IPUsageResponseItem{
			IP:           usage.IP,
			RequestCount: usage.RequestCount,
			FirstSeenAt:  usage.FirstSeenAt.Unix(),
			LastSeenAt:   usage.LastSeenAt.Unix(),
		})
	}
	return &IPUsageResponse{
		SubjectID:     tokenId,
		SubjectLabel:  fmt.Sprintf("Token #%d", tokenId),
		SubjectType:   "token",
		Window:        window,
		UniqueCount:   len(usages),
		TotalRequests: totalRequests,
		Items:         items,
	}, nil
}

func GetUserIPUsageSummary(userId int, window string) (*IPUsageResponse, error) {
	start, _, err := getWindowStart(window)
	if err != nil {
		return nil, err
	}
	usages, totalRequests, err := model.GetUserIPUsage(userId, start)
	if err != nil {
		return nil, err
	}
	items := make([]IPUsageResponseItem, 0, len(usages))
	for _, usage := range usages {
		items = append(items, IPUsageResponseItem{
			IP:           usage.IP,
			RequestCount: usage.RequestCount,
			FirstSeenAt:  usage.FirstSeenAt.Unix(),
			LastSeenAt:   usage.LastSeenAt.Unix(),
		})
	}
	return &IPUsageResponse{
		SubjectID:     userId,
		SubjectLabel:  fmt.Sprintf("User #%d", userId),
		SubjectType:   "user",
		Window:        window,
		UniqueCount:   len(usages),
		TotalRequests: totalRequests,
		Items:         items,
	}, nil
}

func BuildExportURL(basePath string, query PublicLogQuery) string {
	params := url.Values{}
	if query.Window != "" {
		params.Set("window", query.Window)
	}
	if query.Model != "" {
		params.Set("model", query.Model)
	}
	if query.Search != "" {
		params.Set("search", query.Search)
	}
	encoded := params.Encode()
	if encoded == "" {
		return basePath
	}
	return fmt.Sprintf("%s?%s", basePath, encoded)
}
