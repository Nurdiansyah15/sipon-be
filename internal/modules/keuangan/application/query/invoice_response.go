package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application/dto"
	bpRepo "sipon-be/internal/modules/keuangan/domain/billingperiod/repository"
	feeRepo "sipon-be/internal/modules/keuangan/domain/feecomponent/repository"
	invEntity "sipon-be/internal/modules/keuangan/domain/invoice/entity"
)

// buildInvoiceResponse maps an invoice entity to its response DTO, best-effort
// enriching it with the fee component and billing period briefs when their
// repositories are provided.
func buildInvoiceResponse(ctx context.Context, inv *invEntity.Invoice, feeComponentRepo feeRepo.FeeComponentRepository, billingPeriodRepo bpRepo.BillingPeriodRepository) dto.InvoiceResponse {
	resp := dto.InvoiceResponse{
		ID:              inv.ID,
		InvoiceNumber:   inv.InvoiceNumber,
		SantriID:        inv.SantriID,
		UserID:          inv.UserID,
		BillingSchemeID: inv.BillingSchemeID,
		FeeComponentID:  inv.FeeComponentID,
		BillingPeriodID: inv.BillingPeriodID,
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
	if billingPeriodRepo != nil {
		if bp, err := billingPeriodRepo.FindByID(ctx, inv.BillingPeriodID); err == nil {
			resp.BillingPeriod = &dto.BillingPeriodBriefResponse{
				ID:     bp.ID,
				Name:   bp.Name,
				Status: string(bp.Status),
			}
		}
	}
	return resp
}
