package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type RegisterUseCase struct {
	userRepo       domain.UserRepository
	verifRepo      domain.VerificationRepository
	roleRepo       domain.RoleRepository
	userRoleRepo   domain.UserRoleRepository
	hasher         application.PasswordHasher
	otpGen         application.OTPGenerator
	emailSender    application.EmailSender
	smsSender      application.SMSSender
	tokenGen       application.TokenGenerator
	transactor     application.Transactor
	roleAssignment *domain.UserRoleAssignmentService
}

func NewRegisterUseCase(
	userRepo domain.UserRepository,
	verifRepo domain.VerificationRepository,
	roleRepo domain.RoleRepository,
	userRoleRepo domain.UserRoleRepository,
	hasher application.PasswordHasher,
	otpGen application.OTPGenerator,
	emailSender application.EmailSender,
	smsSender application.SMSSender,
	tokenGen application.TokenGenerator,
	transactor application.Transactor,
	roleAssignment *domain.UserRoleAssignmentService,
) *RegisterUseCase {
	return &RegisterUseCase{
		userRepo:       userRepo,
		verifRepo:      verifRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		hasher:         hasher,
		otpGen:         otpGen,
		emailSender:    emailSender,
		smsSender:      smsSender,
		tokenGen:       tokenGen,
		transactor:     transactor,
		roleAssignment: roleAssignment,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	email, err := domain.NewEmail(req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeEmailEmpty, domain.ErrCodeEmailInvalidFormat:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	username, err := domain.NewUsername(req.Username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUsernameEmpty, domain.ErrCodeUsernameTooLong, domain.ErrCodeUsernameTooShort, domain.ErrCodeUsernameInvalidChar:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	var phone *domain.PhoneNumber
	if req.Phone != nil {
		pn, err := domain.NewPhoneNumber(*req.Phone)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case domain.ErrCodePhoneNumberEmpty, domain.ErrCodePhoneNumberInvalidFormat:
					return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
				}
			}
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		phone = &pn
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodePlainPasswordEmpty, domain.ErrCodePlainPasswordTooShort, domain.ErrCodePlainPasswordNoUppercase, domain.ErrCodePlainPasswordNoDigit:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
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

	usernameExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindUsername, username.String())
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return nil, err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeHashedPasswordTooShort:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	userID := uuid.NewString()
	credentialID := uuid.NewString()

	user, err := domain.NewUser(userID, username, req.Fullname, email, phone)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserIDRequired, domain.ErrCodeUserEmailRequired, domain.ErrCodeUserPhoneNumberInvalid:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	credential := domain.NewLocalCredential(credentialID, userID, hashedPw, true)

	emailLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindEmail, email.String(), true, nil)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	credential.AddLoginIdentity(emailLI)

	now := time.Now()
	usernameLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindUsername, username.String(), true, &now)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	credential.AddLoginIdentity(usernameLI)

	if phone != nil {
		phoneLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindPhone, phone.String(), true, nil)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		credential.AddLoginIdentity(phoneLI)
	}

	user.AddCredential(credential)

	memberRole, err := uc.roleRepo.FindByName(ctx, domain.MemberRoleName)
	if err != nil {
		return nil, err
	}

	userRole, err := domain.NewUserRole(uuid.NewString(), userID, memberRole.ID, domain.ScopeTypeGlobal, nil, userID, nil, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserRoleIDRequired, domain.ErrCodeUserRoleUserIDRequired, domain.ErrCodeUserRoleRoleIDRequired, domain.ErrCodeInvalidScopeType, domain.ErrCodeUserRoleScopeIDEmpty, domain.ErrCodeUserRoleScopeIDRequired:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Save(txCtx, user); err != nil {
			return err
		}
		return uc.userRoleRepo.Save(txCtx, userRole)
	}); err != nil {
		return nil, err
	}

	if err := user.EnsureCanLogin(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserBanned, domain.ErrCodeUserNotActive, domain.ErrCodeUserLockedOut:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, string(ke.Code), ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	sessionID := uuid.NewString()
	deviceID := req.DeviceID
	if deviceID == "" {
		deviceID = uuid.NewString()
	}

	accessToken, err := uc.tokenGen.GenerateAccessToken(userID, sessionID, deviceID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	refreshToken, err := uc.tokenGen.GenerateRefreshToken(userID, deviceID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	isEmailVerified := false
	if li := user.FindLoginIdentity(domain.LoginIdentifierKindEmail, email.String()); li != nil {
		isEmailVerified = li.IsVerified()
	}

	var phoneStr *string
	var isPhoneVerified bool
	if phone != nil {
		s := phone.String()
		phoneStr = &s
		if li := user.FindLoginIdentity(domain.LoginIdentifierKindPhone, phone.String()); li != nil {
			isPhoneVerified = li.IsVerified()
		}
	}

	return &dto.RegisterResponse{
		UserID: userID,
		LoginResponse: dto.LoginResponse{
			Token:        accessToken,
			RefreshToken: refreshToken,
			User: dto.UserMe{
				ID:              user.ID,
				Username:        user.Username.String(),
				Email:           user.Email.String(),
				IsEmailVerified: isEmailVerified,
				Fullname:        user.Fullname,
				Phone:           phoneStr,
				IsPhoneVerified: isPhoneVerified,
				Status:          string(user.Status),
				CreatedAt:       user.CreatedAt,
				HasPassword:     user.HasLocalPassword(),
				AvatarURL:       nil,
			},
		},
	}, nil
}
