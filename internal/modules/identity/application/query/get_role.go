package query

import (
	"context"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
	"sipon-be/internal/shared/kernel"
)

type GetRoleUseCase struct {
	roleRepo     domain.RoleRepository
	rolePermRepo domain.RolePermissionRepository
}

func NewGetRoleUseCase(
	roleRepo domain.RoleRepository,
	rolePermRepo domain.RolePermissionRepository,
) *GetRoleUseCase {
	return &GetRoleUseCase{
		roleRepo:     roleRepo,
		rolePermRepo: rolePermRepo,
	}
}

func (uc *GetRoleUseCase) Execute(ctx context.Context, roleID string) (*dto.RoleItem, error) {
	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, kernel.Wrap(domain.ErrCodeRoleNotFound, err)
	}

	rps, err := uc.rolePermRepo.ListByRoleID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	permItems := make([]dto.PermissionItem, 0, len(rps))
	for _, rp := range rps {
		for _, def := range domain.AllPermissionDefinitions {
			if def.Key == rp.PermissionKey {
				permItems = append(permItems, dto.PermissionItem{
					Key:         string(def.Key),
					DisplayName: def.DisplayName,
					Description: def.Description,
				})
				break
			}
		}
	}

	_ = permItems

	return &dto.RoleItem{
		ID:          role.ID,
		Name:        string(role.Name),
		DisplayName: role.DisplayName,
		Description: role.Description,
		RoleType:    string(role.RoleType),
		ScopeType:   string(role.ScopeType),
		Assignable:  role.Assignable,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}, nil
}
