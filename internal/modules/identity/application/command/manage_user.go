package command

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

const (
	generatedPasswordLength = 12
	generatedPasswordUpper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	generatedPasswordLower  = "abcdefghijklmnopqrstuvwxyz"
	generatedPasswordDigits = "0123456789"
)

func generateRandomPassword() (string, error) {
	alphabet := generatedPasswordUpper + generatedPasswordLower + generatedPasswordDigits
	buf := make([]byte, generatedPasswordLength)

	pick := func(set string) (byte, error) {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(set))))
		if err != nil {
			return 0, err
		}
		return set[idx.Int64()], nil
	}

	upper, err := pick(generatedPasswordUpper)
	if err != nil {
		return "", err
	}
	digit, err := pick(generatedPasswordDigits)
	if err != nil {
		return "", err
	}
	buf[0] = upper
	buf[1] = digit

	for i := 2; i < generatedPasswordLength; i++ {
		c, err := pick(alphabet)
		if err != nil {
			return "", err
		}
		buf[i] = c
	}

	return string(buf), nil
}

type CreateUserUseCase struct {
	userRepo userrepo.UserRepository
	hasher   ports.PasswordHasher
}

func NewCreateUserUseCase(
	userRepo userrepo.UserRepository,
	hasher ports.PasswordHasher,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	username, err := uservo.NewUsername(req.Username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	email, err := uservo.NewEmail(req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	var phone *uservo.PhoneNumber
	if req.Phone != nil && strings.TrimSpace(*req.Phone) != "" {
		pn, err := uservo.NewPhoneNumber(*req.Phone)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}
		phone = &pn
	}

	generatedPassword, err := generateRandomPassword()
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	plainPw, err := uservo.NewPlainPassword(generatedPassword)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	hashedPw, err := uservo.NewHashedPassword(hashedPassword)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	userID := uuid.NewString()
	user, err := userentity.NewUser(userID, username, req.Fullname, email, phone)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	credentialID := uuid.NewString()
	cred := userentity.NewLocalCredential(credentialID, userID, hashedPw, true)

	emailLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindEmail, email.String(), true, nil)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	cred.AddLoginIdentity(emailLI)

	usernameLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindUsername, username.String(), true, nil)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	cred.AddLoginIdentity(usernameLI)

	if phone != nil {
		phoneLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindPhone, phone.String(), false, nil)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		cred.AddLoginIdentity(phoneLI)
	}

	user.AddCredential(cred)

	if err := uc.userRepo.Save(ctx, user); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	phoneStr := (*string)(nil)
	if phone != nil {
		s := phone.String()
		phoneStr = &s
	}

	return &dto.CreateUserResponse{
		UserManagementResponse: dto.UserManagementResponse{
			ID:        userID,
			Username:  username.String(),
			Fullname:  req.Fullname,
			Email:     email.String(),
			Phone:     phoneStr,
			Status:    string(userconstant.UserStatusActive),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		GeneratedPassword: plainPw.String(),
	}, nil
}

type ResetUserPasswordUseCase struct {
	userRepo userrepo.UserRepository
	hasher   ports.PasswordHasher
}

func NewResetUserPasswordUseCase(
	userRepo userrepo.UserRepository,
	hasher ports.PasswordHasher,
) *ResetUserPasswordUseCase {
	return &ResetUserPasswordUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *ResetUserPasswordUseCase) Execute(ctx context.Context, userID string) (*dto.ResetUserPasswordResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	generatedPassword, err := generateRandomPassword()
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	plainPw, err := uservo.NewPlainPassword(generatedPassword)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	newHashed, err := uservo.NewHashedPassword(hashedPassword)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	local := user.FindCredential(userconstant.CredentialTypeLocal)
	if local == nil || local.DeletedAt != nil {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	now := time.Now()
	local.SecretHash = &newHashed
	local.LastChangedAt = &now
	local.UpdatedAt = now
	user.UpdatedAt = now
	user.ResetFailedAttempts()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.ResetUserPasswordResponse{GeneratedPassword: plainPw.String()}, nil
}

type DeactivateUserUseCase struct {
	userRepo userrepo.UserRepository
}

func NewDeactivateUserUseCase(userRepo userrepo.UserRepository) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{userRepo: userRepo}
}

func (uc *DeactivateUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := user.Deactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserAlreadyBanned:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return userToManagementResponse(user), nil
}

type ReactivateUserUseCase struct {
	userRepo userrepo.UserRepository
}

func NewReactivateUserUseCase(userRepo userrepo.UserRepository) *ReactivateUserUseCase {
	return &ReactivateUserUseCase{userRepo: userRepo}
}

func (uc *ReactivateUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := user.Reactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotActive:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return userToManagementResponse(user), nil
}

func userToManagementResponse(user *userentity.User) *dto.UserManagementResponse {
	phoneStr := (*string)(nil)
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
	}

	return &dto.UserManagementResponse{
		ID:          user.ID,
		Username:    user.Username.String(),
		Fullname:    user.Fullname,
		Email:       user.Email.String(),
		Phone:       phoneStr,
		Status:      string(user.Status),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}
