package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/billingbatch/constant"
	"sipon-be/internal/shared/kernel"
)

type BillingBatch struct {
	ID              string
	Name            string
	BillingSchemeID string
	BillingPeriodID string
	Status          constant.BillingBatchStatus
	CreatedBy       string
	CreatedAt       time.Time
	CompletedAt     *time.Time
	TotalCreated    int
	TotalSkipped    int
	TotalError      int
}

func NewBillingBatch(id, name, billingSchemeID, billingPeriodID, createdBy string) (*BillingBatch, error) {
	if id == "" || name == "" || billingSchemeID == "" || billingPeriodID == "" || createdBy == "" {
		return nil, kernel.New(constant.CodeBillingBatchNotFound)
	}
	return &BillingBatch{
		ID:              id,
		Name:            name,
		BillingSchemeID: billingSchemeID,
		BillingPeriodID: billingPeriodID,
		Status:          constant.BillingBatchProcessing,
		CreatedBy:       createdBy,
		CreatedAt:       time.Now(),
	}, nil
}

func (b *BillingBatch) Complete(created, skipped, errored int) {
	now := time.Now()
	b.Status = constant.BillingBatchCompleted
	b.CompletedAt = &now
	b.TotalCreated = created
	b.TotalSkipped = skipped
	b.TotalError = errored
}
