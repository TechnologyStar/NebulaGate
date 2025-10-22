package governance

import (
    "crypto/sha256"
    "encoding/hex"
    "math"
    "strconv"
    "strings"
    "sync"
    "sync/atomic"
    "time"

    "github.com/QuantumNous/new-api/service"
)

// Severity indicates the impact level of a detector result.
// - malicious: abuse or behavior that warrants strict enforcement or reroute
// - violation: policy violation or suspicious content
const (
    SeverityMalicious = "malicious"
    SeverityViolation = "violation"
)

// DetectorResult encapsulates the outcome of a detector evaluation.
// Reasons are short machine-parsable codes (e.g., "high_rpm", "low_entropy").
// Metadata can include anonymized or hashed snippets for later analysis.
// Triggered indicates whether any rule has fired.

type DetectorResult struct {
    Triggered bool              `json:"triggered"`
    Severity  string            `json:"severity"`
    Reasons   []string          `json:"reasons"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// metrics skeleton (wired in M7)
var (
    detectionsFlagged uint64
    detectionsPassed  uint64
)

// RPMMonitor provides a simple in-memory sliding window RPM counter per subject key.
// It keeps timestamps in the last `window` duration and returns the number of requests
// in that window on each Record call.
type RPMMonitor struct {
    mu      sync.Mutex
    window  time.Duration
    buckets map[string][]time.Time
}

func NewRPMMonitor(window time.Duration) *RPMMonitor {
    return &RPMMonitor{
        window:  window,
        buckets: make(map[string][]time.Time),
    }
}

func (m *RPMMonitor) Record(key string, at time.Time) int {
    m.mu.Lock()
    defer m.mu.Unlock()
    arr := m.buckets[key]
    cutoff := at.Add(-m.window)
    // drop old entries
    i := 0
    for ; i < len(arr); i++ {
        if arr[i].After(cutoff) {
            break
        }
    }
    if i > 0 {
        arr = arr[i:]
    }
    arr = append(arr, at)
    m.buckets[key] = arr
    return len(arr)
}

func (m *RPMMonitor) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.buckets = make(map[string][]time.Time)
}

var (
    rpmOnce    sync.Once
    rpmMonitor *RPMMonitor
)

func getRPMMonitor() *RPMMonitor {
    rpmOnce.Do(func() {
        rpmMonitor = NewRPMMonitor(time.Minute)
    })
    return rpmMonitor
}

// DetectHighRPM checks whether the subject has exceeded the configured RPM threshold.
// Returns a malicious severity result when triggered.
func DetectHighRPM(subjectKey string, now time.Time) DetectorResult {
    count := getRPMMonitor().Record(subjectKey, now)
    threshold := AbuseRPMThreshold()
    if threshold > 0 && count > threshold {
        // triggered as malicious abuse
        atomic.AddUint64(&detectionsFlagged, 1)
        return DetectorResult{
            Triggered: true,
            Severity:  SeverityMalicious,
            Reasons:   []string{"high_rpm"},
            Metadata: map[string]string{
                "subject": subjectKey,
                "count":   intToString(count),
                "window":  "1m",
            },
        }
    }
    atomic.AddUint64(&detectionsPassed, 1)
    return DetectorResult{Triggered: false}
}

// DetectPromptSanity runs heuristic checks on prompt content: length, entropy, repetition.
// It flags low entropy or high repetition as violations; extremely long content may also
// be flagged according to configuration.
func DetectPromptSanity(prompt string) DetectorResult {
    // Normalize newlines for analysis but do not alter length significantly
    normalized := strings.ReplaceAll(prompt, "\r\n", "\n")
    reasons := make([]string, 0, 3)

    maxLen := PromptMaxLength()
    if maxLen > 0 && len(normalized) > maxLen {
        reasons = append(reasons, "prompt_too_long")
    }

    entropy := shannonEntropy(normalized)
    if entropy > 0 && entropy < PromptMinEntropyBits() {
        reasons = append(reasons, "low_entropy")
    }

    repRatio := maxRuneRatio(normalized)
    if repRatio > PromptMaxRepetitionRatio() {
        reasons = append(reasons, "high_repetition")
    }

    if len(reasons) > 0 {
        atomic.AddUint64(&detectionsFlagged, 1)
        return DetectorResult{
            Triggered: true,
            Severity:  SeverityViolation,
            Reasons:   reasons,
            Metadata: map[string]string{
                "entropy":    floatToString(entropy),
                "rep_ratio":  floatToString(repRatio),
                "len":        intToString(len(normalized)),
            },
        }
    }
    atomic.AddUint64(&detectionsPassed, 1)
    return DetectorResult{Triggered: false}
}

// DetectKeywordPolicy checks the prompt text against the built-in sensitive word
// list and additional governance violation keywords.
func DetectKeywordPolicy(prompt string) DetectorResult {
    text := strings.ToLower(prompt)
    // First, reuse existing sensitive word detector
    if ok, words := service.SensitiveWordContains(text); ok {
        hash := ""
        if len(words) > 0 {
            hash = hashSnippet(words[0])
        }
        atomic.AddUint64(&detectionsFlagged, 1)
        return DetectorResult{
            Triggered: true,
            Severity:  SeverityViolation,
            Reasons:   []string{"keyword_violation"},
            Metadata: map[string]string{
                "hash": hash,
                "type": "sensitive",
            },
        }
    }

    // Merge additional violation keywords from governance config
    extra := ViolationKeywords()
    if len(extra) > 0 {
        for _, w := range extra {
            w = strings.TrimSpace(strings.ToLower(w))
            if w == "" {
                continue
            }
            if strings.Contains(text, w) {
                atomic.AddUint64(&detectionsFlagged, 1)
                return DetectorResult{
                    Triggered: true,
                    Severity:  SeverityViolation,
                    Reasons:   []string{"keyword_violation"},
                    Metadata: map[string]string{
                        "hash": hashSnippet(w),
                        "type": "policy",
                    },
                }
            }
        }
    }

    atomic.AddUint64(&detectionsPassed, 1)
    return DetectorResult{Triggered: false}
}

// Utility functions

func shannonEntropy(s string) float64 {
    if s == "" {
        return 0
    }
    // frequency of runes
    freq := make(map[rune]int)
    total := 0
    for _, r := range s {
        freq[r]++
        total++
    }
    if total == 0 {
        return 0
    }
    var entropy float64
    for _, c := range freq {
        p := float64(c) / float64(total)
        if p > 0 {
            entropy -= p * math.Log2(p)
        }
    }
    return entropy
}

func maxRuneRatio(s string) float64 {
    if s == "" {
        return 0
    }
    freq := make(map[rune]int)
    maxC := 0
    total := 0
    for _, r := range s {
        freq[r]++
        if freq[r] > maxC {
            maxC = freq[r]
        }
        total++
    }
    if total == 0 {
        return 0
    }
    return float64(maxC) / float64(total)
}

func hashSnippet(s string) string {
    sum := sha256.Sum256([]byte(s))
    return hex.EncodeToString(sum[:8]) // short hash for metadata brevity
}

func intToString(n int) string {
    return strconv.Itoa(n)
}

func floatToString(f float64) string {
    return strconv.FormatFloat(f, 'f', 4, 64)
}
