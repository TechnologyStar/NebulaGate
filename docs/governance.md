# Governance detectors and configuration

Milestone M4 introduces request governance primitives focused on detection and (future) reroute.
This document summarizes the detectors implemented in `service/governance` and the configuration
surfaces used to tune behavior.

Detectors
- High RPM (requests per minute)
  - Purpose: detect abusive concurrency from the same subject (user/token/IP).
  - Implementation: in-memory sliding window counter. A Redis backend can be wired later.
  - Result: severity "malicious", reason code `high_rpm`.
- Prompt sanity
  - Purpose: detect obviously problematic prompts (very low entropy or excessive repetition) and
    optionally guard against extremely long payloads.
  - Implementation: Shannon entropy over runes; repetition measured by the share of the most
    frequent rune; max-length check.
  - Result: severity "violation", reason codes include `low_entropy`, `high_repetition`, `prompt_too_long`.
- Keyword policy
  - Purpose: enforce policy- or content-based keyword rules in addition to the existing sensitive
    word list.
  - Implementation: combines `setting.SensitiveWords` with additional `governance.violation_keywords`.
  - Result: severity "violation", reason code `keyword_violation` with a short hash of the hit.

Detector outputs
- All detectors return a DetectorResult with fields:
  - `triggered`: boolean
  - `severity`: "malicious" or "violation"
  - `reasons`: list of short reason codes
  - `metadata`: optional key/value (e.g. short hash of offending snippet)

Configuration keys
Configuration is layered through the config manager (`setting/config`) with environment
variable fallbacks. Defaults are applied even when governance is disabled.

- governance.enabled: feature gate (default: false)
- governance.abuse_rpm_threshold: RPM ceiling before `high_rpm` triggers (default: 3000)
- governance.reroute_model_alias: model alias used for reroute in later milestones (default: "")
- governance.flag_ttl_hours: default TTL for request flags, used by scheduler (default: 24)
- governance.prompt_max_length: max prompt size in bytes/runes before `prompt_too_long` (default: 8192)
- governance.prompt_min_entropy: minimum Shannon entropy (bits/char) before `low_entropy` (default: 1.0)
- governance.prompt_max_repetition: maximum ratio for the most frequent rune before `high_repetition` (default: 0.6)
- governance.violation_keywords: additional keywords (comma/newline separated). Merged with `setting.SensitiveWords`.

Current defaults are applied inside `service/governance/config.go` and can also be overridden via
environment variables with the following names if DB-config is not present:
- GOVERNANCE_ABUSE_RPM_THRESHOLD (int)
- GOVERNANCE_PROMPT_MAX_LENGTH (int)
- GOVERNANCE_PROMPT_MIN_ENTROPY (float)
- GOVERNANCE_PROMPT_MAX_REPETITION (float 0..1)
- GOVERNANCE_VIOLATION_KEYWORDS (comma/newline separated string)

Metrics
A minimal metrics skeleton has been added in `service/governance/detector.go` with counters for
flagged vs passed evaluations; these will be exported and wired to the metrics collector in M7.

Fallback behavior
- If a configuration value is not set, detectors fall back to safe defaults.
- The keyword detector always honors `setting.SensitiveWords` even if the governance module is disabled.
- When thresholds are not configured, detectors use the defaults listed above.
