package query

import (
	"context"

	"sipon-be/internal/modules/identity/application/dto"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
)

type ListPermissionsUseCase struct{}

func NewListPermissionsUseCase() *ListPermissionsUseCase {
	return &ListPermissionsUseCase{}
}

func (uc *ListPermissionsUseCase) Execute(ctx context.Context) []dto.PermissionItem {
	items := make([]dto.PermissionItem, 0, len(roleconstant.AllPermissionDefinitions))
	for _, def := range roleconstant.AllPermissionDefinitions {
		items = append(items, dto.PermissionItem{
			Key:         string(def.Key),
			DisplayName: def.DisplayName,
			Description: def.Description,
		})
	}
	return items
}
