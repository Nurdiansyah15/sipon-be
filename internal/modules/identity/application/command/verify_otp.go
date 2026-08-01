package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	verificationconstant "sipon-be/internal/modules/identity/domain/verification/constant"
	verificationrepo "sipon-be/internal/modules/identity/domain/verification/repository"
	"sipon-be/internal/shared/kernel"
)

type VerifyIdentityOTPUseCase struct {
	userRepo  userrepo.UserRepository
	verifRepo verificationrepo.VerificationRepository
}

func NewVerifyIdentityOTPUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
) *VerifyIdentityOTPUseCase {
	return &VerifyIdentityOTPUseCase{
		userRepo:  userRepo,
		verifRepo: verifRepo,
	}
}

func (uc *VerifyIdentityOTPUseCase) Execute(ctx context.Context, req dto.VerifyOTPRequest) (*dto.VerifyOTPResponse, error) {
	identifier, err := uservo.NewLoginIdentifier(req.Identifier)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeInvalidLoginIdentityValue:
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

	var purpose verificationconstant.CodePurpose
	switch identifier.Kind {
	case userconstant.LoginIdentifierKindEmail:
		purpose = verificationconstant.PurposeEmailVerification
	case userconstant.LoginIdentifierKindPhone:
		purpose = verificationconstant.PurposePhoneVerification
	default:
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, purpose)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case verificationconstant.ErrCodeVerificationCodeNotFound:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := verifCode.Verify(req.OTP); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case verificationconstant.ErrCodeVerificationCodeExpired, verificationconstant.ErrCodeVerificationCodeUsed, verificationconstant.ErrCodeVerificationCodeMismatch:
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
