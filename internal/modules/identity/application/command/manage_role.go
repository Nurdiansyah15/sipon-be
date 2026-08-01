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
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type CreateRoleUseCase struct {
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewCreateRoleUseCase(
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *CreateRoleUseCase {
	return &CreateRoleUseCase{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *CreateRoleUseCase) Execute(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleItem, error) {
	roleName := roleconstant.RoleName(strings.TrimSpace(req.Name))
	roleType := roleconstant.RoleType(req.RoleType)
	scopeType := roleconstant.ScopeType(req.ScopeType)

	var desc *string
	if req.Description != nil && strings.TrimSpace(*req.Description) != "" {
		desc = req.Description
	}

	role, err := roleentity.NewRole(uuid.NewString(), roleName, strings.TrimSpace(req.DisplayName), desc, roleType, scopeType, req.Assignable)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.roleRepo.Save(ctx, role); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeRoleNotFound:
				return nil, kernel.New(application.ErrCodeConflict)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildRoleResponse(ctx, uc.roleRepo, uc.rolePermRepo, role.ID)
}

type UpdateRoleUseCase struct {
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewUpdateRoleUseCase(
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *UpdateRoleUseCase {
	return &UpdateRoleUseCase{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *UpdateRoleUseCase) Execute(ctx context.Context, roleID string, req dto.UpdateRoleRequest) (*dto.RoleItem, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(roleID))
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := role.UpdateDetails(req.DisplayName, req.Description, req.Assignable); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildRoleResponse(ctx, uc.roleRepo, uc.rolePermRepo, role.ID)
}

type AssignUserRoleUseCase struct {
	roleRepo     rolerepo.RoleRepository
	userRoleRepo rolerepo.UserRoleRepository
	userRepo     userrepo.UserRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewAssignUserRoleUseCase(
	roleRepo rolerepo.RoleRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *AssignUserRoleUseCase {
	return &AssignUserRoleUseCase{
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
		userRepo:     userRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *AssignUserRoleUseCase) Execute(ctx context.Context, assignedBy string, req dto.AssignUserRoleRequest) (*dto.UserRoleItem, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(req.RoleID))
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := role.EnsureAssignable(); err != nil {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	scopeType := roleconstant.ScopeType(strings.TrimSpace(req.ScopeType))

	userRole, err := roleentity.NewUserRole(uuid.NewString(), strings.TrimSpace(req.UserID), role.ID, scopeType, req.ScopeID, strings.TrimSpace(assignedBy), req.ExpiredAt, req.Notes)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.userRoleRepo.Save(ctx, userRole); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type UpdateUserRoleUseCase struct {
	userRoleRepo rolerepo.UserRoleRepository
	userRepo     userrepo.UserRepository
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewUpdateUserRoleUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *UpdateUserRoleUseCase {
	return &UpdateUserRoleUseCase{
		userRoleRepo: userRoleRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *UpdateUserRoleUseCase) Execute(ctx context.Context, userRoleID string, req dto.UpdateUserRoleRequest) (*dto.UserRoleItem, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	userRole.UpdateExpiration(req.ExpiredAt)

	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type DeactivateUserRoleUseCase struct {
	userRoleRepo rolerepo.UserRoleRepository
	userRepo     userrepo.UserRepository
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewDeactivateUserRoleUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *DeactivateUserRoleUseCase {
	return &DeactivateUserRoleUseCase{
		userRoleRepo: userRoleRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *DeactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleItem, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := userRole.Deactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeUserRoleNotActive:
				return nil, kernel.New(application.ErrCodeBadRequest)
			}
		}
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type ReactivateUserRoleUseCase struct {
	userRoleRepo rolerepo.UserRoleRepository
	userRepo     userrepo.UserRepository
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewReactivateUserRoleUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *ReactivateUserRoleUseCase {
	return &ReactivateUserRoleUseCase{
		userRoleRepo: userRoleRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *ReactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleItem, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := userRole.Reactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeUserRoleNotActive:
				return nil, kernel.New(application.ErrCodeBadRequest)
			case roleconstant.ErrCodeUserRoleExpired:
				return nil, kernel.New(application.ErrCodeGone)
			}
		}
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type DeleteUserRoleUseCase struct {
	userRoleRepo rolerepo.UserRoleRepository
}

func NewDeleteUserRoleUseCase(userRoleRepo rolerepo.UserRoleRepository) *DeleteUserRoleUseCase {
	return &DeleteUserRoleUseCase{userRoleRepo: userRoleRepo}
}

func (uc *DeleteUserRoleUseCase) Execute(ctx context.Context, userRoleID string) error {
	if err := uc.userRoleRepo.Delete(ctx, strings.TrimSpace(userRoleID)); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	return nil
}
