package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	accRepo "sipon-be/internal/modules/keuangan/domain/account/repository"
	periodRepo "sipon-be/internal/modules/keuangan/domain/period/repository"
	"sipon-be/internal/shared/kernel"
)

type ReportTrialBalanceUseCase struct {
	reportReader ports.ReportReader
	accountRepo  accRepo.AccountRepository
	periodRepo   periodRepo.AccountingPeriodRepository
}

func NewReportTrialBalanceUseCase(reportReader ports.ReportReader, accountRepo accRepo.AccountRepository, periodRepo periodRepo.AccountingPeriodRepository) *ReportTrialBalanceUseCase {
	return &ReportTrialBalanceUseCase{reportReader: reportReader, accountRepo: accountRepo, periodRepo: periodRepo}
}

func (uc *ReportTrialBalanceUseCase) Execute(ctx context.Context, query dto.TrialBalanceQuery) (*dto.TrialBalanceResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, query.PeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	asOfDate := period.EndDate.Format("2006-01-02")
	balances, err := uc.reportReader.AccountBalancesToDate(ctx, &asOfDate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	accountMap := make(map[string]string)
	allAccounts, err := uc.accountRepo.ListAll(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	for _, acc := range allAccounts {
		accountMap[acc.ID] = acc.Code + "|" + acc.Name + "|" + string(acc.Type)
	}

	var totalDebit, totalCredit float64
	lines := make([]dto.TrialBalanceLine, 0, len(balances))
	for _, br := range balances {
		info, ok := accountMap[br.AccountID]
		if !ok {
			info = "||"
		}
		parts := splitN(info, "|", 3)
		line := dto.TrialBalanceLine{
			AccountID:   br.AccountID,
			AccountCode: parts[0],
			AccountName: parts[1],
			AccountType: parts[2],
			Debit:       br.Debit,
			Credit:      br.Credit,
		}
		totalDebit += br.Debit
		totalCredit += br.Credit
		lines = append(lines, line)
	}

	return &dto.TrialBalanceResponse{
		PeriodID:    period.ID,
		PeriodName:  period.Name,
		Lines:       lines,
		TotalDebit:  totalDebit,
		TotalCredit: totalCredit,
	}, nil
}

func splitN(s, sep string, n int) []string {
	result := make([]string, 0, n)
	start := 0
	for i := 0; i < n-1; i++ {
		idx := -1
		for j := start; j < len(s); j++ {
			if s[j:j+len(sep)] == sep {
				idx = j
				break
			}
		}
		if idx == -1 {
			result = append(result, s[start:])
			start = len(s)
			break
		}
		result = append(result, s[start:idx])
		start = idx + len(sep)
	}
	if start <= len(s) {
		result = append(result, s[start:])
	}
	for len(result) < n {
		result = append(result, "")
	}
	return result
}
