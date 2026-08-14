package dto

import scopesentity "sipon-be/internal/modules/system/domain/scope/entity"

func ToScopeItem(s *scopesentity.Scope) *ScopeItem {
	return &ScopeItem{
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
