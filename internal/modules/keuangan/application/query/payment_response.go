package query

import (
	"sipon-be/internal/modules/keuangan/application/dto"
	adjEntity "sipon-be/internal/modules/keuangan/domain/adjustment/entity"
	payEntity "sipon-be/internal/modules/keuangan/domain/payment/entity"
)

// buildPaymentResponse maps a payment entity to its base response DTO,
// without the nested invoice/debit account brief (callers enrich those
// themselves when needed, to avoid redundant nesting when a payment is
// already being built as part of its parent invoice's response).
func buildPaymentResponse(p *payEntity.Payment) dto.PaymentResponse {
	resp := dto.PaymentResponse{
		ID:              p.ID,
		PaymentNumber:   p.PaymentNumber,
		InvoiceID:       p.InvoiceID,
		DebitAccountID:  p.DebitAccountID,
		Amount:          p.Amount,
		Method:          string(p.Method),
		ReferenceNumber: p.ReferenceNumber,
		PaymentDate:     p.PaymentDate.Format("2006-01-02"),
		Status:          string(p.Status),
		VerifiedBy:      p.VerifiedBy,
		Notes:           p.Notes,
		ProofKey:        p.ProofKey,
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.VerifiedAt != nil {
		s := p.VerifiedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.VerifiedAt = &s
	}
	return resp
}

func buildAdjustmentResponse(a *adjEntity.InvoiceAdjustment) dto.InvoiceAdjustmentResponse {
	return dto.InvoiceAdjustmentResponse{
		ID:          a.ID,
		InvoiceID:   a.InvoiceID,
		Type:        string(a.Type),
		Amount:      a.Amount,
		Percentage:  a.Percentage,
		Description: a.Description,
		AppliedBy:   a.AppliedBy,
		AppliedAt:   a.AppliedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
