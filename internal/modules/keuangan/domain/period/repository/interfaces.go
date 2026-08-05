package repository

import (
	"context"
	"time"

	"sipon-be/internal/modules/keuangan/domain/period/entity"
)

type PeriodListQuery struct {
	Status *string
	Page   int
	Limit  int
}

type PeriodListResult struct {
	Items []*entity.AccountingPeriod
	Total int64
}

type AccountingPeriodRepository interface {
	Save(ctx context.Context, period *entity.AccountingPeriod) error
	Update(ctx context.Context, period *entity.AccountingPeriod) error
	FindByID(ctx context.Context, id string) (*entity.AccountingPeriod, error)
	FindActive(ctx context.Context) (*entity.AccountingPeriod, error)
	List(ctx context.Context, query PeriodListQuery) (*PeriodListResult, error)
	FindByDate(ctx context.Context, date time.Time) (*entity.AccountingPeriod, error)
	HasOverlap(ctx context.Context, startDate, endDate time.Time, excludeID string) (bool, error)
}
