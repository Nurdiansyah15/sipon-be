package command

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accEntity "sipon-be/internal/modules/keuangan/domain/account/entity"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	journalConst "sipon-be/internal/modules/keuangan/domain/journal/constant"
	journalEntity "sipon-be/internal/modules/keuangan/domain/journal/entity"
	journalRepo "sipon-be/internal/modules/keuangan/domain/journal/repository"
	periodConst "sipon-be/internal/modules/keuangan/domain/period/constant"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

const (
	closingAccountRetainedEarnings     = "3200"
	closingAccountCurrentYearEarnings  = "3201"
)

type ClosePeriodUseCase struct {
	periodRepo  periodRepo.AccountingPeriodRepository
	accountRepo accRepo.AccountRepository
	journalRepo journalRepo.JournalRepository
	transactor  ports.Transactor
}

func NewClosePeriodUseCase(periodRepo periodRepo.AccountingPeriodRepository, accountRepo accRepo.AccountRepository, journalRepo journalRepo.JournalRepository, transactor ports.Transactor) *ClosePeriodUseCase {
	return &ClosePeriodUseCase{periodRepo: periodRepo, accountRepo: accountRepo, journalRepo: journalRepo, transactor: transactor}
}

func (uc *ClosePeriodUseCase) Execute(ctx context.Context, periodID string, closedBy string) (*dto.PeriodResponse, error) {
	var resp *dto.PeriodResponse
	err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		period, err := uc.periodRepo.FindByID(txCtx, periodID)
		if err != nil {
			return err
		}
		if !period.IsOpen() {
			return kernel.WrapMsg(periodConst.CodePeriodInvalidStatus, "Hanya periode berstatus open yang dapat ditutup", nil)
		}

		retainedEarnings, err := uc.findClosingAccount(txCtx, closingAccountRetainedEarnings)
		if err != nil {
			return err
		}
		currentYearEarnings, err := uc.findClosingAccount(txCtx, closingAccountCurrentYearEarnings)
		if err != nil {
			return err
		}

		balances, err := uc.journalRepo.ComputeAccountBalances(txCtx, periodID)
		if err != nil {
			return err
		}

		allAccounts, err := uc.accountRepo.ListAll(txCtx)
		if err != nil {
			return err
		}
		accountByID := make(map[string]*accEntity.Account, len(allAccounts))
		for _, acc := range allAccounts {
			accountByID[acc.ID] = acc
		}

		var totalRevenue, totalExpense float64
		var revenueLines, expenseLines []*journalEntity.JournalEntryLine
		for accountID, bal := range balances {
			acc := accountByID[accountID]
			if acc == nil {
				continue
			}
			switch acc.Type {
			case accConst.TypeRevenue:
				if bal.Credit > bal.Debit {
					revenueLines = append(revenueLines, journalEntity.NewJournalEntryLine(
						uuid.New().String(), "", acc.ID, acc.Code, bal.Credit-bal.Debit, 0, nil,
					))
					totalRevenue += bal.Credit - bal.Debit
				}
			case accConst.TypeExpense:
				if bal.Debit > bal.Credit {
					expenseLines = append(expenseLines, journalEntity.NewJournalEntryLine(
						uuid.New().String(), "", acc.ID, acc.Code, 0, bal.Debit-bal.Credit, nil,
					))
					totalExpense += bal.Debit - bal.Credit
				}
			}
		}

		if totalRevenue > 0 || totalExpense > 0 {
			jn, err := uc.journalRepo.NextJournalNumber(txCtx)
			if err != nil {
				return err
			}
			entryDate := period.EndDate
			entry, err := journalEntity.NewJournalEntry(
				uuid.New().String(), jn.String(), entryDate,
				fmt.Sprintf("Closing periode %s", period.Name),
				period.ID, closedBy,
			)
			if err != nil {
				return err
			}
			entry.SetSource(journalConst.SourceClosing, period.ID)

			desc := "Penutupan pendapatan"
			for _, line := range revenueLines {
				line.JournalEntryID = entry.ID
				line.Description = &desc
				entry.AddLine(line)
			}
			if totalRevenue > 0 {
				entry.AddLine(journalEntity.NewJournalEntryLine(
					uuid.New().String(), entry.ID, currentYearEarnings.ID, currentYearEarnings.Code,
					0, totalRevenue, nil,
				))
			}
			descExp := "Penutupan beban"
			for _, line := range expenseLines {
				line.JournalEntryID = entry.ID
				line.Description = &descExp
				entry.AddLine(line)
			}
			if totalExpense > 0 {
				entry.AddLine(journalEntity.NewJournalEntryLine(
					uuid.New().String(), entry.ID, currentYearEarnings.ID, currentYearEarnings.Code,
					totalExpense, 0, nil,
				))
			}

			diff := totalRevenue - totalExpense
			if diff > 0 {
				entry.AddLine(journalEntity.NewJournalEntryLine(
					uuid.New().String(), entry.ID, currentYearEarnings.ID, currentYearEarnings.Code,
					diff, 0, nil,
				))
				entry.AddLine(journalEntity.NewJournalEntryLine(
					uuid.New().String(), entry.ID, retainedEarnings.ID, retainedEarnings.Code,
					0, diff, nil,
				))
			} else if diff < 0 {
				entry.AddLine(journalEntity.NewJournalEntryLine(
					uuid.New().String(), entry.ID, retainedEarnings.ID, retainedEarnings.Code,
					-diff, 0, nil,
				))
				entry.AddLine(journalEntity.NewJournalEntryLine(
					uuid.New().String(), entry.ID, currentYearEarnings.ID, currentYearEarnings.Code,
					0, -diff, nil,
				))
			}

			if err := entry.Post(); err != nil {
				return err
			}
			if err := uc.journalRepo.Save(txCtx, entry); err != nil {
				return err
			}
		}

		if err := period.Close(closedBy); err != nil {
			return err
		}
		if err := uc.periodRepo.Update(txCtx, period); err != nil {
			return err
		}
		resp = toPeriodResponse(period)
		return nil
	})
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case periodConst.CodePeriodNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			case periodConst.CodePeriodInvalidStatus:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case periodConst.CodePeriodClosingAccountMissing:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case accConst.CodeAccountNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return resp, nil
}

func (uc *ClosePeriodUseCase) findClosingAccount(ctx context.Context, code string) (*accEntity.Account, error) {
	acc, err := uc.accountRepo.FindByCode(ctx, code)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == accConst.CodeAccountNotFound {
			return nil, kernel.WrapMsg(periodConst.CodePeriodClosingAccountMissing,
				fmt.Sprintf("Akun %s (untuk closing periode) belum tersedia di chart of accounts", code), ke)
		}
		return nil, err
	}
	return acc, nil
}
