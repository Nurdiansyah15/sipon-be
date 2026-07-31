package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
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
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := role.EnsureCustom(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeInvalidRoleType:
				return kernel.New(application.ErrCodeBadRequest)
			}
		}
		return kernel.New(application.ErrCodeForbidden)
	}

	permissionKey := domain.PermissionKey(req.PermissionKey)
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	rp, err := domain.NewRolePermission(uuid.NewString(), roleID, permissionKey, assignedBy, notes)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	return uc.rolePermRepo.Save(ctx, rp)
}

type DeleteRolePermissionUseCase struct {
	roleRepo     domain.RoleRepository
	rolePermRepo domain.RolePermissionRepository
}

func NewDeleteRolePermissionUseCase(roleRepo domain.RoleRepository, rolePermRepo domain.RolePermissionRepository) *DeleteRolePermissionUseCase {
	return &DeleteRolePermissionUseCase{roleRepo: roleRepo, rolePermRepo: rolePermRepo}
}

func (uc *DeleteRolePermissionUseCase) Execute(ctx context.Context, roleID, permissionKey string) error {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := role.EnsureCustom(); err != nil {
		return kernel.New(application.ErrCodeConflict)
	}

	rps, err := uc.rolePermRepo.ListByRoleID(ctx, roleID)
	if err != nil {
		return err
	}

	found := false
	for _, rp := range rps {
		if string(rp.PermissionKey) == permissionKey {
			found = true
			break
		}
	}
	if !found {
		return kernel.New(application.ErrCodeNotFound)
	}

	return uc.rolePermRepo.Delete(ctx, roleID, domain.PermissionKey(permissionKey))
}
