package securityanalytics

import (
	"fmt"
	"testing"
	"time"
)

func TestQuotaSpikeDetector(t *testing.T) {
	detector := NewQuotaSpikeDetector(150.0)

	if detector.GetRuleType() != "quota_spike" {
		t.Errorf("expected rule type 'quota_spike', got %s", detector.GetRuleType())
	}

	// Test case: Normal activity (no anomaly)
	ctx := &DetectionContext{
		UserId:           1,
		RequestCount:     100,
		QuotaUsed:        1000,
		WindowStartTime:  time.Now().Add(-1 * time.Hour),
		WindowEndTime:    time.Now(),
	}

	result, err := detector.Detect(ctx)
	if err != nil {
		t.Errorf("detector failed: %v", err)
	}

	if result.Detected {
		t.Errorf("expected no anomaly for normal activity, but detected one")
	}

	// Test case: High quota spike
	ctx.QuotaUsed = 10000
	result, err = detector.Detect(ctx)
	if err != nil {
		t.Errorf("detector failed: %v", err)
	}

	// Should not detect if no baseline is set
	if result.Detected {
		t.Errorf("expected no detection without baseline")
	}
}

func TestAbnormalLoginRatioDetector(t *testing.T) {
	detector := NewAbnormalLoginRatioDetector(1000.0)

	if detector.GetRuleType() != "abnormal_login_ratio" {
		t.Errorf("expected rule type 'abnormal_login_ratio', got %s", detector.GetRuleType())
	}

	// Test case: No logins
	ctx := &DetectionContext{
		UserId:           1,
		RequestCount:     100,
		LoginCount:       0,
		WindowStartTime:  time.Now().Add(-1 * time.Hour),
		WindowEndTime:    time.Now(),
	}

	result, err := detector.Detect(ctx)
	if err != nil {
		t.Errorf("detector failed: %v", err)
	}

	if result.Detected {
		t.Errorf("expected no anomaly with no logins")
	}

	// Test case: Normal ratio
	ctx.LoginCount = 1
	result, err = detector.Detect(ctx)
	if err != nil {
		t.Errorf("detector failed: %v", err)
	}

	if result.Detected {
		t.Errorf("expected no anomaly for normal ratio")
	}
}

func TestHighRequestRatioDetector(t *testing.T) {
	detector := NewHighRequestRatioDetector(500.0)

	if detector.GetRuleType() != "high_request_ratio" {
		t.Errorf("expected rule type 'high_request_ratio', got %s", detector.GetRuleType())
	}

	// Test case: High request ratio
	ctx := &DetectionContext{
		UserId:           1,
		RequestCount:     1000,
		LoginCount:       1,
		WindowStartTime:  time.Now().Add(-1 * time.Hour),
		WindowEndTime:    time.Now(),
	}

	result, err := detector.Detect(ctx)
	if err != nil {
		t.Errorf("detector failed: %v", err)
	}

	if !result.Detected {
		t.Errorf("expected anomaly detection for high request ratio")
	}

	if result.Severity != "high" {
		t.Errorf("expected severity 'high', got %s", result.Severity)
	}

	// Test case: No login with high requests
	ctx.LoginCount = 0
	ctx.RequestCount = 1001
	result, err = detector.Detect(ctx)
	if err != nil {
		t.Errorf("detector failed: %v", err)
	}

	if !result.Detected {
		t.Errorf("expected anomaly detection for high request without login")
	}
}

func TestUnusualDeviceActivityDetector(t *testing.T) {
	detector := NewUnusualDeviceActivityDetector(100, 5)

	if detector.GetRuleType() != "unusual_device_activity" {
		t.Errorf("expected rule type 'unusual_device_activity', got %s", detector.GetRuleType())
	}

	// Test case: Too many IPs per device
	now := time.Now()
	ips := make([]string, 0)
	for i := 0; i < 6; i++ {
		ips = append(ips, fmt.Sprintf("192.168.1.%d", i))
	}

	ctx := &DetectionContext{
		UserId:           1,
		WindowStartTime:  now.Add(-1 * time.Hour),
		WindowEndTime:    now,
		DeviceAggregations: []*DeviceAggregationResult{
			{
				NormalizedDeviceId: "device_123",
				RequestCount:       150,
				UniqueIPs:          ips,
				FirstSeenAt:        now.Add(-2 * time.Hour),
				LastSeenAt:         now,
			},
		},
	}

	result, err := detector.Detect(ctx)
	if err != nil {
		t.Errorf("detector failed: %v", err)
	}

	if !result.Detected {
		t.Errorf("expected anomaly detection for too many IPs")
	}

	if result.Severity != "high" {
		t.Errorf("expected severity 'high', got %s", result.Severity)
	}
}

func TestCalculateSeverity(t *testing.T) {
	tests := []struct {
		deviation           float64
		mediumThreshold     float64
		criticalThreshold   float64
		expectedSeverity    string
	}{
		{50.0, 100.0, 300.0, "low"},
		{150.0, 100.0, 300.0, "medium"},
		{250.0, 100.0, 300.0, "high"},
		{350.0, 100.0, 300.0, "critical"},
	}

	for _, test := range tests {
		result := calculateSeverity(test.deviation, test.mediumThreshold, test.criticalThreshold)
		if result != test.expectedSeverity {
			t.Errorf("calculateSeverity(%.1f, %.1f, %.1f) = %s, want %s",
				test.deviation, test.mediumThreshold, test.criticalThreshold, result, test.expectedSeverity)
		}
	}
}

func TestEngineInitialization(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("engine creation failed")
	}

	if len(engine.detectors) == 0 {
		t.Errorf("engine should have default detectors")
	}

	if engine.running {
		t.Errorf("engine should not be running on initialization")
	}
}

func TestEngineAddDetector(t *testing.T) {
	engine := NewEngine()
	initialCount := len(engine.detectors)

	customDetector := NewQuotaSpikeDetector(100.0)
	engine.AddDetector(customDetector)

	if len(engine.detectors) != initialCount+1 {
		t.Errorf("expected %d detectors after adding one, got %d", initialCount+1, len(engine.detectors))
	}
}
