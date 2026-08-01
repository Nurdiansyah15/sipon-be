package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type ListScopesUseCase struct {
	roleScopeRepo domain.RoleScopeRepository
}

func NewListScopesUseCase(roleScopeRepo domain.RoleScopeRepository) *ListScopesUseCase {
	return &ListScopesUseCase{roleScopeRepo: roleScopeRepo}
}

func (uc *ListScopesUseCase) Execute(ctx context.Context, roleID string) ([]dto.ScopeItem, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	scopes, err := uc.roleScopeRepo.FindByRoleID(ctx, roleID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ScopeItem, 0, len(scopes))
	for _, scope := range scopes {
		items = append(items, dto.ScopeItem{
			ID:         scope.ID,
			ScopeType:  string(scope.ScopeType),
			ScopeValue: scope.ScopeValue,
		})
	}

	return items, nil
}
