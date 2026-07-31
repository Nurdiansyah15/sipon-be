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

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	otp, err := domain.NewOTPCode(otpCode)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	verifCode, err := domain.NewVerificationCode(uuid.NewString(), user.ID, otp, domain.PurposeResetPassword, time.Now().Add(10*time.Minute))
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	username := user.Username.String()
	if err := uc.emailSender.SendPasswordResetOTP(email.String(), username, otpCode); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	return nil
}
