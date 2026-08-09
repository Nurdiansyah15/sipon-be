package query

import (
	"context"
	"database/sql"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ReportIncomeStatementUseCase struct {
	db          *sql.DB
	accountRepo accRepo.AccountRepository
	periodRepo  periodRepo.AccountingPeriodRepository
}

func NewReportIncomeStatementUseCase(db *sql.DB, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *ReportIncomeStatementUseCase {
	return &ReportIncomeStatementUseCase{db: db, accountRepo: accountRepo, periodRepo: periodRepo}
}

func (uc *ReportIncomeStatementUseCase) Execute(ctx context.Context, query dto.IncomeStatementQuery) (*dto.IncomeStatementResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, query.PeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	allAccounts, err := uc.accountRepo.ListAll(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	sqlQuery := `SELECT 
		jel.account_id,
		COALESCE(SUM(jel.debit), 0) as total_debit,
		COALESCE(SUM(jel.credit), 0) as total_credit
	FROM journal_entry_lines jel
	JOIN journal_entries je ON je.id = jel.journal_entry_id
	WHERE je.period_id = $1
		AND jel.deleted_at IS NULL
		AND je.deleted_at IS NULL
	GROUP BY jel.account_id`

	rows, err := uc.db.QueryContext(ctx, sqlQuery, query.PeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	defer rows.Close()

	balances := make(map[string]float64)
	for rows.Next() {
		var accountID string
		var totalDebit, totalCredit float64
		if err := rows.Scan(&accountID, &totalDebit, &totalCredit); err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		balances[accountID] = totalCredit - totalDebit
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	var revenues, expenses []dto.IncomeStatementLine
	var totalRevenue, totalExpense float64

	for _, acc := range allAccounts {
		if !acc.IsPostable || !acc.IsActive {
			continue
		}
		bal := balances[acc.ID]
		line := dto.IncomeStatementLine{
			AccountID:   acc.ID,
			AccountCode: acc.Code,
			AccountName: acc.Name,
			Amount:      bal,
		}
		switch acc.Type {
		case accConst.TypeRevenue:
			revenues = append(revenues, line)
			totalRevenue += bal
		case accConst.TypeExpense:
			expenses = append(expenses, line)
			totalExpense += bal
		}
	}

	return &dto.IncomeStatementResponse{
		PeriodID:     period.ID,
		PeriodName:   period.Name,
		Revenues:     revenues,
		TotalRevenue: totalRevenue,
		Expenses:     expenses,
		TotalExpense: totalExpense,
		NetIncome:    totalRevenue - totalExpense,
	}, nil
}
