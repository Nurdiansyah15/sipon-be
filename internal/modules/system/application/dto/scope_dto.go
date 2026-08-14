package dto

import "time"

type ScopeItem struct {
	ID          string    `json:"id"`
	ScopeType   string    `json:"scope_type"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateScopeRequest struct {
	ScopeType   string  `json:"scope_type" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type UpdateScopeRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

type ListScopesRequest struct {
	ScopeType       string `form:"scope_type"`
	IncludeInactive bool   `form:"include_inactive"`
}

type UserScopeAccessResponse struct {
	UserID        string   `json:"user_id"`
	ScopeType     string   `json:"scope_type"`
	HasAccess     bool     `json:"has_access"`
	HasFullAccess bool     `json:"has_full_access"`
	AllowedCodes  []string `json:"allowed_codes"`
}
