package middleware

import (
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// IPProtection middleware checks IP against blacklist, whitelist, and bans
func IPProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := GetClientIP(c)
		if ip == "" {
			c.Next()
			return
		}

		allowed, reason, err := service.CheckIPProtection(ip)
		if err != nil {
			// Log error but don't block on error
			c.Next()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"message": reason,
					"type":    "ip_blocked",
					"code":    "ip_protection_violation",
				},
			})
			c.Abort()
			return
		}

		// If whitelisted, skip rate limiting by setting a flag
		if reason == "whitelisted" {
			c.Set("ip_whitelisted", true)
		}

		c.Next()
	}
}

// IPRateLimit middleware enforces IP-based rate limiting
func IPRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if IP is whitelisted
		if whitelisted, exists := c.Get("ip_whitelisted"); exists && whitelisted.(bool) {
			c.Next()
			return
		}

		ip := GetClientIP(c)
		if ip == "" {
			c.Next()
			return
		}

		allowed, remaining, resetTime, err := service.CheckIPRateLimit(ip)
		if err != nil {
			// Log error but don't block on error
			c.Next()
			return
		}

		// Set rate limit headers
		if remaining > 0 {
			c.Header("X-RateLimit-Remaining", string(rune(remaining)))
			c.Header("X-RateLimit-Reset", resetTime.Format("2006-01-02T15:04:05Z07:00"))
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"message": "Rate limit exceeded. Please try again later.",
					"type":    "rate_limit_error",
					"code":    "rate_limit_exceeded",
					"reset_at": resetTime.Format("2006-01-02T15:04:05Z07:00"),
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetClientIP extracts the real client IP from the request
// Handles X-Forwarded-For, X-Real-IP headers and supports proxy scenarios
func GetClientIP(c *gin.Context) string {
	// Try X-Forwarded-For first (standard proxy header)
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Try X-Real-IP (common in nginx)
	xri := c.GetHeader("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Try CF-Connecting-IP (Cloudflare)
	cfip := c.GetHeader("CF-Connecting-IP")
	if cfip != "" {
		return strings.TrimSpace(cfip)
	}

	// Fallback to remote address
	ip := c.ClientIP()
	return ip
}
