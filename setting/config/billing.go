package config

import "github.com/QuantumNous/new-api/common"

// BillingConfig controls feature flags and defaults for upcoming billing features.
// These options are feature-gated and are NOT active unless explicitly enabled.
type BillingConfig struct {
    Enabled       bool   `json:"enabled"`
    DefaultMode   string `json:"default_mode"`
    AutoFallback  bool   `json:"auto_fallback"`
    // ResetHourUTC optionally forces daily/monthly reset checks to run at a given hour (0-23) in UTC.
    // If negative, scheduler uses server-local midnight by default.
    ResetHourUTC  int    `json:"reset_hour_utc"`
    // ResetTimezone is a TZ database name like "America/Los_Angeles"; if set and valid,
    // it is used together with ResetHourUTC to schedule resets in that zone.
    ResetTimezone string `json:"reset_timezone"`
}

var billingConfig = BillingConfig{
    Enabled:       common.GetEnvOrDefaultBool("BILLING_ENABLED", false),
    // default_mode: the default account charging mode for runtime when feature is enabled
    // balance | prepaid | voucher | plan | fallback (placeholder, subject to change)
    DefaultMode:   common.GetEnvOrDefaultString("BILLING_DEFAULT_MODE", "balance"),
    AutoFallback:  common.GetEnvOrDefaultBool("BILLING_AUTO_FALLBACK", false),
    ResetHourUTC:  common.GetEnvOrDefault("BILLING_RESET_HOUR_UTC", -1),
    ResetTimezone: common.GetEnvOrDefaultString("BILLING_RESET_TIMEZONE", ""),
}

func init() {
    // Register under the key "billing" so it appears as billing.* in options
    GlobalConfig.Register("billing", &billingConfig)
}

func GetBillingConfig() *BillingConfig {
    return &billingConfig
}
