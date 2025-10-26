package dto

// BillingFeatureConfig describes runtime billing feature flags exposed via the API.
type BillingFeatureConfig struct {
	Enabled       bool   `json:"enabled"`
	DefaultMode   string `json:"defaultMode,omitempty"`
	AutoFallback  bool   `json:"autoFallback,omitempty"`
	ResetHourUTC  int    `json:"resetHourUtc,omitempty"`
	ResetTimezone string `json:"resetTimezone,omitempty"`
	NotifyOnReset bool   `json:"notifyOnReset,omitempty"`
}

// BillingFeatureUpdate represents a partial update payload for billing config.
type BillingFeatureUpdate struct {
	Enabled       *bool   `json:"enabled"`
	DefaultMode   *string `json:"defaultMode"`
	AutoFallback  *bool   `json:"autoFallback"`
	ResetHourUTC  *int    `json:"resetHourUtc"`
	ResetTimezone *string `json:"resetTimezone"`
	NotifyOnReset *bool   `json:"notifyOnReset"`
}

// GovernanceFeatureConfig describes governance feature flags for the UI.
type GovernanceFeatureConfig struct {
	Enabled           bool   `json:"enabled"`
	AbuseRPMThreshold int    `json:"abuseRpmThreshold,omitempty"`
	RerouteModelAlias string `json:"rerouteModelAlias,omitempty"`
	FlagTTLHours      int    `json:"flagTtlHours,omitempty"`
}

// GovernanceFeatureUpdate represents governance config partial updates.
type GovernanceFeatureUpdate struct {
	Enabled           *bool   `json:"enabled"`
	AbuseRPMThreshold *int    `json:"abuseRpmThreshold"`
	RerouteModelAlias *string `json:"rerouteModelAlias"`
	FlagTTLHours      *int    `json:"flagTtlHours"`
}

// PublicLogsFeatureConfig exposes public log feature switches to clients.
type PublicLogsFeatureConfig struct {
	Enabled          bool   `json:"enabled"`
	PublicVisibility string `json:"publicVisibility,omitempty"`
	RetentionDays    int    `json:"retention_days,omitempty"`
}

// PublicLogsFeatureUpdate represents partial updates for public logs config.
type PublicLogsFeatureUpdate struct {
	Enabled          *bool   `json:"enabled"`
	PublicVisibility *string `json:"publicVisibility"`
	RetentionDays    *int    `json:"retention_days"`
}

// FeatureConfigResponse groups the available feature flag configurations.
type FeatureConfigResponse struct {
	Billing    *BillingFeatureConfig    `json:"billing,omitempty"`
	Governance *GovernanceFeatureConfig `json:"governance,omitempty"`
	PublicLogs *PublicLogsFeatureConfig `json:"public_logs,omitempty"`
}

// FeatureConfigUpdateRequest carries partial updates for feature flags.
type FeatureConfigUpdateRequest struct {
	Billing    *BillingFeatureUpdate    `json:"billing"`
	Governance *GovernanceFeatureUpdate `json:"governance"`
	PublicLogs *PublicLogsFeatureUpdate `json:"public_logs"`
}
