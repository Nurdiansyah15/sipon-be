package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/academic_period/entity"
)

type AcademicPeriodListQuery struct {
	Status *string
	Search *string
	Page   int
	Limit  int
}

type AcademicPeriodListResult struct {
	Items []*entity.AcademicPeriod
	Total int64
}

type AcademicPeriodRepository interface {
	Save(ctx context.Context, period *entity.AcademicPeriod) error
	Update(ctx context.Context, period *entity.AcademicPeriod) error
	FindByID(ctx context.Context, id string) (*entity.AcademicPeriod, error)
	FindByCode(ctx context.Context, code string) (*entity.AcademicPeriod, error)
	FindByIDs(ctx context.Context, ids []string) ([]*entity.AcademicPeriod, error)
	// FindOpen returns the most recent open period (highest start_date).
	FindOpen(ctx context.Context) (*entity.AcademicPeriod, error)
	List(ctx context.Context, query AcademicPeriodListQuery) (*AcademicPeriodListResult, error)
	HasData(ctx context.Context, id string) (bool, error)
}
