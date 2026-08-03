package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	"sipon-be/internal/shared/kernel"
)

type ListPendaftaranUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
}

func NewListPendaftaranUseCase(pendaftarRepo prepo.PendaftarRepository) *ListPendaftaranUseCase {
	return &ListPendaftaranUseCase{pendaftarRepo: pendaftarRepo}
}

func (uc *ListPendaftaranUseCase) Execute(ctx context.Context, q dto.ListPendaftarQuery) ([]dto.ListPendaftarItem, *dto.PaginationMeta, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 || q.Limit > 100 {
		q.Limit = 10
	}

	var statusPtr *string
	if q.Status != "" {
		statusPtr = &q.Status
	}

	result, err := uc.pendaftarRepo.List(ctx, prepo.PendaftarListQuery{
		PsbSettingID: q.PsbSettingID,
		Status:       statusPtr,
		Page:         q.Page,
		Limit:        q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ListPendaftarItem, len(result.Items))
	for i, p := range result.Items {
		items[i] = dto.ListPendaftarItem{
			ID:           p.ID,
			UserID:       p.UserID,
			PsbSettingID: p.PsbSettingID,
			Gender:       p.Gender,
			Program:      p.Program,
			Status:       string(p.Status),
			NIS:          p.NIS,
			CreatedAt:    p.CreatedAt.Format("2006-01-02"),
		}
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(q.Limit)))
	meta := &dto.PaginationMeta{
		CurrentPage: q.Page,
		PerPage:     q.Limit,
		Total:       int(result.Total),
		TotalPages:  totalPages,
	}

	return items, meta, nil
}
