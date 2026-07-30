package domain

import (
	"context"
	"time"

	"sipon-be/internal/shared/kernel"
)

type AssignRoleInput struct {
	UserID     string
	RoleName   RoleName
	ScopeType  ScopeType
	ScopeID    *string
	AssignedBy string
	ExpiredAt  *time.Time
}

type UserRoleAssignmentService struct {
	roleRepo     RoleRepository
	userRoleRepo UserRoleRepository
}

func NewUserRoleAssignmentService(roleRepo RoleRepository, userRoleRepo UserRoleRepository) *UserRoleAssignmentService {
	return &UserRoleAssignmentService{
		roleRepo:     roleRepo,
		userRoleRepo: userRoleRepo,
	}
}

func (s *UserRoleAssignmentService) AssignByRoleName(ctx context.Context, input AssignRoleInput) error {
	role, err := s.roleRepo.FindByName(ctx, input.RoleName)
	if err != nil {
		return kernel.Wrap(kernel.Code("ROLE_NOT_FOUND"), err)
	}

	if err := role.EnsureAssignable(); err != nil {
		return err
	}

	if err := role.EnsureAssignmentScopeMatch(input.ScopeType); err != nil {
		return err
	}

	existingRoles, err := s.userRoleRepo.FindActiveByUserID(ctx, input.UserID)
	if err != nil {
		return err
	}

	for _, ur := range existingRoles {
		if ur.RoleID == role.ID && ur.IsUsable() {
			return kernel.New(ErrCodeUserRoleAlreadyAssigned)
		}
	}

	return nil
}
