package governance

import (
    "testing"
    "time"

    cfg "github.com/QuantumNous/new-api/setting/config"
)

func TestDetectorHighRPMTrigger(t *testing.T) {
    // Override threshold for test
    gc := cfg.GetGovernanceConfig()
    old := gc.AbuseRPMThreshold
    gc.AbuseRPMThreshold = 3
    defer func() { gc.AbuseRPMThreshold = old }()

    getRPMMonitor().Reset()

    now := time.Now()
    var res DetectorResult
    // 4th call in the same minute should trigger (threshold 3)
    for i := 0; i < 4; i++ {
        res = DetectHighRPM("user-1", now)
    }

    if !res.Triggered {
        t.Fatalf("expected high RPM to trigger, got: %+v", res)
    }
    if res.Severity != SeverityMalicious {
        t.Fatalf("expected severity 'malicious', got: %s", res.Severity)
    }
    found := false
    for _, r := range res.Reasons {
        if r == "high_rpm" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected reason 'high_rpm', got: %+v", res.Reasons)
    }
}

func TestDetectorLowEntropyFlagged(t *testing.T) {
    sample := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
    res := DetectPromptSanity(sample)
    if !res.Triggered {
        t.Fatalf("expected low entropy to trigger, got: %+v", res)
    }
    if res.Severity != SeverityViolation {
        t.Fatalf("expected violation severity, got: %s", res.Severity)
    }
    hasLowEntropy := false
    for _, r := range res.Reasons {
        if r == "low_entropy" {
            hasLowEntropy = true
            break
        }
    }
    if !hasLowEntropy {
        t.Fatalf("expected reason 'low_entropy', got: %+v", res.Reasons)
    }
}

func TestDetectorKeywordHit(t *testing.T) {
    // 'test_sensitive' is present in default setting.SensitiveWords
    text := "this contains test_sensitive keyword"
    res := DetectKeywordPolicy(text)
    if !res.Triggered {
        t.Fatalf("expected keyword violation to trigger, got: %+v", res)
    }
    if res.Severity != SeverityViolation {
        t.Fatalf("expected violation severity, got: %s", res.Severity)
    }
    found := false
    for _, r := range res.Reasons {
        if r == "keyword_violation" {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected reason 'keyword_violation', got: %+v", res.Reasons)
    }
}

func TestDetectorNegativeCases(t *testing.T) {
    // RPM negative case
    gc := cfg.GetGovernanceConfig()
    old := gc.AbuseRPMThreshold
    gc.AbuseRPMThreshold = 100
    defer func() { gc.AbuseRPMThreshold = old }()
    getRPMMonitor().Reset()
    now := time.Now()
    res := DetectHighRPM("user-2", now)
    if res.Triggered {
        t.Fatalf("did not expect RPM detector to trigger, got: %+v", res)
    }

    // Prompt sanity negative case
    okText := "Hello world, this is a normal prompt asking for guidelines."
    res2 := DetectPromptSanity(okText)
    if res2.Triggered {
        t.Fatalf("did not expect sanity detector to trigger, got: %+v", res2)
    }

    // Keyword negative case
    res3 := DetectKeywordPolicy(okText)
    if res3.Triggered {
        t.Fatalf("did not expect keyword detector to trigger, got: %+v", res3)
    }
}
