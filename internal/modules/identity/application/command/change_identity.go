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
		return err
	}

	exists, err := uc.userRepo.ExistsByLoginIdentity(ctx, identifier.Kind, identifier.Value)
	if err != nil {
		return err
	}
	if exists {
		return kernel.New(application.ErrCodeDuplicateEmail)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
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
		purpose = domain.PurposeChangeEmail
	case domain.LoginIdentifierKindPhone:
		purpose = domain.PurposeChangePhone
	default:
		purpose = domain.PurposeChangeEmail
	}

	verifCode, err := domain.NewVerificationCode(uuid.NewString(), userID, otp, purpose, time.Now().Add(10*time.Minute))
	if err != nil {
		return err
	}

	if err := verifCode.SetNewIdentityValue(identifier.Value); err != nil {
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

type ConfirmChangeIdentityUseCase struct {
	userRepo     domain.UserRepository
	verifRepo    domain.VerificationRepository
	transactor   application.Transactor
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

func (uc *ConfirmChangeIdentityUseCase) Execute(ctx context.Context, userID string, req dto.ChangeIdentityConfirmRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	changeEmailCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, userID, domain.PurposeChangeEmail)
	changePhoneCode, err2 := uc.verifRepo.FindLatestByUserAndPurpose(ctx, userID, domain.PurposeChangePhone)

	var verifCode *domain.VerificationCode
	var newIdentityKind domain.LoginIdentifierKind

	inputOTP, err := domain.NewOTPCode(req.Code)
	if err != nil {
		return err
	}

	if changePhoneCode == nil && err2 != nil {
		verifCode = changeEmailCode
	} else if changeEmailCode == nil && err != nil {
		verifCode = changePhoneCode
	} else if changeEmailCode != nil && changePhoneCode != nil {
		if changeEmailCode.CreatedAt.After(changePhoneCode.CreatedAt) {
			verifCode = changeEmailCode
		} else {
			verifCode = changePhoneCode
		}
	} else if changeEmailCode != nil {
		verifCode = changeEmailCode
	} else {
		verifCode = changePhoneCode
	}

	if verifCode == nil {
		return kernel.New(domain.ErrCodeVerificationCodeNotFound)
	}

	if err := verifCode.Verify(inputOTP); err != nil {
		return err
	}

	if err := uc.verifRepo.Update(ctx, verifCode); err != nil {
		return err
	}

	if verifCode.NewIdentityValue == nil || *verifCode.NewIdentityValue == "" {
		return kernel.New(domain.ErrCodeVerificationNewIdentityEmpty)
	}

	newValue := *verifCode.NewIdentityValue

	switch verifCode.Purpose {
	case domain.PurposeChangeEmail:
		newIdentityKind = domain.LoginIdentifierKindEmail
	case domain.PurposeChangePhone:
		newIdentityKind = domain.LoginIdentifierKindPhone
	default:
		return kernel.New(domain.ErrCodeVerificationInvalidPurpose)
	}

	identifier, err := domain.NewLoginIdentifier(newValue)
	if err != nil {
		return err
	}

	exists, err := uc.userRepo.ExistsByLoginIdentity(ctx, identifier.Kind, identifier.Value)
	if err != nil {
		return err
	}
	if exists {
		return kernel.New(application.ErrCodeDuplicateEmail)
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
			return err
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
				return err
			}
			user.Email = newEmail
		case domain.LoginIdentifierKindPhone:
			newPhone, err := domain.NewPhoneNumber(newValue)
			if err != nil {
				return err
			}
			user.PhoneNumber = &newPhone
		}

		user.UpdatedAt = time.Now()
		return uc.userRepo.Update(txCtx, user)
	})
}
