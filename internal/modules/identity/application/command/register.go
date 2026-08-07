package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	roleservice "sipon-be/internal/modules/identity/domain/role/service"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	uservo "sipon-be/internal/modules/identity/domain/user/valueobject"
	verificationrepo "sipon-be/internal/modules/identity/domain/verification/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type RegisterUseCase struct {
	userRepo       userrepo.UserRepository
	verifRepo      verificationrepo.VerificationRepository
	roleRepo       rolerepo.RoleRepository
	userRoleRepo   rolerepo.UserRoleRepository
	hasher         ports.PasswordHasher
	otpGen         ports.OTPGenerator
	emailSender    ports.EmailSender
	smsSender      ports.SMSSender
	tokenGen       ports.TokenGenerator
	transactor     ports.Transactor
	roleAssignment *roleservice.UserRoleAssignmentService
}

func NewRegisterUseCase(
	userRepo userrepo.UserRepository,
	verifRepo verificationrepo.VerificationRepository,
	roleRepo rolerepo.RoleRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	hasher ports.PasswordHasher,
	otpGen ports.OTPGenerator,
	emailSender ports.EmailSender,
	smsSender ports.SMSSender,
	tokenGen ports.TokenGenerator,
	transactor ports.Transactor,
	roleAssignment *roleservice.UserRoleAssignmentService,
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
	email, err := uservo.NewEmail(req.Email)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeEmailEmpty, userconstant.ErrCodeEmailInvalidFormat:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	username, err := uservo.NewUsername(req.Username)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUsernameEmpty, userconstant.ErrCodeUsernameTooLong, userconstant.ErrCodeUsernameTooShort, userconstant.ErrCodeUsernameInvalidChar:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	var phone *uservo.PhoneNumber
	if req.Phone != nil {
		pn, err := uservo.NewPhoneNumber(*req.Phone)
		if err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case userconstant.ErrCodePhoneNumberEmpty, userconstant.ErrCodePhoneNumberInvalidFormat:
					return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
				}
			}
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		phone = &pn
	}

	plainPw, err := uservo.NewPlainPassword(req.Password)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodePlainPasswordEmpty, userconstant.ErrCodePlainPasswordTooShort, userconstant.ErrCodePlainPasswordNoUppercase, userconstant.ErrCodePlainPasswordNoDigit:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	emailExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, userconstant.LoginIdentifierKindEmail, email.String())
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if phone != nil {
		phoneExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, userconstant.LoginIdentifierKindPhone, phone.String())
		if err != nil {
			return nil, err
		}
		if phoneExists {
			return nil, kernel.New(application.ErrCodeConflict)
		}
	}

	usernameExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, userconstant.LoginIdentifierKindUsername, username.String())
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

	hashedPw, err := uservo.NewHashedPassword(hashedPassword)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeHashedPasswordTooShort:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	userID := uuid.NewString()
	credentialID := uuid.NewString()

	user, err := userentity.NewUser(userID, username, req.Fullname, email, phone)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserIDRequired, userconstant.ErrCodeUserEmailRequired, userconstant.ErrCodeUserPhoneNumberInvalid:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	credential := userentity.NewLocalCredential(credentialID, userID, hashedPw, true)

	emailLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindEmail, email.String(), true, nil)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	credential.AddLoginIdentity(emailLI)

	now := time.Now()
	usernameLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindUsername, username.String(), true, &now)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	credential.AddLoginIdentity(usernameLI)

	if phone != nil {
		phoneLI, err := userentity.NewLoginIdentity(uuid.NewString(), userID, credentialID, userconstant.LoginIdentifierKindPhone, phone.String(), true, nil)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		credential.AddLoginIdentity(phoneLI)
	}

	user.AddCredential(credential)

	memberRole, err := uc.roleRepo.FindByName(ctx, roleconstant.MemberRoleName)
	if err != nil {
		return nil, err
	}

	userRole, err := roleentity.NewUserRole(uuid.NewString(), userID, memberRole.ID, roleconstant.ScopeTypeGlobal, nil, userID, nil, nil)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeUserRoleIDRequired, roleconstant.ErrCodeUserRoleUserIDRequired, roleconstant.ErrCodeUserRoleRoleIDRequired, roleconstant.ErrCodeInvalidScopeType, roleconstant.ErrCodeUserRoleScopeIDEmpty, roleconstant.ErrCodeUserRoleScopeIDRequired:
				return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
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
			case userconstant.ErrCodeUserBanned, userconstant.ErrCodeUserNotActive, userconstant.ErrCodeUserLockedOut:
				return nil, kernel.WrapMsg(application.ErrCodeForbidden, ke.Message, ke)
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
	if li := user.FindLoginIdentity(userconstant.LoginIdentifierKindEmail, email.String()); li != nil {
		isEmailVerified = li.IsVerified()
	}

	var phoneStr *string
	var isPhoneVerified bool
	if phone != nil {
		s := phone.String()
		phoneStr = &s
		if li := user.FindLoginIdentity(userconstant.LoginIdentifierKindPhone, phone.String()); li != nil {
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
