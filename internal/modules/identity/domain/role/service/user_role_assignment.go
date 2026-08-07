package service

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/identity/domain/role/constant"
	"sipon-be/internal/modules/identity/domain/role/entity"
	"sipon-be/internal/modules/identity/domain/role/repository"
	"sipon-be/internal/shared/kernel"
)

type AssignRoleInput struct {
	UserID     string
	RoleName   constant.RoleName
	ScopeType  constant.ScopeType
	ScopeID    *string
	AssignedBy string
	ExpiredAt  *time.Time
}

type UserRoleAssignmentService struct {
	roleRepo     repository.RoleRepository
	userRoleRepo repository.UserRoleRepository
}

func NewUserRoleAssignmentService(roleRepo repository.RoleRepository, userRoleRepo repository.UserRoleRepository) *UserRoleAssignmentService {
	return &UserRoleAssignmentService{
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
	}
}

type AssignRoleByIDInput struct {
	UserID     string
	RoleID     string
	ScopeType  constant.ScopeType
	ScopeID    *string
	AssignedBy string
	ExpiredAt  *time.Time
}

func (s *UserRoleAssignmentService) AssignByRoleID(ctx context.Context, input AssignRoleByIDInput) (*entity.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, input.RoleID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(kernel.Code("ERR_NOT_FOUND"), ke.Message, ke)
		}
		return nil, kernel.Wrap(kernel.Code("ERR_NOT_FOUND"), err)
	}

	if err := s.checkAssignable(ctx, role, input.ScopeType, input.UserID); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *UserRoleAssignmentService) checkAssignable(ctx context.Context, role *entity.Role, scopeType constant.ScopeType, userID string) error {
	if err := role.EnsureAssignable(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(kernel.Code("ERR_FORBIDDEN"), ke.Message, ke)
		}
		return kernel.Wrap(kernel.Code("ERR_FORBIDDEN"), err)
	}

	if err := role.EnsureAssignmentScopeMatch(scopeType); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(kernel.Code("ERR_BAD_REQUEST"), ke.Message, ke)
		}
		return kernel.Wrap(kernel.Code("ERR_BAD_REQUEST"), err)
	}

	existingRoles, err := s.userRoleRepo.FindActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, ur := range existingRoles {
		if ur.RoleID == role.ID && ur.IsUsable() {
			return kernel.WrapMsg(kernel.Code("ERR_CONFLICT"), "User role sudah ditugaskan", nil)
		}
	}

	return nil
}

func (s *UserRoleAssignmentService) AssignByRoleName(ctx context.Context, input AssignRoleInput) error {
	role, err := s.roleRepo.FindByName(ctx, input.RoleName)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return kernel.WrapMsg(kernel.Code("ERR_NOT_FOUND"), ke.Message, ke)
		}
		return kernel.Wrap(kernel.Code("ERR_NOT_FOUND"), err)
	}

	return s.checkAssignable(ctx, role, input.ScopeType, input.UserID)
}
