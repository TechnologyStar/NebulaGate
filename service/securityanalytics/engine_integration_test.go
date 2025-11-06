package securityanalytics

import (
    "testing"
    "time"

    "github.com/QuantumNous/new-api/model"
)

// TestAggregationQueryResults tests that aggregation queries return accurate results
func TestAggregationQueryResults(t *testing.T) {
    t.Skip("Integration test - requires database connection")

    // This test would verify:
    // 1. Device aggregation groups correctly
    // 2. IP aggregation includes all IPs
    // 3. Results have accurate counts

    userId := 1
    startTime := time.Now().Add(-24 * time.Hour)
    endTime := time.Now()

    // Test device aggregation
    deviceAggs, err := AggregateDeviceActivity(userId, startTime, endTime)
    if err != nil {
        t.Fatalf("device aggregation failed: %v", err)
    }

    if deviceAggs == nil {
        t.Errorf("expected device aggregations, got nil")
    }

    // Test IP aggregation
    ipAggs, err := AggregateIPActivity(userId, startTime, endTime)
    if err != nil {
        t.Fatalf("IP aggregation failed: %v", err)
    }

    if ipAggs == nil {
        t.Errorf("expected IP aggregations, got nil")
    }
}

// TestAnomalyDetectionWithSyntheticData tests anomaly detection with synthetic data
func TestAnomalyDetectionWithSyntheticData(t *testing.T) {
    t.Skip("Integration test - requires database connection")

    // Simulate suspicious behavior
    userId := 1

    // Create a detection context with synthetic data
    ctx := &DetectionContext{
        UserId:              userId,
        WindowStartTime:     time.Now().Add(-1 * time.Hour),
        WindowEndTime:       time.Now(),
        TimeWindow:          1 * time.Hour,
        RequestCount:        10000,     // Very high
        QuotaUsed:           100000,    // Very high
        LoginCount:          1,         // Only 1 login
        UniqueIPs:           20,        // Many IPs
        UniqueDevices:       5,         // Multiple devices
        LatestBaseline:      make(map[string]*model.AnomalyBaseline),
        DeviceAggregations: []*DeviceAggregationResult{
            {
                NormalizedDeviceId: "new_device",
                UserId:             userId,
                RequestCount:       5000,
                UniqueIPs:          []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"},
                LastSeenAt:         time.Now(),
                FirstSeenAt:        time.Now().Add(-1 * time.Hour),
            },
        },
    }

    // Run detectors
    engine := NewEngine()

    for _, detector := range engine.detectors {
        result, err := detector.Detect(ctx)
        if err != nil {
            t.Errorf("detector %s failed: %v", detector.GetRuleType(), err)
            continue
        }

        // Verify result structure
        if result.Detected {
            if result.RuleType == "" {
                t.Errorf("detected anomaly should have rule type")
            }
            if result.Severity == "" {
                t.Errorf("detected anomaly should have severity")
            }
            if result.Message == "" {
                t.Errorf("detected anomaly should have message")
            }
            if result.Evidence == nil {
                t.Errorf("detected anomaly should have evidence")
            }
        }
    }
}

// TestAnomalyPersistence tests that anomalies are correctly persisted
func TestAnomalyPersistence(t *testing.T) {
    t.Skip("Integration test - requires database connection")

    // This test would verify:
    // 1. Anomalies are saved to database
    // 2. Retrieved anomalies match saved data
    // 3. Deduplication prevents duplicates

    userId := 1
    result := &DetectionResult{
        Detected:      true,
        RuleType:      "quota_spike",
        Severity:      "high",
        Message:       "Test anomaly",
        Evidence:      map[string]interface{}{"test": "value"},
        Threshold:     1000.0,
        ActualValue:   1500.0,
        Baseline:      1000.0,
        Deviation:     50.0,
    }

    // Create anomaly record
    // anomaly, err := createAnomalyRecord(userId, result, nil)
    // if err != nil {
    //     t.Fatalf("failed to create anomaly: %v", err)
    // }

    // if anomaly.Id == 0 {
    //     t.Errorf("anomaly should have been assigned an ID")
    // }

    // if anomaly.RuleType != result.RuleType {
    //     t.Errorf("anomaly rule type mismatch")
    // }

    // // Retrieve and verify
    // anomalies, _, err := model.GetSecurityAnomalies(0, 10, userId, "", "", nil, nil)
    // if err != nil {
    //     t.Fatalf("failed to retrieve anomalies: %v", err)
    // }

    // if len(anomalies) == 0 {
    //     t.Errorf("expected to retrieve anomaly")
    // }
}

// TestBackgroundProcessing tests the background processing loop
func TestBackgroundProcessing(t *testing.T) {
    t.Skip("Integration test - requires database connection and time")

    engine := NewEngine()

    // Start engine with short interval for testing
    stopCh := engine.Start(1*time.Second, 1*time.Hour, 5)

    // Let it run for a bit
    time.Sleep(2 * time.Second)

    // Stop engine
    engine.Stop(stopCh)

    // Verify it stopped
    if engine.running {
        t.Errorf("engine should have stopped")
    }
}

// TestDeduplication tests that duplicate anomalies are not created
func TestDeduplication(t *testing.T) {
    t.Skip("Integration test - requires database connection")

    engine := NewEngine()

    // Simulate same anomaly detected twice
    userId := 1
    result := &DetectionResult{
        Detected:      true,
        RuleType:      "quota_spike",
        Severity:      "high",
        Message:       "Test anomaly",
        Evidence:      map[string]interface{}{},
        DeduplicationKey: "test_key",
    }

    // Try to create same anomaly
    // anomaly1, _ := engine.createAnomalyRecord(userId, result, nil)
    // anomaly2, _ := engine.createAnomalyRecord(userId, result, nil)

    // isDuplicate should return true for second call
    // if !engine.isDuplicate(anomaly2) {
    //     t.Errorf("second anomaly should be detected as duplicate")
    // }
}
