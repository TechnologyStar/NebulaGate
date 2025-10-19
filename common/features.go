package common

// Global runtime flags exposed for billing, governance and public logs features.
// These are updated by the option/config manager and can also be overridden via env.

var (
	// Billing
	BillingFeatureEnabled bool   = false
	BillingDefaultMode    string = BillingModeBalance

	// Governance
	GovernanceFeatureEnabled  bool   = false
	GovernanceAbuseRPMThreshold int    = 3000
	GovernanceRerouteModelAlias string = ""

	// Public Logs
	PublicLogsFeatureEnabled bool   = false
	PublicLogsVisibility     string = "anonymous"
	PublicLogsRetentionDays  int    = 3
)
