package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	ports "sipon-be/internal/modules/identity/application/ports"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	"sipon-be/internal/shared/kernel"
)

type ListUsersUseCase struct {
	userListRepo ports.UserListRepository
	userRoleRepo rolerepo.UserRoleRepository
	roleRepo     rolerepo.RoleRepository
}

func NewListUsersUseCase(
	userListRepo ports.UserListRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	roleRepo rolerepo.RoleRepository,
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

	users, total, err := uc.userListRepo.List(ctx, req.Status, req.RoleID, req.Search, req.SortBy, req.SortType, req.Page, req.Limit)
	if err != nil {
		return nil, dto.Meta{}, kernel.WrapMsg(application.ErrCodeInternal, "gagal mencari daftar pengguna", err)
	}

	userIDs := make([]string, 0, len(users))
	items := make([]dto.UserManagementResponse, 0, len(users))
	for _, user := range users {
		var phone *string
		if user.PhoneNumber != nil {
			s := user.PhoneNumber.String()
			phone = &s
		}

		items = append(items, dto.UserManagementResponse{
			ID:          user.ID,
			Username:    user.Username.String(),
			Fullname:    user.Fullname,
			Email:       user.Email.String(),
			Phone:       phone,
			Status:      string(user.Status),
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			LastLoginAt: user.LastLoginAt,
		})
		userIDs = append(userIDs, user.ID)
	}

	if len(userIDs) > 0 {
		for i := range items {
			userRoles, err := uc.userRoleRepo.FindActiveByUserID(ctx, items[i].ID)
			if err != nil {
				continue
			}
			roles := make([]dto.UserRoleSummaryResponse, 0)
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
			items[i].Roles = roles
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return items, dto.Meta{
		CurrentPage: req.Page,
		PerPage:     req.Limit,
		Total:       int(total),
		TotalPages:  totalPages,
	}, nil
}
