package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	verificationconstant "sipon-be/internal/modules/identity/domain/verification/constant"
	verificationentity "sipon-be/internal/modules/identity/domain/verification/entity"
	verificationrepo "sipon-be/internal/modules/identity/domain/verification/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type ForgotPasswordUseCase struct {
	userRepo    userrepo.UserRepository
	verifRepo   verificationrepo.VerificationRepository
	otpGen      ports.OTPGenerator
	emailSender ports.EmailSender
}

func NewForgotPasswordUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	otpGen ports.OTPGenerator,
	emailSender ports.EmailSender,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		userRepo:    userRepo,
		verifRepo:   verifRepo,
		otpGen:      otpGen,
		emailSender: emailSender,
	}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, req dto.ForgotPasswordRequest) (*dto.ForgotPasswordResponse, error) {
	successMsg := &dto.ForgotPasswordResponse{Message: "jika email terdaftar, OTP reset password telah dikirim"}

	user, err := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierKindEmail, req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeInvalidLoginIdentityValue:
				return successMsg, nil
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	localCred := user.FindCredential(userconstant.CredentialTypeLocal)
	if localCred == nil {
		return successMsg, nil
	}

	emailIdentity := localCred.FindLoginIdentity(userconstant.LoginIdentifierKindEmail, user.Email.String())
	if emailIdentity == nil {
		return successMsg, nil
	}

	if err := emailIdentity.EnsureVerified(); err != nil {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	verifCode, err := verificationentity.NewVerificationCode(uuid.NewString(), user.ID, otpCode, verificationconstant.PurposeResetPassword, 15*time.Minute)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.emailSender.SendPasswordResetOTP(user.Email.String(), user.Username.String(), otpCode); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return successMsg, nil
}
