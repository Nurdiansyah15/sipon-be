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

type ForgotPasswordUseCase struct {
	userRepo    domain.UserRepository
	verifRepo   domain.VerificationRepository
	otpGen      application.OTPGenerator
	emailSender application.EmailSender
}

func NewForgotPasswordUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
	otpGen application.OTPGenerator,
	emailSender application.EmailSender,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:    userRepo,
		verifRepo:   verifRepo,
		otpGen:      otpGen,
		emailSender: emailSender,
	}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, req dto.ForgotPasswordRequest) error {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return err
	}

	user, err := uc.userRepo.FindByIdentity(ctx, domain.LoginIdentifierKindEmail, email.String())
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

	verifCode, err := domain.NewVerificationCode(uuid.NewString(), user.ID, otp, domain.PurposeResetPassword, time.Now().Add(10*time.Minute))
	if err != nil {
		return err
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return err
	}

	username := user.Username.String()
	return uc.emailSender.SendPasswordResetOTP(email.String(), username, otpCode)
}
