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
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "ID pengguna tidak boleh kosong", nil)
	}

	if req.CurrentPassword == req.NewPassword {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Kata sandi baru tidak boleh sama dengan kata sandi saat ini", nil)
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

	localCred := user.FindCredential(userconstant.CredentialTypeLocal)
	if !user.HasLocalPassword() {
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "Pengguna tidak memiliki kata sandi lokal", nil)
	}

	if err := uc.hasher.Verify(localCred.SecretHash.String(), req.CurrentPassword); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Kata sandi saat ini salah", err)
	}

	hashedStr, err := uc.hasher.Hash(newPlain.String())
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mengenkripsi kata sandi", err)
	}

	newHashed, err := uservo.NewHashedPassword(hashedStr)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memvalidasi hash kata sandi", err)
	}

	now := time.Now()
	localCred.SecretHash = &newHashed
	localCred.LastChangedAt = &now
	localCred.UpdatedAt = now
	user.UpdatedAt = now

	if err := uc.userRepo.Update(ctx, user); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui kata sandi", err)
	}

	return &dto.ChangePasswordResponse{Message: "password berhasil diubah"}, nil
}
