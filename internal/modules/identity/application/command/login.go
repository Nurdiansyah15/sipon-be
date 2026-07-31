package command

import (
	"context"
	"errors"

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
	fileUploader application.FileUploader
}

func NewLoginUseCase(
	userRepo domain.UserRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	hasher application.PasswordHasher,
	tokenGen application.TokenGenerator,
	fileUploader application.FileUploader,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
		hasher:       hasher,
		tokenGen:     tokenGen,
		fileUploader: fileUploader,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	identifier, err := domain.NewLoginIdentifier(req.Identifier)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := user.EnsureCanLogin(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserBanned:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, string(ke.Code), ke)
			case domain.ErrCodeUserLockedOut:
				return nil, kernel.WrapMsg(application.ErrCodeTooManyRequests, string(ke.Code), ke)
			case domain.ErrCodeUserNotActive:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, string(ke.Code), ke)
			case domain.ErrCodeIdentityNotVerified:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, string(ke.Code), ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "unknown domain error", err)
	}

	credential := user.FindCredential(domain.CredentialTypeLocal)
	if credential == nil || credential.SecretHash == nil {
		user.IncrementFailedAttempts()
		_ = uc.userRepo.Update(ctx, user)
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.hasher.Verify(credential.SecretHash.String(), plainPw.String()); err != nil {
		user.IncrementFailedAttempts()
		_ = uc.userRepo.Update(ctx, user)
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	user.ResetFailedAttempts()
	user.MarkLogin()

	if err := uc.userRepo.Update(ctx, user); err != nil {
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

	isEmailVerified := false
	if li := user.FindLoginIdentityByKind(domain.LoginIdentifierKindEmail); li != nil {
		isEmailVerified = li.IsVerified()
	}
	isPhoneVerified := false
	if li := user.FindLoginIdentityByKind(domain.LoginIdentifierKindPhone); li != nil {
		isPhoneVerified = li.IsVerified()
	}

	avatarURL := (*string)(nil)
	if user.AvatarKey != nil && *user.AvatarKey != "" && uc.fileUploader != nil {
		url := uc.fileUploader.PublicURL(*user.AvatarKey)
		avatarURL = &url
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: dto.UserMe{
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
		},
	}, nil
}
