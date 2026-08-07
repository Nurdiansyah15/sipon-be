package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type SetPasswordLocalUseCase struct {
	userRepo userrepo.UserRepository
	hasher   ports.PasswordHasher
}

func NewSetPasswordLocalUseCase(
	userRepo userrepo.UserRepository,
	hasher ports.PasswordHasher,
) *SetPasswordLocalUseCase {
	return &SetPasswordLocalUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *SetPasswordLocalUseCase) Execute(ctx context.Context, userID string, req dto.SetPasswordRequest) (*dto.SetPasswordResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "ID pengguna tidak boleh kosong", nil)
	}

	newPlain, err := uservo.NewPlainPassword(req.NewPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodePlainPasswordEmpty, userconstant.ErrCodePlainPasswordTooShort, userconstant.ErrCodePlainPasswordNoUppercase, userconstant.ErrCodePlainPasswordNoDigit:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	hashedStr, err := uc.hasher.Hash(newPlain.String())
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mengenkripsi kata sandi", err)
	}

	newHashed, err := uservo.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memvalidasi hash kata sandi", err)
	}

	if err := user.SetLocalPassword(newHashed); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeCredentialNotLocal:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mengatur kata sandi", err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui data pengguna", err)
	}

	return &dto.SetPasswordResponse{Message: "password berhasil ditambahkan"}, nil
}
