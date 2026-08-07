package query

import (
	"context"
	"math"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	userrepo "sipon-be/internal/modules/identity/domain/user/repository"
	"sipon-be/internal/shared/kernel"
)

type UserRoleListRepository interface {
	List(ctx context.Context, userID, roleID, scopeType, scopeID string, isActive *bool, sortBy string, sortType string, page, limit int) ([]*roleentity.UserRole, int64, error)
}

func BuildUserRoleItem(
	ctx context.Context,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
	ur *roleentity.UserRole,
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
	userRepo         userrepo.UserRepository
	roleRepo         rolerepo.RoleRepository
	rolePermRepo     rolerepo.RolePermissionRepository
}

func NewListUserRolesUseCase(
	userRoleListRepo UserRoleListRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
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

	userRoles, total, err := uc.userRoleListRepo.List(ctx,
		strings.TrimSpace(req.UserID),
		strings.TrimSpace(req.RoleID),
		strings.TrimSpace(req.ScopeType),
		strings.TrimSpace(req.ScopeID),
		req.IsActive,
		req.SortBy,
		req.SortType,
		req.Page,
		req.Limit,
	)
	if err != nil {
		return nil, dto.Meta{}, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	items := make([]dto.UserRoleItem, 0, len(userRoles))
	for _, ur := range userRoles {
		item, err := BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, ur)
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
	userRoleRepo rolerepo.UserRoleRepository
	userRepo     userrepo.UserRepository
	roleRepo     rolerepo.RoleRepository
	rolePermRepo rolerepo.RolePermissionRepository
}

func NewGetUserRoleUseCase(
	userRoleRepo rolerepo.UserRoleRepository,
	userRepo userrepo.UserRepository,
	roleRepo rolerepo.RoleRepository,
	rolePermRepo rolerepo.RolePermissionRepository,
) *GetUserRoleUseCase {
	return &GetUserRoleUseCase{
		userRoleRepo: userRoleRepo,
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *GetUserRoleUseCase) Execute(ctx context.Context, userRoleID string) (*dto.UserRoleItem, error) {
	ur, err := uc.userRoleRepo.FindByID(ctx, strings.TrimSpace(userRoleID))
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "data tidak ditemukan", err)
	}

	return BuildUserRoleItem(ctx, uc.userRepo, uc.roleRepo, uc.rolePermRepo, ur)
}
