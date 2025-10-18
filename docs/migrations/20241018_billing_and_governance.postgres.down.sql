BEGIN;

DROP TABLE IF EXISTS request_aggregates;
DROP TABLE IF EXISTS request_logs;
DROP TABLE IF EXISTS request_flags;
DROP TABLE IF EXISTS voucher_redemptions;
DROP TABLE IF EXISTS voucher_batches;
DROP TABLE IF EXISTS usage_counters;
DROP TABLE IF EXISTS plan_assignments;
DROP TABLE IF EXISTS plans;

COMMIT;
