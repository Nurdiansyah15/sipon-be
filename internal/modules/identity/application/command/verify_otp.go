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

func (uc *VerifyIdentityOTPUseCase) Execute(ctx context.Context, req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error) {
	identifier, err := domain.NewLoginIdentifier(req.Identifier)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
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

	identity := user.FindLoginIdentity(identifier.Kind, identifier.Value)
	if identity == nil {
		return nil, kernel.New(application.ErrCodeNotFound)
	}

	if identity.IsVerified() {
		return &dto.VerifyOTPResponse{Message: "identity sudah terverifikasi"}, nil
	}

	var purpose domain.CodePurpose
	switch identifier.Kind {
	case domain.LoginIdentifierKindEmail:
		purpose = domain.PurposeEmailVerification
	case domain.LoginIdentifierKindPhone:
		purpose = domain.PurposePhoneVerification
	default:
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, purpose)
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

	if err := verifCode.Verify(req.OTP); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeVerificationCodeExpired, domain.ErrCodeVerificationCodeUsed, domain.ErrCodeVerificationCodeMismatch:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	identity.MarkVerified()
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.verifRepo.Update(ctx, verifCode); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.VerifyOTPResponse{Message: "identity berhasil diverifikasi"}, nil
}
