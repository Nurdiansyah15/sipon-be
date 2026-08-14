package query

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	"sipon-be/internal/shared/kernel"
)

type UserScopeValue struct {
	ScopeType  string
	ScopeValue string
}

type UserScopeSet struct {
	// HasSystemRole true ketika user memegang role superuser (role_type system
	// yang TIDAK assignable — usergod/superadmin). Role "system" lain (admin,
	// member) tetap mengikuti resolusi scope berbasis role_scopes.
	HasSystemRole bool
	Values        []UserScopeValue
}

type GetUserScopeSetUseCase struct {
	userRoleRepo  rolerepo.UserRoleRepository
	roleRepo      rolerepo.RoleRepository
	roleScopeRepo rolerepo.RoleScopeRepository
}

func NewGetUserScopeSetUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	roleRepo rolerepo.RoleRepository,
	roleScopeRepo rolerepo.RoleScopeRepository,
) *GetUserScopeSetUseCase {
	return &GetUserScopeSetUseCase{
		userRoleRepo:  userRoleRepo,
		roleRepo:      roleRepo,
		roleScopeRepo: roleScopeRepo,
	}
}

func (uc *GetUserScopeSetUseCase) Execute(ctx context.Context, userID string) (*UserScopeSet, error) {
	userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mencari role pengguna", err)
	}

	result := &UserScopeSet{}
	seen := make(map[string]struct{})

	for _, ur := range userRoles {
		if !ur.IsUsable() {
			continue
		}

		role, err := uc.roleRepo.FindByID(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		if role.RoleType == roleconstant.RoleTypeSystem && !role.Assignable {
			result.HasSystemRole = true
		}

		scopes, err := uc.roleScopeRepo.FindByRoleID(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		for _, s := range scopes {
			key := string(s.ScopeType) + ":" + s.ScopeValue
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result.Values = append(result.Values, UserScopeValue{
				ScopeType:  string(s.ScopeType),
				ScopeValue: s.ScopeValue,
			})
		}
	}

	return result, nil
}
