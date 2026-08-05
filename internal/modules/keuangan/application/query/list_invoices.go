package query

import (
	"context"

	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type ListInvoicesUseCase struct {
	invoiceRepo      invRepo.InvoiceRepository
	feeComponentRepo feeRepo.FeeComponentRepository
}

func NewListInvoicesUseCase(invoiceRepo invRepo.InvoiceRepository, feeComponentRepo feeRepo.FeeComponentRepository) *ListInvoicesUseCase {
	return &ListInvoicesUseCase{invoiceRepo: invoiceRepo, feeComponentRepo: feeComponentRepo}
}

func (uc *ListInvoicesUseCase) Execute(ctx context.Context, query dto.InvoiceListQuery) ([]dto.InvoiceResponse, *dto.Meta, error) {
	repoQuery := invRepo.InvoiceListQuery{
		SantriID:    query.SantriID,
		UserID:      query.UserID,
		Status:      query.Status,
		Periode:     query.Periode,
		TahunAjaran: query.TahunAjaran,
		Page:        query.Page,
		Limit:       query.Limit,
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
		items[i] = buildInvoiceResponse(ctx, inv, uc.feeComponentRepo)
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
