package service

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestGenerateVoucherCode(t *testing.T) {
	prefix := "TEST"
	code := generateVoucherCode(prefix)
	if len(code) < len(prefix)+1 {
		t.Errorf("Expected code to be longer than prefix, got %s", code)
	}
	if code[:len(prefix)] != prefix {
		t.Errorf("Expected code to start with %s, got %s", prefix, code)
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"TEST-abc123", "TEST"},
		{"VCH-xyz789", "VCH"},
		{"NODELIMITER", "NODELIMITER"},
	}

	for _, tt := range tests {
		result := extractPrefix(tt.code)
		if result != tt.expected {
			t.Errorf("extractPrefix(%s) = %s; want %s", tt.code, result, tt.expected)
		}
	}
}

func TestGenerateVoucherBatch_InvalidCount(t *testing.T) {
	_, err := GenerateVoucherBatch("TEST", 0, common.VoucherGrantTypeCredit, 100, 0, 0, "admin", "test")
	if err == nil {
		t.Error("Expected error for count 0, got nil")
	}

	_, err = GenerateVoucherBatch("TEST", 1001, common.VoucherGrantTypeCredit, 100, 0, 0, "admin", "test")
	if err == nil {
		t.Error("Expected error for count > 1000, got nil")
	}
}

func TestGenerateVoucherBatch_InvalidGrantType(t *testing.T) {
	_, err := GenerateVoucherBatch("TEST", 10, "invalid", 100, 0, 0, "admin", "test")
	if err == nil {
		t.Error("Expected error for invalid grant type, got nil")
	}
}

func TestGenerateVoucherBatch_CreditValidation(t *testing.T) {
	_, err := GenerateVoucherBatch("TEST", 10, common.VoucherGrantTypeCredit, 0, 0, 0, "admin", "test")
	if err == nil {
		t.Error("Expected error for credit amount 0, got nil")
	}

	_, err = GenerateVoucherBatch("TEST", 10, common.VoucherGrantTypeCredit, -10, 0, 0, "admin", "test")
	if err == nil {
		t.Error("Expected error for negative credit amount, got nil")
	}
}

func TestGenerateVoucherBatch_PlanValidation(t *testing.T) {
	_, err := GenerateVoucherBatch("TEST", 10, common.VoucherGrantTypePlan, 0, 0, 0, "admin", "test")
	if err == nil {
		t.Error("Expected error for invalid plan ID, got nil")
	}

	_, err = GenerateVoucherBatch("TEST", 10, common.VoucherGrantTypePlan, 0, -1, 0, "admin", "test")
	if err == nil {
		t.Error("Expected error for negative plan ID, got nil")
	}
}

func TestRedeemVoucher_EmptyCode(t *testing.T) {
	_, err := RedeemVoucher("", 1, "testuser")
	if err == nil {
		t.Error("Expected error for empty voucher code, got nil")
	}
}

func TestRedeemVoucher_InvalidUserId(t *testing.T) {
	_, err := RedeemVoucher("TEST-abc123", 0, "testuser")
	if err == nil {
		t.Error("Expected error for invalid user ID, got nil")
	}

	_, err = RedeemVoucher("TEST-abc123", -1, "testuser")
	if err == nil {
		t.Error("Expected error for negative user ID, got nil")
	}
}
