package dto

// GovernanceSummary describes a quick snapshot of governance feature flags
// and enforcement thresholds to be surfaced in admin APIs.
type GovernanceSummary struct {
	Enabled           bool   `json:"enabled"`
	AbuseRPMThreshold int    `json:"abuse_rpm_threshold"`
	RerouteModelAlias string `json:"reroute_model_alias,omitempty"`
}

// DetectorThreshold is a generic shape for detector name and threshold.
type DetectorThreshold struct {
	Name      string  `json:"name"`
	Threshold float64 `json:"threshold"`
}

// GovernanceStatusResponse bundles summary and detectors for UI rendering.
type GovernanceStatusResponse struct {
	Summary   GovernanceSummary   `json:"summary"`
	Detectors []DetectorThreshold `json:"detectors,omitempty"`
}
