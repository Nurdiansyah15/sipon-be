package query

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
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
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, 0, len(userRoles))
	permSet := make(map[string]struct{})
	scopes := make([]dto.ScopeResponse, 0)

	for _, ur := range userRoles {
		if !ur.IsUsable() {
			continue
		}

		role, err := uc.roleRepo.FindByID(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		roleNames = append(roleNames, string(role.Name))

		for _, pk := range domain.PermissionsForRole(role.Name) {
			permSet[string(pk)] = struct{}{}
		}

		rps, _ := uc.rolePermRepo.ListByRoleID(ctx, ur.RoleID)
		for _, rp := range rps {
			permSet[string(rp.PermissionKey)] = struct{}{}
		}

		rs, _ := uc.roleScopeRepo.FindByRoleID(ctx, ur.RoleID)
		for _, scope := range rs {
			scopes = append(scopes, dto.ScopeResponse{
				ScopeType: string(scope.ScopeType),
				ScopeID:   &scope.ScopeValue,
			})
		}
	}

	permList := make([]string, 0, len(permSet))
	for p := range permSet {
		permList = append(permList, p)
	}

	phoneStr := (*string)(nil)
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
	}

	return &dto.SessionResponse{
		UserID:      user.ID,
		Username:    user.Username.String(),
		Fullname:    user.Fullname,
		Email:       user.Email.String(),
		Phone:       phoneStr,
		Roles:       roleNames,
		Permissions: permList,
		Scopes:      scopes,
	}, nil
}
