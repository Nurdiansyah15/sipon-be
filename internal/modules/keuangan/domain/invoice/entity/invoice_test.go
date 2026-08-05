package entity_test

import (
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/domain/invoice/constant"
	"sipon-be/internal/modules/keuangan/domain/invoice/entity"
)

func createTestInvoice() *entity.Invoice {
	inv, _ := entity.NewInvoice(
		"inv-1", "INV/2026/08/000001", "santri-1", "user-1",
		"fc-1", "2026-08", "2025/2026",
		500000, time.Now().AddDate(0, 1, 0), "user-1",
	)
	return inv
}

func TestNewInvoice(t *testing.T) {
	inv := createTestInvoice()
	if inv.Status != constant.StatusDraft {
		t.Errorf("expected draft status, got %s", inv.Status)
	}
	if inv.Amount != 500000 {
		t.Errorf("expected amount 500000, got %f", inv.Amount)
	}
}

func TestInvoiceIssue(t *testing.T) {
	inv := createTestInvoice()
	if err := inv.Issue(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != constant.StatusIssued {
		t.Errorf("expected issued status, got %s", inv.Status)
	}
	if inv.IssuedAt == nil {
		t.Error("issuedAt should not be nil after issuing")
	}

	if err := inv.Issue(); err == nil {
		t.Error("expected error when issuing twice")
	}
}

func TestInvoiceAddPayment(t *testing.T) {
	inv := createTestInvoice()
	if err := inv.Issue(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := inv.AddPayment(200000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != constant.StatusPartial {
		t.Errorf("expected partial status, got %s", inv.Status)
	}
	if inv.PaidAmount != 200000 {
		t.Errorf("expected paidAmount 200000, got %f", inv.PaidAmount)
	}

	if err := inv.AddPayment(300000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != constant.StatusPaid {
		t.Errorf("expected paid status, got %s", inv.Status)
	}
}

func TestInvoiceExpire(t *testing.T) {
	inv := createTestInvoice()
	if err := inv.Issue(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := inv.Expire(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != constant.StatusExpired {
		t.Errorf("expected expired status, got %s", inv.Status)
	}

	if err := inv.Expire(); err == nil {
		t.Error("expected error when expiring twice")
	}
}

func TestInvoiceCancel(t *testing.T) {
	inv := createTestInvoice()
	if err := inv.Cancel(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.Status != constant.StatusCancelled {
		t.Errorf("expected cancelled status, got %s", inv.Status)
	}

	inv2 := createTestInvoice()
	if err := inv2.Issue(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := inv2.Cancel(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	inv3 := createTestInvoice()
	if err := inv3.Issue(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := inv3.AddPayment(500000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := inv3.Cancel(); err == nil {
		t.Error("expected error when cancelling paid invoice")
	}
}

func TestInvoiceRemainingAmount(t *testing.T) {
	inv := createTestInvoice()
	if err := inv.Issue(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	inv.ApplyDiscount(50000)

	if inv.RemainingAmount() != 450000 {
		t.Errorf("expected 450000 remaining, got %f", inv.RemainingAmount())
	}

	if err := inv.AddPayment(450000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.RemainingAmount() != 0 {
		t.Errorf("expected 0 remaining, got %f", inv.RemainingAmount())
	}
}

func TestInvoiceHasOutstanding(t *testing.T) {
	inv := createTestInvoice()
	if inv.HasOutstanding() {
		t.Error("draft should not have outstanding")
	}
	if err := inv.Issue(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !inv.HasOutstanding() {
		t.Error("issued should have outstanding")
	}
	if err := inv.AddPayment(500000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.HasOutstanding() {
		t.Error("paid should not have outstanding")
	}
}
