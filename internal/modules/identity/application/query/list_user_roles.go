package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type UserRoleListRepository interface {
	List(ctx context.Context, userID, roleID, scopeType, scopeID string, isActive *bool, page, limit int) ([]*domain.UserRole, int64, error)
}

// buildUserRoleItem resolves a domain.UserRole into its full wire representation
// (nested user/role summaries + permissions) — shared by ListUserRoles and GetUserRole.
func buildUserRoleItem(
	ctx context.Context,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
	ur *domain.UserRole,
) (*dto.UserRoleItem, error) {
	role, err := roleRepo.FindByID(ctx, ur.RoleID)
	if err != nil {
		return nil, err
	}

	user, err := userRepo.FindByID(ctx, ur.UserID)
	if err != nil {
		return nil, err
	}

	rps, _ := rolePermRepo.ListByRoleID(ctx, ur.RoleID)
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

	return &dto.UserRoleItem{
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
	}, nil
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

func (uc *ListUserRolesUseCase) Execute(ctx context.Context, req dto.ListUserRolesRequest) ([]dto.UserRoleItem, dto.Meta, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	userRoles, total, err := uc.userRoleListRepo.List(ctx, req.UserID, req.RoleID, req.ScopeType, req.ScopeID, req.IsActive, req.Page, req.Limit)
	if err != nil {
		return nil, dto.Meta{}, err
	}

	items := make([]dto.UserRoleItem, 0, len(userRoles))
	for _, ur := range userRoles {
		item, err := buildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, ur)
		if err != nil {
			continue
		}
		items = append(items, *item)
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return items, dto.Meta{
		CurrentPage: req.Page,
		PerPage:     req.Limit,
		Total:       int(total),
		TotalPages:  totalPages,
	}, nil
}

type GetUserRoleUseCase struct {
	userRoleRepo domain.UserRoleRepository
	userRepo     domain.UserRepository
	roleRepo     domain.RoleRepository
	rolePermRepo domain.RolePermissionRepository
}

func NewGetUserRoleUseCase(
	userRoleRepo domain.UserRoleRepository,
	userRepo domain.UserRepository,
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
) *GetUserRoleUseCase {
	return &GetUserRoleUseCase{
		userRoleRepo: userRoleRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *GetUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleItem, error) {
	ur, err := uc.userRoleRepo.FindByID(ctx, userRoleID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	return buildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, ur)
}
