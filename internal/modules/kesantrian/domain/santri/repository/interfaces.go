package repository

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/santri/entity"
)

type SantriListQuery struct {
	NIS      *string
	Page     int
	Limit    int
	SortBy   string
	SortType string
}

type SantriListResult struct {
	Items []*entity.Santri
	Total int64
}

type SantriRepository interface {
	Save(ctx context.Context, santri *entity.Santri) error
	Update(ctx context.Context, santri *entity.Santri) error
	FindByID(ctx context.Context, id string) (*entity.Santri, error)
	FindByUserID(ctx context.Context, userID string) (*entity.Santri, error)
	FindByNIS(ctx context.Context, nis string) (*entity.Santri, error)
	List(ctx context.Context, query SantriListQuery) (*SantriListResult, error)
}
