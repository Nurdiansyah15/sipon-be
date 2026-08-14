package query

import (
	"context"
	"log/slog"
	"strings"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	"sipon-be/internal/shared/kernel"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

func resolvePagination(page, limit int) (int, int) {
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit
}

func buildMeta(page, limit int, total int64) dto.Meta {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	return dto.Meta{
		CurrentPage: page,
		PerPage:     limit,
		Total:       int(total),
		TotalPages:  totalPages,
	}
}

type ListSantriUseCase struct {
	santriRepo  santrirepo.SantriRepository
	provisioner ports.AccountProvisioner
	scopeReader ports.ScopeReader
}

func NewListSantriUseCase(santriRepo santrirepo.SantriRepository, provisioner ports.AccountProvisioner, scopeReader ports.ScopeReader) *ListSantriUseCase {
	return &ListSantriUseCase{santriRepo: santriRepo, provisioner: provisioner, scopeReader: scopeReader}
}

func (uc *ListSantriUseCase) Execute(ctx context.Context, userID string, req dto.ListSantriQuery) ([]dto.ListSantriItem, dto.Meta, error) {
	page, limit := resolvePagination(req.Page, req.Limit)

	var nisFilter *string
	if trimmed := strings.TrimSpace(req.NIS); trimmed != "" {
		nisFilter = &trimmed
	}

	scopeSet, err := uc.scopeReader.GetSantriScopeSet(ctx, userID)
	if err != nil {
		return nil, dto.Meta{}, kernel.Wrap(application.ErrCodeInternal, err)
	}

	result, err := uc.santriRepo.List(ctx, santrirepo.SantriListQuery{
		NIS:      nisFilter,
		Page:     page,
		Limit:    limit,
		SortBy:   req.SortBy,
		SortType: req.SortType,
		Scope:    scopeSet,
	})
	if err != nil {
		return nil, dto.Meta{}, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ListSantriItem, 0, len(result.Items))
	for _, s := range result.Items {
		item := dto.ListSantriItem{
			ID:        s.ID,
			UserID:    s.UserID,
			CreatedAt: s.CreatedAt,
		}
		if s.NIS != nil {
			nis := s.NIS.String()
			item.NIS = &nis
		}

		// N+1 by design: kesantrian cannot batch-query identity's users
		// table directly (module boundary). A future
		// identity.Contract.GetUserSummaries(ctx, []string) batch method
		// would fix this, but is deliberately not added yet (YAGNI) — see
		// docs/plans notes. Enrichment failure is logged, not fatal.
		summary, err := uc.provisioner.GetUserSummary(ctx, s.UserID)
		if err != nil {
			slog.Warn("kesantrian: user summary enrichment failed", "user_id", s.UserID, "error", err)
		} else {
			item.Username = summary.Username
			item.Email = summary.Email
			item.Fullname = summary.Fullname
			item.Status = boolToStatus(summary.IsActive)
		}

		items = append(items, item)
	}

	return items, buildMeta(page, limit, result.Total), nil
}

func boolToStatus(isActive bool) string {
	if isActive {
		return "ACTIVE"
	}
	return "BANNED"
}
