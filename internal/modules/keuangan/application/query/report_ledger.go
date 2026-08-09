package query

import (
	"context"
	"database/sql"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/shared/kernel"
)

type ReportLedgerUseCase struct {
	db          *sql.DB
	accountRepo accRepo.AccountRepository
}

func NewReportLedgerUseCase(db *sql.DB, accountRepo accRepo.AccountRepository) *ReportLedgerUseCase {
	return &ReportLedgerUseCase{db: db, accountRepo: accountRepo}
}

func (uc *ReportLedgerUseCase) Execute(ctx context.Context, query dto.LedgerQuery) (*dto.LedgerResponse, error) {
	account, err := uc.accountRepo.FindByID(ctx, query.AccountID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	isDebitNormal := account.NormalBalance == accConst.BalanceDebit

	sqlQuery := `SELECT 
		je.entry_date::date::text,
		je.journal_number,
		COALESCE(je.description, ''),
		COALESCE(jel.debit, 0),
		COALESCE(jel.credit, 0)
	FROM journal_entry_lines jel
	JOIN journal_entries je ON je.id = jel.journal_entry_id
	WHERE jel.account_id = $1 
		AND je.period_id = $2
		AND jel.deleted_at IS NULL
		AND je.deleted_at IS NULL
	ORDER BY je.entry_date ASC, je.journal_number ASC`

	rows, err := uc.db.QueryContext(ctx, sqlQuery, query.AccountID, query.PeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	defer rows.Close()

	var runningBalance float64
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
		AccountID:   account.ID,
		AccountCode: account.Code,
		AccountName: account.Name,
		Lines:       lines,
	}, nil
}
