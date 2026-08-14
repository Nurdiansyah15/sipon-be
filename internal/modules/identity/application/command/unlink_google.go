package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"
)

type UnlinkGoogleUseCase struct {
	userRepo userrepo.UserRepository
}

func NewUnlinkGoogleUseCase(userRepo userrepo.UserRepository) *UnlinkGoogleUseCase {
	return &UnlinkGoogleUseCase{userRepo: userRepo}
}

func (uc *UnlinkGoogleUseCase) Execute(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := user.UnlinkGoogleCredential(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeGoogleUnlinkRequiresPassword:
				return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			case userconstant.ErrCodeGoogleNotLinked:
				return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui data pengguna", err)
	}
	return nil
}
