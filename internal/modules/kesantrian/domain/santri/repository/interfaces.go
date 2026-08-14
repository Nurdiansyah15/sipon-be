package repository

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
)

type SantriListQuery struct {
	NIS      *string
	Page     int
	Limit    int
	SortBy   string
	SortType string
	// Scope membatasi daftar santri sesuai akses scope user. Zero value
	// (santriscope.ScopeSet{}) diperlakukan sebagai tanpa filter.
	Scope santriscope.ScopeSet
}

type SantriListResult struct {
	Items []*entity.Santri
	Total int64
}

type SantriBasicInfo struct {
	SantriID string
	UserID   string
	NIS      *string
	Status   string
}

type SantriRepository interface {
	Save(ctx context.Context, santri *entity.Santri) error
	Update(ctx context.Context, santri *entity.Santri) error
	FindByID(ctx context.Context, id string) (*entity.Santri, error)
	FindByUserID(ctx context.Context, userID string) (*entity.Santri, error)
	FindByNIS(ctx context.Context, nis string) (*entity.Santri, error)
	FindMaxSequence(ctx context.Context, prefix string) (int, error)
	List(ctx context.Context, query SantriListQuery) (*SantriListResult, error)
	ListActiveIDs(ctx context.Context) ([]string, error)
	FindBasicByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error)
	FindBasicByID(ctx context.Context, santriID string) (*SantriBasicInfo, error)
	FindBasicByNIS(ctx context.Context, nis string) (*SantriBasicInfo, error)
	ListActiveWithUserID(ctx context.Context) ([]SantriBasicInfo, error)
}
