package query

import (
	"context"
	"math"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
)

type RoleListRepository interface {
	List(ctx context.Context, roleType string, scopeType string, assignable *bool, page, limit int) ([]*domain.Role, int64, error)
}

type ListRolesUseCase struct {
	roleListRepo RoleListRepository
}

func NewListRolesUseCase(roleListRepo RoleListRepository) *ListRolesUseCase {
	return &ListRolesUseCase{roleListRepo: roleListRepo}
}

func (uc *ListRolesUseCase) Execute(ctx context.Context, req dto.ListRolesRequest) (*dto.ListRolesResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}

	roles, total, err := uc.roleListRepo.List(ctx, req.RoleType, req.ScopeType, req.Assignable, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.RoleItem, 0, len(roles))
	for _, role := range roles {
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
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.Limit)))

	return &dto.ListRolesResponse{
		Roles: items,
		Meta: dto.Meta{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: int(total),
			TotalPages: totalPages,
		},
	}, nil
}
