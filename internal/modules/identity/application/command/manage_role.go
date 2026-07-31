package command

import (
	"context"
	"errors"

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
	roleType := domain.RoleType(req.RoleType)

	if _, err := uc.roleRepo.FindByName(ctx, roleName); err == nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	var desc *string
	if req.Description != nil && *req.Description != "" {
		desc = req.Description
	}

	role, err := domain.NewRole(uuid.NewString(), roleName, req.DisplayName, desc, roleType, scopeType, req.Assignable)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
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
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := role.EnsureCustom(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeInvalidRoleType:
				return nil, kernel.New(application.ErrCodeBadRequest)
			}
		}
		return nil, kernel.New(application.ErrCodeForbidden)
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
	userRepo       domain.UserRepository
	roleRepo       domain.RoleRepository
	userRoleRepo   domain.UserRoleRepository
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
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	scopeType := domain.ScopeType(req.ScopeType)

	role, err := uc.roleAssignment.AssignByRoleID(ctx, domain.AssignRoleByIDInput{
		UserID:     req.UserID,
		RoleID:     req.RoleID,
		ScopeType:  scopeType,
		ScopeID:    req.ScopeID,
		AssignedBy: assignedBy,
		ExpiredAt:  req.ExpiredAt,
	})
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case kernel.Code("ERR_NOT_FOUND"):
				return kernel.New(application.ErrCodeNotFound)
			case kernel.Code("ERR_FORBIDDEN"):
				return kernel.New(application.ErrCodeForbidden)
			case kernel.Code("ERR_BAD_REQUEST"):
				return kernel.New(application.ErrCodeBadRequest)
			case kernel.Code("ERR_CONFLICT"):
				return kernel.New(application.ErrCodeConflict)
			}
		}
		return kernel.New(application.ErrCodeConflict)
	}

	userRole, err := domain.NewUserRole(uuid.NewString(), req.UserID, role.ID, scopeType, req.ScopeID, assignedBy, req.ExpiredAt, req.Notes)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, string(ke.Code), ke)
		}
		return kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
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
		return kernel.Wrap(application.ErrCodeNotFound, err)
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
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := userRole.Deactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserRoleNotActive:
				return kernel.New(application.ErrCodeBadRequest)
			}
		}
		return kernel.New(application.ErrCodeBadRequest)
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
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}

	if err := userRole.Reactivate(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case domain.ErrCodeUserRoleNotActive:
				return kernel.New(application.ErrCodeBadRequest)
			case domain.ErrCodeUserRoleExpired:
				return kernel.New(application.ErrCodeGone)
			}
		}
		return kernel.New(application.ErrCodeBadRequest)
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
	if _, err := uc.userRoleRepo.FindByID(ctx, userRoleID); err != nil {
		return kernel.Wrap(application.ErrCodeNotFound, err)
	}
	return uc.userRoleRepo.Delete(ctx, userRoleID)
}
