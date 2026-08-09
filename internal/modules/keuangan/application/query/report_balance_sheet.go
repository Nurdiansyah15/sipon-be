package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accConst "sipon-be/internal/modules/keuangan/domain/account/constant"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ReportBalanceSheetUseCase struct {
	reportReader ports.ReportReader
	accountRepo  accRepo.AccountRepository
	periodRepo   periodRepo.AccountingPeriodRepository
}

func NewReportBalanceSheetUseCase(reportReader ports.ReportReader, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *ReportBalanceSheetUseCase {
	return &ReportBalanceSheetUseCase{reportReader: reportReader, accountRepo: accountRepo, periodRepo: periodRepo}
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
	var asOfDate *string
	if query.PeriodID != nil && *query.PeriodID != "" {
		period, err := uc.periodRepo.FindByID(ctx, *query.PeriodID)
		if err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
		}
		d := period.EndDate.Format("2006-01-02")
		asOfDate = &d
	} else if query.AsOfDate != nil {
		asOfDate = query.AsOfDate
	}

	balances, err := uc.reportReader.AccountBalancesToDate(ctx, asOfDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	balanceMap := make(map[string]float64, len(balances))
	for _, bal := range balances {
		balanceMap[bal.AccountID] = bal.Debit - bal.Credit
	}

	var assets, liabilities, equities []dto.BalanceSheetLine
	var totalAssets, totalLiabilities, totalEquities float64

	for _, acc := range allAccounts {
		if !acc.IsPostable || !acc.IsActive {
			continue
		}
		bal := balanceMap[acc.ID]
		// balanceMap menghitung debit - credit. Untuk akun credit-normal
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

	asOfDateStr := ""
	if asOfDate != nil {
		asOfDateStr = *asOfDate
	}

	return &dto.BalanceSheetResponse{
		AsOfDate:         asOfDateStr,
		Assets:           assets,
		TotalAssets:      totalAssets,
		Liabilities:      liabilities,
		TotalLiabilities: totalLiabilities,
		Equities:         equities,
		TotalEquities:    totalEquities,
	}, nil
}
