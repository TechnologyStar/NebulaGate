package common

// Global runtime flags exposed for billing, governance and public logs features.
// These are updated by the option/config manager and can also be overridden via env.

var (
    // Billing
    BillingFeatureEnabled     bool   = true
    BillingDefaultMode        string = BillingModeBalance
    BillingAutoFallbackEnabled bool  = false

    // Governance
    GovernanceFeatureEnabled  bool   = true
    GovernanceAbuseRPMThreshold int    = 3000
    GovernanceRerouteModelAlias string = ""

    // Public Logs
    PublicLogsFeatureEnabled bool   = false
    PublicLogsVisibility     string = "anonymous"
    PublicLogsRetentionDays  int    = 3
)
