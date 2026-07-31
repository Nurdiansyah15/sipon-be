package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type VerifyIdentityOTPUseCase struct {
	userRepo  domain.UserRepository
	verifRepo domain.VerificationRepository
}

func NewVerifyIdentityOTPUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
) *VerifyIdentityOTPUseCase {
	return &VerifyIdentityOTPUseCase{
		userRepo:  userRepo,
		verifRepo: verifRepo,
	}
}

func (uc *VerifyIdentityOTPUseCase) Execute(ctx context.Context, req dto.VerifyOTPRequest) error {
	identifier, err := domain.NewLoginIdentifier(req.Identifier)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	var purpose domain.CodePurpose
	switch identifier.Kind {
	case domain.LoginIdentifierKindEmail:
		purpose = domain.PurposeEmailVerification
	case domain.LoginIdentifierKindPhone:
		purpose = domain.PurposePhoneVerification
	default:
		purpose = domain.PurposeEmailVerification
	}

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, purpose)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	inputOTP, err := domain.NewOTPCode(req.OTP)
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

	li := user.FindLoginIdentity(identifier.Kind, identifier.Value)
	if li == nil {
		li = user.FindLoginIdentityByKind(identifier.Kind)
	}

	if li == nil {
		return kernel.New(application.ErrCodeForbidden)
	}

	li.MarkVerified()

	return uc.userRepo.Update(ctx, user)
}
