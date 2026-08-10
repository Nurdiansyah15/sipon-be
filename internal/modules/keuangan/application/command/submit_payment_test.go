package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	invValue "sipon-be/internal/modules/keuangan/domain/invoice/valueobject"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	payValue "sipon-be/internal/modules/keuangan/domain/payment/valueobject"
	"sipon-be/internal/shared/kernel"
)

type submitInvoiceRepo struct {
	inv *invEntity.Invoice
}

func (s *submitInvoiceRepo) FindByID(ctx context.Context, id string) (*invEntity.Invoice, error) {
	if s.inv == nil {
		return nil, kernel.WrapMsg(invConst.CodeInvoiceNotFound, "Invoice tidak ditemukan", nil)
	}
	return s.inv, nil
}
func (s *submitInvoiceRepo) Save(ctx context.Context, inv *invEntity.Invoice) error { return nil }
func (s *submitInvoiceRepo) Update(ctx context.Context, inv *invEntity.Invoice) error {
	return nil
}
func (s *submitInvoiceRepo) FindByNumber(ctx context.Context, number string) (*invEntity.Invoice, error) {
	return nil, nil
}
func (s *submitInvoiceRepo) List(ctx context.Context, q invRepo.InvoiceListQuery) (*invRepo.InvoiceListResult, error) {
	return &invRepo.InvoiceListResult{}, nil
}
func (s *submitInvoiceRepo) FindBySantriComponentPeriod(ctx context.Context, santriID, feeComponentID, billingPeriodID string) (*invEntity.Invoice, error) {
	return nil, nil
}
func (s *submitInvoiceRepo) FindOutstandingBySantriID(ctx context.Context, santriID string) ([]*invEntity.Invoice, error) {
	return nil, nil
}
func (s *submitInvoiceRepo) FindSummaryByUserID(ctx context.Context, userID string) (*invRepo.InvoiceSummary, error) {
	return &invRepo.InvoiceSummary{}, nil
}
func (s *submitInvoiceRepo) HasPaidComponent(ctx context.Context, santriID, componentCode, billingPeriodID string) (bool, error) {
	return false, nil
}
func (s *submitInvoiceRepo) NextInvoiceNumber(ctx context.Context) (invValue.InvoiceNumber, error) {
	return invValue.InvoiceNumber{}, nil
}

type submitPaymentRepo struct {
	saved *payEntity.Payment
}

func (s *submitPaymentRepo) FindByID(ctx context.Context, id string) (*payEntity.Payment, error) {
	return nil, nil
}
func (s *submitPaymentRepo) Save(ctx context.Context, p *payEntity.Payment) error {
	s.saved = p
	return nil
}
func (s *submitPaymentRepo) Update(ctx context.Context, p *payEntity.Payment) error { return nil }
func (s *submitPaymentRepo) FindByNumber(ctx context.Context, number string) (*payEntity.Payment, error) {
	return nil, nil
}
func (s *submitPaymentRepo) List(ctx context.Context, q payRepo.PaymentListQuery) (*payRepo.PaymentListResult, error) {
	return &payRepo.PaymentListResult{}, nil
}
func (s *submitPaymentRepo) FindByInvoiceID(ctx context.Context, invoiceID string) ([]*payEntity.Payment, error) {
	return nil, nil
}
func (s *submitPaymentRepo) FindVerifiedByInvoiceID(ctx context.Context, invoiceID string) ([]*payEntity.Payment, error) {
	return nil, nil
}
func (s *submitPaymentRepo) NextPaymentNumber(ctx context.Context) (payValue.PaymentNumber, error) {
	return payValue.NewPaymentNumber("2026", "08", 1), nil
}

func newIssuedInvoice(amount float64, userID string) *invEntity.Invoice {
	inv, _ := invEntity.NewInvoice(
		"inv-1", "INV-1", "santri-1", userID, "fc-1", nil, amount,
		time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), "user-1",
	)
	_ = inv.Issue(time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	return inv
}

func TestSubmitPaymentSuccess(t *testing.T) {
	inv := newIssuedInvoice(150000, "user-1")
	payRepo := &submitPaymentRepo{}
	uc := NewSubmitPaymentUseCase(payRepo, &submitInvoiceRepo{inv: inv})

	proofKey := "payment-proofs/user-1/x.jpg"
	resp, err := uc.Execute(context.Background(), "user-1", dto.SubmitPaymentRequest{
		InvoiceID:   "inv-1",
		Amount:      100000,
		Method:      "transfer",
		PaymentDate: "2026-08-12",
		ProofKey:    proofKey,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "pending" {
		t.Errorf("expected status pending, got %s", resp.Status)
	}
	if payRepo.saved == nil {
		t.Fatal("expected payment to be saved")
	}
	if payRepo.saved.DebitAccountID != nil {
		t.Errorf("expected debit_account_id nil on submit, got %v", *payRepo.saved.DebitAccountID)
	}
	if payRepo.saved.ProofKey == nil || *payRepo.saved.ProofKey != proofKey {
		t.Errorf("expected proof key %s, got %v", proofKey, payRepo.saved.ProofKey)
	}
}

func TestSubmitPaymentNotOwner(t *testing.T) {
	inv := newIssuedInvoice(150000, "user-1")
	uc := NewSubmitPaymentUseCase(&submitPaymentRepo{}, &submitInvoiceRepo{inv: inv})

	_, err := uc.Execute(context.Background(), "user-2", dto.SubmitPaymentRequest{
		InvoiceID:   "inv-1",
		Amount:      100000,
		Method:      "transfer",
		PaymentDate: "2026-08-12",
		ProofKey:    "key",
	})
	if err == nil {
		t.Fatal("expected forbidden error")
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kernel.AppError, got %T", err)
	}
	if ke.Code != application.ErrCodeForbidden {
		t.Errorf("expected code %s, got %s", application.ErrCodeForbidden, ke.Code)
	}
}

func TestSubmitPaymentExceedsOutstanding(t *testing.T) {
	inv := newIssuedInvoice(150000, "user-1")
	uc := NewSubmitPaymentUseCase(&submitPaymentRepo{}, &submitInvoiceRepo{inv: inv})

	_, err := uc.Execute(context.Background(), "user-1", dto.SubmitPaymentRequest{
		InvoiceID:   "inv-1",
		Amount:      200000,
		Method:      "transfer",
		PaymentDate: "2026-08-12",
		ProofKey:    "key",
	})
	if err == nil {
		t.Fatal("expected error when amount exceeds outstanding")
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kernel.AppError, got %T", err)
	}
	if ke.Code != application.ErrCodeBadRequest {
		t.Errorf("expected code %s, got %s", application.ErrCodeBadRequest, ke.Code)
	}
}
