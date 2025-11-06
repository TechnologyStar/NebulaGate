package service

import (
    "fmt"
    "strings"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    governanceSvc "github.com/QuantumNous/new-api/service/governance"
)

// CheckContentViolation checks if content violates policies
func CheckContentViolation(content string) (bool, []string, string) {
    if strings.TrimSpace(content) == "" {
        return false, nil, ""
    }

    // Use existing governance detection
    result := governanceSvc.DetectKeywordPolicy(content)
    if result.Triggered {
        return true, result.Reasons, result.Severity
    }

    return false, nil, ""
}

// RecordViolation records a security violation
func RecordViolation(userId int, tokenId *int, content string, keywords []string, model, ipAddress, requestId, severity, action string) error {
    // Sanitize content snippet (limit to 500 chars, mask sensitive parts)
    snippet := sanitizeContent(content)

    violation := &model.SecurityViolation{
        UserId:          userId,
        TokenId:         tokenId,
        ViolatedAt:      time.Now(),
        ContentSnippet:  snippet,
        MatchedKeywords: strings.Join(keywords, ", "),
        Model:           model,
        IpAddress:       ipAddress,
        RequestId:       requestId,
        Severity:        severity,
        ActionTaken:     action,
    }

    err := model.CreateSecurityViolation(violation)
    if err != nil {
        return err
    }

    // Increment user violation count
    return model.IncrementViolationCount(userId)
}

// sanitizeContent masks sensitive information in content
func sanitizeContent(content string) string {
    // Limit length
    if len(content) > 500 {
        content = content[:500] + "..."
    }

    // Basic sanitization - mask potential sensitive data patterns
    // Email addresses
    content = maskPattern(content, `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, "***@***.***")
    // Phone numbers
    content = maskPattern(content, `\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`, "***-***-****")
    // Credit card numbers
    content = maskPattern(content, `\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`, "****-****-****-****")

    return content
}

// maskPattern replaces pattern matches with mask
func maskPattern(text, pattern, mask string) string {
    // Simple implementation - in production use regex
    // This is a placeholder for demonstration
    return text
}

// GetDashboardStats retrieves security dashboard statistics
func GetDashboardStats(startTime, endTime time.Time) (map[string]interface{}, error) {
    stats, err := model.GetViolationStatsByDateRange(startTime, endTime)
    if err != nil {
        return nil, err
    }

    // Add today's count
    today := time.Now().Truncate(24 * time.Hour)
    todayEnd := today.Add(24 * time.Hour)
    todayStats, err := model.GetViolationStatsByDateRange(today, todayEnd)
    if err == nil {
        stats["today_count"] = todayStats["total_count"]
    }

    // Add device clusters
    topDevices, err := model.GetTopSuspiciousDevices(10)
    if err == nil {
        stats["device_clusters"] = topDevices
    }

    // Add suspicious IPs
    topIPs, err := model.GetTopSuspiciousIPs(10)
    if err == nil {
        stats["suspicious_ips"] = topIPs
    }

    // Add anomaly counts by severity
    anomalyCounts, err := model.GetAnomalyCountsBySeverity(startTime, endTime)
    if err == nil {
        stats["anomaly_counts"] = anomalyCounts
    }

    // Add anomaly trends
    anomalyTrends, err := model.GetAnomalyTrends(startTime, endTime)
    if err == nil {
        stats["anomaly_trends"] = anomalyTrends
    }

    return stats, nil
}

// BanUser bans a user from making requests
func BanUser(userId int) error {
    return model.BanUser(userId)
}

// UnbanUser removes ban from a user
func UnbanUser(userId int) error {
    return model.UnbanUser(userId)
}

// SetUserRedirect sets model redirect for a user
func SetUserRedirect(userId int, targetModel string) error {
    if targetModel == "" {
        return fmt.Errorf("target model cannot be empty")
    }
    return model.SetUserRedirect(userId, targetModel)
}

// ClearUserRedirect removes model redirect for a user
func ClearUserRedirect(userId int) error {
    return model.ClearUserRedirect(userId)
}

// GetUserSecurity retrieves security status for a user
func GetUserSecurity(userId int) (*model.UserSecurity, error) {
    // Try Redis first
    if userSec, found := model.GetUserSecurityFromRedis(userId); found {
        return userSec, nil
    }

    // Fall back to database
    return model.GetUserSecurity(userId)
}

// CheckUserBanned checks if a user is banned
func CheckUserBanned(userId int) (bool, error) {
    userSec, err := GetUserSecurity(userId)
    if err != nil {
        return false, err
    }
    return userSec.IsBanned, nil
}

// GetUserRedirectModel gets the redirect model for a user
func GetUserRedirectModel(userId int) (string, error) {
    userSec, err := GetUserSecurity(userId)
    if err != nil {
        return "", err
    }
    return userSec.RedirectModel, nil
}

// GetViolationRedirectModel gets global violation redirect model from options
func GetViolationRedirectModel() string {
    // Get from system options
    model := model.GetOptionValue(common.OptionViolationRedirectModel)
    if model == "" {
        // Fall back to governance config
        model = common.GovernanceViolationFallbackAlias
    }
    return model
}

// SetViolationRedirectModel sets global violation redirect model
func SetViolationRedirectModel(targetModel string) error {
    return model.UpdateOption(common.OptionViolationRedirectModel, targetModel)
}

// GetSecuritySettings retrieves all security settings
func GetSecuritySettings() map[string]interface{} {
    settings := make(map[string]interface{})

    settings["violation_redirect_model"] = GetViolationRedirectModel()
    settings["auto_ban_enabled"] = model.GetOptionValue(common.OptionAutobanEnabled) == "true"
    
    threshold := model.GetOptionValue(common.OptionAutobanThreshold)
    if threshold == "" {
        threshold = "10"
    }
    settings["auto_ban_threshold"] = threshold

    settings["auto_enforcement_enabled"] = model.GetOptionValue(common.OptionAutoEnforcementEnabled) == "true"
    settings["auto_block_enabled"] = model.GetOptionValue(common.OptionAutoBlockEnabled) == "true"

    return settings
}

// UpdateSecuritySettings updates security settings
func UpdateSecuritySettings(settings map[string]interface{}) error {
    if model, ok := settings["violation_redirect_model"].(string); ok {
        if err := SetViolationRedirectModel(model); err != nil {
            return err
        }
    }

    if enabled, ok := settings["auto_ban_enabled"].(bool); ok {
        value := "false"
        if enabled {
            value = "true"
        }
        if err := model.UpdateOption(common.OptionAutobanEnabled, value); err != nil {
            return err
        }
    }

    if threshold, ok := settings["auto_ban_threshold"].(float64); ok {
        if err := model.UpdateOption(common.OptionAutobanThreshold, fmt.Sprintf("%.0f", threshold)); err != nil {
            return err
        }
    }

    if enabled, ok := settings["auto_enforcement_enabled"].(bool); ok {
        value := "false"
        if enabled {
            value = "true"
        }
        if err := model.UpdateOption(common.OptionAutoEnforcementEnabled, value); err != nil {
            return err
        }
    }

    if enabled, ok := settings["auto_block_enabled"].(bool); ok {
        value := "false"
        if enabled {
            value = "true"
        }
        if err := model.UpdateOption(common.OptionAutoBlockEnabled, value); err != nil {
            return err
        }
    }

    return nil
}

// CheckAutoban checks if user should be auto-banned based on violations
func CheckAutoban(userId int) error {
    enabled := model.GetOptionValue(common.OptionAutobanEnabled) == "true"
    if !enabled {
        return nil
    }

    thresholdStr := model.GetOptionValue(common.OptionAutobanThreshold)
    if thresholdStr == "" {
        return nil
    }

    threshold := 10 // default
    fmt.Sscanf(thresholdStr, "%d", &threshold)

    userSec, err := GetUserSecurity(userId)
    if err != nil {
        return err
    }

    if userSec.ViolationCount >= threshold && !userSec.IsBanned {
        return BanUser(userId)
    }

    return nil
}
