BEGIN;

CREATE TABLE IF NOT EXISTS plans (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    cycle_type VARCHAR(16) NOT NULL CHECK (cycle_type IN ('daily','monthly','custom')),
    cycle_duration_days INTEGER NOT NULL DEFAULT 0,
    quota_metric VARCHAR(16) NOT NULL CHECK (quota_metric IN ('requests','tokens')),
    quota_amount BIGINT NOT NULL,
    upstream_alias_whitelist JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS plan_assignments (
    id BIGSERIAL PRIMARY KEY,
    subject_type VARCHAR(16) NOT NULL,
    subject_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    billing_mode VARCHAR(16) NOT NULL DEFAULT 'plan' CHECK (billing_mode IN ('plan','prepaid','voucher','fallback')),
    activated_at TIMESTAMPTZ NOT NULL,
    deactivated_at TIMESTAMPTZ,
    rollover_policy VARCHAR(16) NOT NULL DEFAULT 'none' CHECK (rollover_policy IN ('none','carry_all','cap')),
    rollover_amount BIGINT NOT NULL DEFAULT 0,
    rollover_expires_at TIMESTAMPTZ,
    auto_fallback_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    fallback_plan_id BIGINT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_plan_assignments_subject ON plan_assignments (subject_type, subject_id);
CREATE INDEX IF NOT EXISTS idx_plan_assignments_window ON plan_assignments (activated_at, deactivated_at);

CREATE TABLE IF NOT EXISTS usage_counters (
    id BIGSERIAL PRIMARY KEY,
    plan_assignment_id BIGINT NOT NULL,
    metric VARCHAR(16) NOT NULL,
    cycle_start TIMESTAMPTZ NOT NULL,
    cycle_end TIMESTAMPTZ NOT NULL,
    consumed_amount BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (plan_assignment_id, metric, cycle_start)
);
CREATE INDEX IF NOT EXISTS idx_usage_counters_cycle_end ON usage_counters (cycle_end);

CREATE TABLE IF NOT EXISTS voucher_batches (
    id BIGSERIAL PRIMARY KEY,
    code_prefix VARCHAR(48) NOT NULL UNIQUE,
    batch_label VARCHAR(128) NOT NULL,
    grant_type VARCHAR(16) NOT NULL DEFAULT 'credit' CHECK (grant_type IN ('credit','plan')),
    credit_amount BIGINT NOT NULL DEFAULT 0,
    plan_grant_id BIGINT,
    plan_grant_duration INTEGER,
    is_stackable BOOLEAN NOT NULL DEFAULT FALSE,
    max_redemptions INTEGER NOT NULL DEFAULT 0,
    max_per_subject INTEGER NOT NULL DEFAULT 0,
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    metadata JSONB,
    created_by VARCHAR(64),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_voucher_batches_validity ON voucher_batches (valid_from, valid_until);

CREATE TABLE IF NOT EXISTS voucher_redemptions (
    id BIGSERIAL PRIMARY KEY,
    voucher_batch_id BIGINT NOT NULL,
    code VARCHAR(96) NOT NULL UNIQUE,
    subject_type VARCHAR(16) NOT NULL,
    subject_id BIGINT NOT NULL,
    plan_assignment_id BIGINT,
    redeemed_at TIMESTAMPTZ NOT NULL,
    redeemed_by VARCHAR(64),
    credit_amount BIGINT NOT NULL DEFAULT 0,
    plan_granted_id BIGINT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_batch ON voucher_redemptions (voucher_batch_id);
CREATE INDEX IF NOT EXISTS idx_voucher_redemptions_subject ON voucher_redemptions (subject_type, subject_id);

CREATE TABLE IF NOT EXISTS request_flags (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(64) NOT NULL,
    subject_type VARCHAR(16) NOT NULL,
    subject_id BIGINT NOT NULL,
    user_id BIGINT,
    token_id BIGINT,
    reason VARCHAR(16) NOT NULL,
    rerouted_model_alias VARCHAR(128),
    ttl_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_request_flags_request ON request_flags (request_id);
CREATE INDEX IF NOT EXISTS idx_request_flags_subject ON request_flags (subject_type, subject_id);

CREATE TABLE IF NOT EXISTS request_logs (
    id BIGSERIAL PRIMARY KEY,
    request_id VARCHAR(64) NOT NULL UNIQUE,
    occurred_at TIMESTAMPTZ NOT NULL,
    model_alias VARCHAR(128),
    upstream_provider VARCHAR(64),
    subject_type VARCHAR(16) NOT NULL,
    anonymized_subject_hash VARCHAR(128) NOT NULL,
    plan_id BIGINT,
    plan_assignment_id BIGINT,
    usage_metric VARCHAR(16) NOT NULL DEFAULT 'requests',
    prompt_tokens BIGINT NOT NULL DEFAULT 0,
    completion_tokens BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    flag_ids JSONB,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_request_logs_model_window ON request_logs (model_alias, upstream_provider, occurred_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_subject ON request_logs (subject_type, anonymized_subject_hash);

CREATE TABLE IF NOT EXISTS request_aggregates (
    id BIGSERIAL PRIMARY KEY,
    model_alias VARCHAR(128) NOT NULL,
    upstream VARCHAR(64) NOT NULL,
    subject_type VARCHAR(16) NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    total_requests BIGINT NOT NULL DEFAULT 0,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    unique_subjects BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (model_alias, upstream, subject_type, window_start, window_end)
);
CREATE INDEX IF NOT EXISTS idx_request_aggregates_window_end ON request_aggregates (window_end);

COMMIT;
