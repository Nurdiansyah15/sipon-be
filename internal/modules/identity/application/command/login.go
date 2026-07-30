package command

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type LoginUseCase struct {
	userRepo     domain.UserRepository
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
	rolePermRepo domain.RolePermissionRepository
	hasher       application.PasswordHasher
	tokenGen     application.TokenGenerator
}

func NewLoginUseCase(
	userRepo domain.UserRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	hasher application.PasswordHasher,
	tokenGen application.TokenGenerator,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
		hasher:       hasher,
		tokenGen:     tokenGen,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	identifier, err := domain.NewLoginIdentifier(req.Identity)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	if err := user.EnsureCanLogin(); err != nil {
		return nil, err
	}

	credential := user.FindCredential(domain.CredentialTypeLocal)
	if credential == nil || credential.SecretHash == nil {
		user.IncrementFailedAttempts()
		_ = uc.userRepo.Update(ctx, user)
		return nil, kernel.New(application.ErrCodeInvalidCredentials)
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		return nil, err
	}

	if err := uc.hasher.Verify(credential.SecretHash.String(), plainPw.String()); err != nil {
		user.IncrementFailedAttempts()
		_ = uc.userRepo.Update(ctx, user)
		return nil, kernel.New(application.ErrCodeInvalidCredentials)
	}

	user.ResetFailedAttempts()
	user.MarkLogin()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	roles, permissions, err := uc.resolveRolesAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	sessionID := uuid.NewString()
	deviceID := req.DeviceID
	if deviceID == "" {
		deviceID = uuid.NewString()
	}

	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID, sessionID, deviceID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenGen.GenerateRefreshToken(user.ID, deviceID)
	if err != nil {
		return nil, err
	}

	phoneStr := (*string)(nil)
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
	}

	permStrs := make([]string, 0, len(permissions))
	for p := range permissions {
		permStrs = append(permStrs, p)
	}

	return &dto.LoginResponse{
		UserID:       user.ID,
		Username:     user.Username.String(),
		Email:        user.Email.String(),
		Phone:        phoneStr,
		Roles:        roles,
		Permissions:  permStrs,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

func (uc *LoginUseCase) resolveRolesAndPermissions(ctx context.Context, userID string) ([]string, map[string]struct{}, error) {
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

		rps, err := uc.rolePermRepo.ListByRoleID(ctx, ur.RoleID)
		if err != nil {
			continue
		}
		for _, rp := range rps {
			permSet[string(rp.PermissionKey)] = struct{}{}
		}
	}

	return roleNames, permSet, nil
}
