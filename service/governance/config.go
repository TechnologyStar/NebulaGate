package governance

import (
    "os"
    "strconv"
    "strings"

    "github.com/QuantumNous/new-api/common"
    cfg "github.com/QuantumNous/new-api/setting/config"
)

// Convenience getters for governance-related configuration. These helpers never panic
// and always return sensible defaults even when the governance feature is disabled
// or the configuration has not been loaded yet.

const (
    defaultAbuseRPMThreshold      = 3000
    defaultPromptMaxLength        = 8192
    defaultPromptMinEntropyBits   = 1.0   // bits per character
    defaultPromptMaxRepeatRatio   = 0.6    // 60% of characters being the same triggers
)

// AbuseRPMThreshold returns the max allowed requests per minute for a single subject
// before the high-RPM detector triggers.
func AbuseRPMThreshold() int {
    if gc := cfg.GetGovernanceConfig(); gc != nil {
        if gc.AbuseRPMThreshold > 0 {
            return gc.AbuseRPMThreshold
        }
    }
    if common.GovernanceAbuseRPMThreshold > 0 {
        return common.GovernanceAbuseRPMThreshold
    }
    // Last resort: env var (useful for tests) or default
    if s := os.Getenv("GOVERNANCE_ABUSE_RPM_THRESHOLD"); s != "" {
        if n, err := strconv.Atoi(s); err == nil && n > 0 {
            return n
        }
    }
    return defaultAbuseRPMThreshold
}

// PromptMaxLength returns the maximum allowed prompt length before flagging.
func PromptMaxLength() int {
    if gc := cfg.GetGovernanceConfig(); gc != nil {
        // Support optional dynamic field via reflection-less approach: we rely on
        // environment fallback if not wired in config object.
    }
    if s := os.Getenv("GOVERNANCE_PROMPT_MAX_LENGTH"); s != "" {
        if n, err := strconv.Atoi(s); err == nil && n > 0 {
            return n
        }
    }
    return defaultPromptMaxLength
}

// PromptMinEntropyBits returns the minimum acceptable Shannon entropy (bits per char)
// for a prompt; lower values are considered suspicious.
func PromptMinEntropyBits() float64 {
    if s := os.Getenv("GOVERNANCE_PROMPT_MIN_ENTROPY"); s != "" {
        if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 {
            return f
        }
    }
    return defaultPromptMinEntropyBits
}

// PromptMaxRepetitionRatio returns the maximum allowed ratio for the most frequent
// rune within a prompt. Values above this threshold are flagged.
func PromptMaxRepetitionRatio() float64 {
    if s := os.Getenv("GOVERNANCE_PROMPT_MAX_REPETITION"); s != "" {
        if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 && f <= 1.0 {
            return f
        }
    }
    return defaultPromptMaxRepeatRatio
}

// ViolationKeywords returns additional violation keywords configured for governance,
// combined with any environment overrides. Entries are normalized to lower-case and
// trimmed. It is intended to be merged with the base sensitive word list.
func ViolationKeywords() []string {
    // Prefer config manager if such a key exists later; for now, read from env.
    raw := os.Getenv("GOVERNANCE_VIOLATION_KEYWORDS")
    if raw == "" {
        return nil
    }
    // Support comma or newline separated values
    sep := ","
    if strings.Contains(raw, "\n") {
        sep = "\n"
    }
    parts := strings.Split(raw, sep)
    out := make([]string, 0, len(parts))
    for _, p := range parts {
        w := strings.ToLower(strings.TrimSpace(p))
        if w != "" {
            out = append(out, w)
        }
    }
    return out
}
