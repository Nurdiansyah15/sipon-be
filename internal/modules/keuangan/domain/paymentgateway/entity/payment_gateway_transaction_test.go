package entity_test

import (
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
	"sipon-be/internal/modules/keuangan/domain/paymentgateway/entity"
)

func createTestGatewayTx() *entity.PaymentGatewayTransaction {
	tx, _ := entity.NewPaymentGatewayTransaction(
		"tx-1", "SIPON-TEST-1", "inv-1",
		150000, "snap-token", "https://redirect", nil,
		time.Now().Add(24*time.Hour),
	)
	return tx
}

func TestNewPaymentGatewayTransaction(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		transactionID string
		invoiceID     string
		amount        float64
		snapToken     string
		wantErr       bool
	}{
		{"valid", "tx-1", "SIPON-1", "inv-1", 100000, "token", false},
		{"empty id", "", "SIPON-1", "inv-1", 100000, "token", true},
		{"empty transactionID", "tx-1", "", "inv-1", 100000, "token", true},
		{"empty invoiceID", "tx-1", "SIPON-1", "", 100000, "token", true},
		{"zero amount", "tx-1", "SIPON-1", "inv-1", 0, "token", true},
		{"empty snapToken", "tx-1", "SIPON-1", "inv-1", 100000, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := entity.NewPaymentGatewayTransaction(
				tt.id, tt.transactionID, tt.invoiceID, tt.amount, tt.snapToken,
				"https://redirect", nil, time.Now().Add(time.Hour),
			)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tx.Status != constant.GatewayStatusPending {
				t.Errorf("expected pending status, got %s", tx.Status)
			}
		})
	}
}

func TestApplyNotificationSettlement(t *testing.T) {
	tx := createTestGatewayTx()
	pm := "bank_transfer"

	if err := tx.ApplyNotification(constant.GatewayStatusSettlement, &pm, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Status != constant.GatewayStatusSettlement {
		t.Errorf("expected settlement, got %s", tx.Status)
	}
	if tx.PaymentMethod == nil || *tx.PaymentMethod != "bank_transfer" {
		t.Error("payment method should be recorded")
	}
}

func TestApplyNotificationNeverRegressesAfterSuccess(t *testing.T) {
	tx := createTestGatewayTx()
	pm := "bank_transfer"

	if err := tx.ApplyNotification(constant.GatewayStatusSettlement, &pm, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Notifikasi gagal yang datang setelah settlement harus diabaikan.
	denyPm := "bank_transfer"
	if err := tx.ApplyNotification(constant.GatewayStatusDeny, &denyPm, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Status != constant.GatewayStatusSettlement {
		t.Errorf("status must not regress after success, got %s", tx.Status)
	}
}

func TestApplyNotificationFinalIsSticky(t *testing.T) {
	tx := createTestGatewayTx()
	pm := "bank_transfer"

	if err := tx.ApplyNotification(constant.GatewayStatusExpire, &pm, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := tx.ApplyNotification(constant.GatewayStatusSettlement, &pm, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Status != constant.GatewayStatusExpire {
		t.Errorf("final status must be sticky, got %s", tx.Status)
	}
}

func TestApplyNotificationInvalidStatus(t *testing.T) {
	tx := createTestGatewayTx()
	pm := "bank_transfer"
	if err := tx.ApplyNotification("bogus", &pm, nil); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestLinkPaymentOnce(t *testing.T) {
	tx := createTestGatewayTx()
	if err := tx.LinkPayment("pay-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.PaymentID == nil || *tx.PaymentID != "pay-1" {
		t.Error("payment id should be linked")
	}
	if err := tx.LinkPayment("pay-2"); err == nil {
		t.Error("expected error when linking twice")
	}
}

func TestMarkRejected(t *testing.T) {
	tx := createTestGatewayTx()
	if err := tx.MarkRejected(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Status != constant.GatewayStatusCancel {
		t.Errorf("expected cancel, got %s", tx.Status)
	}

	// Tidak boleh mengubah transaksi yang sudah settlement.
	tx2 := createTestGatewayTx()
	pm := "bank_transfer"
	_ = tx2.ApplyNotification(constant.GatewayStatusSettlement, &pm, nil)
	if err := tx2.MarkRejected(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx2.Status != constant.GatewayStatusSettlement {
		t.Errorf("settlement must not be rejected, got %s", tx2.Status)
	}
}

func TestIsSuccess(t *testing.T) {
	for _, s := range []constant.PaymentGatewayStatus{
		constant.GatewayStatusSettlement,
		constant.GatewayStatusCapture,
	} {
		if !s.IsSuccess() {
			t.Errorf("expected %s to be success", s)
		}
	}
	for _, s := range []constant.PaymentGatewayStatus{
		constant.GatewayStatusPending,
		constant.GatewayStatusPendingChallenge,
		constant.GatewayStatusDeny,
		constant.GatewayStatusFailure,
		constant.GatewayStatusExpire,
		constant.GatewayStatusCancel,
	} {
		if s.IsSuccess() {
			t.Errorf("expected %s to NOT be success", s)
		}
	}
}
