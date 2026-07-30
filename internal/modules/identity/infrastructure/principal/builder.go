package principal

import (
	"context"
	"fmt"

	"sipon-be/internal/modules/identity/domain"
)

type Builder struct {
	userRoleRepo  domain.UserRoleRepository
	roleRepo      domain.RoleRepository
	rolePermRepo  domain.RolePermissionRepository
	roleScopeRepo domain.RoleScopeRepository
}

type Principal struct {
	UserID      string
	Roles       []string
	Permissions []string
	Scopes      []ScopeInfo
}

type ScopeInfo struct {
	ScopeType string
	ScopeID   *string
}

func NewBuilder(
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	roleScopeRepo domain.RoleScopeRepository,
) *Builder {
	return &Builder{
		userRoleRepo:  userRoleRepo,
		roleRepo:      roleRepo,
		rolePermRepo:  rolePermRepo,
		roleScopeRepo: roleScopeRepo,
	}
}

func (b *Builder) Build(ctx context.Context, userID string) (*Principal, error) {
	userRoles, err := b.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("find active user roles: %w", err)
	}

	roleSet := make(map[string]struct{})
	permSet := make(map[string]struct{})
	scopeSet := make(map[string]ScopeInfo)

	for _, ur := range userRoles {
		if !ur.IsUsable() {
			continue
		}

		role, err := b.roleRepo.FindByID(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		roleNameStr := string(role.Name)
		roleSet[roleNameStr] = struct{}{}

		systemPerms := domain.PermissionsForRole(role.Name)
		for _, pk := range systemPerms {
			permSet[string(pk)] = struct{}{}
		}

		customPerms, err := b.rolePermRepo.ListByRoleID(ctx, ur.RoleID)
		if err == nil {
			for _, rp := range customPerms {
				permSet[string(rp.PermissionKey)] = struct{}{}
			}
		}

		scopes, err := b.roleScopeRepo.FindByRoleID(ctx, ur.RoleID)
		if err == nil {
			for _, s := range scopes {
				key := string(s.ScopeType) + ":" + s.ScopeValue
				scopeSet[key] = ScopeInfo{
					ScopeType: string(s.ScopeType),
					ScopeID:   ur.ScopeID,
				}
			}
		}
	}

	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}

	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	scopes := make([]ScopeInfo, 0, len(scopeSet))
	for _, s := range scopeSet {
		scopes = append(scopes, s)
	}

	return &Principal{
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
		Scopes:      scopes,
	}, nil
}
