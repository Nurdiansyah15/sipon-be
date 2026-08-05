package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	accountRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	feeConst "sipon-be/internal/modules/keuangan/domain/feecomponent/constant"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalEntity "sipon-be/internal/modules/keuangan/domain/journal/entity"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type AutoPostingService struct {
	journalRepo journalRepo.JournalRepository
	accountRepo accountRepo.AccountRepository
	periodRepo  periodRepo.AccountingPeriodRepository
}

func NewAutoPostingService(
	journalRepo journalRepo.JournalRepository,
	accountRepo accountRepo.AccountRepository,
	periodRepo periodRepo.AccountingPeriodRepository,
) *AutoPostingService {
	return &AutoPostingService{
		journalRepo: journalRepo,
		accountRepo: accountRepo,
		periodRepo:  periodRepo,
	}
}

var feeTypeRevenueAccount = map[feeConst.FeeComponentType]string{
	feeConst.FeeTypeSPP:         "4100",
	feeConst.FeeTypeUKT:         "4200",
	feeConst.FeeTypeDaftarUlang: "4300",
	feeConst.FeeTypeInsidental:  "4400",
}

func generateJournalNumber(seq int, sourceType journalConst.SourceType) string {
	now := time.Now()
	prefix := "JRN"
	switch sourceType {
	case journalConst.SourceInvoiceIssued:
		prefix = "INV"
	case journalConst.SourcePaymentVerified:
		prefix = "PAY"
	case journalConst.SourceInvoiceCancelled:
		prefix = "CNL"
	case journalConst.SourceAdjustment:
		prefix = "ADJ"
	case journalConst.SourceClosing:
		prefix = "CLS"
	case journalConst.SourceManual:
		prefix = "MAN"
	}
	return fmt.Sprintf("%s/%d/%02d/%06d", prefix, now.Year(), now.Month(), seq)
}

func (s *AutoPostingService) PostInvoiceIssued(ctx context.Context, invoiceID, invoiceNumber, description string, entryDate time.Time, amount, discountAmount float64, feeType feeConst.FeeComponentType, postedBy string, seq int) error {
	revCode, ok := feeTypeRevenueAccount[feeType]
	if !ok {
		return kernel.New(journalConst.CodeJournalNotBalanced)
	}

	piutang, err := s.accountRepo.FindByCode(ctx, "1103")
	if err != nil {
		return fmt.Errorf("find piutang account: %w", err)
	}
	revenue, err := s.accountRepo.FindByCode(ctx, revCode)
	if err != nil {
		return fmt.Errorf("find revenue account %s: %w", revCode, err)
	}

	period, err := s.periodRepo.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("find active period: %w", err)
	}

	netAmount := amount - discountAmount
	jvNum := generateJournalNumber(seq, journalConst.SourceInvoiceIssued)

	entry, err := journalEntity.NewJournalEntry(
		uuid.New().String(),
		jvNum,
		entryDate,
		fmt.Sprintf("Invoice issued %s: %s", invoiceNumber, description),
		period.ID,
		postedBy,
	)
	if err != nil {
		return err
	}
	entry.SetSource(journalConst.SourceInvoiceIssued, invoiceID)

	entry.AddLine(journalEntity.NewJournalEntryLine(
		uuid.New().String(), entry.ID, piutang.ID, piutang.Code, netAmount, 0, nil,
	))
	entry.AddLine(journalEntity.NewJournalEntryLine(
		uuid.New().String(), entry.ID, revenue.ID, revenue.Code, 0, netAmount, nil,
	))

	if err := entry.Post(); err != nil {
		return err
	}

	return s.journalRepo.Save(ctx, entry)
}

func (s *AutoPostingService) PostPaymentVerified(ctx context.Context, paymentID, paymentNumber, description string, entryDate time.Time, amount float64, debitAccountID string, postedBy string, seq int) error {
	debitAcc, err := s.accountRepo.FindByID(ctx, debitAccountID)
	if err != nil {
		return fmt.Errorf("find debit account: %w", err)
	}

	piutang, err := s.accountRepo.FindByCode(ctx, "1103")
	if err != nil {
		return fmt.Errorf("find piutang account: %w", err)
	}

	period, err := s.periodRepo.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("find active period: %w", err)
	}

	jvNum := generateJournalNumber(seq, journalConst.SourcePaymentVerified)
	entry, err := journalEntity.NewJournalEntry(
		uuid.New().String(),
		jvNum,
		entryDate,
		fmt.Sprintf("Payment verified %s: %s", paymentNumber, description),
		period.ID,
		postedBy,
	)
	if err != nil {
		return err
	}
	entry.SetSource(journalConst.SourcePaymentVerified, paymentID)

	entry.AddLine(journalEntity.NewJournalEntryLine(
		uuid.New().String(), entry.ID, debitAcc.ID, debitAcc.Code, amount, 0, nil,
	))
	entry.AddLine(journalEntity.NewJournalEntryLine(
		uuid.New().String(), entry.ID, piutang.ID, piutang.Code, 0, amount, nil,
	))

	if err := entry.Post(); err != nil {
		return err
	}

	return s.journalRepo.Save(ctx, entry)
}

func (s *AutoPostingService) PostInvoiceCancelled(ctx context.Context, invoiceID, invoiceNumber, description string, entryDate time.Time, originalAmount float64, feeType feeConst.FeeComponentType, postedBy string, seq int) error {
	revCode, ok := feeTypeRevenueAccount[feeType]
	if !ok {
		return kernel.New(journalConst.CodeJournalNotBalanced)
	}

	piutang, err := s.accountRepo.FindByCode(ctx, "1103")
	if err != nil {
		return fmt.Errorf("find piutang account: %w", err)
	}
	revenue, err := s.accountRepo.FindByCode(ctx, revCode)
	if err != nil {
		return fmt.Errorf("find revenue account %s: %w", revCode, err)
	}

	period, err := s.periodRepo.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("find active period: %w", err)
	}

	jvNum := generateJournalNumber(seq, journalConst.SourceInvoiceCancelled)
	entry, err := journalEntity.NewJournalEntry(
		uuid.New().String(),
		jvNum,
		entryDate,
		fmt.Sprintf("Invoice cancelled %s: %s", invoiceNumber, description),
		period.ID,
		postedBy,
	)
	if err != nil {
		return err
	}
	entry.SetSource(journalConst.SourceInvoiceCancelled, invoiceID)

	entry.AddLine(journalEntity.NewJournalEntryLine(
		uuid.New().String(), entry.ID, revenue.ID, revenue.Code, originalAmount, 0, nil,
	))
	entry.AddLine(journalEntity.NewJournalEntryLine(
		uuid.New().String(), entry.ID, piutang.ID, piutang.Code, 0, originalAmount, nil,
	))

	if err := entry.Post(); err != nil {
		return err
	}

	return s.journalRepo.Save(ctx, entry)
}
