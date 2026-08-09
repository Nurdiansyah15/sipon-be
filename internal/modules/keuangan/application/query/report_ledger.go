package query

import (
	"context"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ReportLedgerUseCase struct {
	reportReader ports.ReportReader
	accountRepo  accRepo.AccountRepository
	periodRepo   periodRepo.AccountingPeriodRepository
}

func NewReportLedgerUseCase(reportReader ports.ReportReader, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *ReportLedgerUseCase {
	return &ReportLedgerUseCase{reportReader: reportReader, accountRepo: accountRepo, periodRepo: periodRepo}
}

func (uc *ReportLedgerUseCase) Execute(ctx context.Context, query dto.LedgerQuery) (*dto.LedgerResponse, error) {
	account, err := uc.accountRepo.FindByID(ctx, query.AccountID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	period, err := uc.periodRepo.FindByID(ctx, query.PeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	isDebitNormal := account.NormalBalance == accConst.BalanceDebit

	// Saldo awal: kumulatif semua jurnal sebelum start_date periode (carry-forward).
	opening, err := uc.balanceBefore(ctx, account.ID, period.StartDate, isDebitNormal)
	if err != nil {
		return nil, err
	}

	// Baris transaksi dalam periode ini (start_date s/d end_date), urut naik.
	lines, err := uc.reportReader.LedgerLines(ctx, query.AccountID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	var runningBalance = opening
	respLines := make([]dto.LedgerLineResponse, 0, len(lines))
	for _, line := range lines {
		respLine := dto.LedgerLineResponse{
			Date:          line.Date.Format("2006-01-02"),
			JournalNumber: line.JournalNumber,
			Description:   line.Description,
			Debit:         line.Debit,
			Credit:        line.Credit,
		}
		if isDebitNormal {
			runningBalance += line.Debit - line.Credit
		} else {
			runningBalance += line.Credit - line.Debit
		}
		respLine.Balance = runningBalance
		respLines = append(respLines, respLine)
	}

	return &dto.LedgerResponse{
		AccountID:      account.ID,
		AccountCode:    account.Code,
		AccountName:    account.Name,
		OpeningBalance: opening,
		Lines:          respLines,
		ClosingBalance: runningBalance,
	}, nil
}

func (uc *ReportLedgerUseCase) balanceBefore(ctx context.Context, accountID string, startDate time.Time, isDebitNormal bool) (float64, error) {
	debit, credit, err := uc.reportReader.BalanceBefore(ctx, accountID, startDate)
	if err != nil {
		return 0, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if isDebitNormal {
		return debit - credit, nil
	}
	return credit - debit, nil
}
