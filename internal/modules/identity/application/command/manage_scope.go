package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AssignRoleScopeUseCase struct {
	roleRepo      domain.RoleRepository
	roleScopeRepo domain.RoleScopeRepository
}

func NewAssignRoleScopeUseCase(
	roleRepo domain.RoleRepository,
	roleScopeRepo domain.RoleScopeRepository,
) *AssignRoleScopeUseCase {
	return &AssignRoleScopeUseCase{
		roleRepo:      roleRepo,
		roleScopeRepo: roleScopeRepo,
	}
}

func (uc *AssignRoleScopeUseCase) Execute(ctx context.Context, roleID string, req dto.AssignScopeRequest) (*dto.ScopeItem, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if role.IsSystem() {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	scopeType := domain.ScopeType(strings.TrimSpace(strings.ToLower(req.ScopeType)))
	scopeValue := strings.TrimSpace(strings.ToLower(req.ScopeValue))

	existingScopes, err := uc.roleScopeRepo.FindByRoleID(ctx, roleID)
	if err == nil {
		for _, es := range existingScopes {
			if es.ScopeType == scopeType && es.ScopeValue == scopeValue {
				return nil, kernel.New(application.ErrCodeConflict)
			}
		}
	}

	scope, err := domain.NewRoleScope(uuid.NewString(), roleID, scopeType, scopeValue)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.roleScopeRepo.Save(ctx, scope); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ScopeItem{
		ID:         scope.ID,
		ScopeType:  string(scope.ScopeType),
		ScopeValue: scope.ScopeValue,
	}, nil
}

type DeleteRoleScopeUseCase struct {
	roleScopeRepo domain.RoleScopeRepository
}

func NewDeleteRoleScopeUseCase(roleScopeRepo domain.RoleScopeRepository) *DeleteRoleScopeUseCase {
	return &DeleteRoleScopeUseCase{roleScopeRepo: roleScopeRepo}
}

func (uc *DeleteRoleScopeUseCase) Execute(ctx context.Context, scopeID string) error {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return kernel.New(application.ErrCodeUnprocessableEntity)
	}

	if _, err := uc.roleScopeRepo.FindByID(ctx, scopeID); err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	return uc.roleScopeRepo.Delete(ctx, scopeID)
}
