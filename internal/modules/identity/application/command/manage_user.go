package command

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type CreateUserUseCase struct {
	userRepo       domain.UserRepository
	roleRepo       domain.RoleRepository
	userRoleRepo   domain.UserRoleRepository
	hasher         application.PasswordHasher
	transactor     application.Transactor
	roleAssignment *domain.UserRoleAssignmentService
}

func NewCreateUserUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	userRoleRepo domain.UserRoleRepository,
	hasher application.PasswordHasher,
	transactor application.Transactor,
	roleAssignment *domain.UserRoleAssignmentService,
) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		hasher:         hasher,
		transactor:     transactor,
		roleAssignment: roleAssignment,
	}
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req dto.CreateUserRequest, createdBy string) error {
	username, err := domain.NewUsername(req.Username)
	if err != nil {
		return err
	}

	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return err
	}

	var phone *domain.PhoneNumber
	if req.Phone != "" {
		pn, err := domain.NewPhoneNumber(req.Phone)
		if err != nil {
			return err
		}
		phone = &pn
	}

	plainPw, err := domain.NewPlainPassword(req.Password)
	if err != nil {
		return err
	}

	emailExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindEmail, email.String())
	if err != nil {
		return err
	}
	if emailExists {
		return kernel.New(application.ErrCodeDuplicateEmail)
	}

	if phone != nil {
		phoneExists, err := uc.userRepo.ExistsByLoginIdentity(ctx, domain.LoginIdentifierKindPhone, phone.String())
		if err != nil {
			return err
		}
		if phoneExists {
			return kernel.New(application.ErrCodeDuplicatePhone)
		}
	}

	usernameExists, err := uc.userRepo.ExistsByUsername(ctx, username.String())
	if err != nil {
		return err
	}
	if usernameExists {
		return kernel.New(application.ErrCodeDuplicateUsername)
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		return err
	}

	userID := uuid.NewString()
	credentialID := uuid.NewString()
	fullname := strPtr(req.Fullname)

	user, err := domain.NewUser(userID, username, fullname, email, phone)
	if err != nil {
		return err
	}

	credential := domain.NewLocalCredential(credentialID, userID, hashedPw, true)

	emailLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindEmail, email.String(), true, nil)
	if err != nil {
		return err
	}
	credential.AddLoginIdentity(emailLI)

	usernameLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindUsername, username.String(), false, nil)
	if err != nil {
		return err
	}
	credential.AddLoginIdentity(usernameLI)

	if phone != nil {
		phoneLI, err := domain.NewLoginIdentity(uuid.NewString(), userID, credentialID, domain.LoginIdentifierKindPhone, phone.String(), false, nil)
		if err != nil {
			return err
		}
		credential.AddLoginIdentity(phoneLI)
	}

	user.AddCredential(credential)

	return uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.userRepo.Save(txCtx, user); err != nil {
			return err
		}

		roleName := domain.RoleName(req.RoleName)
		role, err := uc.roleRepo.FindByName(txCtx, roleName)
		if err != nil {
			return err
		}

		if err := uc.roleAssignment.AssignByRoleName(txCtx, domain.AssignRoleInput{
			UserID:     userID,
			RoleName:   roleName,
			ScopeType:  domain.ScopeTypeGlobal,
			ScopeID:    nil,
			AssignedBy: createdBy,
			ExpiredAt:  nil,
		}); err != nil {
			return err
		}

		userRole, err := domain.NewUserRole(uuid.NewString(), userID, role.ID, domain.ScopeTypeGlobal, nil, createdBy, nil, nil)
		if err != nil {
			return err
		}

		return uc.userRoleRepo.Save(txCtx, userRole)
	})
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

func (uc *ResetUserPasswordUseCase) Execute(ctx context.Context, userID string, req dto.ResetUserPasswordRequest) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	plainPw, err := domain.NewPlainPassword(req.NewPassword)
	if err != nil {
		return err
	}

	hashedPassword, err := uc.hasher.Hash(plainPw.String())
	if err != nil {
		return err
	}

	hashedPw, err := domain.NewHashedPassword(hashedPassword)
	if err != nil {
		return err
	}

	if err := user.SetLocalPassword(hashedPw); err != nil {
		return err
	}

	return uc.userRepo.Update(ctx, user)
}

type DeactivateUserUseCase struct {
	userRepo domain.UserRepository
}

func NewDeactivateUserUseCase(userRepo domain.UserRepository) *DeactivateUserUseCase {
	return &DeactivateUserUseCase{userRepo: userRepo}
}

func (uc *DeactivateUserUseCase) Execute(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	if err := user.Deactivate(); err != nil {
		return err
	}

	return uc.userRepo.Update(ctx, user)
}

type ReactivateUserUseCase struct {
	userRepo domain.UserRepository
}

func NewReactivateUserUseCase(userRepo domain.UserRepository) *ReactivateUserUseCase {
	return &ReactivateUserUseCase{userRepo: userRepo}
}

func (uc *ReactivateUserUseCase) Execute(ctx context.Context, userID string) error {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	if err := user.Reactivate(); err != nil {
		return err
	}

	return uc.userRepo.Update(ctx, user)
}
