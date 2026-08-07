package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	"sipon-be/internal/shared/kernel"
)

type GetRoleUseCase struct {
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewGetRoleUseCase(
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *GetRoleUseCase {
	return &GetRoleUseCase{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *GetRoleUseCase) Execute(ctx context.Context, roleID string) (*dto.RoleItem, error) {
	roleID = strings.TrimSpace(roleID)
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	permissions, err := resolveRolePermissions(ctx, uc.rolePermRepo, role)
	if err != nil {
		return nil, err
	}

	return &dto.RoleItem{
		ID:          role.ID,
		Name:        string(role.Name),
		DisplayName: role.DisplayName,
		Description: role.Description,
		RoleType:    string(role.RoleType),
		ScopeType:   string(role.ScopeType),
		Assignable:  role.Assignable,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		Permissions: permissions,
	}, nil
}

func resolveRolePermissions(ctx context.Context, rolePermRepo rolerepo.RolePermissionRepository, role *roleentity.Role) ([]string, error) {
	if role.IsSystem() {
		keys := make([]string, 0)
		for _, pk := range roleconstant.PermissionsForRole(role.Name) {
			keys = append(keys, string(pk))
		}
		return keys, nil
	}
	assigned, err := rolePermRepo.ListByRoleID(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(assigned))
	for _, a := range assigned {
		keys = append(keys, string(a.PermissionKey))
	}
	return keys, nil
}

func BuildRoleResponse(ctx context.Context, roleRepo rolerepo.RoleRepository, rolePermRepo rolerepo.RolePermissionRepository, roleID string) (*dto.RoleItem, error) {
	roleID = strings.TrimSpace(roleID)
	role, err := roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}
	permissions, err := resolveRolePermissions(ctx, rolePermRepo, role)
	if err != nil {
		return nil, err
	}
	return &dto.RoleItem{
		ID:          role.ID,
		Name:        string(role.Name),
		DisplayName: role.DisplayName,
		Description: role.Description,
		RoleType:    string(role.RoleType),
		ScopeType:   string(role.ScopeType),
		Assignable:  role.Assignable,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
		Permissions: permissions,
	}, nil
}
