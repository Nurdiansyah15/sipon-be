package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application/dto"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
)

// buildInvoiceResponse maps an invoice entity to its response DTO, best-effort
// enriching it with the fee component brief when feeComponentRepo is provided.
func buildInvoiceResponse(ctx context.Context, inv *invEntity.Invoice, feeComponentRepo feeRepo.FeeComponentRepository) dto.InvoiceResponse {
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
	if feeComponentRepo != nil {
		if fc, err := feeComponentRepo.FindByID(ctx, inv.FeeComponentID); err == nil {
			resp.FeeComponent = &dto.FeeComponentBriefResponse{
				ID:     fc.ID,
				Code:   fc.Code,
				Name:   fc.Name,
				Type:   string(fc.Type),
				Amount: fc.Amount,
			}
		}
	}
	return resp
}
