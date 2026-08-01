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
	userRepo domain.UserRepository
	hasher   application.PasswordHasher
	tokenGen application.TokenGenerator
}

func NewLoginUseCase(
	userRepo domain.UserRepository,
	hasher application.PasswordHasher,
	tokenGen application.TokenGenerator,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		tokenGen: tokenGen,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	identifier, err := domain.NewLoginIdentifier(req.Identifier)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnauthorized, err)
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnauthorized, err)
	}

	if err := user.EnsureNotLockedOut(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == domain.ErrCodeUserLockedOut {
			return nil, kernel.WrapMsg(application.ErrCodeTooManyRequests, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	identity := user.FindLoginIdentity(identifier.Kind, identifier.Value)
	if identity == nil {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	cred := user.FindCredential(domain.CredentialTypeLocal)
	if cred == nil {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	if cred.SecretHash == nil {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	if err := uc.hasher.Verify(cred.SecretHash.String(), req.Password); err != nil {
		user.IncrementFailedAttempts()
		_ = uc.userRepo.Update(ctx, user)
		return nil, kernel.Wrap(application.ErrCodeUnauthorized, err)
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

	emailLI := user.FindLoginIdentity(domain.LoginIdentifierKindEmail, user.Email.String())

	user.ResetFailedAttempts()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	sessionID := uuid.NewString()
	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID, sessionID, req.DeviceID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	refreshToken, err := uc.tokenGen.GenerateRefreshToken(user.ID, req.DeviceID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

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
		},
	}, nil
}
