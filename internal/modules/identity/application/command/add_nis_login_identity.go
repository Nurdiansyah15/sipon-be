package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

// AddNISLoginIdentityUseCase backs identity's cross-module Contract — it
// attaches a NIS login identity to an already existing user's local
// credential (e.g. kesantrian's approve-santri-request flow).
type AddNISLoginIdentityUseCase struct {
	userRepo userrepo.UserRepository
}

func NewAddNISLoginIdentityUseCase(userRepo userrepo.UserRepository) *AddNISLoginIdentityUseCase {
	return &AddNISLoginIdentityUseCase{userRepo: userRepo}
}

func (uc *AddNISLoginIdentityUseCase) Execute(ctx context.Context, userID, nisValue string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return kernel.WrapMsg(application.ErrCodeBadRequest, "ID pengguna tidak boleh kosong", nil)
	}

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

	cred := user.FindCredential(userconstant.CredentialTypeLocal)
	if cred == nil {
		return kernel.WrapMsg(application.ErrCodeConflict, "Kredensial lokal tidak ditemukan", nil)
	}

	nisLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, cred.ID, userconstant.LoginIdentifierKindNIS, nisValue, false, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat identitas login NIS", err)
	}
	cred.AddLoginIdentity(nisLI)

	if err := uc.userRepo.Update(ctx, user); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal menyimpan identitas login NIS", err)
	}

	return nil
}
