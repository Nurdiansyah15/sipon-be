package query

import (
	"context"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
)

type GetSessionUseCase struct {
	userRepo      domain.UserRepository
	userRoleRepo  domain.UserRoleRepository
	roleRepo      domain.RoleRepository
	rolePermRepo  domain.RolePermissionRepository
	roleScopeRepo domain.RoleScopeRepository
}

func NewGetSessionUseCase(
	userRepo domain.UserRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	roleScopeRepo domain.RoleScopeRepository,
) *GetSessionUseCase {
	return &GetSessionUseCase{
		userRepo:      userRepo,
		userRoleRepo:  userRoleRepo,
		roleRepo:      roleRepo,
		rolePermRepo:  rolePermRepo,
		roleScopeRepo: roleScopeRepo,
	}
}

func (uc *GetSessionUseCase) Execute(ctx context.Context, userID string) (*dto.SessionResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]dto.SessionRole, 0, len(userRoles))
	permKeySet := make(map[string]struct{})
	permissions := make([]dto.SessionPermission, 0)
	scopes := make([]dto.SessionUserScope, 0)

	for _, ur := range userRoles {
		if !ur.IsUsable() {
			continue
		}

		role, err := uc.roleRepo.FindByID(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		roles = append(roles, dto.SessionRole{
			Name:      string(role.Name),
			RoleType:  string(role.RoleType),
			ScopeType: string(ur.ScopeType),
			ScopeID:   ur.ScopeID,
		})

		permKeys := make(map[string]struct{})
		for _, pk := range domain.PermissionsForRole(role.Name) {
			permKeys[string(pk)] = struct{}{}
		}

		rps, _ := uc.rolePermRepo.ListByRoleID(ctx, ur.RoleID)
		for _, rp := range rps {
			permKeys[string(rp.PermissionKey)] = struct{}{}
		}

		for key := range permKeys {
			dedupeKey := key + "|" + string(ur.ScopeType)
			if _, seen := permKeySet[dedupeKey]; seen {
				continue
			}
			permKeySet[dedupeKey] = struct{}{}
			permissions = append(permissions, dto.SessionPermission{
				Key:   key,
				Scope: string(ur.ScopeType),
			})
		}

		rs, _ := uc.roleScopeRepo.FindByRoleID(ctx, ur.RoleID)
		for _, scope := range rs {
			scopes = append(scopes, dto.SessionUserScope{
				ScopeType:  string(scope.ScopeType),
				ScopeValue: scope.ScopeValue,
			})
		}
	}

	name := ""
	if user.Fullname != nil {
		name = *user.Fullname
	}

	return &dto.SessionResponse{
		User: dto.SessionUser{
			ID:       user.ID,
			Name:     name,
			Email:    user.Email.String(),
			Username: user.Username.String(),
		},
		Roles:       roles,
		Permissions: permissions,
		Scopes:      scopes,
	}, nil
}
