package command

import (
	"context"
	"errors"
	"strings"

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
}

func NewRefreshTokenUseCase(
	tokenGen application.TokenGenerator,
	sessionRevocationStore application.SessionRevocationStore,
	userRepo domain.UserRepository,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		tokenGen:               tokenGen,
		sessionRevocationStore: sessionRevocationStore,
		userRepo:               userRepo,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req dto.RefreshTokenRequest) (*dto.LoginResponse, error) {
	refreshToken := strings.TrimSpace(req.RefreshToken)
	if refreshToken == "" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	claims, err := uc.tokenGen.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnauthorized, err)
	}

	userID := strings.TrimSpace(claims.UserID)
	if userID == "" {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	if uc.sessionRevocationStore != nil {
		if revokedBefore, revErr := uc.sessionRevocationStore.RevokedBefore(ctx, userID); revErr == nil && revokedBefore != nil {
			if claims.IssuedAt.Before(*revokedBefore) {
				return nil, kernel.New(application.ErrCodeUnauthorized)
			}
		}
		if claims.DeviceID != "" {
			if revokedBefore, revErr := uc.sessionRevocationStore.DeviceRevokedBefore(ctx, userID, claims.DeviceID); revErr == nil && revokedBefore != nil {
				if claims.IssuedAt.Before(*revokedBefore) {
					return nil, kernel.New(application.ErrCodeUnauthorized)
				}
			}
		}
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeInvalidLoginIdentityValue:
				return nil, kernel.Wrap(application.ErrCodeUnauthorized, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := user.EnsureCanLogin(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserBanned:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, string(ke.Code), ke)
			case domain.ErrCodeUserNotActive:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	sessionID := uuid.NewString()
	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID, sessionID, claims.DeviceID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	newRefreshToken, err := uc.tokenGen.GenerateRefreshToken(user.ID, claims.DeviceID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	emailLI := user.FindLoginIdentity(domain.LoginIdentifierKindEmail, user.Email.String())
	isEmailVerified := emailLI != nil && emailLI.IsVerified()
	var phoneStr *string
	var isPhoneVerified bool
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
		if li := user.FindLoginIdentity(domain.LoginIdentifierKindPhone, s); li != nil {
			isPhoneVerified = li.IsVerified()
		}
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
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
		},
	}, nil
}
