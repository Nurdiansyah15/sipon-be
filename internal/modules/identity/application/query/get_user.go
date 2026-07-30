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

func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserItem, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUserNotFound, err)
	}

	userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, userID)
	roleNames := make([]string, 0)
	if err == nil {
		for _, ur := range userRoles {
			if ur.IsUsable() {
				role, err := uc.roleRepo.FindByID(ctx, ur.RoleID)
				if err == nil {
					roleNames = append(roleNames, string(role.Name))
				}
			}
		}
	}

	phoneStr := (*string)(nil)
	if user.PhoneNumber != nil {
		s := user.PhoneNumber.String()
		phoneStr = &s
	}

	return &dto.UserItem{
		ID:          user.ID,
		Username:    user.Username.String(),
		Fullname:    user.Fullname,
		Email:       user.Email.String(),
		Phone:       phoneStr,
		Status:      string(user.Status),
		Roles:       roleNames,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}, nil
}
