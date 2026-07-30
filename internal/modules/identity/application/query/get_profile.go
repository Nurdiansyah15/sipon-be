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
	fileUploader  application.FileUploader
}

func NewGetProfileUseCase(
	userRepo domain.UserRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	fileUploader application.FileUploader,
) *GetProfileUseCase {
	return &GetProfileUseCase{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
		fileUploader: fileUploader,
	}
}

func (uc *GetProfileUseCase) Execute(ctx context.Context, userID string) (*dto.ProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	roles, permissions, err := uc.resolveRolesAndPermissions(ctx, userID)
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

	permList := make([]string, 0, len(permissions))
	for p := range permissions {
		permList = append(permList, p)
	}

	return &dto.ProfileResponse{
		UserID:      user.ID,
		Username:    user.Username.String(),
		Fullname:    user.Fullname,
		Email:       user.Email.String(),
		Phone:       phoneStr,
		AvatarURL:   avatarURL,
		Roles:       roles,
		Permissions: permList,
		Status:      string(user.Status),
		CreatedAt:   user.CreatedAt,
	}, nil
}

func (uc *GetProfileUseCase) resolveRolesAndPermissions(ctx context.Context, userID string) ([]string, map[string]struct{}, error) {
	userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	roleNames := make([]string, 0, len(userRoles))
	permSet := make(map[string]struct{})

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
	}

	return roleNames, permSet, nil
}

type MeUseCase struct {
	userRepo      domain.UserRepository
	userRoleRepo  domain.UserRoleRepository
	roleRepo      domain.RoleRepository
	rolePermRepo  domain.RolePermissionRepository
	fileUploader  application.FileUploader
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

func (uc *MeUseCase) Execute(ctx context.Context, userID string) (*dto.ProfileResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUserNotFound, err)
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

	userRoles, _ := uc.userRoleRepo.FindActiveByUserID(ctx, userID)
	roleNames := make([]string, 0, len(userRoles))
	for _, ur := range userRoles {
		if ur.IsUsable() {
			role, _ := uc.roleRepo.FindByID(ctx, ur.RoleID)
			if role != nil {
				roleNames = append(roleNames, string(role.Name))
			}
		}
	}

	return &dto.ProfileResponse{
		UserID:      user.ID,
		Username:    user.Username.String(),
		Fullname:    user.Fullname,
		Email:       user.Email.String(),
		Phone:       phoneStr,
		AvatarURL:   avatarURL,
		Roles:       roleNames,
		Permissions: nil,
		Status:      string(user.Status),
		CreatedAt:   user.CreatedAt,
	}, nil
}
