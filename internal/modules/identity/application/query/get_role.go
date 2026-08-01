package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type GetRoleUseCase struct {
	roleRepo     domain.RoleRepository
	rolePermRepo domain.RolePermissionRepository
}

func NewGetRoleUseCase(
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
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
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
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

func resolveRolePermissions(ctx context.Context, rolePermRepo domain.RolePermissionRepository, role *domain.Role) ([]string, error) {
	if role.IsSystem() {
		keys := make([]string, 0)
		for _, pk := range domain.PermissionsForRole(role.Name) {
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

func BuildRoleResponse(ctx context.Context, roleRepo domain.RoleRepository, rolePermRepo domain.RolePermissionRepository, roleID string) (*dto.RoleItem, error) {
	roleID = strings.TrimSpace(roleID)
	role, err := roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
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
