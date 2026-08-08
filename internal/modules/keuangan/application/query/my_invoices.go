package query

import (
	"context"

	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type MyInvoicesUseCase struct {
	invoiceRepo       invRepo.InvoiceRepository
	feeComponentRepo  feeRepo.FeeComponentRepository
	billingPeriodRepo bpRepo.BillingPeriodRepository
}

func NewMyInvoicesUseCase(invoiceRepo invRepo.InvoiceRepository, feeComponentRepo feeRepo.FeeComponentRepository, billingPeriodRepo bpRepo.BillingPeriodRepository) *MyInvoicesUseCase {
	return &MyInvoicesUseCase{invoiceRepo: invoiceRepo, feeComponentRepo: feeComponentRepo, billingPeriodRepo: billingPeriodRepo}
}

func (uc *MyInvoicesUseCase) Execute(ctx context.Context, userID string, query dto.InvoiceListQuery) ([]dto.InvoiceResponse, *dto.Meta, error) {
	repoQuery := invRepo.InvoiceListQuery{
		UserID:          &userID,
		Status:          query.Status,
		BillingPeriodID: query.BillingPeriodID,
		Page:            query.Page,
		Limit:           query.Limit,
	}
	if repoQuery.Page == 0 {
		repoQuery.Page = 1
	}
	if repoQuery.Limit == 0 {
		repoQuery.Limit = 20
	}

	result, err := uc.invoiceRepo.List(ctx, repoQuery)
	if err != nil {
		return nil, nil, application.WrapRepoErr(err, invConst.CodeInvoiceQueryFailed)
	}

	items := make([]dto.InvoiceResponse, len(result.Items))
	for i, inv := range result.Items {
		items[i] = buildInvoiceResponse(ctx, inv, uc.feeComponentRepo, uc.billingPeriodRepo)
	}

	totalPages := (result.Total + int64(repoQuery.Limit) - 1) / int64(repoQuery.Limit)
	meta := &dto.Meta{
		Page:       repoQuery.Page,
		Limit:      repoQuery.Limit,
		Total:      result.Total,
		TotalPages: totalPages,
	}

	return items, meta, nil
}
