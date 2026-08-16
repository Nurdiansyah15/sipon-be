package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"
)

type ListSuratUseCase struct {
	suratRepo   suratrepo.SuratRepository
	scopeReader ports.ScopeReader
}

func NewListSuratUseCase(suratRepo suratrepo.SuratRepository, scopeReader ports.ScopeReader) *ListSuratUseCase {
	return &ListSuratUseCase{suratRepo: suratRepo, scopeReader: scopeReader}
}

func (uc *ListSuratUseCase) Execute(ctx context.Context, userID string, q dto.ListSuratQuery) ([]dto.SuratResponse, dto.Meta, error) {
	page, limit := resolvePagination(q.Page, q.Limit)

	var tipeSuratID *string
	if q.TipeSuratID != "" {
		tipeSuratID = &q.TipeSuratID
	}
	var search *string
	if q.Search != "" {
		search = &q.Search
	}

	scopeSet, err := uc.resolveScope(ctx, userID)
	if err != nil {
		return nil, dto.Meta{}, err
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
		Scope:       scopeSet,
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
			ScopeID:     s.ScopeID,
			CreatedAt:   s.CreatedAt,
		}
	}

	return items, buildMeta(page, limit, result.Total), nil
}

// resolveScope membangun ScopeSet untuk filter surat. AllowedOptions berisi
// master scope ID. Akses penuh -> Unrestricted; tanpa akses -> Restricted(nil).
func (uc *ListSuratUseCase) resolveScope(ctx context.Context, userID string) (santriscope.ScopeSet, error) {
	if uc.scopeReader == nil {
		return santriscope.ScopeSet{}, nil
	}
	access, err := uc.scopeReader.GetSuratScopeAccess(ctx, userID)
	if err != nil {
		return santriscope.ScopeSet{}, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if !access.HasAccess {
		return santriscope.Restricted(nil), nil
	}
	if access.HasFullAccess {
		return santriscope.Unrestricted(), nil
	}
	return santriscope.Restricted(access.AllowedScopeIDs), nil
}
