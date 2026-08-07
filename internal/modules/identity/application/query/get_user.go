package query

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	userconstant "sipon-be/internal/modules/identity/domain/user/constant"
	userentity "sipon-be/internal/modules/identity/domain/user/entity"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"
)

type GetUserUseCase struct {
	userRepo     userrepo.UserRepository
	userRoleRepo rolerepo.UserRoleRepository
	roleRepo     rolerepo.RoleRepository
}

func NewGetUserUseCase(
	userRepo userrepo.UserRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	roleRepo rolerepo.RoleRepository,
) *GetUserUseCase {
	return &GetUserUseCase{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
	}
}

func (uc *GetUserUseCase) Execute(ctx context.Context, userID string) (*dto.UserManagementResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case userconstant.ErrCodeUserNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return buildUserManagementResponse(ctx, uc.userRoleRepo, uc.roleRepo, user)
}

func buildUserManagementResponse(ctx context.Context, userRoleRepo rolerepo.UserRoleRepository, roleRepo rolerepo.RoleRepository, user *userentity.User) (*dto.UserManagementResponse, error) {
	userRoles, err := userRoleRepo.FindActiveByUserID(ctx, user.ID)
	roles := make([]dto.UserRoleSummaryResponse, 0)
	if err == nil {
		for _, ur := range userRoles {
			if !ur.IsUsable() {
				continue
			}
			role, err := roleRepo.FindByID(ctx, ur.RoleID)
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
