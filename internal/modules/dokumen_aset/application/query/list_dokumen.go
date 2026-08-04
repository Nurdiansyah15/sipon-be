package query

import (
	"context"

	"sipon-be/internal/modules/dokumen_aset/application"
	"sipon-be/internal/modules/dokumen_aset/application/dto"
	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	entity "sipon-be/internal/modules/dokumen_aset/domain/dokumen/entity"
	repo "sipon-be/internal/modules/dokumen_aset/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

type ListDokumenAsetUseCase struct {
	dokumenRepo repo.DokumenAsetRepository
}

func NewListDokumenAsetUseCase(dokumenRepo repo.DokumenAsetRepository) *ListDokumenAsetUseCase {
	return &ListDokumenAsetUseCase{dokumenRepo: dokumenRepo}
}

func (uc *ListDokumenAsetUseCase) Execute(ctx context.Context, isAuthenticated bool, query dto.DokumenAsetListQuery) ([]dto.DokumenAsetItem, *dto.DokumenAsetMeta, error) {
	if query.Limit <= 0 {
		query.Limit = 10
	}
	if query.Page <= 0 {
		query.Page = 1
	}

	filter := repo.DokumenAsetFilter{
		Search: query.Search,
		Page:   query.Page,
		Limit:  query.Limit,
	}

	if query.Kategori != "" {
		k := constant.Kategori(query.Kategori)
		filter.Kategori = &k
	}

	if !isAuthenticated {
		filter.PublicOnly = true
	}

	dokumens, total, err := uc.dokumenRepo.List(ctx, filter)
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := mapDokumenAsetItems(dokumens)

	totalPages := (total + query.Limit - 1) / query.Limit
	meta := &dto.DokumenAsetMeta{
		CurrentPage: query.Page,
		PerPage:     query.Limit,
		Total:       total,
		TotalPages:  totalPages,
	}

	return items, meta, nil
}

func mapDokumenAsetItems(dokumens []*entity.DokumenAset) []dto.DokumenAsetItem {
	items := make([]dto.DokumenAsetItem, 0, len(dokumens))
	for _, d := range dokumens {
		items = append(items, dto.DokumenAsetItem{
			ID:        d.ID,
			Judul:     d.Judul,
			Deskripsi: d.Deskripsi,
			Kategori:  string(d.Kategori),
			Filename:  d.Filename,
			MimeType:  d.MimeType,
			Size:      d.Size,
			IsPublic:  d.IsPublic,
			CreatedBy: d.CreatedBy,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		})
	}
	return items
}
