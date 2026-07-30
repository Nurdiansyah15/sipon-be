package command

import (
	"context"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AssignRolePermissionUseCase struct {
	roleRepo     domain.RoleRepository
	rolePermRepo domain.RolePermissionRepository
}

func NewAssignRolePermissionUseCase(
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
) *AssignRolePermissionUseCase {
	return &AssignRolePermissionUseCase{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *AssignRolePermissionUseCase) Execute(ctx context.Context, roleID, assignedBy string, req dto.AssignRolePermissionRequest) error {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return kernel.Wrap(domain.ErrCodeRoleNotFound, err)
	}

	if err := role.EnsureCustom(); err != nil {
		return err
	}

	permissionKey := domain.PermissionKey(req.PermissionKey)
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	rp, err := domain.NewRolePermission(uuid.NewString(), roleID, permissionKey, assignedBy, notes)
	if err != nil {
		return err
	}

	return uc.rolePermRepo.Save(ctx, rp)
}

type DeleteRolePermissionUseCase struct {
	rolePermRepo domain.RolePermissionRepository
}

func NewDeleteRolePermissionUseCase(rolePermRepo domain.RolePermissionRepository) *DeleteRolePermissionUseCase {
	return &DeleteRolePermissionUseCase{rolePermRepo: rolePermRepo}
}

func (uc *DeleteRolePermissionUseCase) Execute(ctx context.Context, roleID, permissionKey string) error {
	return uc.rolePermRepo.Delete(ctx, roleID, domain.PermissionKey(permissionKey))
}
