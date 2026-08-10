package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accEntity "sipon-be/internal/modules/keuangan/domain/account/entity"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	feeEntity "sipon-be/internal/modules/keuangan/domain/feecomponent/entity"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	invValue "sipon-be/internal/modules/keuangan/domain/invoice/valueobject"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalEntity "sipon-be/internal/modules/keuangan/domain/journal/entity"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	journalVO "sipon-be/internal/modules/keuangan/domain/journal/valueobject"
	journalService "sipon-be/internal/modules/keuangan/domain/journal/service"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	payValue "sipon-be/internal/modules/keuangan/domain/payment/valueobject"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodEntity "sipon-be/internal/modules/keuangan/domain/period/entity"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type verifyAccountRepo struct {
	byID map[string]*accEntity.Account
}

func (v *verifyAccountRepo) FindByID(ctx context.Context, id string) (*accEntity.Account, error) {
	if acc, ok := v.byID[id]; ok {
		return acc, nil
	}
	return nil, kernel.WrapMsg(accConst.CodeAccountNotFound, "Akun tidak ditemukan", nil)
}
func (v *verifyAccountRepo) Save(ctx context.Context, acc *accEntity.Account) error { return nil }
func (v *verifyAccountRepo) Update(ctx context.Context, acc *accEntity.Account) error {
	return nil
}
func (v *verifyAccountRepo) FindByCode(ctx context.Context, code string) (*accEntity.Account, error) {
	return nil, kernel.WrapMsg(accConst.CodeAccountNotFound, "Akun tidak ditemukan", nil)
}
func (v *verifyAccountRepo) List(ctx context.Context, q accRepo.AccountListQuery) (*accRepo.AccountListResult, error) {
	return &accRepo.AccountListResult{}, nil
}
func (v *verifyAccountRepo) ListAll(ctx context.Context) ([]*accEntity.Account, error) { return nil, nil }
func (v *verifyAccountRepo) ListPostable(ctx context.Context) ([]*accEntity.Account, error) {
	return nil, nil
}
func (v *verifyAccountRepo) FindChildren(ctx context.Context, parentID string) ([]*accEntity.Account, error) {
	return nil, nil
}
func (v *verifyAccountRepo) HasJournalEntries(ctx context.Context, accountID string) (bool, error) {
	return false, nil
}
func (v *verifyAccountRepo) ExistsByCode(ctx context.Context, code, excludeID string) (bool, error) {
	return false, nil
}

type verifyJournalRepo struct {
	savedEntry *journalEntity.JournalEntry
}

func (v *verifyJournalRepo) Save(ctx context.Context, entry *journalEntity.JournalEntry) error {
	v.savedEntry = entry
	return nil
}
func (v *verifyJournalRepo) Update(ctx context.Context, entry *journalEntity.JournalEntry) error {
	return nil
}
func (v *verifyJournalRepo) FindByID(ctx context.Context, id string) (*journalEntity.JournalEntry, error) {
	return nil, kernel.WrapMsg(journalConst.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
}
func (v *verifyJournalRepo) FindByNumber(ctx context.Context, number string) (*journalEntity.JournalEntry, error) {
	return nil, kernel.WrapMsg(journalConst.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
}
func (v *verifyJournalRepo) NextJournalNumber(ctx context.Context) (journalVO.JournalNumber, error) {
	return journalVO.NewJournalNumber("2026", "08", 1), nil
}
func (v *verifyJournalRepo) List(ctx context.Context, q journalRepo.JournalListQuery) (*journalRepo.JournalListResult, error) {
	return nil, nil
}
func (v *verifyJournalRepo) FindBySource(ctx context.Context, sourceType, sourceID string) (*journalEntity.JournalEntry, error) {
	return nil, kernel.WrapMsg(journalConst.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
}
func (v *verifyJournalRepo) SaveLines(ctx context.Context, entryID string, lines []*journalEntity.JournalEntryLine) error {
	return nil
}
func (v *verifyJournalRepo) FindLinesByEntryID(ctx context.Context, entryID string) ([]*journalEntity.JournalEntryLine, error) {
	return nil, nil
}
func (v *verifyJournalRepo) ComputeAccountBalances(ctx context.Context, periodID string) (map[string]journalRepo.AccountBalance, error) {
	return nil, nil
}

type verifyPeriodRepo struct {
	period *periodEntity.AccountingPeriod
}

func (v *verifyPeriodRepo) Save(ctx context.Context, p *periodEntity.AccountingPeriod) error {
	return nil
}
func (v *verifyPeriodRepo) Update(ctx context.Context, p *periodEntity.AccountingPeriod) error {
	return nil
}
func (v *verifyPeriodRepo) FindByID(ctx context.Context, id string) (*periodEntity.AccountingPeriod, error) {
	return nil, kernel.WrapMsg(periodConst.CodePeriodNotFound, "Periode tidak ditemukan", nil)
}
func (v *verifyPeriodRepo) FindActive(ctx context.Context) (*periodEntity.AccountingPeriod, error) {
	return v.period, nil
}
func (v *verifyPeriodRepo) List(ctx context.Context, q periodRepo.PeriodListQuery) (*periodRepo.PeriodListResult, error) {
	return nil, nil
}
func (v *verifyPeriodRepo) FindByDate(ctx context.Context, date time.Time) (*periodEntity.AccountingPeriod, error) {
	return v.period, nil
}
func (v *verifyPeriodRepo) HasOverlap(ctx context.Context, startDate, endDate time.Time, excludeID string) (bool, error) {
	return false, nil
}

type verifyInvoiceRepo struct {
	inv *invEntity.Invoice
	err error
}

func (v *verifyInvoiceRepo) FindByID(ctx context.Context, id string) (*invEntity.Invoice, error) {
	if v.err != nil {
		return nil, v.err
	}
	return v.inv, nil
}
func (v *verifyInvoiceRepo) Save(ctx context.Context, inv *invEntity.Invoice) error { return nil }
func (v *verifyInvoiceRepo) Update(ctx context.Context, inv *invEntity.Invoice) error {
	return nil
}
func (v *verifyInvoiceRepo) FindByNumber(ctx context.Context, number string) (*invEntity.Invoice, error) {
	return nil, nil
}
func (v *verifyInvoiceRepo) List(ctx context.Context, q invRepo.InvoiceListQuery) (*invRepo.InvoiceListResult, error) {
	return &invRepo.InvoiceListResult{}, nil
}
func (v *verifyInvoiceRepo) FindBySantriComponentPeriod(ctx context.Context, santriID, feeComponentID, billingPeriodID string) (*invEntity.Invoice, error) {
	return nil, nil
}
func (v *verifyInvoiceRepo) FindOutstandingBySantriID(ctx context.Context, santriID string) ([]*invEntity.Invoice, error) {
	return nil, nil
}
func (v *verifyInvoiceRepo) FindSummaryByUserID(ctx context.Context, userID string) (*invRepo.InvoiceSummary, error) {
	return &invRepo.InvoiceSummary{}, nil
}
func (v *verifyInvoiceRepo) HasPaidComponent(ctx context.Context, santriID, componentCode, billingPeriodID string) (bool, error) {
	return false, nil
}
func (v *verifyInvoiceRepo) NextInvoiceNumber(ctx context.Context) (invValue.InvoiceNumber, error) {
	return invValue.InvoiceNumber{}, nil
}

type verifyPaymentRepo struct {
	payment *payEntity.Payment
}

func (v *verifyPaymentRepo) FindByID(ctx context.Context, id string) (*payEntity.Payment, error) {
	return v.payment, nil
}
func (v *verifyPaymentRepo) Save(ctx context.Context, p *payEntity.Payment) error { return nil }
func (v *verifyPaymentRepo) Update(ctx context.Context, p *payEntity.Payment) error {
	return nil
}
func (v *verifyPaymentRepo) FindByNumber(ctx context.Context, number string) (*payEntity.Payment, error) {
	return nil, nil
}
func (v *verifyPaymentRepo) List(ctx context.Context, q payRepo.PaymentListQuery) (*payRepo.PaymentListResult, error) {
	return &payRepo.PaymentListResult{}, nil
}
func (v *verifyPaymentRepo) FindByInvoiceID(ctx context.Context, invoiceID string) ([]*payEntity.Payment, error) {
	return nil, nil
}
func (v *verifyPaymentRepo) FindVerifiedByInvoiceID(ctx context.Context, invoiceID string) ([]*payEntity.Payment, error) {
	return nil, nil
}
func (v *verifyPaymentRepo) NextPaymentNumber(ctx context.Context) (payValue.PaymentNumber, error) {
	return payValue.PaymentNumber{}, nil
}

type verifyFeeRepo struct {
	fee *feeEntity.FeeComponent
	err error
}

func (v *verifyFeeRepo) FindByID(ctx context.Context, id string) (*feeEntity.FeeComponent, error) {
	if v.err != nil {
		return nil, v.err
	}
	return v.fee, nil
}
func (v *verifyFeeRepo) Save(ctx context.Context, fc *feeEntity.FeeComponent) error { return nil }
func (v *verifyFeeRepo) Update(ctx context.Context, fc *feeEntity.FeeComponent) error {
	return nil
}
func (v *verifyFeeRepo) FindByCode(ctx context.Context, code string) (*feeEntity.FeeComponent, error) {
	return nil, nil
}
func (v *verifyFeeRepo) List(ctx context.Context, q feeRepo.FeeComponentListQuery) (*feeRepo.FeeComponentListResult, error) {
	return &feeRepo.FeeComponentListResult{}, nil
}
func (v *verifyFeeRepo) ExistsByCode(ctx context.Context, code, excludeID string) (bool, error) {
	return false, nil
}

type verifyTransactor struct{}

func (v *verifyTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func stPtr(st accConst.AccountSubType) *accConst.AccountSubType {
	return &st
}

func TestVerifyPaymentPostsReceivableFromFeeComponent(t *testing.T) {
	debitID := "cash-1"
	payment, err := payEntity.NewPayment(
		"pay-1", "PAY-1", "inv-1", 150000, payConst.MethodCash,
		time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC),
		&debitID, nil, nil, nil, "user-1",
	)
	if err != nil {
		t.Fatalf("unexpected error creating payment: %v", err)
	}

	inv, err := invEntity.NewInvoice(
		"inv-1", "INV-1", "santri-1", "user-1", "fc-1", nil, 150000,
		time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), "user-1",
	)
	if err != nil {
		t.Fatalf("unexpected error creating invoice: %v", err)
	}
	if err := inv.Issue(time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("unexpected error issuing invoice: %v", err)
	}

	fee, err := feeEntity.NewFeeComponent("fc-1", "SPP", "SPP Bulanan", "rev-spp", "recv-spp", 150000, "user-1")
	if err != nil {
		t.Fatalf("unexpected error creating fee component: %v", err)
	}

	cash := &accEntity.Account{ID: "cash-1", Code: "1101", Name: "Kas", Type: accConst.TypeAsset, SubType: stPtr(accConst.SubTypeCashBank), IsPostable: true, IsActive: true}
	recv := &accEntity.Account{ID: "recv-spp", Code: "1104", Name: "Piutang SPP", Type: accConst.TypeAsset, SubType: stPtr(accConst.SubTypeReceivable), IsPostable: true, IsActive: true}

	accountRepo := &verifyAccountRepo{byID: map[string]*accEntity.Account{cash.ID: cash, recv.ID: recv}}
	journalRepo := &verifyJournalRepo{}
	periodRepo := &verifyPeriodRepo{
		period: &periodEntity.AccountingPeriod{
			ID: "period-1", Name: "Agustus 2026", Status: periodConst.PeriodOpen,
			StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	autoPosting := journalService.NewAutoPostingService(journalRepo, accountRepo, periodRepo)
	uc := NewVerifyPaymentUseCase(
		&verifyPaymentRepo{payment: payment},
		&verifyInvoiceRepo{inv: inv},
		&verifyFeeRepo{fee: fee},
		accountRepo,
		&verifyTransactor{},
		autoPosting,
	)

	if _, err := uc.Execute(context.Background(), "pay-1", "verifier-1", "cash-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if journalRepo.savedEntry == nil {
		t.Fatal("expected journal entry to be saved")
	}

	var receivableCredit *journalEntity.JournalEntryLine
	for _, line := range journalRepo.savedEntry.Lines {
		if line.AccountID == "recv-spp" {
			receivableCredit = line
		}
	}
	if receivableCredit == nil {
		t.Fatal("expected journal line crediting fee component receivable account recv-spp")
	}
	if receivableCredit.Credit != 150000 {
		t.Errorf("expected receivable credit 150000, got %v", receivableCredit.Credit)
	}
}

func TestVerifyPaymentFeeComponentNotFound(t *testing.T) {
	debitID := "cash-1"
	payment, err := payEntity.NewPayment(
		"pay-1", "PAY-1", "inv-1", 150000, payConst.MethodCash,
		time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC),
		&debitID, nil, nil, nil, "user-1",
	)
	if err != nil {
		t.Fatalf("unexpected error creating payment: %v", err)
	}
	inv, err := invEntity.NewInvoice(
		"inv-1", "INV-1", "santri-1", "user-1", "fc-missing", nil, 150000,
		time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), "user-1",
	)
	if err != nil {
		t.Fatalf("unexpected error creating invoice: %v", err)
	}
	if err := inv.Issue(time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("unexpected error issuing invoice: %v", err)
	}

	uc := NewVerifyPaymentUseCase(
		&verifyPaymentRepo{payment: payment},
		&verifyInvoiceRepo{inv: inv},
		&verifyFeeRepo{err: kernel.WrapMsg(feeConst.CodeFeeComponentNotFound, "Komponen biaya tidak ditemukan", nil)},
		&verifyAccountRepo{byID: map[string]*accEntity.Account{"cash-1": {ID: "cash-1", Code: "1101", Name: "Kas", Type: accConst.TypeAsset, SubType: stPtr(accConst.SubTypeCashBank), IsPostable: true, IsActive: true}}},
		&verifyTransactor{},
		nil,
	)

	_, err = uc.Execute(context.Background(), "pay-1", "verifier-1", "cash-1")
	if err == nil {
		t.Fatal("expected error when fee component not found")
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kernel.AppError, got %T: %v", err, err)
	}
	if ke.Code != application.ErrCodeNotFound {
		t.Fatalf("expected code %s, got %s (%v)", application.ErrCodeNotFound, ke.Code, err)
	}
}

func TestVerifyPaymentRequiresDebitAccount(t *testing.T) {
	payment, err := payEntity.NewPayment(
		"pay-1", "PAY-1", "inv-1", 150000, payConst.MethodCash,
		time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC),
		nil, nil, nil, nil, "user-1",
	)
	if err != nil {
		t.Fatalf("unexpected error creating payment: %v", err)
	}
	inv, err := invEntity.NewInvoice(
		"inv-1", "INV-1", "santri-1", "user-1", "fc-1", nil, 150000,
		time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), "user-1",
	)
	if err != nil {
		t.Fatalf("unexpected error creating invoice: %v", err)
	}
	if err := inv.Issue(time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("unexpected error issuing invoice: %v", err)
	}
	fee, err := feeEntity.NewFeeComponent("fc-1", "SPP", "SPP Bulanan", "rev-spp", "recv-spp", 150000, "user-1")
	if err != nil {
		t.Fatalf("unexpected error creating fee component: %v", err)
	}
	cash := &accEntity.Account{ID: "cash-1", Code: "1101", Name: "Kas", Type: accConst.TypeAsset, SubType: stPtr(accConst.SubTypeCashBank), IsPostable: true, IsActive: true}
	accountRepo := &verifyAccountRepo{byID: map[string]*accEntity.Account{cash.ID: cash}}
	journalRepo := &verifyJournalRepo{}
	periodRepo := &verifyPeriodRepo{
		period: &periodEntity.AccountingPeriod{
			ID: "period-1", Name: "Agustus 2026", Status: periodConst.PeriodOpen,
			StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		},
	}
	autoPosting := journalService.NewAutoPostingService(journalRepo, accountRepo, periodRepo)
	uc := NewVerifyPaymentUseCase(
		&verifyPaymentRepo{payment: payment},
		&verifyInvoiceRepo{inv: inv},
		&verifyFeeRepo{fee: fee},
		accountRepo,
		&verifyTransactor{},
		autoPosting,
	)

	// Tanpa debit_account_id → ditolak.
	if _, err := uc.Execute(context.Background(), "pay-1", "verifier-1", ""); err == nil {
		t.Fatal("expected error when debit account is empty")
	}

	// Dengan debit_account_id yang bukan kas/bank → ditolak.
	nonCash := &accEntity.Account{ID: "acc-noncash", Code: "5101", Name: "Beban", Type: accConst.TypeExpense, SubType: stPtr(accConst.SubTypeOperatingExpense), IsPostable: true, IsActive: true}
	nonCashRepo := &verifyAccountRepo{byID: map[string]*accEntity.Account{nonCash.ID: nonCash}}
	ucNonCash := NewVerifyPaymentUseCase(
		&verifyPaymentRepo{payment: payment},
		&verifyInvoiceRepo{inv: inv},
		&verifyFeeRepo{fee: fee},
		nonCashRepo,
		&verifyTransactor{},
		autoPosting,
	)
	if _, err := ucNonCash.Execute(context.Background(), "pay-1", "verifier-1", "acc-noncash"); err == nil {
		t.Fatal("expected error when debit account is not cash/bank")
	}
}
