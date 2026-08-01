package query

import (
	"context"

	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/domain"
)

type ListPermissionsUseCase struct{}

func NewListPermissionsUseCase() *ListPermissionsUseCase {
	return &ListPermissionsUseCase{}
}

func (uc *ListPermissionsUseCase) Execute(ctx context.Context) []dto.PermissionItem {
	items := make([]dto.PermissionItem, 0, len(domain.AllPermissionDefinitions))
	for _, def := range domain.AllPermissionDefinitions {
		items = append(items, dto.PermissionItem{
			Key:         string(def.Key),
			DisplayName: def.DisplayName,
			Description: def.Description,
		})
	}
	return items
}
