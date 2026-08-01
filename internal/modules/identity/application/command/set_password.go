package command

import (
	"context"
	"errors"
	"strings"

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

func (uc *SetPasswordLocalUseCase) Execute(ctx context.Context, userID string, req dto.SetPasswordRequest) (*dto.SetPasswordResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	newPlain, err := domain.NewPlainPassword(req.NewPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodePlainPasswordEmpty, domain.ErrCodePlainPasswordTooShort, domain.ErrCodePlainPasswordNoUppercase, domain.ErrCodePlainPasswordNoDigit:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeInvalidLoginIdentityValue:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	hashedStr, err := uc.hasher.Hash(newPlain.String())
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	newHashed, err := domain.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := user.SetLocalPassword(newHashed); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeCredentialNotLocal:
				return nil, kernel.Wrap(application.ErrCodeInternal, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.SetPasswordResponse{Message: "password berhasil ditambahkan"}, nil
}
