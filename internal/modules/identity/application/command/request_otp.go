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
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	verificationconstant "sipon-be/internal/modules/identity/domain/verification/constant"
	verificationentity "sipon-be/internal/modules/identity/domain/verification/entity"
	verificationrepo "sipon-be/internal/modules/identity/domain/verification/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type RequestIdentityOTPUseCase struct {
	userRepo    userrepo.UserRepository
	verifRepo   verificationrepo.VerificationRepository
	otpGen      ports.OTPGenerator
	emailSender ports.EmailSender
	smsSender   ports.SMSSender
}

func NewRequestIdentityOTPUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	otpGen ports.OTPGenerator,
	emailSender ports.EmailSender,
	smsSender ports.SMSSender,
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
	identifier, err := uservo.NewLoginIdentifier(req.Identifier)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
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

	var purpose verificationconstant.CodePurpose
	switch identifier.Kind {
	case userconstant.LoginIdentifierKindEmail:
		purpose = verificationconstant.PurposeEmailVerification
	case userconstant.LoginIdentifierKindPhone:
		purpose = verificationconstant.PurposePhoneVerification
	default:
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	verifCode, err := verificationentity.NewVerificationCode(uuid.NewString(), user.ID, otpCode, purpose, 5*time.Minute)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	switch identifier.Kind {
	case userconstant.LoginIdentifierKindEmail:
		if err := uc.emailSender.SendOTP(identity.Value, user.Username.String(), otpCode); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	case userconstant.LoginIdentifierKindPhone:
		if err := uc.smsSender.SendOTP(identity.Value, otpCode); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	return &dto.RequestOTPResponse{Message: "OTP verifikasi berhasil dikirim"}, nil
}
