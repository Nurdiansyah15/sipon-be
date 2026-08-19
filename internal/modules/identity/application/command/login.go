package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

const routingLoginSucceeded = "identity.user.login_succeeded"

type LoginUseCase struct {
	userRepo     userrepo.UserRepository
	hasher       ports.PasswordHasher
	tokenGen     ports.TokenGenerator
	outboxWriter ports.OutboxWriter
}

func NewLoginUseCase(
	userRepo userrepo.UserRepository,
	hasher ports.PasswordHasher,
	tokenGen ports.TokenGenerator,
) *LoginUseCase {
	return &LoginUseCase{
		userRepo: userRepo,
		hasher:   hasher,
		tokenGen: tokenGen,
	}
}

// SetOutboxWriter memasang outbox writer untuk publikasi event login.
func (uc *LoginUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *LoginUseCase) Execute(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	identifier, err := uservo.NewLoginIdentifier(req.Identifier)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Identitas login tidak valid", err)
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mencari pengguna", err)
	}

	if err := user.EnsureNotLockedOut(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == userconstant.ErrCodeUserLockedOut {
			return nil, kernel.WrapMsg(application.ErrCodeTooManyRequests, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	identity := user.FindLoginIdentity(identifier.Kind, identifier.Value)
	if identity == nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Identitas login tidak ditemukan", nil)
	}

	cred := user.FindCredential(userconstant.CredentialTypeLocal)
	if cred == nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Kredensial lokal tidak ditemukan", nil)
	}

	if cred.SecretHash == nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Kredensial tidak memiliki kata sandi", nil)
	}

	if err := uc.hasher.Verify(cred.SecretHash.String(), req.Password); err != nil {
		user.IncrementFailedAttempts()
		_ = uc.userRepo.Update(ctx, user)
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "Kata sandi salah", err)
	}

	if err := user.EnsureCanLogin(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserBanned:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, ke.Message, ke)
			case userconstant.ErrCodeUserNotActive:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	emailLI := user.FindLoginIdentity(userconstant.LoginIdentifierKindEmail, user.Email.String())

	user.ResetFailedAttempts()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui data pengguna", err)
	}

	sessionID := uuid.NewString()
	accessToken, err := uc.tokenGen.GenerateAccessToken(user.ID, sessionID, req.DeviceID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat access token", err)
	}
	refreshToken, err := uc.tokenGen.GenerateRefreshToken(user.ID, req.DeviceID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat refresh token", err)
	}

	isEmailVerified := emailLI != nil && emailLI.IsVerified()
	var phoneStr *string
	var isPhoneVerified bool
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
		if li := user.FindLoginIdentity(userconstant.LoginIdentifierKindPhone, s); li != nil {
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
