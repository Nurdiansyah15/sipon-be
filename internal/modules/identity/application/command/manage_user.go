package command

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

const (
	generatedPasswordLength = 12
	generatedPasswordUpper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	generatedPasswordLower  = "abcdefghijklmnopqrstuvwxyz"
	generatedPasswordDigits = "0123456789"
)

// generateRandomPassword produces a random password guaranteed to contain at
// least one uppercase letter and one digit, satisfying domain.NewPlainPassword.
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
	userRepo domain.UserRepository
	hasher   application.PasswordHasher
}

func NewCreateUserUseCase(
	userRepo domain.UserRepository,
	hasher application.PasswordHasher,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req dto.CreateUserRequest, createdBy string) (*dto.CreateUserResponse, error) {
	username, err := domain.NewUsername(req.Username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	email, err := domain.NewEmail(req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	var phone *domain.PhoneNumber
	if req.Phone != nil && *req.Phone != "" {
		pn, err := domain.NewPhoneNumber(*req.Phone)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
		}
		phone = &pn
	}

	emailExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindEmail, email.String())
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if phone != nil {
		phoneExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindPhone, phone.String())
		if err != nil {
			return nil, err
		}
		if phoneExists {
			return nil, kernel.New(application.ErrCodeConflict)
		}
	}

	usernameExists, err := uc.userRepo.ExistsByUsername(ctx, username.String())
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	generatedPassword, err := generateRandomPassword()
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	hashedPassword, err := uc.hasher.Hash(generatedPassword)
	if err != nil {
		return nil, err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	userID := uuid.NewString()
	credentialID := uuid.NewString()

	user, err := domain.NewUser(userID, username, req.Fullname, email, phone)
	if err != nil {
		return nil, err
	}

	credential := domain.NewLocalCredential(credentialID, userID, hashedPw, true)

	emailLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindEmail, email.String(), true, nil)
	if err != nil {
		return nil, err
	}
	credential.AddLoginIdentity(emailLI)

	usernameLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindUsername, username.String(), false, nil)
	if err != nil {
		return nil, err
	}
	credential.AddLoginIdentity(usernameLI)

	if phone != nil {
		phoneLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindPhone, phone.String(), false, nil)
		if err != nil {
			return nil, err
		}
		credential.AddLoginIdentity(phoneLI)
	}

	user.AddCredential(credential)

	if err := uc.userRepo.Save(ctx, user); err != nil {
		return nil, err
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
			Status:    string(domain.UserStatusActive),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		GeneratedPassword: generatedPassword,
	}, nil
}

type ResetUserPasswordUseCase struct {
	userRepo domain.UserRepository
	hasher   application.PasswordHasher
}

func NewResetUserPasswordUseCase(
	userRepo domain.UserRepository,
	hasher application.PasswordHasher,
) *ResetUserPasswordUseCase {
	return &ResetUserPasswordUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *ResetUserPasswordUseCase) Execute(ctx context.Context, userID string) (*dto.ResetUserPasswordResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	generatedPassword, err := generateRandomPassword()
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	hashedPassword, err := uc.hasher.Hash(generatedPassword)
	if err != nil {
		return nil, err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if err := user.SetLocalPassword(hashedPw); err != nil {
		return nil, err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return &dto.ResetUserPasswordResponse{GeneratedPassword: generatedPassword}, nil
}

type DeactivateUserUseCase struct {
	userRepo domain.UserRepository
}

func NewDeactivateUserUseCase(userRepo domain.UserRepository) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{userRepo: userRepo}
}

func (uc *DeactivateUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := user.Deactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserAlreadyBanned:
				return nil, kernel.New(application.ErrCodeConflict)
			}
		}
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return userToManagementResponse(user), nil
}

type ReactivateUserUseCase struct {
	userRepo domain.UserRepository
}

func NewReactivateUserUseCase(userRepo domain.UserRepository) *ReactivateUserUseCase {
	return &ReactivateUserUseCase{userRepo: userRepo}
}

func (uc *ReactivateUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := user.Reactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserNotActive:
				return nil, kernel.New(application.ErrCodeConflict)
			}
		}
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return userToManagementResponse(user), nil
}

func userToManagementResponse(user *domain.User) *dto.UserManagementResponse {
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
