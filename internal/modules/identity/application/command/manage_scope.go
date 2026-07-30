package command

import (
	"context"

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

func (uc *AssignRoleScopeUseCase) Execute(ctx context.Context, roleID string, req dto.AssignScopeRequest) error {
	scopeType := domain.ScopeType(req.ScopeType)

	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return kernel.Wrap(domain.ErrCodeRoleNotFound, err)
	}

	if err := role.EnsureCustom(); err != nil {
		return err
	}

	roleScope, err := domain.NewRoleScope(uuid.NewString(), roleID, scopeType, req.ScopeValue)
	if err != nil {
		return err
	}

	return uc.roleScopeRepo.Save(ctx, roleScope)
}

type DeleteRoleScopeUseCase struct {
	roleScopeRepo domain.RoleScopeRepository
}

func NewDeleteRoleScopeUseCase(roleScopeRepo domain.RoleScopeRepository) *DeleteRoleScopeUseCase {
	return &DeleteRoleScopeUseCase{roleScopeRepo: roleScopeRepo}
}

func (uc *DeleteRoleScopeUseCase) Execute(ctx context.Context, scopeID string) error {
	return uc.roleScopeRepo.Delete(ctx, scopeID)
}
