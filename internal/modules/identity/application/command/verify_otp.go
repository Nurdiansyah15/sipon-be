package command

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type VerifyIdentityOTPUseCase struct {
	userRepo  domain.UserRepository
	verifRepo domain.VerificationRepository
}

func NewVerifyIdentityOTPUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
) *VerifyIdentityOTPUseCase {
	return &VerifyIdentityOTPUseCase{
		userRepo:  userRepo,
		verifRepo: verifRepo,
	}
}

func (uc *VerifyIdentityOTPUseCase) Execute(ctx context.Context, req dto.VerifyOTPRequest) error {
	identifier, err := domain.NewLoginIdentifier(req.Identity)
	if err != nil {
		return err
	}

	user, err := uc.userRepo.FindByLoginIdentifier(ctx, identifier)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
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

	verifCode, err := uc.verifRepo.FindLatestByUserAndPurpose(ctx, user.ID, purpose)
	if err != nil {
		return kernel.Wrap(domain.ErrCodeVerificationCodeNotFound, err)
	}

	inputOTP, err := domain.NewOTPCode(req.Code)
	if err != nil {
		return err
	}

	if err := verifCode.Verify(inputOTP); err != nil {
		return err
	}

	if err := uc.verifRepo.Update(ctx, verifCode); err != nil {
		return err
	}

	li := user.FindLoginIdentity(identifier.Kind, identifier.Value)
	if li == nil {
		li = user.FindLoginIdentityByKind(identifier.Kind)
	}

	if li == nil {
		return kernel.New(domain.ErrCodeIdentityNotVerified)
	}

	li.MarkVerified()

	return uc.userRepo.Update(ctx, user)
}
