package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/application/ports"
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
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "data tidak dapat diproses", err)
	}

	if err := uc.roleRepo.Save(ctx, role); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeRoleNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
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
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	if err := role.UpdateDetails(req.DisplayName, req.Description, req.Assignable); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return query.BuildRoleResponse(ctx, uc.roleRepo, uc.rolePermRepo, role.ID)
}

type AssignUserRoleUseCase struct {
	roleRepo       rolerepo.RoleRepository
	userRoleRepo   rolerepo.UserRoleRepository
	userRepo       userrepo.UserRepository
	rolePermRepo   rolerepo.RolePermissionRepository
	principalCache ports.PrincipalCacheInvalidator
}

func NewAssignUserRoleUseCase(
	roleRepo rolerepo.RoleRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
	principalCache ports.PrincipalCacheInvalidator,
) *AssignUserRoleUseCase {
	return &AssignUserRoleUseCase{
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		userRepo:       userRepo,
		rolePermRepo:   rolePermRepo,
		principalCache: principalCache,
	}
}

func (uc *AssignUserRoleUseCase) Execute(ctx context.Context, assignedBy string, req dto.AssignUserRoleRequest) (*dto.UserRoleItem, error) {
	role, err := uc.roleRepo.FindByID(ctx, strings.TrimSpace(req.RoleID))
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	if err := role.EnsureAssignable(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeForbidden, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "tidak memiliki akses", err)
	}

	scopeType := roleconstant.ScopeType(strings.TrimSpace(req.ScopeType))

	userRole, err := roleentity.NewUserRole(uuid.NewString(), strings.TrimSpace(req.UserID), role.ID, scopeType, req.ScopeID, strings.TrimSpace(assignedBy), req.ExpiredAt, req.Notes)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "data tidak dapat diproses", err)
	}

	if err := uc.userRoleRepo.Save(ctx, userRole); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	invalidatePrincipalUsers(ctx, uc.principalCache, []string{userRole.UserID})

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type UpdateUserRoleUseCase struct {
	userRoleRepo   rolerepo.UserRoleRepository
	userRepo       userrepo.UserRepository
	roleRepo       rolerepo.RoleRepository
	rolePermRepo   rolerepo.RolePermissionRepository
	principalCache ports.PrincipalCacheInvalidator
}

func NewUpdateUserRoleUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
	principalCache ports.PrincipalCacheInvalidator,
) *UpdateUserRoleUseCase {
	return &UpdateUserRoleUseCase{
		userRoleRepo:   userRoleRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		rolePermRepo:   rolePermRepo,
		principalCache: principalCache,
	}
}

func (uc *UpdateUserRoleUseCase) Execute(ctx context.Context, userRoleID string, req dto.UpdateUserRoleRequest) (*dto.UserRoleItem, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	userRole.UpdateExpiration(req.ExpiredAt)

	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	invalidatePrincipalUsers(ctx, uc.principalCache, []string{userRole.UserID})

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type DeactivateUserRoleUseCase struct {
	userRoleRepo   rolerepo.UserRoleRepository
	userRepo       userrepo.UserRepository
	roleRepo       rolerepo.RoleRepository
	rolePermRepo   rolerepo.RolePermissionRepository
	principalCache ports.PrincipalCacheInvalidator
}

func NewDeactivateUserRoleUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
	principalCache ports.PrincipalCacheInvalidator,
) *DeactivateUserRoleUseCase {
	return &DeactivateUserRoleUseCase{
		userRoleRepo:   userRoleRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		rolePermRepo:   rolePermRepo,
		principalCache: principalCache,
	}
}

func (uc *DeactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleItem, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	if err := userRole.Deactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "permintaan tidak valid", err)
	}

	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	invalidatePrincipalUsers(ctx, uc.principalCache, []string{userRole.UserID})

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type ReactivateUserRoleUseCase struct {
	userRoleRepo   rolerepo.UserRoleRepository
	userRepo       userrepo.UserRepository
	roleRepo       rolerepo.RoleRepository
	rolePermRepo   rolerepo.RolePermissionRepository
	principalCache ports.PrincipalCacheInvalidator
}

func NewReactivateUserRoleUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
	principalCache ports.PrincipalCacheInvalidator,
) *ReactivateUserRoleUseCase {
	return &ReactivateUserRoleUseCase{
		userRoleRepo:   userRoleRepo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		rolePermRepo:   rolePermRepo,
		principalCache: principalCache,
	}
}

func (uc *ReactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleItem, error) {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	if err := userRole.Reactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeUserRoleNotActive:
				return nil, kernel.WrapMsg(application.ErrCodeBadRequest, ke.Message, ke)
			case roleconstant.ErrCodeUserRoleExpired:
				return nil, kernel.WrapMsg(application.ErrCodeGone, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "permintaan tidak valid", err)
	}

	if err := uc.userRoleRepo.Update(ctx, userRole); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	invalidatePrincipalUsers(ctx, uc.principalCache, []string{userRole.UserID})

	return query.BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, userRole)
}

type DeleteUserRoleUseCase struct {
	userRoleRepo   rolerepo.UserRoleRepository
	principalCache ports.PrincipalCacheInvalidator
}

func NewDeleteUserRoleUseCase(userRoleRepo rolerepo.UserRoleRepository, principalCache ports.PrincipalCacheInvalidator) *DeleteUserRoleUseCase {
	return &DeleteUserRoleUseCase{userRoleRepo: userRoleRepo, principalCache: principalCache}
}

func (uc *DeleteUserRoleUseCase) Execute(ctx context.Context, userRoleID string) error {
	userRole, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.userRoleRepo.Delete(ctx, strings.TrimSpace(userRoleID)); err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	invalidatePrincipalUsers(ctx, uc.principalCache, []string{userRole.UserID})
	return nil
}
