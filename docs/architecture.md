# Billing Engine Architecture (M2)

This document summarizes the core billing engine design introduced in milestone M2. It enables plan-based allowance management with balance fallback, atomic charging with audit logs, and idempotent/concurrent safety.

Key components
- service/billing_engine.go: BillingEngine orchestrates Prepare → Commit → Rollback flow.
- service/billing_errors.go: Typed errors for plan exhaustion and balance insufficiency.
- service/billing_audit.go: Helper to build RequestLog audit entries while keeping legacy Log compatible.
- model helpers: plan_assignment.go and usage_counter.go gained transactional helpers for safe updates.

Decision flow
1) Prepare
   - Input: subject (user or token), RelayInfo.
   - Resolve active PlanAssignment(s) at “now”.
   - Load the attached Plan, derive current cycle window (daily/monthly/custom) and fetch the UsageCounter for the window.
   - Compute remaining allowance = plan quota (+ rollover) − consumed. The engine returns a PreparedCharge describing
     - subject, selected PlanAssignment, mode (plan|balance|auto), cycle window and allowance.

2) Commit
   - Idempotency barrier: create model.RequestLog first using the request_id as unique key. If it already exists,
     the commit becomes a no-op and returns success.
   - If the mode is auto/fallback and allowance is insufficient, commit transparently switches to balance mode.
   - Plan mode: lock the counter row (if exists) and verify remaining allowance. Increment the UsageCounter within the same DB transaction.
   - Balance mode: row-lock user (and token if needed), perform conditional decrement (quota >= amount) to avoid going negative.
   - The RequestLog and all updates are committed atomically.

3) Rollback
   - Best-effort reversal using request_id. Currently implemented as deleting the RequestLog so a subsequent retry can proceed.
     Precise reversal can be extended using richer audit metadata.

Concurrency and idempotency
- Redis locks can be added as a first line of defense when available, but DB transactions are the source of truth.
- The RequestLog unique index serves as a robust idempotency key (subject + request_id). Only one commit per request_id can succeed.
- UsageCounter updates run inside the same transaction with row-level locks (FOR UPDATE where supported).

Model helpers
- model.GetActivePlanAssignmentsTx: transactional loader with optional FOR UPDATE.
- model.GetUsageCounterTx: transactional Select … FOR UPDATE for an existing counter row.
- model.IncrementUsageCounterTx: upsert-with-increment within a transaction.

Public API hook
- service/billing_gate.go exposes Prepare(ctx, relayInfo). When common.BillingFeatureEnabled is false, the gate returns nil and leaves existing quota code paths unchanged.

Testing
- Focused unit tests in service/billing_engine_test.go cover: plan-only, balance-only, auto fallback, idempotent retries, and concurrency guard.
- Model tests in model/usage_counter_test.go verify usage counter increments.

Notes
- The initial implementation covers daily/monthly cycles and simple rollover amount. Scheduling of cycle resets or complex rollover policies can be built on top.
- Upstream provider and model alias are recorded in RequestLog for future analytics.
