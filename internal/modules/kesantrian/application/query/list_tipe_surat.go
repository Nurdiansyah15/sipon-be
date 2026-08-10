package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
	"sipon-be/internal/shared/kernel"
)

type ListTipeSuratUseCase struct {
	tipeSuratRepo tiperepo.TipeSuratRepository
}

func NewListTipeSuratUseCase(tipeSuratRepo tiperepo.TipeSuratRepository) *ListTipeSuratUseCase {
	return &ListTipeSuratUseCase{tipeSuratRepo: tipeSuratRepo}
}

func (uc *ListTipeSuratUseCase) Execute(ctx context.Context, q dto.ListTipeSuratQuery) ([]dto.TipeSuratResponse, error) {
	page, limit := resolvePagination(q.Page, q.Limit)

	result, err := uc.tipeSuratRepo.List(ctx, tiperepo.TipeSuratListQuery{
		Page:     page,
		Limit:    limit,
		SortBy:   q.SortBy,
		SortType: q.SortType,
	})
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.TipeSuratResponse, len(result.Items))
	for i, ts := range result.Items {
		items[i] = dto.TipeSuratResponse{
			ID:        ts.ID,
			Nama:      ts.Nama,
			Kode:      ts.Kode,
			CreatedBy: ts.CreatedBy,
			CreatedAt: ts.CreatedAt,
			UpdatedAt: ts.UpdatedAt,
		}
	}

	return items, nil
}
