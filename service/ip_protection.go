package service

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// CheckIPProtection checks if an IP should be blocked
// Returns: allowed (bool), reason (string), error
func CheckIPProtection(ip string) (bool, string, error) {
	// Check if IP is banned
	isBanned, ban, err := model.IsIPBanned(ip)
	if err != nil {
		return true, "", err // Allow on error to avoid blocking legitimate users
	}
	if isBanned {
		reason := fmt.Sprintf("IP banned: %s", ban.BanReason)
		if ban.BanType == "temporary" && ban.ExpiresAt != nil {
			reason += fmt.Sprintf(" (expires at %s)", ban.ExpiresAt.Format(time.RFC3339))
		}
		return false, reason, nil
	}

	// Check blacklist
	isBlacklisted, blacklistEntry, err := model.IsIPInList(ip, "blacklist")
	if err != nil {
		return true, "", err
	}
	if isBlacklisted {
		reason := fmt.Sprintf("IP in blacklist: %s", blacklistEntry.Reason)
		return false, reason, nil
	}

	// Check whitelist - if in whitelist, skip rate limiting
	isWhitelisted, _, err := model.IsIPInList(ip, "whitelist")
	if err != nil {
		return true, "", err
	}
	if isWhitelisted {
		return true, "whitelisted", nil
	}

	return true, "", nil
}

// CheckIPRateLimit checks if an IP has exceeded rate limits
// Returns: allowed (bool), remaining requests, reset time, error
func CheckIPRateLimit(ip string) (bool, int, time.Time, error) {
	// Get rate limit rule for this IP
	limit, err := model.GetIPRateLimitForIP(ip)
	if err != nil {
		return true, 0, time.Time{}, err
	}
	if limit == nil {
		// No specific limit for this IP
		return true, 0, time.Time{}, nil
	}

	if !common.RedisEnabled {
		// Redis not enabled, can't enforce rate limit
		return true, 0, time.Time{}, nil
	}

	key := fmt.Sprintf("rate_limit:ip:%s", ip)
	count, err := common.RedisGet(key)
	if err != nil {
		// Redis error, allow request
		return true, 0, time.Time{}, nil
	}

	var currentCount int
	if count == "" {
		currentCount = 0
	} else {
		fmt.Sscanf(count, "%d", &currentCount)
	}

	if currentCount >= limit.MaxRequests {
		// Rate limit exceeded
		ttl, _ := common.RedisTTL(key)
		resetTime := time.Now().Add(time.Duration(ttl) * time.Second)
		return false, 0, resetTime, nil
	}

	// Increment counter
	if currentCount == 0 {
		// First request in window
		err = common.RedisSet(key, "1", time.Duration(limit.TimeWindow)*time.Second)
	} else {
		newCount := currentCount + 1
		err = common.RedisSet(key, fmt.Sprintf("%d", newCount), -1) // Keep existing TTL
	}

	if err != nil {
		return true, 0, time.Time{}, err
	}

	remaining := limit.MaxRequests - (currentCount + 1)
	return true, remaining, time.Now().Add(time.Duration(limit.TimeWindow) * time.Second), nil
}

// RecordIPViolation records an IP violation and may trigger auto-ban
func RecordIPViolation(ip string, reason string, userId int) error {
	if !common.RedisEnabled {
		return nil
	}

	key := fmt.Sprintf("ip_violations:%s", ip)
	count, err := common.RedisGet(key)
	if err != nil {
		count = "0"
	}

	var violationCount int
	fmt.Sscanf(count, "%d", &violationCount)
	violationCount++

	// Store violation count for 1 hour
	err = common.RedisSet(key, fmt.Sprintf("%d", violationCount), time.Hour)
	if err != nil {
		return err
	}

	// Auto-ban thresholds
	var banDuration time.Duration
	var banReason string

	if violationCount >= 100 {
		// Permanent ban
		banReason = fmt.Sprintf("Auto-ban: %d violations (%s)", violationCount, reason)
		ban := &model.IPBan{
			IP:        ip,
			BanReason: banReason,
			BanType:   "permanent",
			BannedAt:  time.Now(),
			BannedBy:  0, // System auto-ban
		}
		return model.BanIP(ban)
	} else if violationCount >= 50 {
		// 24 hour ban
		banDuration = 24 * time.Hour
		banReason = fmt.Sprintf("Auto-ban: %d violations (%s)", violationCount, reason)
	} else if violationCount >= 20 {
		// 1 hour ban
		banDuration = time.Hour
		banReason = fmt.Sprintf("Auto-ban: %d violations (%s)", violationCount, reason)
	} else {
		// No ban yet
		return nil
	}

	expiresAt := time.Now().Add(banDuration)
	ban := &model.IPBan{
		IP:        ip,
		BanReason: banReason,
		BanType:   "temporary",
		BannedAt:  time.Now(),
		ExpiresAt: &expiresAt,
		BannedBy:  0, // System auto-ban
	}
	return model.BanIP(ban)
}

// IsValidIP checks if a string is a valid IP address or CIDR
func IsValidIP(ipStr string) bool {
	// Check if it's a CIDR
	if strings.Contains(ipStr, "/") {
		_, _, err := net.ParseCIDR(ipStr)
		return err == nil
	}
	// Check if it's a plain IP
	ip := net.ParseIP(ipStr)
	return ip != nil
}

// MatchIPPattern checks if an IP matches a pattern (supports CIDR)
func MatchIPPattern(ip string, pattern string) bool {
	if pattern == ip {
		return true
	}

	// Check CIDR match
	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		ipAddr := net.ParseIP(ip)
		if ipAddr == nil {
			return false
		}
		return ipNet.Contains(ipAddr)
	}

	return false
}

// GetIPStatistics gets statistics about IP usage
func GetIPStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count active bans
	bans, _, err := model.GetIPBans(1, 1000, false)
	if err != nil {
		return nil, err
	}
	stats["active_bans"] = len(bans)

	// Count blacklist and whitelist entries
	blacklist, blacklistTotal, _ := model.GetIPLists("blacklist", 1, 1)
	stats["blacklist_count"] = blacklistTotal

	whitelist, whitelistTotal, _ := model.GetIPLists("whitelist", 1, 1)
	stats["whitelist_count"] = whitelistTotal

	// Count rate limit rules
	limits, limitsTotal, _ := model.GetIPRateLimits(1, 1)
	stats["rate_limit_rules"] = limitsTotal

	// Additional statistics
	stats["blacklist"] = blacklist
	stats["whitelist"] = whitelist
	stats["limits"] = limits

	return stats, nil
}
