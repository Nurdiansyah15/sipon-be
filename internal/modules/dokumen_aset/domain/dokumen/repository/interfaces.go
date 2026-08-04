package repository

import (
	"context"

	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	entity "sipon-be/internal/modules/dokumen_aset/domain/dokumen/entity"
)

type DokumenAsetRepository interface {
	Save(ctx context.Context, d *entity.DokumenAset) error
	Update(ctx context.Context, d *entity.DokumenAset) error
	FindByID(ctx context.Context, id string) (*entity.DokumenAset, error)
	List(ctx context.Context, filter DokumenAsetFilter) ([]*entity.DokumenAset, int, error)
}

type DokumenAsetFilter struct {
	Kategori *constant.Kategori
	PublicOnly bool
	Search     string
	Page       int
	Limit      int
}
