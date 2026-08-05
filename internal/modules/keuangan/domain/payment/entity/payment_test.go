package entity_test

import (
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/domain/payment/constant"
	"sipon-be/internal/modules/keuangan/domain/payment/entity"
)

func createTestPayment() *entity.Payment {
	p, _ := entity.NewPayment(
		"pay-1", "PAY/2026/08/000001", "inv-1",
		150000, constant.MethodTransfer, time.Now(),
		nil, nil, nil, nil, "user-1",
	)
	return p
}

func TestNewPayment(t *testing.T) {
	debitAcc := "acc-1"
	tests := []struct {
		name          string
		id            string
		paymentNumber string
		invoiceID     string
		amount        float64
		method        constant.PaymentMethod
		createdBy     string
		wantErr       bool
	}{
		{"valid payment", "pay-1", "PAY/001", "inv-1", 100000, constant.MethodCash, "user-1", false},
		{"empty id", "", "PAY/001", "inv-1", 100000, constant.MethodCash, "user-1", true},
		{"empty paymentNumber", "pay-2", "", "inv-1", 100000, constant.MethodCash, "user-1", true},
		{"empty invoiceID", "pay-3", "PAY/001", "", 100000, constant.MethodCash, "user-1", true},
		{"empty createdBy", "pay-4", "PAY/001", "inv-1", 100000, constant.MethodCash, "", true},
		{"zero amount", "pay-5", "PAY/001", "inv-1", 0, constant.MethodCash, "user-1", true},
		{"negative amount", "pay-6", "PAY/001", "inv-1", -50000, constant.MethodCash, "user-1", true},
		{"with debit account", "pay-7", "PAY/001", "inv-1", 100000, constant.MethodTransfer, "user-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var da *string
			if tt.name == "with debit account" {
				da = &debitAcc
			}
			p, err := entity.NewPayment(tt.id, tt.paymentNumber, tt.invoiceID, tt.amount, tt.method, time.Now(), da, nil, nil, nil, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Status != constant.PaymentPending {
				t.Errorf("expected pending status, got %s", p.Status)
			}
		})
	}
}

func TestPaymentVerify(t *testing.T) {
	payment := createTestPayment()

	if err := payment.Verify("verifier-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != constant.PaymentVerified {
		t.Errorf("expected verified status, got %s", payment.Status)
	}
	if payment.VerifiedBy == nil || *payment.VerifiedBy != "verifier-1" {
		t.Error("verifiedBy should be set")
	}
	if payment.VerifiedAt == nil {
		t.Error("verifiedAt should be set")
	}

	if err := payment.Verify("verifier-2"); err == nil {
		t.Error("expected error when verifying already verified payment")
	}
}

func TestPaymentVerifyRejected(t *testing.T) {
	payment := createTestPayment()
	if err := payment.Reject(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := payment.Verify("verifier-1"); err == nil {
		t.Error("expected error when verifying rejected payment")
	}
}

func TestPaymentReject(t *testing.T) {
	payment := createTestPayment()

	if err := payment.Reject(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.Status != constant.PaymentRejected {
		t.Errorf("expected rejected status, got %s", payment.Status)
	}

	if err := payment.Reject(); err == nil {
		t.Error("expected error when rejecting already rejected payment")
	}
}

func TestPaymentRejectVerified(t *testing.T) {
	payment := createTestPayment()
	if err := payment.Verify("verifier-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := payment.Reject(); err == nil {
		t.Error("expected error when rejecting verified payment")
	}
}
