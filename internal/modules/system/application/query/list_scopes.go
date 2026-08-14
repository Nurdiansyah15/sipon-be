package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/system/application"
	"sipon-be/internal/modules/system/application/dto"
	scoperepo "sipon-be/internal/modules/system/domain/scope/repository"
	"sipon-be/internal/shared/kernel"
)

type ListScopesUseCase struct {
	scopeRepo scoperepo.ScopeRepository
}

func NewListScopesUseCase(scopeRepo scoperepo.ScopeRepository) *ListScopesUseCase {
	return &ListScopesUseCase{scopeRepo: scopeRepo}
}

func (uc *ListScopesUseCase) Execute(ctx context.Context, req dto.ListScopesRequest) ([]dto.ScopeItem, error) {
	req.ScopeType = strings.TrimSpace(strings.ToLower(req.ScopeType))

	scopes, err := uc.scopeRepo.ListAll(ctx, req.IncludeInactive)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mendaftar scope", err)
	}

	items := make([]dto.ScopeItem, 0, len(scopes))
	for _, s := range scopes {
		if req.ScopeType != "" && string(s.ScopeType) != req.ScopeType {
			continue
		}
		items = append(items, *dto.ToScopeItem(s))
	}
	return items, nil
}
