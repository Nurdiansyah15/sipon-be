package command

import (
	"context"
	"errors"

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
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	credential := user.FindCredential(domain.CredentialTypeLocal)
	if credential == nil || credential.SecretHash == nil {
		return kernel.New(application.ErrCodeUnauthorized)
	}

	oldPlainPw, err := domain.NewPlainPassword(req.CurrentPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := uc.hasher.Verify(credential.SecretHash.String(), oldPlainPw.String()); err != nil {
		return kernel.New(application.ErrCodeUnauthorized)
	}

	newPlainPw, err := domain.NewPlainPassword(req.NewPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	hashedPassword, err := uc.hasher.Hash(newPlainPw.String())
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
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
		return kernel.Wrap(application.ErrCodeBadRequest, err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	return nil
}
