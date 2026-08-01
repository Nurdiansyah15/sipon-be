package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/application/query"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AssignRolePermissionUseCase struct {
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewAssignRolePermissionUseCase(
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *AssignRolePermissionUseCase {
	return &AssignRolePermissionUseCase{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *AssignRolePermissionUseCase) Execute(ctx context.Context, roleID, assignedBy string, req dto.AssignRolePermissionRequest) (*dto.RoleItem, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := role.EnsureCustom(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeInvalidRoleType:
				return nil, kernel.New(application.ErrCodeBadRequest)
			}
		}
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	permissionKey := roleconstant.PermissionKey(strings.TrimSpace(req.PermissionKey))
	var notes *string
	if strings.TrimSpace(req.Notes) != "" {
		notes = &req.Notes
	}

	rp, err := roleentity.NewRolePermission(uuid.NewString(), role.ID, permissionKey, strings.TrimSpace(assignedBy), notes)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.rolePermRepo.Save(ctx, rp); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildRoleResponse(ctx, uc.roleRepo, uc.rolePermRepo, role.ID)
}

type DeleteRolePermissionUseCase struct {
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewDeleteRolePermissionUseCase(roleRepo rolerepo.RoleRepository, rolePermRepo rolerepo.RolePermissionRepository) *DeleteRolePermissionUseCase {
	return &DeleteRolePermissionUseCase{roleRepo: roleRepo, rolePermRepo: rolePermRepo}
}

func (uc *DeleteRolePermissionUseCase) Execute(ctx context.Context, roleID, permissionKey string) (*dto.RoleItem, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := role.EnsureCustom(); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.rolePermRepo.Delete(ctx, role.ID, roleconstant.PermissionKey(strings.TrimSpace(permissionKey))); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildRoleResponse(ctx, uc.roleRepo, uc.rolePermRepo, role.ID)
}
