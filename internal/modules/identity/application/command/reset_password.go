package command

import (
	"context"
	"errors"

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

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, req dto.ResetPasswordRequest) error {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	user, err := uc.userRepo.FindByIdentity(ctx, domain.LoginIdentifierKindEmail, email.String())
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, domain.PurposeResetPassword)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	inputOTP, err := domain.NewOTPCode(req.Token)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := verifCode.Verify(inputOTP); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeVerificationCodeMismatch:
				return kernel.New(application.ErrCodeBadRequest)
			case domain.ErrCodeVerificationCodeExpired:
				return kernel.New(application.ErrCodeGone)
			case domain.ErrCodeVerificationCodeUsed:
				return kernel.New(application.ErrCodeConflict)
			}
		}
		return kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.verifRepo.Update(ctx, verifCode); err != nil {
		return err
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := user.SetLocalPassword(hashedPw); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeCredentialNotLocal:
				return kernel.WrapMsg(application.ErrCodeBadRequest, string(ke.Code), ke)
			}
		}
		return err
	}

	return uc.userRepo.Update(ctx, user)
}
