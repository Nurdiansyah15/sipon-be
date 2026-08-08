package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingbatch/constant"
)

type BillingBatchTarget struct {
	ID          string
	BatchID     string
	SantriID    string
	Status      constant.BillingBatchTargetStatus
	InvoiceID   *string
	Reason      *string
	ProcessedAt *time.Time
}

func NewBillingBatchTarget(id, batchID, santriID string, status constant.BillingBatchTargetStatus) *BillingBatchTarget {
	return &BillingBatchTarget{
		ID:       id,
		BatchID:  batchID,
		SantriID: santriID,
		Status:   status,
	}
}

func (t *BillingBatchTarget) MarkProcessed(status constant.BillingBatchTargetStatus, invoiceID, reason *string) {
	now := time.Now()
	t.Status = status
	t.InvoiceID = invoiceID
	t.Reason = reason
	t.ProcessedAt = &now
}
