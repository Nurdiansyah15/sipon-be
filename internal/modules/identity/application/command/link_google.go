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
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type LinkGoogleUseCase struct {
	userRepo              userrepo.UserRepository
	googleVerifier        ports.GoogleOAuthVerifier
	allowedGoogleClientID []string
}

func NewLinkGoogleUseCase(
	userRepo userrepo.UserRepository,
	googleVerifier ports.GoogleOAuthVerifier,
	allowedGoogleClientIDs []string,
) *LinkGoogleUseCase {
	ids := make([]string, 0, len(allowedGoogleClientIDs))
	for _, id := range allowedGoogleClientIDs {
		v := strings.TrimSpace(id)
		if v == "" {
			continue
		}
		ids = append(ids, v)
	}
	return &LinkGoogleUseCase{
		userRepo:              userRepo,
		googleVerifier:        googleVerifier,
		allowedGoogleClientID: ids,
	}
}

func (uc *LinkGoogleUseCase) Execute(ctx context.Context, userID string, req dto.LinkGoogleRequest) error {
	idToken := req.ResolveIDToken()
	if idToken == "" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID token wajib diisi", nil)
	}
	if len(uc.allowedGoogleClientID) == 0 {
		return kernel.WrapMsg(application.ErrCodeInternal, "allowed google client ids kosong", errors.New("google client ids belum dikonfigurasi"))
	}

	identityInfo, err := uc.googleVerifier.VerifyIDToken(ctx, idToken, uc.allowedGoogleClientID)
	if err != nil {
		return kernel.WrapMsg(application.ErrCodeUnauthorized, "Token Google tidak valid", err)
	}

	googleSub := strings.TrimSpace(identityInfo.Subject)
	if googleSub == "" {
		return kernel.WrapMsg(application.ErrCodeUnauthorized, "Token Google tidak valid", nil)
	}

	existingUser, err := uc.userRepo.FindByIdentity(ctx, userconstant.LoginIdentifierKindGoogle, googleSub)
	if err != nil {
		if !isUserNotFound(err) {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	} else {
		if existingUser.ID != userID {
			return kernel.WrapMsg(application.ErrCodeConflict, "Akun Google sudah terhubung ke pengguna lain", nil)
		}
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

	// Link Google hanya boleh jika email user sudah terverifikasi — mencegah
	// akun yang belum punya bukti kepemilikan email di-link ke identitas Google.
	emailIdentity := user.FindLoginIdentity(userconstant.LoginIdentifierKindEmail, user.Email.String())
	if emailIdentity == nil || !emailIdentity.IsVerified() {
		return kernel.WrapMsg(application.ErrCodeForbidden, "Email belum terverifikasi", nil)
	}

	if err := user.LinkGoogleCredential(uuid.NewString(), googleSub); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeGoogleAlreadyLinked:
				return kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui data pengguna", err)
	}
	return nil
}
