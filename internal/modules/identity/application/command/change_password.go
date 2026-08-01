package command

import (
	"context"
	"errors"
	"strings"
	"time"

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

func (uc *ChangePasswordLocalUseCase) Execute(ctx context.Context, userID string, req dto.ChangePasswordRequest) (*dto.ChangePasswordResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, kernel.New(application.ErrCodeUnauthorized)
	}

	if req.CurrentPassword == req.NewPassword {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
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

	localCred := user.FindCredential(domain.CredentialTypeLocal)
	if !user.HasLocalPassword() {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	if err := uc.hasher.Verify(localCred.SecretHash.String(), req.CurrentPassword); err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	hashedStr, err := uc.hasher.Hash(newPlain.String())
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	newHashed, err := domain.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	now := time.Now()
	localCred.SecretHash = &newHashed
	localCred.LastChangedAt = &now
	localCred.UpdatedAt = now
	user.UpdatedAt = now

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ChangePasswordResponse{Message: "password berhasil diubah"}, nil
}
