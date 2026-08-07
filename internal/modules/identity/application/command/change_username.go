package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"
)

type ChangeUsernameUseCase struct {
	userRepo userrepo.UserRepository
}

func NewChangeUsernameUseCase(userRepo userrepo.UserRepository) *ChangeUsernameUseCase {
	return &ChangeUsernameUseCase{userRepo: userRepo}
}

func (uc *ChangeUsernameUseCase) Execute(ctx context.Context, userID string, req dto.ChangeUsernameRequest) (*dto.ChangeUsernameResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnauthorized, "ID pengguna tidak boleh kosong", nil)
	}

	newUsernameStr := strings.TrimSpace(req.Username)
	newUsername, err := uservo.NewUsername(newUsernameStr)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "data tidak dapat diproses", err)
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

	if user.Username.String() == newUsername.String() {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Username baru sama dengan username lama", nil)
	}

	existingUser, findErr := uc.userRepo.FindByUsername(ctx, newUsername.String())
	if findErr == nil && existingUser.ID != userID {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Username sudah digunakan", nil)
	}
	if findErr != nil {
		var ke *kernel.AppError
		if !errors.As(findErr, &ke) || ke.Code != userconstant.ErrCodeUserNotFound {
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memeriksa ketersediaan username", findErr)
		}
	}

	user.ChangeUsername(newUsername)
	if err := uc.userRepo.UpdateUsername(ctx, user.ID, newUsername.String()); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui username", err)
	}

	return &dto.ChangeUsernameResponse{
		Message:  "username berhasil diubah",
		Username: newUsername.String(),
	}, nil
}
