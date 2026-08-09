package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	"sipon-be/internal/shared/kernel"
)

type ReportSummaryUseCase struct {
	reportReader ports.ReportReader
}

func NewReportSummaryUseCase(reportReader ports.ReportReader) *ReportSummaryUseCase {
	return &ReportSummaryUseCase{reportReader: reportReader}
}

func (uc *ReportSummaryUseCase) Execute(ctx context.Context, query dto.InvoiceSummaryQuery) ([]dto.InvoiceSummaryResponse, error) {
	items, err := uc.reportReader.InvoiceSummary(ctx, query.BillingPeriodID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	results := make([]dto.InvoiceSummaryResponse, 0, len(items))
	for _, it := range items {
		results = append(results, dto.InvoiceSummaryResponse{
			BillingPeriodID:   it.BillingPeriodID,
			BillingPeriodName: it.BillingPeriodName,
			TotalTagihan:      it.TotalTagihan,
			TotalTerbayar:     it.TotalTerbayar,
			TotalTunggakan:    it.TotalTunggakan,
			JumlahInvoice:     it.JumlahInvoice,
			JumlahLunas:       it.JumlahLunas,
			JumlahBelum:       it.JumlahBelum,
		})
	}
	return results, nil
}
