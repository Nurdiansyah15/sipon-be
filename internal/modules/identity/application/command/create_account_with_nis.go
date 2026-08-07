package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/identity/application"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

// CreateAccountWithNISInput/Result mirror contract.go's CreateAccountInput/
// CreateAccountResult 1:1 — kept as separate types (rather than reusing the
// contract DTOs here) so this use case stays a normal application-layer
// type; contract.go's Module methods are the only place that translates
// between the two.
type CreateAccountWithNISInput struct {
	Username string
	Email    string
	Fullname *string
	NISValue string
}

type CreateAccountWithNISResult struct {
	UserID            string
	GeneratedPassword string
}

// CreateAccountWithNISUseCase backs identity's cross-module Contract — it
// provisions a full login account (User + local Credential + 3
// LoginIdentity: NIS primary, email, username) for callers (e.g.
// kesantrian) that need to onboard a person who has no account yet.
type CreateAccountWithNISUseCase struct {
	userRepo userrepo.UserRepository
	hasher   ports.PasswordHasher
}

func NewCreateAccountWithNISUseCase(userRepo userrepo.UserRepository, hasher ports.PasswordHasher) *CreateAccountWithNISUseCase {
	return &CreateAccountWithNISUseCase{userRepo: userRepo, hasher: hasher}
}

func (uc *CreateAccountWithNISUseCase) Execute(ctx context.Context, in CreateAccountWithNISInput) (*CreateAccountWithNISResult, error) {
	username, err := uservo.NewUsername(in.Username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "data tidak dapat diproses", err)
	}

	email, err := uservo.NewEmail(in.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "data tidak dapat diproses", err)
	}

	generatedPassword, err := generateRandomPassword()
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat kata sandi acak", err)
	}

	plainPw, err := uservo.NewPlainPassword(generatedPassword)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memvalidasi kata sandi yang dibuat", err)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mengenkripsi kata sandi", err)
	}

	hashedPw, err := uservo.NewHashedPassword(hashedPassword)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memvalidasi hash kata sandi", err)
	}

	userID := uuid.NewString()
	user, err := userentity.NewUser(userID, username, in.Fullname, email, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	credentialID := uuid.NewString()
	cred := userentity.NewLocalCredential(credentialID, userID, hashedPw, true)

	nisLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindNIS, in.NISValue, true, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat identitas login NIS", err)
	}
	cred.AddLoginIdentity(nisLI)

	emailLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindEmail, email.String(), false, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat identitas login email", err)
	}
	cred.AddLoginIdentity(emailLI)

	usernameLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindUsername, username.String(), false, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat identitas login username", err)
	}
	cred.AddLoginIdentity(usernameLI)

	user.AddCredential(cred)

	if err := uc.userRepo.Save(ctx, user); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal menyimpan data pengguna", err)
	}

	return &CreateAccountWithNISResult{
		UserID:            userID,
		GeneratedPassword: plainPw.String(),
	}, nil
}
