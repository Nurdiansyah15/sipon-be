package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	"sipon-be/internal/shared/kernel"
)

type ReportOutstandingUseCase struct {
	reportReader ports.ReportReader
}

func NewReportOutstandingUseCase(reportReader ports.ReportReader) *ReportOutstandingUseCase {
	return &ReportOutstandingUseCase{reportReader: reportReader}
}

func (uc *ReportOutstandingUseCase) Execute(ctx context.Context, query dto.OutstandingListQuery) ([]dto.OutstandingSantriResponse, *dto.Meta, error) {
	page := query.Page
	limit := query.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	items, total, err := uc.reportReader.OutstandingBySantri(ctx, query.BillingPeriodID, page, limit)
	if err != nil {
		return nil, nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	results := make([]dto.OutstandingSantriResponse, 0, len(items))
	for _, it := range items {
		results = append(results, dto.OutstandingSantriResponse{
			SantriID:         it.SantriID,
			TotalOutstanding: it.TotalOutstanding,
			JumlahInvoice:    it.JumlahInvoice,
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)
	meta := &dto.Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	return results, meta, nil
}
