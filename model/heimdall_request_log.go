package model

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// HeimdallRequestLog represents telemetry data for Heimdall request tracking
type HeimdallRequestLog struct {
	Id                   int       `json:"id" gorm:"primaryKey;autoIncrement"`
	RequestId            string    `json:"request_id" gorm:"size:64;not null;uniqueIndex"`
	OccurredAt           time.Time `json:"occurred_at" gorm:"not null;index"`
	
	// Authorization & Authentication
	AuthKeyFingerprint   string    `json:"auth_key_fingerprint" gorm:"size:128;index"`
	UserId               *int      `json:"user_id" gorm:"index"`
	TokenId              *int      `json:"token_id" gorm:"index"`
	
	// Request Metadata
	NormalizedURL        string    `json:"normalized_url" gorm:"size:512;index"`
	HTTPMethod           string    `json:"http_method" gorm:"size:16;index"`
	HTTPStatus           int       `json:"http_status" gorm:"index"`
	LatencyMs            int64     `json:"latency_ms" gorm:"index"`
	
	// Client Information
	ClientIP             string    `json:"client_ip" gorm:"size:64;index"`
	ClientUserAgent      string    `json:"client_user_agent" gorm:"size:512"`
	ClientDeviceId       string    `json:"client_device_id" gorm:"size:128;index"`
	
	// Request Characteristics
	RequestSizeBytes     int64     `json:"request_size_bytes"`
	ResponseSizeBytes    int64     `json:"response_size_bytes"`
	ParamDigest          string    `json:"param_digest" gorm:"size:128;index"`
	SanitizedCookies     string    `json:"sanitized_cookies" gorm:"type:text"`
	SanitizedLoginInfo   string    `json:"sanitized_login_info" gorm:"type:text"`
	
	// Geolocation (if enabled)
	CountryCode          string    `json:"country_code" gorm:"size:8;index"`
	Region               string    `json:"region" gorm:"size:64"`
	City                 string    `json:"city" gorm:"size:128"`
	
	// Processing metadata
	ProcessingTimeMs     int64     `json:"processing_time_ms"`
	UpstreamProvider     string    `json:"upstream_provider" gorm:"size:128;index"`
	ModelName            string    `json:"model_name" gorm:"size:128;index"`
	
	// Error information (if any)
	ErrorMessage         string    `json:"error_message" gorm:"type:text"`
	ErrorType            string    `json:"error_type" gorm:"size:64;index"`
	
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// BeforeCreate hook to set default values
func (log *HeimdallRequestLog) BeforeCreate(tx *gorm.DB) error {
	if log.OccurredAt.IsZero() {
		log.OccurredAt = time.Now().UTC()
	}
	return nil
}

// TableName returns the table name for HeimdallRequestLog
func (HeimdallRequestLog) TableName() string {
	return "heimdall_request_logs"
}

// AllowedHeaders defines the whitelist of headers to extract
var AllowedHeaders = map[string]bool{
	"x-forwarded-for":      true,
	"forwarded":            true,
	"client-host":          true,
	"user-agent":           true,
	"x-device-id":          true,
	"x-real-ip":           true,
	"cf-connecting-ip":    true,
	"x-forwarded-host":    true,
	"x-original-uri":      true,
	"accept":              true,
	"accept-language":     true,
	"content-type":        true,
	"content-length":      true,
}

// ExtractClientMetadata extracts and validates client metadata from HTTP headers
func ExtractClientMetadata(headers http.Header) (metadata map[string]string, err error) {
	metadata = make(map[string]string)
	
	// Extract and normalize IP addresses
	if ip := extractIPFromHeaders(headers); ip != "" {
		metadata["client_ip"] = ip
	}
	
	// Extract other allowed headers
	for headerName := range headers {
		lowerName := strings.ToLower(headerName)
		if AllowedHeaders[lowerName] {
			values := headers[headerName]
			if len(values) > 0 {
				metadata[lowerName] = values[0] // Take first value
			}
		}
	}
	
	// Normalize user agent
	if userAgent, exists := metadata["user-agent"]; exists {
		metadata["user_agent"] = sanitizeUserAgent(userAgent)
	}
	
	return metadata, nil
}

// extractIPFromHeaders extracts client IP from various headers with validation
func extractIPFromHeaders(headers http.Header) string {
	// Try different headers in order of preference
	ipHeaders := []string{
		"x-forwarded-for",
		"x-real-ip",
		"cf-connecting-ip",
		"client-host",
	}
	
	for _, headerName := range ipHeaders {
		if values := headers[headerName]; len(values) > 0 {
			ip := extractFirstIP(values[0])
			if ip != "" && isValidIP(ip) {
				return ip
			}
		}
	}
	
	return ""
}

// extractFirstIP extracts the first IP from a comma-separated list
func extractFirstIP(ipList string) string {
	ips := strings.Split(ipList, ",")
	if len(ips) > 0 {
		return strings.TrimSpace(ips[0])
	}
	return ""
}

// isValidIP validates if the IP address is valid and not private (unless configured)
func isValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	
	// Check for private IPs - you might want to allow these based on config
	if !isPrivateIP(parsed) {
		return true
	}
	
	// For now, allow private IPs as they might be legitimate in internal networks
	return true
}

// isPrivateIP checks if an IP is private
func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
	}
	
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	
	return false
}

// sanitizeUserAgent sanitizes user agent string to remove potential attacks
func sanitizeUserAgent(userAgent string) string {
	// Remove potential XSS characters
	sanitized := strings.ReplaceAll(userAgent, "<", "&lt;")
	sanitized = strings.ReplaceAll(sanitized, ">", "&gt;")
	sanitized = strings.ReplaceAll(sanitized, "\"", "&quot;")
	sanitized = strings.ReplaceAll(sanitized, "'", "&#x27;")
	
	// Limit length
	if len(sanitized) > 512 {
		sanitized = sanitized[:512]
	}
	
	return sanitized
}

// CreateParamDigest creates a hash of request parameters for anomaly detection
func CreateParamDigest(params map[string]interface{}) string {
	if len(params) == 0 {
		return ""
	}
	
	// Sort keys to ensure consistent hashing
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Create a deterministic string representation
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteString("=")
		value := params[key]
		if str, ok := value.(string); ok {
			// Truncate long strings and hash them
			if len(str) > 100 {
				hash := sha256.Sum256([]byte(str))
				builder.WriteString(hex.EncodeToString(hash[:8]))
			} else {
				builder.WriteString(str)
			}
		} else {
			builder.WriteString(common.GetJsonString(value))
		}
		builder.WriteString(";")
	}
	
	// Create final hash
	hash := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(hash[:16]) // Use first 16 characters for indexing
}

// SanitizeCookies removes sensitive information from cookies
func SanitizeCookies(cookies string) string {
	if cookies == "" {
		return ""
	}
	
	// List of sensitive cookie names to redact
	sensitivePatterns := []string{
		`(?i)session`,
		`(?i)token`,
		`(?i)auth`,
		`(?i)jwt`,
		`(?i)csrf`,
		`(?i)sess`,
		`(?i)password`,
	}
	
	sanitized := cookies
	for _, pattern := range sensitivePatterns {
		// Replace cookie values for sensitive cookies
		re := regexp.MustCompile(`(` + pattern + `[^=]*)=([^;]*)`)
		sanitized = re.ReplaceAllString(sanitized, `${1}=***`)
	}
	
	return sanitized
}

// NormalizeURL normalizes URL for consistent logging
func NormalizeURL(url, method string) string {
	// Remove query parameters for GET requests
	if method == "GET" {
		if idx := strings.Index(url, "?"); idx != -1 {
			url = url[:idx]
		}
	}
	
	// Normalize path separators
	url = strings.ReplaceAll(url, "//", "/")
	
	// Remove trailing slash unless it's the root
	if len(url) > 1 && strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	
	return url
}

// CreateAuthKeyFingerprint creates a fingerprint of the authorization key
func CreateAuthKeyFingerprint(authKey string) string {
	if authKey == "" {
		return ""
	}
	
	// Create a hash of the auth key for identification
	hash := sha256.Sum256([]byte(authKey))
	return hex.EncodeToString(hash[:16])
}
