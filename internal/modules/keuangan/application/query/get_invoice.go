package query

import (
	"context"

	invConst "sipon-be/internal/modules/keuangan/domain/invoice/constant"
	invRepo "sipon-be/internal/modules/keuangan/domain/invoice/repository"
	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
)

type GetInvoiceUseCase struct {
	invoiceRepo invRepo.InvoiceRepository
}

func NewGetInvoiceUseCase(invoiceRepo invRepo.InvoiceRepository) *GetInvoiceUseCase {
	return &GetInvoiceUseCase{invoiceRepo: invoiceRepo}
}

func (uc *GetInvoiceUseCase) Execute(ctx context.Context, id string) (*dto.InvoiceResponse, error) {
	inv, err := uc.invoiceRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)
	}

	resp := &dto.InvoiceResponse{
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

	return resp, nil
}
