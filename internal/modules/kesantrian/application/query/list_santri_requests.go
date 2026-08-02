package query

import (
	"context"
	"log/slog"
	"strings"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	requestconstant "sipon-be/internal/modules/kesantrian/domain/request/constant"
	requestrepo "sipon-be/internal/modules/kesantrian/domain/request/repository"
	"sipon-be/internal/shared/kernel"
)

type ListSantriRequestsUseCase struct {
	requestRepo requestrepo.SantriRequestRepository
	provisioner ports.AccountProvisioner
}

func NewListSantriRequestsUseCase(requestRepo requestrepo.SantriRequestRepository, provisioner ports.AccountProvisioner) *ListSantriRequestsUseCase {
	return &ListSantriRequestsUseCase{requestRepo: requestRepo, provisioner: provisioner}
}

func (uc *ListSantriRequestsUseCase) Execute(ctx context.Context, req dto.ListSantriRequestsQuery) ([]dto.SantriRequestItem, dto.Meta, error) {
	page, limit := resolvePagination(req.Page, req.Limit)

	var statusFilter *requestconstant.SantriRequestStatus
	if trimmed := strings.TrimSpace(req.Status); trimmed != "" {
		s := requestconstant.SantriRequestStatus(trimmed)
		statusFilter = &s
	}

	result, err := uc.requestRepo.List(ctx, requestrepo.SantriRequestListQuery{
		Status:   statusFilter,
		Page:     page,
		Limit:    limit,
		SortBy:   req.SortBy,
		SortType: req.SortType,
	})
	if err != nil {
		return nil, dto.Meta{}, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.SantriRequestItem, 0, len(result.Items))
	for _, r := range result.Items {
		item := dto.SantriRequestItem{
			ID:        r.ID,
			UserID:    r.UserID,
			Status:    string(r.Status),
			Notes:     r.Notes,
			CreatedAt: r.CreatedAt,
		}

		// N+1 by design — see list_santri.go's comment.
		summary, err := uc.provisioner.GetUserSummary(ctx, r.UserID)
		if err != nil {
			slog.Warn("kesantrian: user summary enrichment failed", "user_id", r.UserID, "error", err)
		} else {
			item.Username = summary.Username
			item.Email = summary.Email
			item.Fullname = summary.Fullname
		}

		items = append(items, item)
	}

	return items, buildMeta(page, limit, result.Total), nil
}
