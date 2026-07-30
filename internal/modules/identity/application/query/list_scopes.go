package query

import (
	"context"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
)

type ListScopesUseCase struct {
	roleScopeRepo domain.RoleScopeRepository
}

func NewListScopesUseCase(roleScopeRepo domain.RoleScopeRepository) *ListScopesUseCase {
	return &ListScopesUseCase{roleScopeRepo: roleScopeRepo}
}

func (uc *ListScopesUseCase) Execute(ctx context.Context, roleID string) (*dto.ListScopesResponse, error) {
	scopes, err := uc.roleScopeRepo.FindByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	items := make([]dto.ScopeItem, 0, len(scopes))
	for _, scope := range scopes {
		items = append(items, dto.ScopeItem{
			ID:         scope.ID,
			ScopeType:  string(scope.ScopeType),
			ScopeValue: scope.ScopeValue,
		})
	}

	return &dto.ListScopesResponse{Scopes: items}, nil
}
