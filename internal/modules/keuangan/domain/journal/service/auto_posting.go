package service

import (
	"context"
	"errors"
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

// alreadyPosted reports whether a journal entry already exists for the given
// source. It returns (false, nil) only when the query confirms no prior
// posting; any other outcome is surfaced so callers fail loudly instead of
// silently double-posting.
func (s *AutoPostingService) alreadyPosted(ctx context.Context, sourceType journalConst.SourceType, sourceID string) (bool, error) {
	existing, err := s.journalRepo.FindBySource(ctx, string(sourceType), sourceID)
	if err == nil {
		return existing != nil, nil
	}
	var ke *kernel.AppError
	if errors.As(err, &ke) && ke.Code == journalConst.CodeJournalNotFound {
		return false, nil
	}
	return false, err
}

func (s *AutoPostingService) PostInvoiceIssued(ctx context.Context, invoiceID, invoiceNumber, description string, entryDate time.Time, amount, discountAmount float64, feeType feeConst.FeeComponentType, postedBy string) error {
	posted, err := s.alreadyPosted(ctx, journalConst.SourceInvoiceIssued, invoiceID)
	if err != nil {
		return err
	}
	if posted {
		return nil
	}

	revCode, ok := feeTypeRevenueAccount[feeType]
	if !ok {
		return kernel.WrapMsg(journalConst.CodeJournalAccountMappingNotFound, "Akun pendapatan untuk jenis biaya ini belum dipetakan", nil)
	}

	piutang, err := s.accountRepo.FindByCode(ctx, "1103")
	if err != nil {
		return fmt.Errorf("find piutang account: %w", err)
	}
	revenue, err := s.accountRepo.FindByCode(ctx, revCode)
	if err != nil {
		return fmt.Errorf("find revenue account %s: %w", revCode, err)
	}

	period, err := s.periodRepo.FindByDate(ctx, entryDate)
	if err != nil {
		return fmt.Errorf("find active period: %w", err)
	}

	netAmount := amount - discountAmount
	jvNum, err := s.journalRepo.NextJournalNumber(ctx)
	if err != nil {
		return err
	}

	entry, err := journalEntity.NewJournalEntry(
		uuid.New().String(),
		jvNum.String(),
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

func (s *AutoPostingService) PostPaymentVerified(ctx context.Context, paymentID, paymentNumber, description string, entryDate time.Time, amount float64, debitAccountID string, postedBy string) error {
	posted, err := s.alreadyPosted(ctx, journalConst.SourcePaymentVerified, paymentID)
	if err != nil {
		return err
	}
	if posted {
		return nil
	}

	debitAcc, err := s.accountRepo.FindByID(ctx, debitAccountID)
	if err != nil {
		return fmt.Errorf("find debit account: %w", err)
	}

	piutang, err := s.accountRepo.FindByCode(ctx, "1103")
	if err != nil {
		return fmt.Errorf("find piutang account: %w", err)
	}

	period, err := s.periodRepo.FindByDate(ctx, entryDate)
	if err != nil {
		return fmt.Errorf("find active period: %w", err)
	}

	jvNum, err := s.journalRepo.NextJournalNumber(ctx)
	if err != nil {
		return err
	}
	entry, err := journalEntity.NewJournalEntry(
		uuid.New().String(),
		jvNum.String(),
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

func (s *AutoPostingService) PostInvoiceCancelled(ctx context.Context, invoiceID, invoiceNumber, description string, entryDate time.Time, originalAmount float64, feeType feeConst.FeeComponentType, postedBy string) error {
	posted, err := s.alreadyPosted(ctx, journalConst.SourceInvoiceCancelled, invoiceID)
	if err != nil {
		return err
	}
	if posted {
		return nil
	}

	revCode, ok := feeTypeRevenueAccount[feeType]
	if !ok {
		return kernel.WrapMsg(journalConst.CodeJournalAccountMappingNotFound, "Akun pendapatan untuk jenis biaya ini belum dipetakan", nil)
	}

	piutang, err := s.accountRepo.FindByCode(ctx, "1103")
	if err != nil {
		return fmt.Errorf("find piutang account: %w", err)
	}
	revenue, err := s.accountRepo.FindByCode(ctx, revCode)
	if err != nil {
		return fmt.Errorf("find revenue account %s: %w", revCode, err)
	}

	period, err := s.periodRepo.FindByDate(ctx, entryDate)
	if err != nil {
		return fmt.Errorf("find active period: %w", err)
	}

	jvNum, err := s.journalRepo.NextJournalNumber(ctx)
	if err != nil {
		return err
	}
	entry, err := journalEntity.NewJournalEntry(
		uuid.New().String(),
		jvNum.String(),
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
