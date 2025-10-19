package config

import "github.com/QuantumNous/new-api/common"

// GovernanceConfig controls request governance and abuse protection parameters.
// These options are feature-gated and disabled by default.
type GovernanceConfig struct {
    Enabled           bool   `json:"enabled"`
    AbuseRPMThreshold int    `json:"abuse_rpm_threshold"`
    RerouteModelAlias string `json:"reroute_model_alias"`
    // FlagTTLHours controls default TTL for RequestFlag rows without an explicit ttl_at.
    FlagTTLHours      int    `json:"flag_ttl_hours"`
}

var governanceConfig = GovernanceConfig{
    Enabled:           common.GetEnvOrDefaultBool("GOVERNANCE_ENABLED", false),
    AbuseRPMThreshold: common.GetEnvOrDefault("GOVERNANCE_ABUSE_RPM_THRESHOLD", 3000),
    RerouteModelAlias: common.GetEnvOrDefaultString("GOVERNANCE_REROUTE_MODEL_ALIAS", ""),
    FlagTTLHours:      common.GetEnvOrDefault("GOVERNANCE_FLAG_TTL_HOURS", 24),
}

func init() {
    GlobalConfig.Register("governance", &governanceConfig)
}

func GetGovernanceConfig() *GovernanceConfig {
    return &governanceConfig
}
