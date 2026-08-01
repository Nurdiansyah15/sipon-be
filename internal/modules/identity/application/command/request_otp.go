package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type RequestIdentityOTPUseCase struct {
	userRepo    domain.UserRepository
	verifRepo   domain.VerificationRepository
	otpGen      application.OTPGenerator
	emailSender application.EmailSender
	smsSender   application.SMSSender
}

func NewRequestIdentityOTPUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
	otpGen application.OTPGenerator,
	emailSender application.EmailSender,
	smsSender application.SMSSender,
) *RequestIdentityOTPUseCase {
	return &RequestIdentityOTPUseCase{
		userRepo:    userRepo,
		verifRepo:   verifRepo,
		otpGen:      otpGen,
		emailSender: emailSender,
		smsSender:   smsSender,
	}
}

func (uc *RequestIdentityOTPUseCase) Execute(ctx context.Context, req dto.RequestOTPRequest) (*dto.RequestOTPResponse, error) {
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
		return nil, kernel.New(application.ErrCodeConflict)
	}

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
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

	verifCode, err := domain.NewVerificationCode(uuid.NewString(), user.ID, otpCode, purpose, 5*time.Minute)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	switch identifier.Kind {
	case domain.LoginIdentifierKindEmail:
		if err := uc.emailSender.SendOTP(identity.Value, user.Username.String(), otpCode); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	case domain.LoginIdentifierKindPhone:
		if err := uc.smsSender.SendOTP(identity.Value, otpCode); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	return &dto.RequestOTPResponse{Message: "OTP verifikasi berhasil dikirim"}, nil
}
