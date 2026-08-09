package query

import (
	"context"
	"database/sql"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ReportLedgerUseCase struct {
	db          *sql.DB
	accountRepo accRepo.AccountRepository
	periodRepo  periodRepo.AccountingPeriodRepository
}

func NewReportLedgerUseCase(db *sql.DB, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *ReportLedgerUseCase {
	return &ReportLedgerUseCase{db: db, accountRepo: accountRepo, periodRepo: periodRepo}
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
	sqlQuery := `SELECT 
		je.entry_date::date::text,
		je.journal_number,
		COALESCE(je.description, ''),
		COALESCE(jel.debit, 0),
		COALESCE(jel.credit, 0)
	FROM journal_entry_lines jel
	JOIN journal_entries je ON je.id = jel.journal_entry_id
	WHERE jel.account_id = $1 
		AND je.entry_date >= $2
		AND je.entry_date <= $3
		AND je.status = 'posted'
	ORDER BY je.entry_date ASC, je.journal_number ASC`

	rows, err := uc.db.QueryContext(ctx, sqlQuery, query.AccountID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	defer rows.Close()

	var runningBalance = opening
	lines := make([]dto.LedgerLineResponse, 0)
	for rows.Next() {
		var line dto.LedgerLineResponse
		if err := rows.Scan(&line.Date, &line.JournalNumber, &line.Description, &line.Debit, &line.Credit); err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		if isDebitNormal {
			runningBalance += line.Debit - line.Credit
		} else {
			runningBalance += line.Credit - line.Debit
		}
		line.Balance = runningBalance
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return &dto.LedgerResponse{
		AccountID:      account.ID,
		AccountCode:    account.Code,
		AccountName:    account.Name,
		OpeningBalance: opening,
		Lines:          lines,
		ClosingBalance: runningBalance,
	}, nil
}

func (uc *ReportLedgerUseCase) balanceBefore(ctx context.Context, accountID string, startDate time.Time, isDebitNormal bool) (float64, error) {
	sqlQuery := `SELECT COALESCE(SUM(jel.debit), 0) AS total_debit, COALESCE(SUM(jel.credit), 0) AS total_credit
		FROM journal_entry_lines jel
		JOIN journal_entries je ON je.id = jel.journal_entry_id
		WHERE jel.account_id = $1
			AND je.entry_date < $2
			AND je.status = 'posted'`

	var totalDebit, totalCredit float64
	if err := uc.db.QueryRowContext(ctx, sqlQuery, accountID, startDate).Scan(&totalDebit, &totalCredit); err != nil {
		return 0, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if isDebitNormal {
		return totalDebit - totalCredit, nil
	}
	return totalCredit - totalDebit, nil
}
