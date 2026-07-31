package command

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type RefreshTokenUseCase struct {
	tokenGen               application.TokenGenerator
	sessionRevocationStore application.SessionRevocationStore
	userRepo               domain.UserRepository
	fileUploader           application.FileUploader
}

func NewRefreshTokenUseCase(
	tokenGen application.TokenGenerator,
	sessionRevocationStore application.SessionRevocationStore,
	userRepo domain.UserRepository,
	fileUploader application.FileUploader,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		tokenGen:               tokenGen,
		sessionRevocationStore: sessionRevocationStore,
		userRepo:               userRepo,
		fileUploader:           fileUploader,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	claims, err := uc.tokenGen.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnauthorized, err)
	}

	revokedBefore, err := uc.sessionRevocationStore.RevokedBefore(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if revokedBefore != nil && claims.IssuedAt.Before(*revokedBefore) {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	deviceRevokedBefore, err := uc.sessionRevocationStore.DeviceRevokedBefore(ctx, claims.UserID, claims.DeviceID)
	if err != nil {
		return nil, err
	}
	if deviceRevokedBefore != nil && claims.IssuedAt.Before(*deviceRevokedBefore) {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	sessionID := uuid.NewString()
	deviceID := uuid.NewString()

	accessToken, err := uc.tokenGen.GenerateAccessToken(claims.UserID, sessionID, deviceID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenGen.GenerateRefreshToken(claims.UserID, deviceID)
	if err != nil {
		return nil, err
	}

	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnauthorized, err)
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
