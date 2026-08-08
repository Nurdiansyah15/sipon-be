package query

import (
	"sipon-be/internal/modules/keuangan/application/dto"
	bbEntity "sipon-be/internal/modules/keuangan/domain/billingbatch/entity"
)

func buildBillingBatchResponse(b *bbEntity.BillingBatch) dto.BillingBatchResponse {
	resp := dto.BillingBatchResponse{
		ID:              b.ID,
		Name:            b.Name,
		BillingSchemeID: b.BillingSchemeID,
		BillingPeriodID: b.BillingPeriodID,
		Status:          string(b.Status),
		CreatedBy:       b.CreatedBy,
		CreatedAt:       b.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		TotalCreated:    b.TotalCreated,
		TotalSkipped:    b.TotalSkipped,
		TotalError:      b.TotalError,
	}
	if b.CompletedAt != nil {
		s := b.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &s
	}
	return resp
}

func buildBillingBatchTargetResponse(t *bbEntity.BillingBatchTarget) dto.BillingBatchTargetResponse {
	resp := dto.BillingBatchTargetResponse{
		ID:        t.ID,
		SantriID:  t.SantriID,
		Status:    string(t.Status),
		InvoiceID: t.InvoiceID,
		Reason:    t.Reason,
	}
	if t.ProcessedAt != nil {
		s := t.ProcessedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.ProcessedAt = &s
	}
	return resp
}
