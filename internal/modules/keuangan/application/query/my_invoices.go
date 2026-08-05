package query

import (
	"context"

	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type MyInvoicesUseCase struct {
	invoiceRepo invRepo.InvoiceRepository
}

func NewMyInvoicesUseCase(invoiceRepo invRepo.InvoiceRepository) *MyInvoicesUseCase {
	return &MyInvoicesUseCase{invoiceRepo: invoiceRepo}
}

func (uc *MyInvoicesUseCase) Execute(ctx context.Context, userID string, query dto.InvoiceListQuery) ([]dto.InvoiceResponse, *dto.Meta, error) {
	repoQuery := invRepo.InvoiceListQuery{
		UserID:      &userID,
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
		resp := dto.InvoiceResponse{
			ID:              inv.ID,
			InvoiceNumber:   inv.InvoiceNumber,
			SantriID:        inv.SantriID,
			UserID:          inv.UserID,
			BillingSchemeID: inv.BillingSchemeID,
			FeeComponentID:  inv.FeeComponentID,
			Periode:         inv.Periode,
			TahunAjaran:     inv.TahunAjaran,
			Amount:          inv.Amount,
			DiscountAmount:  inv.DiscountAmount,
			PaidAmount:      inv.PaidAmount,
			Status:          string(inv.Status),
			DueDate:         inv.DueDate.Format("2006-01-02"),
			Notes:           inv.Notes,
			CreatedAt:       inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:       inv.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if inv.IssuedAt != nil {
			s := inv.IssuedAt.Format("2006-01-02T15:04:05Z07:00")
			resp.IssuedAt = &s
		}
		items[i] = resp
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
