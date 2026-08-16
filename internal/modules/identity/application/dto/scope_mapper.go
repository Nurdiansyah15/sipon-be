package dto

import scopesentity "sipon-be/internal/modules/identity/domain/scope/entity"

func ToMasterScopeItem(s *scopesentity.Scope) *MasterScopeItem {
	return &MasterScopeItem{
		ID:          s.ID,
		ScopeType:   string(s.ScopeType),
		Code:        s.Code,
		Name:        s.Name,
		Description: s.Description,
		IsActive:    s.IsActive,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}
