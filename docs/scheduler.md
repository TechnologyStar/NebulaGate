Scheduler: plan cycle resets and TTL housekeeping

Overview
- The scheduler runs background jobs that maintain plan usage cycles and clean up expired data.
- Jobs are feature-gated and only run when the corresponding features are enabled via config or environment.

Jobs
- Plan cycle reset
  - For each UsageCounter whose cycle_end <= now, reset consumed_amount to 0 and move the anchors (cycle_start/cycle_end) to the current cycle window.
  - The cycle window is determined from the associated Plan.CycleType (daily or monthly).
  - Idempotent: if the reset was missed, the next run will advance to the correct window.
- TTL cleanup
  - RequestFlag: deletes records where ttl_at is in the past, or if ttl_at is NULL, where created_at is older than governance.flag_ttl_hours (default 24h).
  - RequestLog: when public logs are enabled, deletes records older than public_logs.retention_days.
  - VoucherBatch: soft-deletes batches past valid_until.

Configuration
- billing.enabled: enable billing engine and plan cycle resets.
- billing.default_mode, billing.auto_fallback: existing billing controls.
- billing.reset_hour_utc (optional): integer [0-23] to prefer resets at a specific hour in UTC. Scheduler uses hourly sweeps and idempotency; exact hour is primarily informational.
- billing.reset_timezone (optional): TZ name like America/Los_Angeles; can be used for future scheduling rules.
- governance.enabled: enable governance features and RequestFlag cleanup.
- governance.flag_ttl_hours: default TTL hours for RequestFlag without explicit ttl_at (default: 24).
- public_logs.enabled and public_logs.retention_days: enable public logs and set deletion retention in days (default: 3).

Enable
- Environment variables (examples):
  - BILLING_ENABLED=true
  - BILLING_DEFAULT_MODE=balance
  - BILLING_AUTO_FALLBACK=false
  - GOVERNANCE_ENABLED=true
  - GOVERNANCE_FLAG_TTL_HOURS=24
  - PUBLIC_LOGS_ENABLED=true
  - PUBLIC_LOGS_RETENTION_DAYS=3
- Or update via the option/config UI which writes billing.*, governance.*, public_logs.* to the database.

Bootstrapping
- The scheduler starts automatically in main.go after the database and options are initialized.
- It uses hourly tickers and emits log lines like:
  - plan cycle reset: assignment=<id> metric=<metric>

Verification
- Unit tests: go test ./service/scheduler -run TestPlanResetJob
- Manual sqlite simulation:
  1. Create a PlanAssignment and a UsageCounter whose cycle_end is in the past.
  2. Call scheduler.RunPlanCycleResetOnce(context.Background()).
  3. Inspect the UsageCounter row: consumed_amount should be 0, cycle_start advanced to current.
- TTL example:
  - With governance.flag_ttl_hours=24, insert a RequestFlag with created_at older than 24h or ttl_at in the past and run scheduler.RunTTLCleanupOnce(...). The record should be deleted.
