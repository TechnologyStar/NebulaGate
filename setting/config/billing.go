package config

import "github.com/QuantumNous/new-api/common"

// BillingConfig controls feature flags and defaults for upcoming billing features.
// These options are feature-gated and are NOT active unless explicitly enabled.
type BillingConfig struct {
    Enabled      bool   `json:"enabled"`
    DefaultMode  string `json:"default_mode"`
    AutoFallback bool   `json:"auto_fallback"`
}

var billingConfig = BillingConfig{
    Enabled:      common.GetEnvOrDefaultBool("BILLING_ENABLED", false),
    // default_mode: the default account charging mode for runtime when feature is enabled
    // balance | prepaid | voucher | plan | fallback (placeholder, subject to change)
    DefaultMode:  common.GetEnvOrDefaultString("BILLING_DEFAULT_MODE", "balance"),
    AutoFallback: common.GetEnvOrDefaultBool("BILLING_AUTO_FALLBACK", false),
}

func init() {
    // Register under the key "billing" so it appears as billing.* in options
    GlobalConfig.Register("billing", &billingConfig)
}

func GetBillingConfig() *BillingConfig {
    return &billingConfig
}
