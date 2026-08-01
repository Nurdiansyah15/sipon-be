package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type ResetPasswordUseCase struct {
	userRepo  domain.UserRepository
	verifRepo domain.VerificationRepository
	hasher    application.PasswordHasher
}

func NewResetPasswordUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
	hasher application.PasswordHasher,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepo:  userRepo,
		verifRepo: verifRepo,
		hasher:    hasher,
	}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, req dto.ResetPasswordRequest) (*dto.ResetPasswordResponse, error) {
	user, err := uc.userRepo.FindByIdentity(ctx, domain.LoginIdentifierKindEmail, req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeInvalidLoginIdentityValue:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, domain.PurposeResetPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeVerificationCodeNotFound:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := verifCode.Verify(req.Token); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeVerificationCodeExpired, domain.ErrCodeVerificationCodeUsed, domain.ErrCodeVerificationCodeMismatch:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodePlainPasswordEmpty, domain.ErrCodePlainPasswordTooShort, domain.ErrCodePlainPasswordNoUppercase, domain.ErrCodePlainPasswordNoDigit:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	localCred := user.FindCredential(domain.CredentialTypeLocal)
	if localCred == nil {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	hashedStr, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	hashedPw, err := domain.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	now := time.Now()
	localCred.SecretHash = &hashedPw
	localCred.LastChangedAt = &now
	localCred.UpdatedAt = now
	user.UpdatedAt = now
	user.ResetFailedAttempts()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.verifRepo.Update(ctx, verifCode); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ResetPasswordResponse{Message: "password berhasil direset"}, nil
}
