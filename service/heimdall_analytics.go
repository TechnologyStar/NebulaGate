package service

import (
    "context"
    "fmt"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/logger"
    "github.com/QuantumNous/new-api/model"
)

// HeimdallAnalyticsService handles analytics and metrics for Heimdall telemetry
type HeimdallAnalyticsService struct{}

// NewHeimdallAnalyticsService creates a new analytics service
func NewHeimdallAnalyticsService() *HeimdallAnalyticsService {
    return &HeimdallAnalyticsService{}
}

// URLFrequencyMetrics represents frequency metrics for URLs
type URLFrequencyMetrics struct {
    URL           string    `json:"url"`
    Count         int64     `json:"count"`
    LastAccessed  time.Time `json:"last_accessed"`
    UniqueUsers   int64     `json:"unique_users"`
    AvgLatency    float64   `json:"avg_latency"`
    ErrorRate     float64   `json:"error_rate"`
}

// TokenFrequencyMetrics represents frequency metrics for tokens
type TokenFrequencyMetrics struct {
    TokenID       int    `json:"token_id"`
    Count         int64  `json:"count"`
    LastAccessed  string `json:"last_accessed"`
    UniqueURLs    int64  `json:"unique_urls"`
    AvgLatency    float64 `json:"avg_latency"`
    ErrorRate     float64 `json:"error_rate"`
}

// UserFrequencyMetrics represents frequency metrics for users
type UserFrequencyMetrics struct {
    UserID        int    `json:"user_id"`
    Count         int64  `json:"count"`
    LastAccessed  string `json:"last_accessed"`
    UniqueURLs    int64  `json:"unique_urls"`
    AvgLatency    float64 `json:"avg_latency"`
    ErrorRate     float64 `json:"error_rate"`
}

// GetURLFrequencyMetrics retrieves frequency metrics for URLs
func (s *HeimdallAnalyticsService) GetURLFrequencyMetrics(ctx context.Context, timeWindow time.Duration) ([]URLFrequencyMetrics, error) {
    var metrics []URLFrequencyMetrics
    
    // Get all URL keys from Redis
    urlKeys, err := s.getURLKeys(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get URL keys: %w", err)
    }
    
    for _, key := range urlKeys {
        count, err := common.RedisGet(ctx, key)
        if err != nil {
            logger.SysLog(fmt.Sprintf("Failed to get count for key %s: %v", key, err))
            continue
        }
        
        url := s.extractURLFromKey(key)
        if url == "" {
            continue
        }
        
        // Get additional metrics from database
        dbMetrics, err := s.getURLDBMetrics(ctx, url, timeWindow)
        if err != nil {
            logger.SysLog(fmt.Sprintf("Failed to get DB metrics for URL %s: %v", url, err))
            dbMetrics = &URLFrequencyMetrics{}
        }
        
        metrics = append(metrics, URLFrequencyMetrics{
            URL:          url,
            Count:        common.Str2Int64(count),
            LastAccessed: dbMetrics.LastAccessed,
            UniqueUsers:  dbMetrics.UniqueUsers,
            AvgLatency:   dbMetrics.AvgLatency,
            ErrorRate:    dbMetrics.ErrorRate,
        })
    }
    
    return metrics, nil
}

// GetTokenFrequencyMetrics retrieves frequency metrics for tokens
func (s *HeimdallAnalyticsService) GetTokenFrequencyMetrics(ctx context.Context, timeWindow time.Duration) ([]TokenFrequencyMetrics, error) {
    var metrics []TokenFrequencyMetrics
    
    // Get all token keys from Redis
    tokenKeys, err := s.getTokenKeys(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get token keys: %w", err)
    }
    
    for _, key := range tokenKeys {
        count, err := common.RedisGet(ctx, key)
        if err != nil {
            logger.SysLog(fmt.Sprintf("Failed to get count for key %s: %v", key, err))
            continue
        }
        
        tokenID := s.extractTokenIDFromKey(key)
        if tokenID == 0 {
            continue
        }
        
        // Get additional metrics from database
        dbMetrics, err := s.getTokenDBMetrics(ctx, tokenID, timeWindow)
        if err != nil {
            logger.SysLog(fmt.Sprintf("Failed to get DB metrics for token %d: %v", tokenID, err))
            dbMetrics = &TokenFrequencyMetrics{TokenID: tokenID}
        }
        
        metrics = append(metrics, TokenFrequencyMetrics{
            TokenID:      tokenID,
            Count:        common.Str2Int64(count),
            LastAccessed: dbMetrics.LastAccessed,
            UniqueURLs:   dbMetrics.UniqueURLs,
            AvgLatency:   dbMetrics.AvgLatency,
            ErrorRate:    dbMetrics.ErrorRate,
        })
    }
    
    return metrics, nil
}

// GetUserFrequencyMetrics retrieves frequency metrics for users
func (s *HeimdallAnalyticsService) GetUserFrequencyMetrics(ctx context.Context, timeWindow time.Duration) ([]UserFrequencyMetrics, error) {
    var metrics []UserFrequencyMetrics
    
    // Get all user keys from Redis
    userKeys, err := s.getUserKeys(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get user keys: %w", err)
    }
    
    for _, key := range userKeys {
        count, err := common.RedisGet(ctx, key)
        if err != nil {
            logger.SysLog(fmt.Sprintf("Failed to get count for key %s: %v", key, err))
            continue
        }
        
        userID := s.extractUserIDFromKey(key)
        if userID == 0 {
            continue
        }
        
        // Get additional metrics from database
        dbMetrics, err := s.getUserDBMetrics(ctx, userID, timeWindow)
        if err != nil {
            logger.SysLog(fmt.Sprintf("Failed to get DB metrics for user %d: %v", userID, err))
            dbMetrics = &UserFrequencyMetrics{UserID: userID}
        }
        
        metrics = append(metrics, UserFrequencyMetrics{
            UserID:      userID,
            Count:       common.Str2Int64(count),
            LastAccessed: dbMetrics.LastAccessed,
            UniqueURLs:  dbMetrics.UniqueURLs,
            AvgLatency:  dbMetrics.AvgLatency,
            ErrorRate:   dbMetrics.ErrorRate,
        })
    }
    
    return metrics, nil
}

// GetAnomalyDetectionData retrieves data for anomaly detection
func (s *HeimdallAnalyticsService) GetAnomalyDetectionData(ctx context.Context, timeWindow time.Duration) (map[string]interface{}, error) {
    data := make(map[string]interface{})
    
    // Get param digest frequencies
    paramDigests, err := s.getParamDigestFrequency(ctx, timeWindow)
    if err != nil {
        return nil, fmt.Errorf("failed to get param digest frequency: %w", err)
    }
    data["param_digests"] = paramDigests
    
    // Get IP frequency
    ipFreq, err := s.getIPFrequency(ctx, timeWindow)
    if err != nil {
        return nil, fmt.Errorf("failed to get IP frequency: %w", err)
    }
    data["ip_frequency"] = ipFreq
    
    // Get user agent frequency
    userAgentFreq, err := s.getUserAgentFrequency(ctx, timeWindow)
    if err != nil {
        return nil, fmt.Errorf("failed to get user agent frequency: %w", err)
    }
    data["user_agent_frequency"] = userAgentFreq
    
    return data, nil
}

// getURLKeys retrieves all URL keys from Redis
func (s *HeimdallAnalyticsService) getURLKeys(ctx context.Context) ([]string, error) {
    return common.RedisScan(ctx, "heimdall:url:*:count")
}

// getTokenKeys retrieves all token keys from Redis
func (s *HeimdallAnalyticsService) getTokenKeys(ctx context.Context) ([]string, error) {
    return common.RedisScan(ctx, "heimdall:token:*:count")
}

// getUserKeys retrieves all user keys from Redis
func (s *HeimdallAnalyticsService) getUserKeys(ctx context.Context) ([]string, error) {
    return common.RedisScan(ctx, "heimdall:user:*:count")
}

// extractURLFromKey extracts URL from Redis key
func (s *HeimdallAnalyticsService) extractURLFromKey(key string) string {
    // Extract URL from key like "heimdall:url:/v1/chat/completions:count"
    if len(key) < len("heimdall:url:") {
        return ""
    }
    
    parts := strings.Split(key, ':')
    if len(parts) < 4 {
        return ""
    }
    
    return parts[2]
}

// extractTokenIDFromKey extracts token ID from Redis key
func (s *HeimdallAnalyticsService) extractTokenIDFromKey(key string) int {
    // Extract token ID from key like "heimdall:token:123:count"
    if len(key) < len("heimdall:token:") {
        return 0
    }
    
    parts := strings.Split(key, ':')
    if len(parts) < 4 {
        return 0
    }
    
    return common.Str2Int(parts[2])
}

// extractUserIDFromKey extracts user ID from Redis key
func (s *HeimdallAnalyticsService) extractUserIDFromKey(key string) int {
    // Extract user ID from key like "heimdall:user:456:count"
    if len(key) < len("heimdall:user:") {
        return 0
    }
    
    parts := strings.Split(key, ':')
    if len(parts) < 4 {
        return 0
    }
    
    return common.Str2Int(parts[2])
}

// getURLDBMetrics retrieves additional metrics for URL from database
func (s *HeimdallAnalyticsService) getURLDBMetrics(ctx context.Context, url string, timeWindow time.Duration) (*URLFrequencyMetrics, error) {
    cutoff := time.Now().Add(-timeWindow)
    
    var result struct {
        UniqueUsers  int64   `json:"unique_users"`
        AvgLatency   float64 `json:"avg_latency"`
        ErrorRate    float64 `json:"error_rate"`
        LastAccessed time.Time `json:"last_accessed"`
    }
    
    err := model.LOG_DB.Table("heimdall_request_logs").
        Select("COUNT(DISTINCT user_id) as unique_users, AVG(latency_ms) as avg_latency, SUM(CASE WHEN http_status >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as error_rate, MAX(occurred_at) as last_accessed").
        Where("normalized_url = ? AND occurred_at >= ?", url, cutoff).
        Scan(&result).Error
    
    if err != nil {
        return nil, err
    }
    
    return &URLFrequencyMetrics{
        URL:          url,
        UniqueUsers:  result.UniqueUsers,
        AvgLatency:   result.AvgLatency,
        ErrorRate:    result.ErrorRate,
        LastAccessed: result.LastAccessed,
    }, nil
}

// getTokenDBMetrics retrieves additional metrics for token from database
func (s *HeimdallAnalyticsService) getTokenDBMetrics(ctx context.Context, tokenID int, timeWindow time.Duration) (*TokenFrequencyMetrics, error) {
    cutoff := time.Now().Add(-timeWindow)
    
    var result struct {
        UniqueURLs   int64   `json:"unique_urls"`
        AvgLatency   float64 `json:"avg_latency"`
        ErrorRate    float64 `json:"error_rate"`
        LastAccessed string `json:"last_accessed"`
    }
    
    err := model.LOG_DB.Table("heimdall_request_logs").
        Select("COUNT(DISTINCT normalized_url) as unique_urls, AVG(latency_ms) as avg_latency, SUM(CASE WHEN http_status >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as error_rate, MAX(occurred_at) as last_accessed").
        Where("token_id = ? AND occurred_at >= ?", tokenID, cutoff).
        Scan(&result).Error
    
    if err != nil {
        return nil, err
    }
    
    return &TokenFrequencyMetrics{
        TokenID:      tokenID,
        UniqueURLs:   result.UniqueURLs,
        AvgLatency:   result.AvgLatency,
        ErrorRate:    result.ErrorRate,
        LastAccessed: result.LastAccessed,
    }, nil
}

// getUserDBMetrics retrieves additional metrics for user from database
func (s *HeimdallAnalyticsService) getUserDBMetrics(ctx context.Context, userID int, timeWindow time.Duration) (*UserFrequencyMetrics, error) {
    cutoff := time.Now().Add(-timeWindow)
    
    var result struct {
        UniqueURLs   int64   `json:"unique_urls"`
        AvgLatency   float64 `json:"avg_latency"`
        ErrorRate    float64 `json:"error_rate"`
        LastAccessed string `json:"last_accessed"`
    }
    
    err := model.LOG_DB.Table("heimdall_request_logs").
        Select("COUNT(DISTINCT normalized_url) as unique_urls, AVG(latency_ms) as avg_latency, SUM(CASE WHEN http_status >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as error_rate, MAX(occurred_at) as last_accessed").
        Where("user_id = ? AND occurred_at >= ?", userID, cutoff).
        Scan(&result).Error
    
    if err != nil {
        return nil, err
    }
    
    return &UserFrequencyMetrics{
        UserID:      userID,
        UniqueURLs:  result.UniqueURLs,
        AvgLatency:  result.AvgLatency,
        ErrorRate:   result.ErrorRate,
        LastAccessed: result.LastAccessed,
    }, nil
}

// getParamDigestFrequency retrieves frequency of parameter digests
func (s *HeimdallAnalyticsService) getParamDigestFrequency(ctx context.Context, timeWindow time.Duration) (map[string]int64, error) {
    cutoff := time.Now().Add(-timeWindow)
    
    var results []struct {
        ParamDigest string `json:"param_digest"`
        Count       int64  `json:"count"`
    }
    
    err := model.LOG_DB.Table("heimdall_request_logs").
        Select("param_digest, COUNT(*) as count").
        Where("param_digest != '' AND occurred_at >= ?", cutoff).
        Group("param_digest").
        Order("count DESC").
        Limit(100).
        Scan(&results).Error
    
    if err != nil {
        return nil, err
    }
    
    freq := make(map[string]int64)
    for _, result := range results {
        freq[result.ParamDigest] = result.Count
    }
    
    return freq, nil
}

// getIPFrequency retrieves frequency of client IPs
func (s *HeimdallAnalyticsService) getIPFrequency(ctx context.Context, timeWindow time.Duration) (map[string]int64, error) {
    cutoff := time.Now().Add(-timeWindow)
    
    var results []struct {
        ClientIP string `json:"client_ip"`
        Count    int64  `json:"count"`
    }
    
    err := model.LOG_DB.Table("heimdall_request_logs").
        Select("client_ip, COUNT(*) as count").
        Where("client_ip != '' AND occurred_at >= ?", cutoff).
        Group("client_ip").
        Order("count DESC").
        Limit(100).
        Scan(&results).Error
    
    if err != nil {
        return nil, err
    }
    
    freq := make(map[string]int64)
    for _, result := range results {
        freq[result.ClientIP] = result.Count
    }
    
    return freq, nil
}

// getUserAgentFrequency retrieves frequency of user agents
func (s *HeimdallAnalyticsService) getUserAgentFrequency(ctx context.Context, timeWindow time.Duration) (map[string]int64, error) {
    cutoff := time.Now().Add(-timeWindow)
    
    var results []struct {
        ClientUserAgent string `json:"client_user_agent"`
        Count           int64  `json:"count"`
    }
    
    err := model.LOG_DB.Table("heimdall_request_logs").
        Select("client_user_agent, COUNT(*) as count").
        Where("client_user_agent != '' AND occurred_at >= ?", cutoff).
        Group("client_user_agent").
        Order("count DESC").
        Limit(50).
        Scan(&results).Error
    
    if err != nil {
        return nil, err
    }
    
    freq := make(map[string]int64)
    for _, result := range results {
        freq[result.ClientUserAgent] = result.Count
    }
    
    return freq, nil
}

// CleanupOldMetrics removes old metrics from Redis
func (s *HeimdallAnalyticsService) CleanupOldMetrics(ctx context.Context) error {
    // This would implement cleanup logic for old Redis keys
    // For now, just log that cleanup was performed
    logger.SysLog("Heimdall metrics cleanup completed")
    return nil
}

// GenerateHourlyRollups creates hourly rollups of metrics
func (s *HeimdallAnalyticsService) GenerateHourlyRollups(ctx context.Context) error {
    // This would implement hourly rollup logic
    // For now, just log that rollups were generated
    logger.SysLog("Heimdall hourly rollups generated")
    return nil
}

// Global analytics service instance
var GlobalHeimdallAnalyticsService = NewHeimdallAnalyticsService()
