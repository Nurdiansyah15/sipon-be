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

type ReportBalanceSheetUseCase struct {
	db          *sql.DB
	accountRepo accRepo.AccountRepository
	periodRepo  periodRepo.AccountingPeriodRepository
}

func NewReportBalanceSheetUseCase(db *sql.DB, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *ReportBalanceSheetUseCase {
	return &ReportBalanceSheetUseCase{db: db, accountRepo: accountRepo, periodRepo: periodRepo}
}

func (uc *ReportBalanceSheetUseCase) Execute(ctx context.Context, query dto.BalanceSheetQuery) (*dto.BalanceSheetResponse, error) {
	allAccounts, err := uc.accountRepo.ListAll(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	normalBalance := make(map[string]accConst.NormalBalance, len(allAccounts))
	for _, acc := range allAccounts {
		normalBalance[acc.ID] = acc.NormalBalance
	}

	// Selalu kumulatif sampai tanggal tertentu (carry-forward akun riil).
	// Kalau caller kirim period_id, resolve jadi end_date periode tsb.
	var asOfDate string
	if query.PeriodID != nil && *query.PeriodID != "" {
		period, err := uc.periodRepo.FindByID(ctx, *query.PeriodID)
		if err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
		}
		asOfDate = period.EndDate.Format("2006-01-02")
	} else if query.AsOfDate != nil {
		asOfDate = *query.AsOfDate
	}

	balances, err := uc.computeBalancesToDate(ctx, asOfDate)
	if err != nil {
		return nil, err
	}

	var assets, liabilities, equities []dto.BalanceSheetLine
	var totalAssets, totalLiabilities, totalEquities float64

	for _, acc := range allAccounts {
		if !acc.IsPostable || !acc.IsActive {
			continue
		}
		bal := balances[acc.ID]
		// queryBalances menghitung debit - credit. Untuk akun credit-normal
		// (liability/equity) saldo yang benar = credit - debit.
		if normalBalance[acc.ID] != accConst.BalanceDebit {
			bal = -bal
		}
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

func (uc *ReportBalanceSheetUseCase) computeBalancesToDate(ctx context.Context, asOfDate string) (map[string]float64, error) {
	sqlQuery := `SELECT 
		jel.account_id,
		COALESCE(SUM(jel.debit), 0) as total_debit,
		COALESCE(SUM(jel.credit), 0) as total_credit
	FROM journal_entry_lines jel
	JOIN journal_entries je ON je.id = jel.journal_entry_id
	WHERE je.status = 'posted'`
	args := []interface{}{}
	if asOfDate != "" {
		sqlQuery += ` AND je.entry_date <= $1`
		args = append(args, asOfDate)
	}
	sqlQuery += ` GROUP BY jel.account_id`

	return uc.queryBalances(ctx, sqlQuery, args...)
}

func (uc *ReportBalanceSheetUseCase) queryBalances(ctx context.Context, sqlQuery string, args ...interface{}) (map[string]float64, error) {
	rows, err := uc.db.QueryContext(ctx, sqlQuery, args...)
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
		balances[accountID] = totalDebit - totalCredit
	}
		if err := rows.Err(); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	return balances, nil
}
