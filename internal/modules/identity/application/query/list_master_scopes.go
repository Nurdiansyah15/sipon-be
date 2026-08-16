package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	scoperepo "sipon-be/internal/modules/identity/domain/scope/repository"
	"sipon-be/internal/shared/kernel"
)

type ListMasterScopesUseCase struct {
	scopeRepo scoperepo.ScopeRepository
}

func NewListMasterScopesUseCase(scopeRepo scoperepo.ScopeRepository) *ListMasterScopesUseCase {
	return &ListMasterScopesUseCase{scopeRepo: scopeRepo}
}

func (uc *ListMasterScopesUseCase) Execute(ctx context.Context, req dto.ListScopesRequest) ([]dto.MasterScopeItem, error) {
	req.ScopeType = strings.TrimSpace(strings.ToLower(req.ScopeType))

	scopes, err := uc.scopeRepo.ListAll(ctx, req.IncludeInactive)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mendaftar scope", err)
	}

	items := make([]dto.MasterScopeItem, 0, len(scopes))
	for _, s := range scopes {
		if req.ScopeType != "" && string(s.ScopeType) != req.ScopeType {
			continue
		}
		items = append(items, *dto.ToMasterScopeItem(s))
	}
	return items, nil
}
