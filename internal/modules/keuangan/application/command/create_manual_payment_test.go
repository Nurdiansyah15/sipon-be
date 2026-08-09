package command

import (
	"context"
	"testing"

	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accEntity "sipon-be/internal/modules/keuangan/domain/account/entity"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	invValue "sipon-be/internal/modules/keuangan/domain/invoice/valueobject"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	payValue "sipon-be/internal/modules/keuangan/domain/payment/valueobject"
	"sipon-be/internal/shared/kernel"
)

type fakeAccountRepo struct {
	acc *accEntity.Account
	err error
}

func (f *fakeAccountRepo) FindByID(ctx context.Context, id string) (*accEntity.Account, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.acc, nil
}

func (f *fakeAccountRepo) Save(ctx context.Context, acc *accEntity.Account) error { return nil }
func (f *fakeAccountRepo) Update(ctx context.Context, acc *accEntity.Account) error {
	return nil
}
func (f *fakeAccountRepo) FindByCode(ctx context.Context, code string) (*accEntity.Account, error) {
	return nil, nil
}
func (f *fakeAccountRepo) List(ctx context.Context, q accRepo.AccountListQuery) (*accRepo.AccountListResult, error) {
	return &accRepo.AccountListResult{}, nil
}
func (f *fakeAccountRepo) ListAll(ctx context.Context) ([]*accEntity.Account, error) { return nil, nil }
func (f *fakeAccountRepo) ListPostable(ctx context.Context) ([]*accEntity.Account, error) {
	return nil, nil
}
func (f *fakeAccountRepo) FindChildren(ctx context.Context, parentID string) ([]*accEntity.Account, error) {
	return nil, nil
}
func (f *fakeAccountRepo) HasJournalEntries(ctx context.Context, accountID string) (bool, error) {
	return false, nil
}
func (f *fakeAccountRepo) ExistsByCode(ctx context.Context, code string, excludeID string) (bool, error) {
	return false, nil
}

type fakeInvoiceRepo struct {
	err error
}

func (f *fakeInvoiceRepo) FindByID(ctx context.Context, id string) (*invEntity.Invoice, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &invEntity.Invoice{ID: id}, nil
}

func (f *fakeInvoiceRepo) Save(ctx context.Context, inv *invEntity.Invoice) error { return nil }
func (f *fakeInvoiceRepo) Update(ctx context.Context, inv *invEntity.Invoice) error {
	return nil
}
func (f *fakeInvoiceRepo) FindByNumber(ctx context.Context, number string) (*invEntity.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) List(ctx context.Context, q invRepo.InvoiceListQuery) (*invRepo.InvoiceListResult, error) {
	return &invRepo.InvoiceListResult{}, nil
}
func (f *fakeInvoiceRepo) FindBySantriComponentPeriod(ctx context.Context, santriID, feeComponentID, billingPeriodID string) (*invEntity.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) FindOutstandingBySantriID(ctx context.Context, santriID string) ([]*invEntity.Invoice, error) {
	return nil, nil
}
func (f *fakeInvoiceRepo) HasPaidComponent(ctx context.Context, santriID, componentCode, billingPeriodID string) (bool, error) {
	return false, nil
}
func (f *fakeInvoiceRepo) NextInvoiceNumber(ctx context.Context) (invValue.InvoiceNumber, error) {
	return invValue.InvoiceNumber{}, nil
}

type fakePaymentRepo struct {
	saved  *payEntity.Payment
	number payValue.PaymentNumber
}

func (f *fakePaymentRepo) Save(ctx context.Context, p *payEntity.Payment) error {
	f.saved = p
	return nil
}
func (f *fakePaymentRepo) Update(ctx context.Context, p *payEntity.Payment) error { return nil }
func (f *fakePaymentRepo) FindByID(ctx context.Context, id string) (*payEntity.Payment, error) {
	return nil, nil
}
func (f *fakePaymentRepo) FindByNumber(ctx context.Context, number string) (*payEntity.Payment, error) {
	return nil, nil
}
func (f *fakePaymentRepo) List(ctx context.Context, q payRepo.PaymentListQuery) (*payRepo.PaymentListResult, error) {
	return &payRepo.PaymentListResult{}, nil
}
func (f *fakePaymentRepo) FindByInvoiceID(ctx context.Context, invoiceID string) ([]*payEntity.Payment, error) {
	return nil, nil
}
func (f *fakePaymentRepo) FindVerifiedByInvoiceID(ctx context.Context, invoiceID string) ([]*payEntity.Payment, error) {
	return nil, nil
}
func (f *fakePaymentRepo) NextPaymentNumber(ctx context.Context) (payValue.PaymentNumber, error) {
	return f.number, nil
}

func TestCreateManualPaymentDebitAccountValidation(t *testing.T) {
	req := dto.CreateManualPaymentRequest{
		InvoiceID:      "inv-1",
		DebitAccountID: "acc-1",
		Amount:         100000,
		Method:         "cash",
		PaymentDate:    "2026-08-10",
	}

	tests := []struct {
		name    string
		account *accEntity.Account
		wantErr bool
	}{
		{
			name:    "cash bank account accepted",
			account: postableAccount(accConst.TypeAsset, accConst.SubTypeCashBank),
			wantErr: false,
		},
		{
			name:    "receivable account rejected",
			account: postableAccount(accConst.TypeAsset, accConst.SubTypeReceivable),
			wantErr: true,
		},
		{
			name:    "payable account rejected",
			account: postableAccount(accConst.TypeLiability, accConst.SubTypePayable),
			wantErr: true,
		},
		{
			name:    "capital account rejected",
			account: postableAccount(accConst.TypeEquity, accConst.SubTypeCapital),
			wantErr: true,
		},
		{
			name:    "revenue account rejected",
			account: postableAccount(accConst.TypeRevenue, accConst.SubTypeOperatingRevenue),
			wantErr: true,
		},
		{
			name:    "expense account rejected",
			account: postableAccount(accConst.TypeExpense, accConst.SubTypeOperatingExpense),
			wantErr: true,
		},
		{
			name:    "non-postable account rejected",
			account: &accEntity.Account{ID: "acc-1", Code: "1100", Name: "Aset Lancar", Type: accConst.TypeAsset, IsPostable: false, IsActive: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accRepo := &fakeAccountRepo{acc: tt.account}
			invoiceRepo := &fakeInvoiceRepo{}
			paymentRepo := &fakePaymentRepo{number: payValue.NewPaymentNumberNow(1)}
			uc := NewCreateManualPaymentUseCase(paymentRepo, invoiceRepo, accRepo)

			_, err := uc.Execute(context.Background(), req, "user-1")
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if paymentRepo.saved == nil {
				t.Error("payment should be saved")
			}
		})
	}
}

func TestCreateManualPaymentDebitAccountNotFound(t *testing.T) {
	accRepo := &fakeAccountRepo{acc: nil, err: kernel.WrapMsg(accConst.CodeAccountNotFound, "Akun tidak ditemukan", nil)}
	invoiceRepo := &fakeInvoiceRepo{}
	paymentRepo := &fakePaymentRepo{number: payValue.NewPaymentNumberNow(1)}
	uc := NewCreateManualPaymentUseCase(paymentRepo, invoiceRepo, accRepo)

	req := dto.CreateManualPaymentRequest{
		InvoiceID:      "inv-1",
		DebitAccountID: "missing",
		Amount:         100000,
		Method:         "cash",
		PaymentDate:    "2026-08-10",
	}
	if _, err := uc.Execute(context.Background(), req, "user-1"); err == nil {
		t.Error("expected error when debit account not found")
	}
}

func postableAccount(accType accConst.AccountType, subType accConst.AccountSubType) *accEntity.Account {
	return &accEntity.Account{
		ID: "acc-1", Code: "1101", Name: "Kas",
		Type: accType, SubType: &subType,
		IsPostable: true, IsActive: true,
	}
}
