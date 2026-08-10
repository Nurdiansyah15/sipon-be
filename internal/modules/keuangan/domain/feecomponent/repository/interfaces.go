package repository

import (
	"context"

	"sipon-be/internal/modules/keuangan/domain/feecomponent/entity"
)

type FeeComponentListQuery struct {
	Active *bool
	Page   int
	Limit  int
}

type FeeComponentListResult struct {
	Items []*entity.FeeComponent
	Total int64
}

type FeeComponentRepository interface {
	Save(ctx context.Context, fc *entity.FeeComponent) error
	Update(ctx context.Context, fc *entity.FeeComponent) error
	FindByID(ctx context.Context, id string) (*entity.FeeComponent, error)
	FindByCode(ctx context.Context, code string) (*entity.FeeComponent, error)
	List(ctx context.Context, query FeeComponentListQuery) (*FeeComponentListResult, error)
	ExistsByCode(ctx context.Context, code string, excludeID string) (bool, error)
}
