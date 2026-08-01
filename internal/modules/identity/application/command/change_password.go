package command

import (
	"context"
	"errors"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type ChangePasswordLocalUseCase struct {
	userRepo userrepo.UserRepository
	hasher   ports.PasswordHasher
}

func NewChangePasswordLocalUseCase(
	userRepo userrepo.UserRepository,
	hasher ports.PasswordHasher,
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

	newPlain, err := uservo.NewPlainPassword(req.NewPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodePlainPasswordEmpty, userconstant.ErrCodePlainPasswordTooShort, userconstant.ErrCodePlainPasswordNoUppercase, userconstant.ErrCodePlainPasswordNoDigit:
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
			case userconstant.ErrCodeInvalidLoginIdentityValue:
				return nil, kernel.Wrap(application.ErrCodeNotFound, err)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	localCred := user.FindCredential(userconstant.CredentialTypeLocal)
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

	newHashed, err := uservo.NewHashedPassword(hashedStr)
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
