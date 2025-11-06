package securityanalytics

import (
	"fmt"
	"math"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// AnomalyDetector defines the interface for anomaly detection rules
type AnomalyDetector interface {
	Detect(ctx *DetectionContext) (*DetectionResult, error)
	GetRuleType() string
	GetDescription() string
}

// DetectionContext contains user activity data for analysis
type DetectionContext struct {
	UserId                int
	TimeWindow            time.Duration
	WindowStartTime       time.Time
	WindowEndTime         time.Time
	RequestCount          int64
	QuotaUsed             int64
	LoginCount            int64
	UniqueIPs             int64
	UniqueDevices         int64
	DeviceAggregations    []*DeviceAggregationResult
	IPAggregations        []*IPAggregationWindow
	ConversationLinkages  []*ConversationLinkageResult
	LatestBaseline        map[string]*model.AnomalyBaseline
}

// DetectionResult represents the result of an anomaly detection check
type DetectionResult struct {
	Detected      bool
	RuleType      string
	Severity      string
	Message       string
	Evidence      map[string]interface{}
	Threshold     float64
	ActualValue   float64
	Baseline      float64
	Deviation     float64 // percentage
	DeduplicationKey string
}

// QuotaSpikeDetector detects sudden quota/log volume changes without corresponding API calls
type QuotaSpikeDetector struct {
	ThresholdPercentage float64 // default 150% increase
}

func NewQuotaSpikeDetector(threshold float64) *QuotaSpikeDetector {
	if threshold <= 0 {
		threshold = 150.0
	}
	return &QuotaSpikeDetector{
		ThresholdPercentage: threshold,
	}
}

func (d *QuotaSpikeDetector) GetRuleType() string {
	return "quota_spike"
}

func (d *QuotaSpikeDetector) GetDescription() string {
	return "Detects sudden quota consumption without corresponding API requests"
}

func (d *QuotaSpikeDetector) Detect(ctx *DetectionContext) (*DetectionResult, error) {
	// Get baseline for quota usage
	baseline, err := model.GetAnomalyBaseline(ctx.UserId, "quota_usage")
	if err != nil {
		return nil, err
	}

	if baseline == nil || baseline.BaselineValue == 0 {
		// No baseline yet, skip detection
		return &DetectionResult{Detected: false}, nil
	}

	// Calculate expected quota based on request count
	expectedQuotaPerRequest := baseline.BaselineValue / float64(baseline.SampleSize)
	expectedQuota := expectedQuotaPerRequest * float64(ctx.RequestCount)
	tolerance := expectedQuota * (d.ThresholdPercentage / 100.0)

	actual := float64(ctx.QuotaUsed)
	deviation := ((actual - expectedQuota) / expectedQuota) * 100.0

	if actual > expectedQuota+tolerance {
		return &DetectionResult{
			Detected:      true,
			RuleType:      d.GetRuleType(),
			Severity:      calculateSeverity(deviation, 200.0, 400.0), // 200% is medium, 400% is critical
			Message:       fmt.Sprintf("Quota spike detected: used %d, expected ~%.0f", ctx.QuotaUsed, expectedQuota),
			Threshold:     expectedQuota + tolerance,
			ActualValue:   actual,
			Baseline:      expectedQuota,
			Deviation:     deviation,
			Evidence: map[string]interface{}{
				"expected_quota":    expectedQuota,
				"actual_quota":      actual,
				"request_count":     ctx.RequestCount,
				"deviation_percent": deviation,
				"baseline_value":    baseline.BaselineValue,
			},
			DeduplicationKey: fmt.Sprintf("%d_quota_spike_%d", ctx.UserId, ctx.WindowStartTime.Unix()/3600), // Hourly dedup
		}, nil
	}

	return &DetectionResult{Detected: false}, nil
}

// AbnormalLoginRatioDetector detects abnormal login frequency vs API usage ratio
type AbnormalLoginRatioDetector struct {
	MinLoginThreshold float64 // minimum logins to trigger
	MaxRatioThreshold float64 // requests per login
}

func NewAbnormalLoginRatioDetector(maxRatio float64) *AbnormalLoginRatioDetector {
	if maxRatio <= 0 {
		maxRatio = 1000.0
	}
	return &AbnormalLoginRatioDetector{
		MinLoginThreshold: 1,
		MaxRatioThreshold: maxRatio,
	}
}

func (d *AbnormalLoginRatioDetector) GetRuleType() string {
	return "abnormal_login_ratio"
}

func (d *AbnormalLoginRatioDetector) GetDescription() string {
	return "Detects abnormal login frequency relative to API usage"
}

func (d *AbnormalLoginRatioDetector) Detect(ctx *DetectionContext) (*DetectionResult, error) {
	// Skip if few logins
	if ctx.LoginCount < 1 {
		return &DetectionResult{Detected: false}, nil
	}

	ratio := float64(ctx.RequestCount) / float64(ctx.LoginCount)

	// Get baseline
	baseline, err := model.GetAnomalyBaseline(ctx.UserId, "login_ratio")
	if err != nil {
		return nil, err
	}

	var baselineRatio float64 = 100.0 // default: 100 requests per login
	if baseline != nil && baseline.BaselineValue > 0 {
		baselineRatio = baseline.BaselineValue
	}

	threshold := baselineRatio * 2.0 // 2x the baseline
	deviation := ((ratio - baselineRatio) / baselineRatio) * 100.0

	if ratio > threshold {
		return &DetectionResult{
			Detected:      true,
			RuleType:      d.GetRuleType(),
			Severity:      calculateSeverity(deviation, 150.0, 400.0),
			Message:       fmt.Sprintf("Abnormal login ratio: %.0f requests per login (baseline: %.0f)", ratio, baselineRatio),
			Threshold:     threshold,
			ActualValue:   ratio,
			Baseline:      baselineRatio,
			Deviation:     deviation,
			Evidence: map[string]interface{}{
				"request_count":      ctx.RequestCount,
				"login_count":        ctx.LoginCount,
				"actual_ratio":       ratio,
				"baseline_ratio":     baselineRatio,
				"deviation_percent":  deviation,
			},
			DeduplicationKey: fmt.Sprintf("%d_login_ratio_%d", ctx.UserId, ctx.WindowStartTime.Unix()/3600),
		}, nil
	}

	return &DetectionResult{Detected: false}, nil
}

// HighRequestRatioDetector detects unusually high request-to-login ratio
type HighRequestRatioDetector struct {
	Threshold float64 // requests-to-login threshold
}

func NewHighRequestRatioDetector(threshold float64) *HighRequestRatioDetector {
	if threshold <= 0 {
		threshold = 500.0 // default: 500 requests per login
	}
	return &HighRequestRatioDetector{
		Threshold: threshold,
	}
}

func (d *HighRequestRatioDetector) GetRuleType() string {
	return "high_request_ratio"
}

func (d *HighRequestRatioDetector) GetDescription() string {
	return "Detects unusually high request-to-login ratio"
}

func (d *HighRequestRatioDetector) Detect(ctx *DetectionContext) (*DetectionResult, error) {
	if ctx.LoginCount == 0 {
		// No logins, high request count could be suspicious
		if ctx.RequestCount > 1000 {
			return &DetectionResult{
				Detected:      true,
				RuleType:      d.GetRuleType(),
				Severity:      "high",
				Message:       fmt.Sprintf("High request volume without login: %d requests", ctx.RequestCount),
				Threshold:     d.Threshold,
				ActualValue:   float64(ctx.RequestCount),
				Baseline:      0,
				Deviation:     100.0,
				Evidence: map[string]interface{}{
					"request_count": ctx.RequestCount,
					"login_count":   ctx.LoginCount,
				},
				DeduplicationKey: fmt.Sprintf("%d_high_requests_%d", ctx.UserId, ctx.WindowStartTime.Unix()/3600),
			}, nil
		}
		return &DetectionResult{Detected: false}, nil
	}

	ratio := float64(ctx.RequestCount) / float64(ctx.LoginCount)

	if ratio > d.Threshold {
		deviation := ((ratio - d.Threshold) / d.Threshold) * 100.0
		return &DetectionResult{
			Detected:      true,
			RuleType:      d.GetRuleType(),
			Severity:      calculateSeverity(deviation, 100.0, 300.0),
			Message:       fmt.Sprintf("High request-to-login ratio: %.0f requests per login", ratio),
			Threshold:     d.Threshold,
			ActualValue:   ratio,
			Baseline:      d.Threshold,
			Deviation:     deviation,
			Evidence: map[string]interface{}{
				"request_count":     ctx.RequestCount,
				"login_count":       ctx.LoginCount,
				"request_ratio":     ratio,
				"threshold":         d.Threshold,
				"deviation_percent": deviation,
			},
			DeduplicationKey: fmt.Sprintf("%d_high_ratio_%d", ctx.UserId, ctx.WindowStartTime.Unix()/3600),
		}, nil
	}

	return &DetectionResult{Detected: false}, nil
}

// UnusualDeviceActivityDetector detects unusual device activity patterns
type UnusualDeviceActivityDetector struct {
	NewDeviceThreshold int // minimum requests threshold for new device
	IPChangeThreshold  int // maximum IPs per device
}

func NewUnusualDeviceActivityDetector(newDeviceThreshold, ipChangeThreshold int) *UnusualDeviceActivityDetector {
	if newDeviceThreshold <= 0 {
		newDeviceThreshold = 100 // minimum requests from new device
	}
	if ipChangeThreshold <= 0 {
		ipChangeThreshold = 5 // max 5 different IPs per device
	}
	return &UnusualDeviceActivityDetector{
		NewDeviceThreshold: newDeviceThreshold,
		IPChangeThreshold:  ipChangeThreshold,
	}
}

func (d *UnusualDeviceActivityDetector) GetRuleType() string {
	return "unusual_device_activity"
}

func (d *UnusualDeviceActivityDetector) GetDescription() string {
	return "Detects unusual device activity patterns"
}

func (d *UnusualDeviceActivityDetector) Detect(ctx *DetectionContext) (*DetectionResult, error) {
	if len(ctx.DeviceAggregations) == 0 {
		return &DetectionResult{Detected: false}, nil
	}

	for _, device := range ctx.DeviceAggregations {
		// Check for new device with high activity
		if device.FirstSeenAt.After(ctx.WindowStartTime.Add(24 * time.Hour)) && device.RequestCount > int64(d.NewDeviceThreshold) {
			return &DetectionResult{
				Detected:      true,
				RuleType:      d.GetRuleType(),
				Severity:      "medium",
				Message:       fmt.Sprintf("Suspicious activity from new device %s: %d requests", device.NormalizedDeviceId, device.RequestCount),
				Threshold:     float64(d.NewDeviceThreshold),
				ActualValue:   float64(device.RequestCount),
				Baseline:      float64(d.NewDeviceThreshold),
				Deviation:     ((float64(device.RequestCount) - float64(d.NewDeviceThreshold)) / float64(d.NewDeviceThreshold)) * 100.0,
				Evidence: map[string]interface{}{
					"device_id":      device.NormalizedDeviceId,
					"request_count":  device.RequestCount,
					"unique_ips":     len(device.UniqueIPs),
					"ips":            device.UniqueIPs,
					"first_seen":     device.FirstSeenAt,
					"last_seen":      device.LastSeenAt,
				},
				DeduplicationKey: fmt.Sprintf("%d_device_%s_%d", ctx.UserId, device.NormalizedDeviceId, ctx.WindowStartTime.Unix()/3600),
			}, nil
		}

		// Check for device with too many IP changes
		if len(device.UniqueIPs) > d.IPChangeThreshold {
			return &DetectionResult{
				Detected:      true,
				RuleType:      d.GetRuleType(),
				Severity:      "high",
				Message:       fmt.Sprintf("Device %s using too many IPs: %d different IPs", device.NormalizedDeviceId, len(device.UniqueIPs)),
				Threshold:     float64(d.IPChangeThreshold),
				ActualValue:   float64(len(device.UniqueIPs)),
				Baseline:      float64(d.IPChangeThreshold),
				Deviation:     ((float64(len(device.UniqueIPs)) - float64(d.IPChangeThreshold)) / float64(d.IPChangeThreshold)) * 100.0,
				Evidence: map[string]interface{}{
					"device_id":    device.NormalizedDeviceId,
					"ip_count":     len(device.UniqueIPs),
					"ips":          device.UniqueIPs,
					"request_count": device.RequestCount,
				},
				DeduplicationKey: fmt.Sprintf("%d_device_multi_ip_%s_%d", ctx.UserId, device.NormalizedDeviceId, ctx.WindowStartTime.Unix()/3600),
			}, nil
		}
	}

	return &DetectionResult{Detected: false}, nil
}

// Helper functions

func calculateSeverity(deviation, mediumThreshold, criticalThreshold float64) string {
	if math.IsNaN(deviation) || math.IsInf(deviation, 0) {
		return "medium"
	}
	
	if deviation >= criticalThreshold {
		return "critical"
	}
	if deviation >= mediumThreshold {
		return "high"
	}
	if deviation >= 100.0 {
		return "medium"
	}
	return "low"
}

// GetConfiguredThreshold retrieves a threshold from options or returns default
func GetConfiguredThreshold(key string, defaultValue float64) float64 {
	value := model.GetOptionValue("anomaly_" + key)
	if value == "" {
		return defaultValue
	}
	
	var result float64
	_, err := fmt.Sscanf(value, "%f", &result)
	if err != nil {
		return defaultValue
	}
	return result
}
