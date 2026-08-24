package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	ports "sipon-be/internal/modules/psb/application/ports"
	pendaftarscope "sipon-be/internal/modules/psb/domain/pendaftar/scope"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	"sipon-be/internal/shared/kernel"
)

type ListPendaftaranUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	scopeReader   ports.ScopeReader
}

func NewListPendaftaranUseCase(pendaftarRepo prepo.PendaftarRepository, scopeReader ports.ScopeReader) *ListPendaftaranUseCase {
	return &ListPendaftaranUseCase{pendaftarRepo: pendaftarRepo, scopeReader: scopeReader}
}

func (uc *ListPendaftaranUseCase) Execute(ctx context.Context, userID string, q dto.ListPendaftarQuery) ([]dto.ListPendaftarItem, *dto.PaginationMeta, error) {
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

	scopeSet, err := uc.resolveScope(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	result, err := uc.pendaftarRepo.List(ctx, prepo.PendaftarListQuery{
		PsbSettingID: q.PsbSettingID,
		Status:       statusPtr,
		Page:         q.Page,
		Limit:        q.Limit,
		Scope:        scopeSet,
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
			ProgramID:    p.ProgramID,
			Status:       string(p.Status),
			NIS:          p.NIS,
			NoRegis:      p.NoRegis,
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

// resolveScope membangun ScopeSet untuk filter daftar pendaftar. Akses penuh ->
// Unrestricted; tanpa akses -> Restricted(nil) sehingga repository mengembalikan
// daftar kosong.
func (uc *ListPendaftaranUseCase) resolveScope(ctx context.Context, userID string) (pendaftarscope.ScopeSet, error) {
	if uc.scopeReader == nil {
		return pendaftarscope.ScopeSet{}, nil
	}
	scopeSet, err := uc.scopeReader.GetPendaftarScopeSet(ctx, userID)
	if err != nil {
		return pendaftarscope.ScopeSet{}, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return scopeSet, nil
}