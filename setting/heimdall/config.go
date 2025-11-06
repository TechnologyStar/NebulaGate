package heimdall

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	ListenAddr string
	TLSEnabled bool
	TLSCertPath string
	TLSKeyPath string
	ACMEEnabled bool
	ACMEDomain string
	ACMEEmail string
	ACMECacheDir string

	// API configuration
	BackendURL string
	APIKeyHeader string
	APIKeyValue string

	// Security configuration
	CORSOrigins []string
	RateLimitEnabled bool
	RateLimitRequests int
	RateLimitWindow time.Duration

	// Logging configuration
	LogLevel string
	LogFormat string
}

var GlobalConfig *Config

func LoadConfig() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load(".env")

	config := &Config{
		ListenAddr: getEnvWithDefault("HEIMDALL_LISTEN_ADDR", ":8443"),
		TLSEnabled: getEnvBoolWithDefault("HEIMDALL_TLS_ENABLED", true),
		TLSCertPath: getEnvWithDefault("HEIMDALL_TLS_CERT", ""),
		TLSKeyPath: getEnvWithDefault("HEIMDALL_TLS_KEY", ""),
		ACMEEnabled: getEnvBoolWithDefault("HEIMDALL_ACME_ENABLED", false),
		ACMEDomain: getEnvWithDefault("HEIMDALL_ACME_DOMAIN", ""),
		ACMEEmail: getEnvWithDefault("HEIMDALL_ACME_EMAIL", ""),
		ACMECacheDir: getEnvWithDefault("HEIMDALL_ACME_CACHE_DIR", "/tmp/heimdall-acme"),
		
		BackendURL: getEnvWithDefault("HEIMDALL_BACKEND_URL", "http://localhost:3000"),
		APIKeyHeader: getEnvWithDefault("HEIMDALL_API_KEY_HEADER", "Authorization"),
		APIKeyValue: getEnvWithDefault("HEIMDALL_API_KEY_VALUE", ""),
		
		CORSOrigins: getEnvStringSlice("HEIMDALL_CORS_ORIGINS", []string{"*"}),
		RateLimitEnabled: getEnvBoolWithDefault("HEIMDALL_RATE_LIMIT_ENABLED", false),
		RateLimitRequests: getEnvIntWithDefault("HEIMDALL_RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow: time.Duration(getEnvIntWithDefault("HEIMDALL_RATE_LIMIT_WINDOW_MINUTES", 1)) * time.Minute,
		
		LogLevel: getEnvWithDefault("HEIMDALL_LOG_LEVEL", "info"),
		LogFormat: getEnvWithDefault("HEIMDALL_LOG_FORMAT", "json"),
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	GlobalConfig = config
	return config, nil
}

func validateConfig(config *Config) error {
	// TLS validation
	if config.TLSEnabled {
		if !config.ACMEEnabled {
			if config.TLSCertPath == "" || config.TLSKeyPath == "" {
				return fmt.Errorf("TLS is enabled but no certificate/key paths provided and ACME is disabled")
			}
			
			// Check if cert/key files exist
			if _, err := os.Stat(config.TLSCertPath); os.IsNotExist(err) {
				return fmt.Errorf("TLS certificate file does not exist: %s", config.TLSCertPath)
			}
			if _, err := os.Stat(config.TLSKeyPath); os.IsNotExist(err) {
				return fmt.Errorf("TLS key file does not exist: %s", config.TLSKeyPath)
			}
		} else {
			// ACME validation
			if config.ACMEDomain == "" {
				return fmt.Errorf("ACME is enabled but no domain specified")
			}
			if config.ACMEEmail == "" {
				return fmt.Errorf("ACME is enabled but no email specified")
			}
		}
	} else {
		// If TLS is disabled, we should warn about security
		fmt.Println("WARNING: TLS is disabled. This is not recommended for production.")
	}

	// Backend URL validation
	if config.BackendURL == "" {
		return fmt.Errorf("backend URL is required")
	}

	return nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		return []string{value}
	}
	return defaultValue
}