package query

import (
	"context"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type GetUserUseCase struct {
	userRepo     domain.UserRepository
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
}

func NewGetUserUseCase(
	userRepo domain.UserRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
	}
}

func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, userID)
	roles := make([]dto.UserRoleSummaryResponse, 0)
	if err == nil {
		for _, ur := range userRoles {
			if !ur.IsUsable() {
				continue
			}
			role, err := uc.roleRepo.FindByID(ctx, ur.RoleID)
			if err != nil {
				continue
			}
			roles = append(roles, dto.UserRoleSummaryResponse{
				ID:        ur.ID,
				RoleID:    role.ID,
				RoleName:  string(role.Name),
				ScopeType: string(ur.ScopeType),
				ScopeID:   ur.ScopeID,
				IsActive:  ur.IsActive,
			})
		}
	}

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
		Roles:       roles,
	}, nil
}
