package config

import "github.com/QuantumNous/new-api/common"

// PublicLogsConfig controls public logging/leaderboard related switches.
// These options are placeholders for upcoming features and default to disabled.
type PublicLogsConfig struct {
	Enabled          bool   `json:"enabled"`
	PublicVisibility string `json:"public_visibility"`
	RetentionDays    int    `json:"retention_days"`
}

var publicLogsConfig = PublicLogsConfig{
	Enabled:          common.GetEnvOrDefaultBool("PUBLIC_LOGS_ENABLED", false),
	PublicVisibility: common.GetEnvOrDefaultString("PUBLIC_LOGS_VISIBILITY", "anonymous"),
	RetentionDays:    common.GetEnvOrDefault("PUBLIC_LOGS_RETENTION_DAYS", 3),
}

func init() {
	GlobalConfig.Register("public_logs", &publicLogsConfig)
}

func GetPublicLogsConfig() *PublicLogsConfig {
	return &publicLogsConfig
}
