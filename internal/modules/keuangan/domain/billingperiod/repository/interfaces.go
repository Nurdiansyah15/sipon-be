package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/billingperiod/entity"
)

type BillingPeriodListQuery struct {
	Status *string
	Page   int
	Limit  int
}

type BillingPeriodListResult struct {
	Items []*entity.BillingPeriod
	Total int64
}

type BillingPeriodRepository interface {
	Save(ctx context.Context, period *entity.BillingPeriod) error
	Update(ctx context.Context, period *entity.BillingPeriod) error
	FindByID(ctx context.Context, id string) (*entity.BillingPeriod, error)
	List(ctx context.Context, query BillingPeriodListQuery) (*BillingPeriodListResult, error)
}
