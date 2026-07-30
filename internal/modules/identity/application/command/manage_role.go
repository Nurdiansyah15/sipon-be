package command

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type CreateRoleUseCase struct {
	roleRepo domain.RoleRepository
}

func NewCreateRoleUseCase(roleRepo domain.RoleRepository) *CreateRoleUseCase {
	return &CreateRoleUseCase{roleRepo: roleRepo}
}

func (uc *CreateRoleUseCase) Execute(ctx context.Context, req dto.CreateRoleRequest) (*dto.RoleItem, error) {
	roleName := domain.RoleName(req.Name)
	scopeType := domain.ScopeType(req.ScopeType)

	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}

	role, err := domain.NewRole(uuid.NewString(), roleName, req.DisplayName, desc, domain.RoleTypeCustom, scopeType, req.Assignable)
	if err != nil {
		return nil, err
	}

	if err := uc.roleRepo.Save(ctx, role); err != nil {
		return nil, err
	}

	return &dto.RoleItem{
		ID:          role.ID,
		Name:        string(role.Name),
		DisplayName: role.DisplayName,
		Description: role.Description,
		RoleType:    string(role.RoleType),
		ScopeType:   string(role.ScopeType),
		Assignable:  role.Assignable,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

type UpdateRoleUseCase struct {
	roleRepo domain.RoleRepository
}

func NewUpdateRoleUseCase(roleRepo domain.RoleRepository) *UpdateRoleUseCase {
	return &UpdateRoleUseCase{roleRepo: roleRepo}
}

func (uc *UpdateRoleUseCase) Execute(ctx context.Context, roleID string, req dto.UpdateRoleRequest) (*dto.RoleItem, error) {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, kernel.Wrap(domain.ErrCodeRoleNotFound, err)
	}

	if req.DisplayName != "" {
		role.DisplayName = req.DisplayName
	}

	if req.Description != "" {
		role.Description = &req.Description
	}

	if req.Assignable != nil {
		role.Assignable = *req.Assignable
	}

	role.Touch()

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	return &dto.RoleItem{
		ID:          role.ID,
		Name:        string(role.Name),
		DisplayName: role.DisplayName,
		Description: role.Description,
		RoleType:    string(role.RoleType),
		ScopeType:   string(role.ScopeType),
		Assignable:  role.Assignable,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}

type AssignUserRoleUseCase struct {
	userRepo      domain.UserRepository
	roleRepo      domain.RoleRepository
	userRoleRepo  domain.UserRoleRepository
	roleAssignment *domain.UserRoleAssignmentService
}

func NewAssignUserRoleUseCase(
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	userRoleRepo domain.UserRoleRepository,
	roleAssignment *domain.UserRoleAssignmentService,
) *AssignUserRoleUseCase {
	return &AssignUserRoleUseCase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		roleAssignment: roleAssignment,
	}
}

func (uc *AssignUserRoleUseCase) Execute(ctx context.Context, assignedBy string, req dto.AssignUserRoleRequest) error {
	_, err := uc.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	roleName := domain.RoleName(req.RoleName)
	scopeType := domain.ScopeType(req.ScopeType)
	if scopeType == "" {
		scopeType = domain.ScopeTypeGlobal
	}

	if err := uc.roleAssignment.AssignByRoleName(ctx, domain.AssignRoleInput{
		UserID:     req.UserID,
		RoleName:   roleName,
		ScopeType:  scopeType,
		ScopeID:    req.ScopeID,
		AssignedBy: assignedBy,
		ExpiredAt:  req.ExpiredAt,
	}); err != nil {
		return err
	}

	role, err := uc.roleRepo.FindByName(ctx, roleName)
	if err != nil {
		return err
	}

	userRole, err := domain.NewUserRole(uuid.NewString(), req.UserID, role.ID, scopeType, req.ScopeID, assignedBy, req.ExpiredAt, nil)
	if err != nil {
		return err
	}

	return uc.userRoleRepo.Save(ctx, userRole)
}

type UpdateUserRoleUseCase struct {
	userRoleRepo domain.UserRoleRepository
}

func NewUpdateUserRoleUseCase(userRoleRepo domain.UserRoleRepository) *UpdateUserRoleUseCase {
	return &UpdateUserRoleUseCase{userRoleRepo: userRoleRepo}
}

func (uc *UpdateUserRoleUseCase) Execute(ctx context.Context, userRoleID string, req dto.UpdateUserRoleRequest) error {
	userRole, err := uc.userRoleRepo.FindByID(ctx, userRoleID)
	if err != nil {
		return kernel.New(domain.ErrCodeUserRoleNotActive)
	}

	userRole.UpdateExpiration(req.ExpiredAt)

	return uc.userRoleRepo.Update(ctx, userRole)
}

type DeactivateUserRoleUseCase struct {
	userRoleRepo domain.UserRoleRepository
}

func NewDeactivateUserRoleUseCase(userRoleRepo domain.UserRoleRepository) *DeactivateUserRoleUseCase {
	return &DeactivateUserRoleUseCase{userRoleRepo: userRoleRepo}
}

func (uc *DeactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) error {
	userRole, err := uc.userRoleRepo.FindByID(ctx, userRoleID)
	if err != nil {
		return kernel.New(domain.ErrCodeUserRoleNotActive)
	}

	if err := userRole.Deactivate(); err != nil {
		return err
	}

	return uc.userRoleRepo.Update(ctx, userRole)
}

type ReactivateUserRoleUseCase struct {
	userRoleRepo domain.UserRoleRepository
}

func NewReactivateUserRoleUseCase(userRoleRepo domain.UserRoleRepository) *ReactivateUserRoleUseCase {
	return &ReactivateUserRoleUseCase{userRoleRepo: userRoleRepo}
}

func (uc *ReactivateUserRoleUseCase) Execute(ctx context.Context, userRoleID string) error {
	userRole, err := uc.userRoleRepo.FindByID(ctx, userRoleID)
	if err != nil {
		return kernel.New(domain.ErrCodeUserRoleNotActive)
	}

	if err := userRole.Reactivate(); err != nil {
		return err
	}

	return uc.userRoleRepo.Update(ctx, userRole)
}

type DeleteUserRoleUseCase struct {
	userRoleRepo domain.UserRoleRepository
}

func NewDeleteUserRoleUseCase(userRoleRepo domain.UserRoleRepository) *DeleteUserRoleUseCase {
	return &DeleteUserRoleUseCase{userRoleRepo: userRoleRepo}
}

func (uc *DeleteUserRoleUseCase) Execute(ctx context.Context, userRoleID string) error {
	return uc.userRoleRepo.Delete(ctx, userRoleID)
}
