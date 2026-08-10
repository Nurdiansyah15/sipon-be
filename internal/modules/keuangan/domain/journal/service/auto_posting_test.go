package service

import (
	"context"
	"errors"
	"testing"
	"time"

	accountConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accountEntity "sipon-be/internal/modules/keuangan/domain/account/entity"
	accountRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalEntity "sipon-be/internal/modules/keuangan/domain/journal/entity"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	journalVO "sipon-be/internal/modules/keuangan/domain/journal/valueobject"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodEntity "sipon-be/internal/modules/keuangan/domain/period/entity"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type mockAccountRepo struct {
	byID   map[string]*accountEntity.Account
	byCode map[string]*accountEntity.Account
}

func (m *mockAccountRepo) Save(ctx context.Context, acc *accountEntity.Account) error { return nil }
func (m *mockAccountRepo) Update(ctx context.Context, acc *accountEntity.Account) error {
	return nil
}
func (m *mockAccountRepo) FindByID(ctx context.Context, id string) (*accountEntity.Account, error) {
	if acc, ok := m.byID[id]; ok {
		return acc, nil
	}
	return nil, kernel.WrapMsg(accountConst.CodeAccountNotFound, "Akun tidak ditemukan", nil)
}
func (m *mockAccountRepo) FindByCode(ctx context.Context, code string) (*accountEntity.Account, error) {
	if acc, ok := m.byCode[code]; ok {
		return acc, nil
	}
	return nil, kernel.WrapMsg(accountConst.CodeAccountNotFound, "Akun tidak ditemukan", nil)
}
func (m *mockAccountRepo) List(ctx context.Context, q accountRepo.AccountListQuery) (*accountRepo.AccountListResult, error) {
	return nil, nil
}
func (m *mockAccountRepo) ListAll(ctx context.Context) ([]*accountEntity.Account, error) {
	return nil, nil
}
func (m *mockAccountRepo) ListPostable(ctx context.Context) ([]*accountEntity.Account, error) {
	return nil, nil
}
func (m *mockAccountRepo) FindChildren(ctx context.Context, parentID string) ([]*accountEntity.Account, error) {
	return nil, nil
}
func (m *mockAccountRepo) HasJournalEntries(ctx context.Context, accountID string) (bool, error) {
	return false, nil
}
func (m *mockAccountRepo) ExistsByCode(ctx context.Context, code, excludeID string) (bool, error) {
	return false, nil
}

type mockJournalRepo struct {
	seq        int
	savedEntry *journalEntity.JournalEntry
}

func (m *mockJournalRepo) Save(ctx context.Context, entry *journalEntity.JournalEntry) error {
	m.savedEntry = entry
	return nil
}
func (m *mockJournalRepo) Update(ctx context.Context, entry *journalEntity.JournalEntry) error {
	return nil
}
func (m *mockJournalRepo) FindByID(ctx context.Context, id string) (*journalEntity.JournalEntry, error) {
	return nil, kernel.WrapMsg(journalConst.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
}
func (m *mockJournalRepo) FindByNumber(ctx context.Context, number string) (*journalEntity.JournalEntry, error) {
	return nil, kernel.WrapMsg(journalConst.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
}
func (m *mockJournalRepo) NextJournalNumber(ctx context.Context) (journalVO.JournalNumber, error) {
	m.seq++
	return journalVO.NewJournalNumber("2026", "08", m.seq), nil
}
func (m *mockJournalRepo) List(ctx context.Context, q journalRepo.JournalListQuery) (*journalRepo.JournalListResult, error) {
	return nil, nil
}
func (m *mockJournalRepo) FindBySource(ctx context.Context, sourceType, sourceID string) (*journalEntity.JournalEntry, error) {
	return nil, kernel.WrapMsg(journalConst.CodeJournalNotFound, "Jurnal tidak ditemukan", nil)
}
func (m *mockJournalRepo) SaveLines(ctx context.Context, entryID string, lines []*journalEntity.JournalEntryLine) error {
	return nil
}
func (m *mockJournalRepo) FindLinesByEntryID(ctx context.Context, entryID string) ([]*journalEntity.JournalEntryLine, error) {
	return nil, nil
}
func (m *mockJournalRepo) ComputeAccountBalances(ctx context.Context, periodID string) (map[string]journalRepo.AccountBalance, error) {
	return nil, nil
}

type mockPeriodRepo struct {
	period *periodEntity.AccountingPeriod
}

func (m *mockPeriodRepo) Save(ctx context.Context, p *periodEntity.AccountingPeriod) error { return nil }
func (m *mockPeriodRepo) Update(ctx context.Context, p *periodEntity.AccountingPeriod) error {
	return nil
}
func (m *mockPeriodRepo) FindByID(ctx context.Context, id string) (*periodEntity.AccountingPeriod, error) {
	return nil, kernel.WrapMsg(periodConst.CodePeriodNotFound, "Periode tidak ditemukan", nil)
}
func (m *mockPeriodRepo) FindActive(ctx context.Context) (*periodEntity.AccountingPeriod, error) {
	return m.period, nil
}
func (m *mockPeriodRepo) List(ctx context.Context, q periodRepo.PeriodListQuery) (*periodRepo.PeriodListResult, error) {
	return nil, nil
}
func (m *mockPeriodRepo) FindByDate(ctx context.Context, date time.Time) (*periodEntity.AccountingPeriod, error) {
	return m.period, nil
}
func (m *mockPeriodRepo) HasOverlap(ctx context.Context, startDate, endDate time.Time, excludeID string) (bool, error) {
	return false, nil
}

func testAccount(id, code string, accType accountConst.AccountType, subType *accountConst.AccountSubType) *accountEntity.Account {
	return &accountEntity.Account{
		ID:         id,
		Code:       code,
		Name:       code,
		Type:       accType,
		SubType:    subType,
		Level:      3,
		IsPostable: true,
		IsActive:   true,
	}
}

func testPeriod() *periodEntity.AccountingPeriod {
	p, _ := periodEntity.NewAccountingPeriod(
		"period-1", "Agustus 2026",
		time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		"user-1",
	)
	return p
}

func newTestAutoPosting(accountRepo accountRepo.AccountRepository, period *periodEntity.AccountingPeriod) (*AutoPostingService, *mockJournalRepo) {
	journalMock := &mockJournalRepo{}
	return NewAutoPostingService(journalMock, accountRepo, &mockPeriodRepo{period: period}), journalMock
}

func assertErrorCode(t *testing.T, err error, code kernel.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %s, got nil", code)
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) {
		t.Fatalf("expected *kernel.AppError, got %T: %v", err, err)
	}
	if ke.Code != code {
		t.Fatalf("expected code %s, got %s (%v)", code, ke.Code, err)
	}
}

func findLine(t *testing.T, entry *journalEntity.JournalEntry, accountID string) *journalEntity.JournalEntryLine {
	t.Helper()
	for _, line := range entry.Lines {
		if line.AccountID == accountID {
			return line
		}
	}
	t.Fatalf("line for account %s not found in journal entry", accountID)
	return nil
}

func subTypePtr(st accountConst.AccountSubType) *accountConst.AccountSubType {
	return &st
}

func TestPostInvoiceIssued_UsesFeeComponentAccounts(t *testing.T) {
	revSPP := testAccount("rev-spp", "4100", accountConst.TypeRevenue, subTypePtr(accountConst.SubTypeOperatingRevenue))
	recvSPP := testAccount("recv-spp", "1104", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeReceivable))
	revUKT := testAccount("rev-ukt", "4200", accountConst.TypeRevenue, subTypePtr(accountConst.SubTypeOperatingRevenue))
	recvUKT := testAccount("recv-ukt", "1105", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeReceivable))

	accounts := &mockAccountRepo{
		byID: map[string]*accountEntity.Account{
			revSPP.ID:  revSPP,
			recvSPP.ID: recvSPP,
			revUKT.ID:  revUKT,
			recvUKT.ID: recvUKT,
		},
	}

	entryDate := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	svc, journalMock := newTestAutoPosting(accounts, testPeriod())

	if err := svc.PostInvoiceIssued(context.Background(), "inv-1", "INV-1", "", entryDate, 150000, 0, revSPP.ID, recvSPP.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	saved := journalMock.savedEntry
	if saved == nil {
		t.Fatal("no journal entry was saved")
	}
	debit := findLine(t, saved, recvSPP.ID)
	credit := findLine(t, saved, revSPP.ID)
	if debit.Debit != 150000 || debit.Credit != 0 {
		t.Errorf("expected receivable line 150000 debit, got debit=%v credit=%v", debit.Debit, debit.Credit)
	}
	if credit.Credit != 150000 || credit.Debit != 0 {
		t.Errorf("expected revenue line 150000 credit, got debit=%v credit=%v", credit.Debit, credit.Credit)
	}

	// Fee component kedua memakai akun piutang berbeda — jurnal harus memakai akun itu.
	if err := svc.PostInvoiceIssued(context.Background(), "inv-2", "INV-2", "", entryDate, 200000, 0, revUKT.ID, recvUKT.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	saved2 := journalMock.savedEntry
	debit2 := findLine(t, saved2, recvUKT.ID)
	if debit2.Debit != 200000 {
		t.Errorf("expected UKT receivable line 200000 debit on recv-ukt, got debit=%v", debit2.Debit)
	}
}

func TestPostInvoiceIssued_RevenueAccountInvalid(t *testing.T) {
	badRev := testAccount("rev-bad", "9999", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeCashBank))
	recv := testAccount("recv-ok", "1103", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeReceivable))
	accounts := &mockAccountRepo{
		byID: map[string]*accountEntity.Account{
			badRev.ID: badRev,
			recv.ID:   recv,
		},
	}
	svc, _ := newTestAutoPosting(accounts, testPeriod())

	err := svc.PostInvoiceIssued(context.Background(), "inv-1", "INV-1", "", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), 100000, 0, badRev.ID, recv.ID, "user-1")
	assertErrorCode(t, err, journalConst.CodeJournalAccountMappingNotFound)
}

func TestPostInvoiceIssued_RevenueAccountNotFound(t *testing.T) {
	recv := testAccount("recv-ok", "1103", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeReceivable))
	accounts := &mockAccountRepo{byID: map[string]*accountEntity.Account{recv.ID: recv}}
	svc, _ := newTestAutoPosting(accounts, testPeriod())

	err := svc.PostInvoiceIssued(context.Background(), "inv-1", "INV-1", "", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), 100000, 0, "rev-missing", recv.ID, "user-1")
	assertErrorCode(t, err, journalConst.CodeJournalAccountMappingNotFound)
}

func TestPostInvoiceIssued_ReceivableAccountNotReceivable(t *testing.T) {
	rev := testAccount("rev-ok", "4100", accountConst.TypeRevenue, subTypePtr(accountConst.SubTypeOperatingRevenue))
	cash := testAccount("cash-1", "1101", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeCashBank))
	accounts := &mockAccountRepo{byID: map[string]*accountEntity.Account{rev.ID: rev, cash.ID: cash}}
	svc, _ := newTestAutoPosting(accounts, testPeriod())

	err := svc.PostInvoiceIssued(context.Background(), "inv-1", "INV-1", "", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), 100000, 0, rev.ID, cash.ID, "user-1")
	assertErrorCode(t, err, journalConst.CodeJournalAccountMappingNotFound)
}

func TestPostInvoiceIssued_ReceivableAccountNotPostable(t *testing.T) {
	rev := testAccount("rev-ok", "4100", accountConst.TypeRevenue, subTypePtr(accountConst.SubTypeOperatingRevenue))
	notPostable := &accountEntity.Account{
		ID:         "recv-not-postable",
		Code:       "1106",
		Name:       "Piutang (tidak postable)",
		Type:       accountConst.TypeAsset,
		SubType:    subTypePtr(accountConst.SubTypeReceivable),
		IsPostable: false,
		IsActive:   true,
	}
	accounts := &mockAccountRepo{byID: map[string]*accountEntity.Account{rev.ID: rev, notPostable.ID: notPostable}}
	svc, _ := newTestAutoPosting(accounts, testPeriod())

	err := svc.PostInvoiceIssued(context.Background(), "inv-1", "INV-1", "", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), 100000, 0, rev.ID, notPostable.ID, "user-1")
	assertErrorCode(t, err, journalConst.CodeJournalAccountMappingNotFound)
}

func TestPostPaymentVerified_UsesFeeComponentReceivable(t *testing.T) {
	cash := testAccount("cash-1", "1101", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeCashBank))
	recv := testAccount("recv-spp", "1104", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeReceivable))
	accounts := &mockAccountRepo{byID: map[string]*accountEntity.Account{cash.ID: cash, recv.ID: recv}}
	svc, journalMock := newTestAutoPosting(accounts, testPeriod())

	if err := svc.PostPaymentVerified(context.Background(), "pay-1", "PAY-1", "", time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), 150000, cash.ID, recv.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	saved := journalMock.savedEntry
	if saved == nil {
		t.Fatal("no journal entry was saved")
	}
	debit := findLine(t, saved, cash.ID)
	credit := findLine(t, saved, recv.ID)
	if debit.Debit != 150000 {
		t.Errorf("expected cash line 150000 debit, got %v", debit.Debit)
	}
	if credit.Credit != 150000 {
		t.Errorf("expected receivable line 150000 credit, got %v", credit.Credit)
	}
}

func TestPostPaymentVerified_ReceivableAccountInvalid(t *testing.T) {
	cash := testAccount("cash-1", "1101", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeCashBank))
	rev := testAccount("rev-ok", "4100", accountConst.TypeRevenue, subTypePtr(accountConst.SubTypeOperatingRevenue))
	accounts := &mockAccountRepo{byID: map[string]*accountEntity.Account{cash.ID: cash, rev.ID: rev}}
	svc, _ := newTestAutoPosting(accounts, testPeriod())

	err := svc.PostPaymentVerified(context.Background(), "pay-1", "PAY-1", "", time.Date(2026, 8, 12, 0, 0, 0, 0, time.UTC), 150000, cash.ID, rev.ID, "user-1")
	assertErrorCode(t, err, journalConst.CodeJournalAccountMappingNotFound)
}

func TestPostInvoiceCancelled_ReversesWithFeeComponentAccounts(t *testing.T) {
	rev := testAccount("rev-spp", "4100", accountConst.TypeRevenue, subTypePtr(accountConst.SubTypeOperatingRevenue))
	recv := testAccount("recv-spp", "1104", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeReceivable))
	accounts := &mockAccountRepo{byID: map[string]*accountEntity.Account{rev.ID: rev, recv.ID: recv}}
	svc, journalMock := newTestAutoPosting(accounts, testPeriod())

	if err := svc.PostInvoiceCancelled(context.Background(), "inv-1", "INV-1", "", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), 150000, rev.ID, recv.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	saved := journalMock.savedEntry
	if saved == nil {
		t.Fatal("no journal entry was saved")
	}
	revLine := findLine(t, saved, rev.ID)
	recvLine := findLine(t, saved, recv.ID)
	if revLine.Debit != 150000 {
		t.Errorf("expected revenue reversal line 150000 debit, got %v", revLine.Debit)
	}
	if recvLine.Credit != 150000 {
		t.Errorf("expected receivable reversal line 150000 credit, got %v", recvLine.Credit)
	}
}

func TestPostAdjustment_UsesFeeComponentAccounts(t *testing.T) {
	rev := testAccount("rev-spp", "4100", accountConst.TypeRevenue, subTypePtr(accountConst.SubTypeOperatingRevenue))
	recv := testAccount("recv-spp", "1104", accountConst.TypeAsset, subTypePtr(accountConst.SubTypeReceivable))
	accounts := &mockAccountRepo{byID: map[string]*accountEntity.Account{rev.ID: rev, recv.ID: recv}}
	svc, journalMock := newTestAutoPosting(accounts, testPeriod())

	if err := svc.PostAdjustment(context.Background(), "adj-1", "INV-1", "", time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), 25000, rev.ID, recv.ID, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	saved := journalMock.savedEntry
	if saved == nil {
		t.Fatal("no journal entry was saved")
	}
	revLine := findLine(t, saved, rev.ID)
	recvLine := findLine(t, saved, recv.ID)
	if revLine.Debit != 25000 {
		t.Errorf("expected revenue adjustment line 25000 debit, got %v", revLine.Debit)
	}
	if recvLine.Credit != 25000 {
		t.Errorf("expected receivable adjustment line 25000 credit, got %v", recvLine.Credit)
	}
}
