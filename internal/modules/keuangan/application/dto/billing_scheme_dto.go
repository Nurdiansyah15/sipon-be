package dto

type CreateBillingSchemeRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type UpdateBillingSchemeRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
}

type BillingSchemeListQuery struct {
	Active *bool `form:"is_active"`
	Page   int   `form:"page"`
	Limit  int   `form:"limit"`
}

type BillingSchemeResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description *string                    `json:"description,omitempty"`
	IsActive    bool                       `json:"is_active"`
	Items       []BillingSchemeItemResponse `json:"items,omitempty"`
	CreatedAt   string                     `json:"created_at"`
	UpdatedAt   string                     `json:"updated_at"`
}

type BillingSchemeItemResponse struct {
	ID              string                    `json:"id"`
	FeeComponentID  string                    `json:"fee_component_id"`
	FeeComponent    *FeeComponentBriefResponse `json:"fee_component,omitempty"`
	AmountOverride  *float64                  `json:"amount_override,omitempty"`
	IsRequired      bool                      `json:"is_required"`
	SortOrder       int                       `json:"sort_order"`
}

type FeeComponentBriefResponse struct {
	ID     string  `json:"id"`
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}

type AddSchemeItemRequest struct {
	FeeComponentID  string   `json:"fee_component_id" binding:"required"`
	AmountOverride  *float64 `json:"amount_override,omitempty"`
	IsRequired      bool     `json:"is_required"`
	SortOrder       int      `json:"sort_order"`
}

type AssignSchemeRequest struct {
	SantriID        string  `json:"santri_id" binding:"required"`
	BillingSchemeID string  `json:"billing_scheme_id" binding:"required"`
	EffectiveFrom   string  `json:"effective_from" binding:"required"`
	EffectiveUntil  *string `json:"effective_until,omitempty"`
}
