BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    cycle_type TEXT NOT NULL,
    cycle_duration_days INTEGER NOT NULL DEFAULT 0,
    quota_metric TEXT NOT NULL,
    quota_amount INTEGER NOT NULL,
    upstream_alias_whitelist TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    is_public INTEGER NOT NULL DEFAULT 0,
    is_system INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    UNIQUE (code),
    CHECK (cycle_type IN ('daily','monthly','custom')),
    CHECK (quota_metric IN ('requests','tokens'))
);

CREATE TABLE IF NOT EXISTS plan_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subject_type TEXT NOT NULL,
    subject_id INTEGER NOT NULL,
    plan_id INTEGER NOT NULL,
    billing_mode TEXT NOT NULL DEFAULT 'plan',
    activated_at DATETIME NOT NULL,
    deactivated_at DATETIME,
    rollover_policy TEXT NOT NULL DEFAULT 'none',
    rollover_amount INTEGER NOT NULL DEFAULT 0,
    rollover_expires_at DATETIME,
    auto_fallback_enabled INTEGER NOT NULL DEFAULT 0,
    fallback_plan_id INTEGER,
    metadata TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    CHECK (billing_mode IN ('plan','prepaid','voucher','fallback')),
    CHECK (rollover_policy IN ('none','carry_all','cap'))
);
CREATE INDEX IF NOT EXISTS idx_plan_assignments_subject ON plan_assignments (subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_plan_assignments_window ON plan_assignments (activated_at, deactivated_at);

CREATE TABLE IF NOT EXISTS usage_counters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    plan_assignment_id INTEGER NOT NULL,
    metric TEXT NOT NULL,
    cycle_start DATETIME NOT NULL,
    cycle_end DATETIME NOT NULL,
    consumed_amount INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (plan_assignment_id, metric, cycle_start)
);
CREATE INDEX IF NOT EXISTS idx_usage_counters_cycle_end ON usage_counters (cycle_end);

CREATE TABLE IF NOT EXISTS voucher_batches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code_prefix TEXT NOT NULL,
    batch_label TEXT NOT NULL,
    grant_type TEXT NOT NULL DEFAULT 'credit',
    credit_amount INTEGER NOT NULL DEFAULT 0,
    plan_grant_id INTEGER,
    plan_grant_duration INTEGER,
    is_stackable INTEGER NOT NULL DEFAULT 0,
    max_redemptions INTEGER NOT NULL DEFAULT 0,
    max_per_subject INTEGER NOT NULL DEFAULT 0,
    valid_from DATETIME,
    valid_until DATETIME,
    metadata TEXT,
    created_by TEXT,
    notes TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    UNIQUE (code_prefix),
    CHECK (grant_type IN ('credit','plan'))
);
CREATE INDEX IF NOT EXISTS idx_voucher_batches_validity ON voucher_batches (valid_from, valid_until);

CREATE TABLE IF NOT EXISTS voucher_redemptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    voucher_batch_id INTEGER NOT NULL,
    code TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id INTEGER NOT NULL,
    plan_assignment_id INTEGER,
    redeemed_at DATETIME NOT NULL,
    redeemed_by TEXT,
    credit_amount INTEGER NOT NULL DEFAULT 0,
    plan_granted_id INTEGER,
    metadata TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (code)
);
CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_batch ON voucher_redemptions (voucher_batch_id);
CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_subject ON voucher_redemptions (subject_type, subject_id);

CREATE TABLE IF NOT EXISTS request_flags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    subject_id INTEGER NOT NULL,
    user_id INTEGER,
    token_id INTEGER,
    reason TEXT NOT NULL,
    rerouted_model_alias TEXT,
    ttl_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_request_flags_request ON request_flags (request_id);
CREATE INDEX IF NOT EXISTS idx_request_flags_subject ON request_flags (subject_type, subject_id);

CREATE TABLE IF NOT EXISTS request_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id TEXT NOT NULL,
    occurred_at DATETIME NOT NULL,
    model_alias TEXT,
    upstream_provider TEXT,
    subject_type TEXT NOT NULL,
    anonymized_subject_hash TEXT NOT NULL,
    plan_id INTEGER,
    plan_assignment_id INTEGER,
    usage_metric TEXT NOT NULL DEFAULT 'requests',
    prompt_tokens INTEGER NOT NULL DEFAULT 0,
    completion_tokens INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    flag_ids TEXT,
    metadata TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (request_id)
);
CREATE INDEX IF NOT EXISTS idx_request_logs_model_window ON request_logs (model_alias, upstream_provider, occurred_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_subject ON request_logs (subject_type, anonymized_subject_hash);

CREATE TABLE IF NOT EXISTS request_aggregates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_alias TEXT NOT NULL,
    upstream TEXT NOT NULL,
    subject_type TEXT NOT NULL,
    window_start DATETIME NOT NULL,
    window_end DATETIME NOT NULL,
    total_requests INTEGER NOT NULL DEFAULT 0,
    total_tokens INTEGER NOT NULL DEFAULT 0,
    unique_subjects INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (model_alias, upstream, subject_type, window_start, window_end)
);
CREATE INDEX IF NOT EXISTS idx_request_aggregates_window_end ON request_aggregates (window_end);

COMMIT;
