package service

import (
	"context"

	"github.com/QuantumNous/new-api/common"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
)

// Prepare is a small wrapper for middleware to get billing decision when billing is enabled.
// When billing is disabled, it returns nil prepared decision and leaves existing quota logic intact.
func Prepare(ctx context.Context, relayInfo *relaycommon.RelayInfo) (*PreparedCharge, error) {
	if !common.BillingFeatureEnabled {
		return nil, nil
	}
	engine := NewBillingEngine(nil)
	return engine.PrepareCharge(ctx, "", relayInfo)
}
