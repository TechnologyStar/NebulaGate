package middleware

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// HeimdallSettings holds the runtime configuration for Heimdall
var HeimdallSettings HeimdallConfig

// InitHeimdallConfig initializes Heimdall configuration from environment variables
func InitHeimdallConfig() {
	config := DefaultHeimdallConfig()
	
	// Authentication settings
	if val := os.Getenv("HEIMDALL_AUTH_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.AuthEnabled = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_API_KEY_VALIDATION"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.APIKeyValidation = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_JWT_VALIDATION"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.JWTValidation = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_MUTUAL_TLS_VALIDATION"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.MutualTLSValidation = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_JWT_SECRET"); val != "" {
		config.JWTSecret = val
	}
	
	if val := os.Getenv("HEIMDALL_JWT_SIGNING_METHOD"); val != "" {
		config.JWTSigningMethod = val
	}
	
	// Request validation settings
	if val := os.Getenv("HEIMDALL_SCHEMA_VALIDATION"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.SchemaValidation = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_REPLAY_PROTECTION"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.ReplayProtection = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_REPLAY_WINDOW"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ReplayWindow = duration
		}
	}
	
	// Rate limiting settings
	if val := os.Getenv("HEIMDALL_RATE_LIMIT_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.RateLimitEnabled = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_PER_KEY_RATE_LIMIT"); val != "" {
		if limit, err := strconv.Atoi(val); err == nil {
			config.PerKeyRateLimit = limit
		}
	}
	
	if val := os.Getenv("HEIMDALL_PER_IP_RATE_LIMIT"); val != "" {
		if limit, err := strconv.Atoi(val); err == nil {
			config.PerIPRateLimit = limit
		}
	}
	
	if val := os.Getenv("HEIMDALL_RATE_LIMIT_WINDOW"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.RateLimitWindow = duration
		}
	}
	
	// Audit logging settings
	if val := os.Getenv("HEIMDALL_AUDIT_LOGGING_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.AuditLoggingEnabled = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_LOG_PAYLOAD_TRUNCATE"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.LogPayloadTruncate = enabled
		}
	}
	
	if val := os.Getenv("HEIMDALL_MAX_PAYLOAD_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.MaxPayloadSize = size
		}
	}
	
	HeimdallSettings = config
	
	// Log the configuration (without sensitive data)
	logConfig := config
	logConfig.JWTSecret = "***REDACTED***"
	configJSON, _ := json.MarshalIndent(logConfig, "", "  ")
	common.SysLog("Heimdall configuration initialized:")
	common.SysLog(string(configJSON))
}

// GetHeimdallConfig returns the current Heimdall configuration
func GetHeimdallConfig() HeimdallConfig {
	return HeimdallSettings
}

// UpdateHeimdallConfig updates the Heimdall configuration at runtime
func UpdateHeimdallConfig(newConfig HeimdallConfig) {
	HeimdallSettings = newConfig
	common.SysLog("Heimdall configuration updated")
}

// IsHeimdallEnabled returns true if Heimdall authentication is enabled
func IsHeimdallEnabled() bool {
	return HeimdallSettings.AuthEnabled
}