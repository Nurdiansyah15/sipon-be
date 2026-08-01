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

func (uc *RequestChangeIdentityUseCase) Execute(ctx context.Context, userID string, kind domain.LoginIdentifierKind, newValue string) (*dto.ChangeIdentityResponse, error) {
	var normalizedValue string
	switch kind {
	case domain.LoginIdentifierKindEmail:
		email, err := domain.NewEmail(newValue)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}
		normalizedValue = email.String()
	case domain.LoginIdentifierKindPhone:
		phone, err := domain.NewPhoneNumber(newValue)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}
		normalizedValue = phone.String()
	default:
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	existingUser, findErr := uc.userRepo.FindByIdentity(ctx, kind, normalizedValue)
	if findErr == nil {
		if existingUser.ID != userID {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		identity := existingUser.FindLoginIdentityByKind(kind)
		if identity != nil && identity.IsVerified() {
			return nil, kernel.New(application.ErrCodeConflict)
		}
	} else {
		var ke *kernel.AppError
		if !errors.As(findErr, &ke) || ke.Code != domain.ErrCodeInvalidLoginIdentityValue {
			return nil, kernel.Wrap(application.ErrCodeInternal, findErr)
		}
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == domain.ErrCodeInvalidLoginIdentityValue {
			return nil, kernel.Wrap(application.ErrCodeNotFound, err)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	var purpose domain.CodePurpose
	switch kind {
	case domain.LoginIdentifierKindEmail:
		purpose = domain.PurposeChangeEmail
	case domain.LoginIdentifierKindPhone:
		purpose = domain.PurposeChangePhone
	default:
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	verifCode, err := domain.NewVerificationCode(uuid.NewString(), user.ID, otpCode, purpose, 5*time.Minute)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	verifCode.SetNewIdentityValue(normalizedValue)

	if err := uc.verifRepo.Save(ctx, verifCode); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	label := "email"
	switch kind {
	case domain.LoginIdentifierKindEmail:
		if err := uc.emailSender.SendOTP(normalizedValue, user.Username.String(), otpCode); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	case domain.LoginIdentifierKindPhone:
		label = "phone"
		if err := uc.smsSender.SendOTP(normalizedValue, otpCode); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	return &dto.ChangeIdentityResponse{Message: "OTP berhasil dikirim ke " + label + " baru"}, nil
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

func (uc *ConfirmChangeIdentityUseCase) Execute(ctx context.Context, userID string, kind domain.LoginIdentifierKind, otp string) (*dto.ChangeIdentityResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == domain.ErrCodeInvalidLoginIdentityValue {
			return nil, kernel.Wrap(application.ErrCodeNotFound, err)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	var purpose domain.CodePurpose
	switch kind {
	case domain.LoginIdentifierKindEmail:
		purpose = domain.PurposeChangeEmail
	case domain.LoginIdentifierKindPhone:
		purpose = domain.PurposeChangePhone
	default:
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	code, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, purpose)
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

	if err := code.Verify(otp); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeVerificationCodeExpired, domain.ErrCodeVerificationCodeUsed, domain.ErrCodeVerificationCodeMismatch:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if code.NewIdentityValue == nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, nil)
	}
	newValue := *code.NewIdentityValue

	identity := user.FindLoginIdentityByKind(kind)
	if identity != nil {
		identity.Value = newValue
		identity.MarkVerified()
	} else {
		cred := user.FindCredential(domain.CredentialTypeLocal)
		if cred == nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, nil)
		}
		now := time.Now()
		newIdentity, err := domain.NewLoginIdentity(
			uuid.NewString(), user.ID, cred.ID, kind, newValue, true, &now,
		)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		cred.AddLoginIdentity(newIdentity)
	}

	switch kind {
	case domain.LoginIdentifierKindEmail:
		email, err := domain.NewEmail(newValue)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		user.Email = email
	case domain.LoginIdentifierKindPhone:
		phone, err := domain.NewPhoneNumber(newValue)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		user.PhoneNumber = &phone
	}
	user.UpdatedAt = time.Now()

	label := "email"
	if kind == domain.LoginIdentifierKindPhone {
		label = "phone"
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Update(txCtx, user); err != nil {
			return err
		}
		return uc.verifRepo.Update(txCtx, code)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ChangeIdentityResponse{Message: label + " berhasil diperbarui"}, nil
}
