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

func (uc *ListUsersUseCase) Execute(ctx context.Context, req dto.ListUsersRequest) (*dto.ListUsersResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	users, total, err := uc.userListRepo.List(ctx, req.Status, req.RoleID, req.Search, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.UserItem, 0, len(users))
	for _, user := range users {
		userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, user.ID)
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

		items = append(items, dto.UserItem{
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
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &dto.ListUsersResponse{
		Users: items,
		Meta: dto.Meta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: int(total),
			TotalPages: totalPages,
		},
	}, nil
}
