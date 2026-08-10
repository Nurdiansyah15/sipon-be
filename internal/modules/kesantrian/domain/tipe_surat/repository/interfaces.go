package repository

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/tipe_surat/entity"
)

type TipeSuratListQuery struct {
	Page     int
	Limit    int
	SortBy   string
	SortType string
}

type TipeSuratListResult struct {
	Items []*entity.TipeSurat
	Total int64
}

type TipeSuratRepository interface {
	Save(ctx context.Context, ts *entity.TipeSurat) error
	Update(ctx context.Context, ts *entity.TipeSurat) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.TipeSurat, error)
	List(ctx context.Context, q TipeSuratListQuery) (*TipeSuratListResult, error)
	IsInUse(ctx context.Context, tipeSuratID string) (bool, error)
}
