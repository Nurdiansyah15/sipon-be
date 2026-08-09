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

type ReportIncomeStatementUseCase struct {
	reportReader ports.ReportReader
	accountRepo  accRepo.AccountRepository
	periodRepo   periodRepo.AccountingPeriodRepository
}

func NewReportIncomeStatementUseCase(reportReader ports.ReportReader, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *ReportIncomeStatementUseCase {
	return &ReportIncomeStatementUseCase{reportReader: reportReader, accountRepo: accountRepo, periodRepo: periodRepo}
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

	balanceRows, err := uc.reportReader.AccountBalancesByPeriod(ctx, query.PeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	balances := make(map[string]float64, len(balanceRows))
	for _, bal := range balanceRows {
		balances[bal.AccountID] = bal.Credit - bal.Debit
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
