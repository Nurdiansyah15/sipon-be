package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/billingbatch/entity"
)

type BillingBatchListQuery struct {
	Status *string
	Page   int
	Limit  int
}

type BillingBatchListResult struct {
	Items []*entity.BillingBatch
	Total int64
}

type BillingBatchRepository interface {
	Save(ctx context.Context, batch *entity.BillingBatch) error
	Update(ctx context.Context, batch *entity.BillingBatch) error
	FindByID(ctx context.Context, id string) (*entity.BillingBatch, error)
	List(ctx context.Context, query BillingBatchListQuery) (*BillingBatchListResult, error)
}

type BillingBatchTargetRepository interface {
	SaveMany(ctx context.Context, targets []*entity.BillingBatchTarget) error
	UpdateTarget(ctx context.Context, target *entity.BillingBatchTarget) error
	FindByBatchID(ctx context.Context, batchID string) ([]*entity.BillingBatchTarget, error)
}
