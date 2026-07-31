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

type RequestChangeIdentityUseCase struct {
	userRepo    domain.UserRepository
	verifRepo   domain.VerificationRepository
	otpGen      application.OTPGenerator
	emailSender application.EmailSender
	smsSender   application.SMSSender
}

func NewRequestChangeIdentityUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
	otpGen application.OTPGenerator,
	emailSender application.EmailSender,
	smsSender application.SMSSender,
) *RequestChangeIdentityUseCase {
	return &RequestChangeIdentityUseCase{
		userRepo:    userRepo,
		verifRepo:   verifRepo,
		otpGen:      otpGen,
		emailSender: emailSender,
		smsSender:   smsSender,
	}
}

func (uc *RequestChangeIdentityUseCase) Execute(ctx context.Context, userID string, req dto.ChangeIdentityRequest) error {
	identifier, err := domain.NewLoginIdentifier(req.NewValue)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	exists, err := uc.userRepo.ExistsByLoginIdentity(ctx, identifier.Kind, identifier.Value)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	if exists {
		return kernel.New(application.ErrCodeConflict)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
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

	var purpose domain.CodePurpose
	switch identifier.Kind {
	case domain.LoginIdentifierKindEmail:
		purpose = domain.PurposeChangeEmail
	case domain.LoginIdentifierKindPhone:
		purpose = domain.PurposeChangePhone
	default:
		purpose = domain.PurposeChangeEmail
	}

	verifCode, err := domain.NewVerificationCode(uuid.NewString(), userID, otp, purpose, time.Now().Add(10*time.Minute))
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := verifCode.SetNewIdentityValue(identifier.Value); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeVerificationNewIdentityEmpty:
				return kernel.New(application.ErrCodeBadRequest)
			}
		}
		return kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	switch identifier.Kind {
	case domain.LoginIdentifierKindEmail:
		if err := uc.emailSender.SendOTP(identifier.Value, user.Username.String(), otpCode); err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
	case domain.LoginIdentifierKindPhone:
		if err := uc.smsSender.SendOTP(identifier.Value, otpCode); err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
	default:
		if err := uc.emailSender.SendOTP(identifier.Value, user.Username.String(), otpCode); err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
	}
	return nil
}

type ConfirmChangeIdentityUseCase struct {
	userRepo   domain.UserRepository
	verifRepo  domain.VerificationRepository
	transactor application.Transactor
}

func NewConfirmChangeIdentityUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
	transactor application.Transactor,
) *ConfirmChangeIdentityUseCase {
	return &ConfirmChangeIdentityUseCase{
		userRepo:   userRepo,
		verifRepo:  verifRepo,
		transactor: transactor,
	}
}

func (uc *ConfirmChangeIdentityUseCase) Execute(ctx context.Context, userID string, purpose domain.CodePurpose, req dto.ChangeIdentityConfirmRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, userID, purpose)
	if err != nil {
		return kernel.New(application.ErrCodeNotFound)
	}

	inputOTP, err := domain.NewOTPCode(req.Code)
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
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	if verifCode.NewIdentityValue == nil || *verifCode.NewIdentityValue == "" {
		return kernel.New(application.ErrCodeBadRequest)
	}

	newValue := *verifCode.NewIdentityValue

	var newIdentityKind domain.LoginIdentifierKind
	switch verifCode.Purpose {
	case domain.PurposeChangeEmail:
		newIdentityKind = domain.LoginIdentifierKindEmail
	case domain.PurposeChangePhone:
		newIdentityKind = domain.LoginIdentifierKindPhone
	default:
		return kernel.New(application.ErrCodeBadRequest)
	}

	identifier, err := domain.NewLoginIdentifier(newValue)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	exists, err := uc.userRepo.ExistsByLoginIdentity(ctx, identifier.Kind, identifier.Value)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	if exists {
		return kernel.New(application.ErrCodeConflict)
	}

	return uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		oldLI := user.FindLoginIdentityByKind(newIdentityKind)
		if oldLI != nil {
			now := time.Now()
			oldLI.DeletedAt = &now
		}

		nowTime := time.Now()
		newLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, "", newIdentityKind, newValue, oldLI != nil && oldLI.IsPrimary, &nowTime)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
			return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}
		newLI.MarkVerified()

		credential := user.FindCredential(domain.CredentialTypeLocal)
		if credential == nil {
			credential = domain.NewLocalCredentialWithoutPassword(uuid.NewString(), userID, true)
			user.AddCredential(credential)
		}
		newLI.CredentialID = credential.ID
		credential.AddLoginIdentity(newLI)

		switch newIdentityKind {
		case domain.LoginIdentifierKindEmail:
			newEmail, err := domain.NewEmail(newValue)
			if err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) {
					return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
				}
				return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
			}
			user.Email = newEmail
		case domain.LoginIdentifierKindPhone:
			newPhone, err := domain.NewPhoneNumber(newValue)
			if err != nil {
				var ke *kernel.AppError
				if errors.As(err, &ke) {
					return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
				}
				return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
			}
			user.PhoneNumber = &newPhone
		}

		user.UpdatedAt = time.Now()
		return uc.userRepo.Update(txCtx, user)
	})
}
