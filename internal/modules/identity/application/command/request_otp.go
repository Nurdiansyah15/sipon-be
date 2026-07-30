package command

import (
	"context"
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

func (uc *RequestIdentityOTPUseCase) Execute(ctx context.Context, req dto.RequestOTPRequest) error {
	identifier, err := domain.NewLoginIdentifier(req.Identity)
	if err != nil {
		return err
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return err
	}

	otp, err := domain.NewOTPCode(otpCode)
	if err != nil {
		return err
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

	verifCode, err := domain.NewVerificationCode(uuid.NewString(), user.ID, otp, purpose, time.Now().Add(10*time.Minute))
	if err != nil {
		return err
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return err
	}

	switch identifier.Kind {
	case domain.LoginIdentifierKindEmail:
		return uc.emailSender.SendOTP(identifier.Value, user.Username.String(), otpCode)
	case domain.LoginIdentifierKindPhone:
		return uc.smsSender.SendOTP(identifier.Value, otpCode)
	default:
		return uc.emailSender.SendOTP(identifier.Value, user.Username.String(), otpCode)
	}
}
