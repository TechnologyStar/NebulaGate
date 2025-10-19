package service

import (
    "net/http/httptest"
    "testing"
    "time"

    "github.com/QuantumNous/new-api/common"
    "github.com/QuantumNous/new-api/model"
    relaycommon "github.com/QuantumNous/new-api/relay/common"
    "github.com/gin-gonic/gin"
)

// TestBillingPreConsume_PlanExhaustion_ReturnsError verifies that when billing is enabled and
// the selected mode is plan without fallback, pre-consume rejects if allowance is insufficient.
func TestBillingPreConsume_PlanExhaustion_ReturnsError(t *testing.T) {
    db := setupServiceTestDB(t)
    // Feature flags
    common.BillingFeatureEnabled = true
    common.BillingDefaultMode = common.BillingModePlan
    common.BillingAutoFallbackEnabled = false

    // Create user & token
    user, token := createUserAndToken(t, db, 100, false)
    // Create plan with small allowance (e.g., 5 requests/tokens)
    _, assignment := createPlanAndAssignment(t, db, user.Id, 5, common.BillingModePlan, false)
    _ = assignment

    // Minimal gin context
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    // Minimal relay info
    info := &relaycommon.RelayInfo{UserId: user.Id, TokenId: token.Id, TokenUnlimited: token.UnlimitedQuota, StartTime: time.Now(), BillingMode: common.BillingModePlan}

    // Ask to pre-consume more than allowance
    err := PreConsumeQuota(c, 10, info)
    if err == nil {
        t.Fatalf("expected pre-consume error due to plan exhaustion, got nil")
    }
    if err.StatusCode != 402 {
        t.Fatalf("expected HTTP 402 Payment Required, got %d (%s)", err.StatusCode, err.Error())
    }
}

// TestBillingPreConsume_PlanFallback_AllowsRequest verifies that with fallback enabled
// the pre-consume passes even when the plan allowance is insufficient.
func TestBillingPreConsume_PlanFallback_AllowsRequest(t *testing.T) {
    db := setupServiceTestDB(t)
    // Feature flags
    common.BillingFeatureEnabled = true
    common.BillingDefaultMode = common.BillingModePlan
    common.BillingAutoFallbackEnabled = false

    // Create user & token
    user, token := createUserAndToken(t, db, 100, false)
    // Create plan with small allowance and assignment enabling fallback
    _, assignment := createPlanAndAssignment(t, db, user.Id, 5, common.BillingModePlan, true)
    _ = assignment

    // Minimal gin context
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    info := &relaycommon.RelayInfo{UserId: user.Id, TokenId: token.Id, TokenUnlimited: token.UnlimitedQuota, StartTime: time.Now(), BillingMode: common.BillingModePlan}

    err := PreConsumeQuota(c, 10, info)
    if err != nil {
        t.Fatalf("expected pre-consume success with fallback enabled, got error: %v", err)
    }
    // Verify prepared was attached for later commit
    if _, ok := c.Get("billing_prepared"); !ok {
        t.Fatalf("expected billing_prepared in context")
    }
    if !info.BillingFeatureEnabled {
        t.Fatalf("expected relayInfo.BillingFeatureEnabled=true")
    }
}
