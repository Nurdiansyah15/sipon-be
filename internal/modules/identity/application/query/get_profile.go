package query

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type GetProfileUseCase struct {
	userRepo      domain.UserRepository
	userRoleRepo  domain.UserRoleRepository
	roleRepo      domain.RoleRepository
	rolePermRepo  domain.RolePermissionRepository
	roleScopeRepo domain.RoleScopeRepository
	fileUploader  application.FileUploader
}

func NewGetProfileUseCase(
	userRepo domain.UserRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	roleScopeRepo domain.RoleScopeRepository,
	fileUploader application.FileUploader,
) *GetProfileUseCase {
	return &GetProfileUseCase{
		userRepo:      userRepo,
		userRoleRepo:  userRoleRepo,
		roleRepo:      roleRepo,
		rolePermRepo:  rolePermRepo,
		roleScopeRepo: roleScopeRepo,
		fileUploader:  fileUploader,
	}
}

func (uc *GetProfileUseCase) Execute(ctx context.Context, userID string) (*dto.ProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	roles, permissions, scopes, err := resolveSessionRolesPermsScopes(ctx, uc.userRoleRepo, uc.roleRepo, uc.rolePermRepo, uc.roleScopeRepo, userID)
	if err != nil {
		return nil, err
	}

	avatarURL := (*string)(nil)
	if user.AvatarKey != nil && *user.AvatarKey != "" {
		url := uc.fileUploader.PublicURL(*user.AvatarKey)
		avatarURL = &url
	}

	phoneStr := (*string)(nil)
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
	}

	isEmailVerified := false
	if li := user.FindLoginIdentityByKind(domain.LoginIdentifierKindEmail); li != nil {
		isEmailVerified = li.IsVerified()
	}
	isPhoneVerified := false
	if li := user.FindLoginIdentityByKind(domain.LoginIdentifierKindPhone); li != nil {
		isPhoneVerified = li.IsVerified()
	}

	return &dto.ProfileResponse{
		ID:              user.ID,
		Username:        user.Username.String(),
		Fullname:        user.Fullname,
		Email:           user.Email.String(),
		IsEmailVerified: isEmailVerified,
		Phone:           phoneStr,
		IsPhoneVerified: isPhoneVerified,
		Status:          string(user.Status),
		HasPassword:     user.HasLocalPassword(),
		CreatedAt:       user.CreatedAt,
		AvatarURL:       avatarURL,
		Roles:           roles,
		Permissions:     permissions,
		Scopes:          scopes,
	}, nil
}

// resolveSessionRolesPermsScopes is shared by GetSession and GetProfile — both
// resolve the same rich role/permission/scope objects for the current user.
func resolveSessionRolesPermsScopes(
	ctx context.Context,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	roleScopeRepo domain.RoleScopeRepository,
	userID string,
) ([]dto.SessionRole, []dto.SessionPermission, []dto.SessionUserScope, error) {
	userRoles, err := userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	roles := make([]dto.SessionRole, 0, len(userRoles))
	permSeen := make(map[string]struct{})
	permissions := make([]dto.SessionPermission, 0)
	scopes := make([]dto.SessionUserScope, 0)

	for _, ur := range userRoles {
		if !ur.IsUsable() {
			continue
		}

		role, err := roleRepo.FindByID(ctx, ur.RoleID)
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

		rps, _ := rolePermRepo.ListByRoleID(ctx, ur.RoleID)
		for _, rp := range rps {
			permKeys[string(rp.PermissionKey)] = struct{}{}
		}

		for key := range permKeys {
			dedupeKey := key + "|" + string(ur.ScopeType)
			if _, seen := permSeen[dedupeKey]; seen {
				continue
			}
			permSeen[dedupeKey] = struct{}{}
			permissions = append(permissions, dto.SessionPermission{Key: key, Scope: string(ur.ScopeType)})
		}

		rs, _ := roleScopeRepo.FindByRoleID(ctx, ur.RoleID)
		for _, scope := range rs {
			scopes = append(scopes, dto.SessionUserScope{
				ScopeType:  string(scope.ScopeType),
				ScopeValue: scope.ScopeValue,
			})
		}
	}

	return roles, permissions, scopes, nil
}

type MeUseCase struct {
	userRepo     domain.UserRepository
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
	rolePermRepo domain.RolePermissionRepository
	fileUploader application.FileUploader
}

func NewMeUseCase(
	userRepo domain.UserRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	fileUploader application.FileUploader,
) *MeUseCase {
	return &MeUseCase{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
		fileUploader: fileUploader,
	}
}

func (uc *MeUseCase) Execute(ctx context.Context, userID string) (*dto.UserMe, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	avatarURL := (*string)(nil)
	if user.AvatarKey != nil && *user.AvatarKey != "" {
		url := uc.fileUploader.PublicURL(*user.AvatarKey)
		avatarURL = &url
	}

	phoneStr := (*string)(nil)
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
	}

	isEmailVerified := false
	if li := user.FindLoginIdentityByKind(domain.LoginIdentifierKindEmail); li != nil {
		isEmailVerified = li.IsVerified()
	}
	isPhoneVerified := false
	if li := user.FindLoginIdentityByKind(domain.LoginIdentifierKindPhone); li != nil {
		isPhoneVerified = li.IsVerified()
	}

	return &dto.UserMe{
		ID:              user.ID,
		Username:        user.Username.String(),
		Email:           user.Email.String(),
		IsEmailVerified: isEmailVerified,
		Fullname:        user.Fullname,
		Phone:           phoneStr,
		IsPhoneVerified: isPhoneVerified,
		Status:          string(user.Status),
		CreatedAt:       user.CreatedAt,
		HasPassword:     user.HasLocalPassword(),
		AvatarURL:       avatarURL,
	}, nil
}
