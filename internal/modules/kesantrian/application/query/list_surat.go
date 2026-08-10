package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"
)

type ListSuratUseCase struct {
	suratRepo suratrepo.SuratRepository
}

func NewListSuratUseCase(suratRepo suratrepo.SuratRepository) *ListSuratUseCase {
	return &ListSuratUseCase{suratRepo: suratRepo}
}

func (uc *ListSuratUseCase) Execute(ctx context.Context, q dto.ListSuratQuery) ([]dto.SuratResponse, dto.Meta, error) {
	page, limit := resolvePagination(q.Page, q.Limit)

	var tipeSuratID *string
	if q.TipeSuratID != "" {
		tipeSuratID = &q.TipeSuratID
	}
	var search *string
	if q.Search != "" {
		search = &q.Search
	}

	result, err := uc.suratRepo.List(ctx, suratrepo.SuratListQuery{
		TipeSuratID: tipeSuratID,
		Bulan:       q.Bulan,
		Tahun:       q.Tahun,
		Search:      search,
		Page:        page,
		Limit:       limit,
		SortBy:      q.SortBy,
		SortType:    q.SortType,
	})
	if err != nil {
		return nil, dto.Meta{}, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.SuratResponse, len(result.Items))
	for i, s := range result.Items {
		items[i] = dto.SuratResponse{
			ID:          s.ID,
			Nomor:       s.Nomor,
			TipeSuratID: s.TipeSuratID,
			Keterangan:  s.Keterangan,
			Tanggal:     s.Tanggal.Format("2006-01-02"),
			CreatedBy:   s.CreatedBy,
			CreatedAt:   s.CreatedAt,
		}
	}

	return items, buildMeta(page, limit, result.Total), nil
}
