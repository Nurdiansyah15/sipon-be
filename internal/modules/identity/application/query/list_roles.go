package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	"sipon-be/internal/shared/kernel"
)

type RoleListRepository interface {
	List(ctx context.Context, roleType string, scopeType string, assignable *bool, sortBy string, sortType string, page, limit int) ([]*roleentity.Role, int64, error)
}

type ListRolesUseCase struct {
	roleListRepo RoleListRepository
}

func NewListRolesUseCase(roleListRepo RoleListRepository) *ListRolesUseCase {
	return &ListRolesUseCase{roleListRepo: roleListRepo}
}

func (uc *ListRolesUseCase) Execute(ctx context.Context, req dto.ListRolesRequest) ([]dto.RoleItem, dto.Meta, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	roles, total, err := uc.roleListRepo.List(ctx, req.RoleType, req.ScopeType, req.Assignable, req.SortBy, req.SortType, req.Page, req.Limit)
	if err != nil {
		return nil, dto.Meta{}, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	items := make([]dto.RoleItem, 0, len(roles))
	for _, role := range roles {
		var permKeys []string
		if role.IsSystem() {
			for _, pk := range roleconstant.PermissionsForRole(role.Name) {
				permKeys = append(permKeys, string(pk))
			}
		}

		items = append(items, dto.RoleItem{
			ID:          role.ID,
			Name:        string(role.Name),
			DisplayName: role.DisplayName,
			Description: role.Description,
			RoleType:    string(role.RoleType),
			ScopeType:   string(role.ScopeType),
			Assignable:  role.Assignable,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
			Permissions: permKeys,
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
