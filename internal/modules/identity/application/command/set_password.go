package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type SetPasswordLocalUseCase struct {
	userRepo domain.UserRepository
	hasher   application.PasswordHasher
}

func NewSetPasswordLocalUseCase(
	userRepo domain.UserRepository,
	hasher application.PasswordHasher,
) *SetPasswordLocalUseCase {
	return &SetPasswordLocalUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *SetPasswordLocalUseCase) Execute(ctx context.Context, userID string, req dto.SetPasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	credential := user.FindCredential(domain.CredentialTypeLocal)
	if credential == nil {
		return kernel.New(application.ErrCodeConflict)
	}

	if credential.SecretHash != nil {
		return kernel.New(application.ErrCodeConflict)
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := user.SetLocalPassword(hashedPw); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeCredentialNotLocal:
				return kernel.WrapMsg(application.ErrCodeBadRequest, string(ke.Code), ke)
			}
		}
		return err
	}

	return uc.userRepo.Update(ctx, user)
}
