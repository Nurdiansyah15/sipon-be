package query

import (
	"context"
	"database/sql"

	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/shared/kernel"
)

type ReportBalanceSheetUseCase struct {
	db          *sql.DB
	accountRepo accRepo.AccountRepository
}

func NewReportBalanceSheetUseCase(db *sql.DB, accountRepo accRepo.AccountRepository) *ReportBalanceSheetUseCase {
	return &ReportBalanceSheetUseCase{db: db, accountRepo: accountRepo}
}

func (uc *ReportBalanceSheetUseCase) Execute(ctx context.Context, query dto.BalanceSheetQuery) (*dto.BalanceSheetResponse, error) {
	allAccounts, err := uc.accountRepo.ListAll(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	accountMap := make(map[string]*accountInfo)
	for _, acc := range allAccounts {
		accountMap[acc.ID] = &accountInfo{
			Code:     acc.Code,
			Name:     acc.Name,
			Type:     acc.Type,
			IsDebitNormal: acc.NormalBalance == accConst.BalanceDebit,
		}
	}

	var balances map[string]float64
	if query.PeriodID != nil && *query.PeriodID != "" {
		balances, err = uc.computePeriodBalances(ctx, *query.PeriodID)
	} else {
		balances, err = uc.computeBalancesToDate(ctx, query.AsOfDate)
	}
	if err != nil {
		return nil, err
	}

	asOfDate := ""
	if query.AsOfDate != nil {
		asOfDate = *query.AsOfDate
	}

	var assets, liabilities, equities []dto.BalanceSheetLine
	var totalAssets, totalLiabilities, totalEquities float64

	for _, acc := range allAccounts {
		if !acc.IsPostable || !acc.IsActive {
			continue
		}
		bal := balances[acc.ID]
		line := dto.BalanceSheetLine{
			AccountID:   acc.ID,
			AccountCode: acc.Code,
			AccountName: acc.Name,
			Amount:      bal,
		}
		switch acc.Type {
		case accConst.TypeAsset:
			assets = append(assets, line)
			totalAssets += bal
		case accConst.TypeLiability:
			liabilities = append(liabilities, line)
			totalLiabilities += bal
		case accConst.TypeEquity:
			equities = append(equities, line)
			totalEquities += bal
		}
	}

	return &dto.BalanceSheetResponse{
		AsOfDate:         asOfDate,
		Assets:           assets,
		TotalAssets:      totalAssets,
		Liabilities:      liabilities,
		TotalLiabilities: totalLiabilities,
		Equities:         equities,
		TotalEquities:    totalEquities,
	}, nil
}

func (uc *ReportBalanceSheetUseCase) computePeriodBalances(ctx context.Context, periodID string) (map[string]float64, error) {
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

	return uc.queryBalances(ctx, sqlQuery, periodID)
}

func (uc *ReportBalanceSheetUseCase) computeBalancesToDate(ctx context.Context, asOfDate *string) (map[string]float64, error) {
	sqlQuery := `SELECT 
		jel.account_id,
		COALESCE(SUM(jel.debit), 0) as total_debit,
		COALESCE(SUM(jel.credit), 0) as total_credit
	FROM journal_entry_lines jel
	JOIN journal_entries je ON je.id = jel.journal_entry_id
	WHERE jel.deleted_at IS NULL
		AND je.deleted_at IS NULL`
	args := []interface{}{}
	if asOfDate != nil && *asOfDate != "" {
		sqlQuery += ` AND je.entry_date <= $1`
		args = append(args, *asOfDate)
	}
	sqlQuery += ` GROUP BY jel.account_id`

	return uc.queryBalances(ctx, sqlQuery, args...)
}

func (uc *ReportBalanceSheetUseCase) queryBalances(ctx context.Context, sqlQuery string, args ...interface{}) (map[string]float64, error) {
	rows, err := uc.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	defer rows.Close()

	balances := make(map[string]float64)
	for rows.Next() {
		var accountID string
		var totalDebit, totalCredit float64
		if err := rows.Scan(&accountID, &totalDebit, &totalCredit); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		balances[accountID] = totalDebit - totalCredit
	}
	if err := rows.Err(); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return balances, nil
}

type accountInfo struct {
	Code          string
	Name          string
	Type          accConst.AccountType
	IsDebitNormal bool
}
