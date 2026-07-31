package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
)

type ListUsersUseCase struct {
	userListRepo application.UserListRepository
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
}

func NewListUsersUseCase(
	userListRepo application.UserListRepository,
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
) *ListUsersUseCase {
	return &ListUsersUseCase{
		userListRepo: userListRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
	}
}

func (uc *ListUsersUseCase) Execute(ctx context.Context, req dto.ListUsersRequest) ([]dto.UserManagementResponse, dto.Meta, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	users, total, err := uc.userListRepo.List(ctx, req.Status, req.RoleID, req.Search, req.Page, req.Limit)
	if err != nil {
		return nil, dto.Meta{}, err
	}

	items := make([]dto.UserManagementResponse, 0, len(users))
	for _, user := range users {
		userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, user.ID)
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

		items = append(items, dto.UserManagementResponse{
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
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return items, dto.Meta{
		CurrentPage: req.Page,
		PerPage:     req.Limit,
		Total:       int(total),
		TotalPages:  totalPages,
	}, nil
}
