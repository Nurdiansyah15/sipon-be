package command

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type ChangePasswordLocalUseCase struct {
	userRepo domain.UserRepository
	hasher   application.PasswordHasher
}

func NewChangePasswordLocalUseCase(
	userRepo domain.UserRepository,
	hasher application.PasswordHasher,
) *ChangePasswordLocalUseCase {
	return &ChangePasswordLocalUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *ChangePasswordLocalUseCase) Execute(ctx context.Context, userID string, req dto.ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	credential := user.FindCredential(domain.CredentialTypeLocal)
	if credential == nil || credential.SecretHash == nil {
		return kernel.New(application.ErrCodeInvalidCredentials)
	}

	oldPlainPw, err := domain.NewPlainPassword(req.OldPassword)
	if err != nil {
		return err
	}

	if err := uc.hasher.Verify(credential.SecretHash.String(), oldPlainPw.String()); err != nil {
		return kernel.New(application.ErrCodeInvalidCredentials)
	}

	newPlainPw, err := domain.NewPlainPassword(req.NewPassword)
	if err != nil {
		return err
	}

	hashedPassword, err := uc.hasher.Hash(newPlainPw.String())
	if err != nil {
		return err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		return err
	}

	if err := user.SetLocalPassword(hashedPw); err != nil {
		return err
	}

	return uc.userRepo.Update(ctx, user)
}
