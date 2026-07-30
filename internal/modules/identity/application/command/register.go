package command

import (
	"context"
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
	username, err := domain.NewUsername(req.Username)
	if err != nil {
		return nil, err
	}

	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return nil, err
	}

	var phone *domain.PhoneNumber
	if req.Phone != "" {
		pn, err := domain.NewPhoneNumber(req.Phone)
		if err != nil {
			return nil, err
		}
		phone = &pn
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		return nil, err
	}

	emailExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindEmail, email.String())
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, kernel.New(application.ErrCodeDuplicateEmail)
	}

	if phone != nil {
		phoneExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindPhone, phone.String())
		if err != nil {
			return nil, err
		}
		if phoneExists {
			return nil, kernel.New(application.ErrCodeDuplicatePhone)
		}
	}

	usernameExists, err := uc.userRepo.ExistsByUsername(ctx, username.String())
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, kernel.New(application.ErrCodeDuplicateUsername)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return nil, err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		return nil, err
	}

	userID := uuid.NewString()
	credentialID := uuid.NewString()

	fullname := strPtr(req.Fullname)

	user, err := domain.NewUser(userID, username, fullname, email, phone)
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

	var roleName string
	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Save(txCtx, user); err != nil {
			return err
		}

		role, err := uc.roleRepo.FindByName(txCtx, domain.MemberRoleName)
		if err != nil {
			return err
		}
		roleName = string(role.Name)

		if err := uc.roleAssignment.AssignByRoleName(txCtx, domain.AssignRoleInput{
			UserID:     userID,
			RoleName:   domain.MemberRoleName,
			ScopeType:  domain.ScopeTypeGlobal,
			ScopeID:    nil,
			AssignedBy: userID,
			ExpiredAt:  nil,
		}); err != nil {
			return err
		}

		userRole, err := domain.NewUserRole(uuid.NewString(), userID, role.ID, domain.ScopeTypeGlobal, nil, userID, nil, nil)
		if err != nil {
			return err
		}

		return uc.userRoleRepo.Save(txCtx, userRole)
	})
	if err != nil {
		return nil, err
	}

	otpCode, err := uc.otpGen.Generate()
	if err != nil {
		return nil, err
	}

	otp, err := domain.NewOTPCode(otpCode)
	if err != nil {
		return nil, err
	}

	sessionID := uuid.NewString()
	deviceID := uuid.NewString()

	accessToken, err := uc.tokenGen.GenerateAccessToken(userID, sessionID, deviceID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenGen.GenerateRefreshToken(userID, deviceID)
	if err != nil {
		return nil, err
	}

	go func() {
		bgCtx := context.Background()
		verifCode, err := domain.NewVerificationCode(uuid.NewString(), userID, otp, domain.PurposeEmailVerification, time.Now().Add(10*time.Minute))
		if err != nil {
			return
		}
		_ = uc.verifRepo.Save(bgCtx, verifCode)
		_ = uc.emailSender.SendOTP(email.String(), req.Fullname, otpCode)
	}()

	phoneStr := (*string)(nil)
	if phone != nil {
		s := phone.String()
		phoneStr = &s
	}

	return &dto.RegisterResponse{
		UserID:       userID,
		Username:     username.String(),
		Email:        email.String(),
		Phone:        phoneStr,
		Roles:        []string{roleName},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
