package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
)

type UserRoleListRepository interface {
	List(ctx context.Context, userID, roleID, scopeType string, isActive *bool, page, limit int) ([]*domain.UserRole, int64, error)
}

type ListUserRolesUseCase struct {
	userRoleListRepo UserRoleListRepository
	userRepo         domain.UserRepository
	roleRepo         domain.RoleRepository
	rolePermRepo     domain.RolePermissionRepository
}

func NewListUserRolesUseCase(
	userRoleListRepo UserRoleListRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
) *ListUserRolesUseCase {
	return &ListUserRolesUseCase{
		userRoleListRepo: userRoleListRepo,
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		rolePermRepo:     rolePermRepo,
	}
}

func (uc *ListUserRolesUseCase) Execute(ctx context.Context, req dto.ListUserRolesRequest) (*dto.ListUserRolesResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	userRoles, total, err := uc.userRoleListRepo.List(ctx, req.UserID, req.RoleID, req.ScopeType, req.IsActive, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.UserRoleItem, 0, len(userRoles))
	for _, ur := range userRoles {
		role, err := uc.roleRepo.FindByID(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		user, err := uc.userRepo.FindByID(ctx, ur.UserID)
		if err != nil {
			continue
		}

		rps, _ := uc.rolePermRepo.ListByRoleID(ctx, ur.RoleID)
		perms := make([]string, 0, len(rps))
		for _, rp := range rps {
			perms = append(perms, string(rp.PermissionKey))
		}

		emailStr := user.Email.String()
		userSummary := dto.UserSummary{
			ID:    user.ID,
			Name:  user.Fullname,
			Email: &emailStr,
		}
		if user.PhoneNumber != nil {
			phone := user.PhoneNumber.String()
			userSummary.Phone = &phone
		}

		items = append(items, dto.UserRoleItem{
			ID:     ur.ID,
			UserID: ur.UserID,
			User:   userSummary,
			RoleID: ur.RoleID,
			Role: dto.RoleSummary{
				ID:          role.ID,
				Name:        string(role.Name),
				DisplayName: role.DisplayName,
				RoleType:    string(role.RoleType),
				Assignable:  role.Assignable,
			},
			ScopeType:     string(ur.ScopeType),
			ScopeID:       ur.ScopeID,
			AssignedAt:    ur.AssignedAt,
			AssignedBy:    &ur.AssignedBy,
			ExpiredAt:     ur.ExpiredAt,
			IsActive:      ur.IsActive,
			DeactivatedAt: ur.DeactivatedAt,
			Permissions:   perms,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &dto.ListUserRolesResponse{
		UserRoles: items,
		Meta: dto.Meta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: int(total),
			TotalPages: totalPages,
		},
	}, nil
}
