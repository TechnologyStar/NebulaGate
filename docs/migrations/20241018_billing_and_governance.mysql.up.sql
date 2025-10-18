START TRANSACTION;

CREATE TABLE IF NOT EXISTS plans (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code VARCHAR(64) NOT NULL,
  name VARCHAR(128) NOT NULL,
  description TEXT NULL,
  cycle_type VARCHAR(16) NOT NULL,
  cycle_duration_days INT NOT NULL DEFAULT 0,
  quota_metric VARCHAR(16) NOT NULL,
  quota_amount BIGINT NOT NULL,
  upstream_alias_whitelist JSON NULL,
  is_active TINYINT(1) NOT NULL DEFAULT 1,
  is_public TINYINT(1) NOT NULL DEFAULT 0,
  is_system TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_at DATETIME(6) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_plans_code (code)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS plan_assignments (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  subject_type VARCHAR(16) NOT NULL,
  subject_id BIGINT NOT NULL,
  plan_id BIGINT NOT NULL,
  billing_mode VARCHAR(16) NOT NULL DEFAULT 'plan',
  activated_at DATETIME(6) NOT NULL,
  deactivated_at DATETIME(6) NULL,
  rollover_policy VARCHAR(16) NOT NULL DEFAULT 'none',
  rollover_amount BIGINT NOT NULL DEFAULT 0,
  rollover_expires_at DATETIME(6) NULL,
  auto_fallback_enabled TINYINT(1) NOT NULL DEFAULT 0,
  fallback_plan_id BIGINT NULL,
  metadata JSON NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_at DATETIME(6) NULL,
  PRIMARY KEY (id),
  KEY idx_plan_assignments_plan (plan_id)
) ENGINE=InnoDB;
CREATE INDEX idx_plan_assignments_subject ON plan_assignments (subject_type, subject_id);
CREATE INDEX idx_plan_assignments_window ON plan_assignments (activated_at, deactivated_at);

CREATE TABLE IF NOT EXISTS usage_counters (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  plan_assignment_id BIGINT NOT NULL,
  metric VARCHAR(16) NOT NULL,
  cycle_start DATETIME(6) NOT NULL,
  cycle_end DATETIME(6) NOT NULL,
  consumed_amount BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_usage_assignment_metric_cycle (plan_assignment_id, metric, cycle_start)
) ENGINE=InnoDB;
CREATE INDEX idx_usage_counters_cycle_end ON usage_counters (cycle_end);

CREATE TABLE IF NOT EXISTS voucher_batches (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  code_prefix VARCHAR(48) NOT NULL,
  batch_label VARCHAR(128) NOT NULL,
  grant_type VARCHAR(16) NOT NULL DEFAULT 'credit',
  credit_amount BIGINT NOT NULL DEFAULT 0,
  plan_grant_id BIGINT NULL,
  plan_grant_duration INT NULL,
  is_stackable TINYINT(1) NOT NULL DEFAULT 0,
  max_redemptions INT NOT NULL DEFAULT 0,
  max_per_subject INT NOT NULL DEFAULT 0,
  valid_from DATETIME(6) NULL,
  valid_until DATETIME(6) NULL,
  metadata JSON NULL,
  created_by VARCHAR(64) NULL,
  notes TEXT NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  deleted_at DATETIME(6) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_voucher_code_prefix (code_prefix)
) ENGINE=InnoDB;
CREATE INDEX idx_voucher_batches_validity ON voucher_batches (valid_from, valid_until);

CREATE TABLE IF NOT EXISTS voucher_redemptions (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  voucher_batch_id BIGINT NOT NULL,
  code VARCHAR(96) NOT NULL,
  subject_type VARCHAR(16) NOT NULL,
  subject_id BIGINT NOT NULL,
  plan_assignment_id BIGINT NULL,
  redeemed_at DATETIME(6) NOT NULL,
  redeemed_by VARCHAR(64) NULL,
  credit_amount BIGINT NOT NULL DEFAULT 0,
  plan_granted_id BIGINT NULL,
  metadata JSON NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_voucher_redemption_code (code),
  KEY idx_voucher_redemptions_batch (voucher_batch_id)
) ENGINE=InnoDB;
CREATE INDEX idx_voucher_redemptions_subject ON voucher_redemptions (subject_type, subject_id);

CREATE TABLE IF NOT EXISTS request_flags (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  request_id VARCHAR(64) NOT NULL,
  subject_type VARCHAR(16) NOT NULL,
  subject_id BIGINT NOT NULL,
  user_id BIGINT NULL,
  token_id BIGINT NULL,
  reason VARCHAR(16) NOT NULL,
  rerouted_model_alias VARCHAR(128) NULL,
  ttl_at DATETIME(6) NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_request_flags_request (request_id)
) ENGINE=InnoDB;
CREATE INDEX idx_request_flags_subject ON request_flags (subject_type, subject_id);

CREATE TABLE IF NOT EXISTS request_logs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  request_id VARCHAR(64) NOT NULL,
  occurred_at DATETIME(6) NOT NULL,
  model_alias VARCHAR(128) NULL,
  upstream_provider VARCHAR(64) NULL,
  subject_type VARCHAR(16) NOT NULL,
  anonymized_subject_hash VARCHAR(128) NOT NULL,
  plan_id BIGINT NULL,
  plan_assignment_id BIGINT NULL,
  usage_metric VARCHAR(16) NOT NULL DEFAULT 'requests',
  prompt_tokens BIGINT NOT NULL DEFAULT 0,
  completion_tokens BIGINT NOT NULL DEFAULT 0,
  total_tokens BIGINT NOT NULL DEFAULT 0,
  latency_ms BIGINT NOT NULL DEFAULT 0,
  flag_ids JSON NULL,
  metadata JSON NULL,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_request_logs_request (request_id)
) ENGINE=InnoDB;
CREATE INDEX idx_request_logs_model_window ON request_logs (model_alias, upstream_provider, occurred_at);
CREATE INDEX idx_request_logs_subject ON request_logs (subject_type, anonymized_subject_hash);

CREATE TABLE IF NOT EXISTS request_aggregates (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  model_alias VARCHAR(128) NOT NULL,
  upstream VARCHAR(64) NOT NULL,
  subject_type VARCHAR(16) NOT NULL,
  window_start DATETIME(6) NOT NULL,
  window_end DATETIME(6) NOT NULL,
  total_requests BIGINT NOT NULL DEFAULT 0,
  total_tokens BIGINT NOT NULL DEFAULT 0,
  unique_subjects BIGINT NOT NULL DEFAULT 0,
  created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  UNIQUE KEY uk_request_aggregates_window (model_alias, upstream, subject_type, window_start, window_end)
) ENGINE=InnoDB;
CREATE INDEX idx_request_aggregates_window_end ON request_aggregates (window_end);

COMMIT;
