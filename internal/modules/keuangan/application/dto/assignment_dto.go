package dto

type AssignmentResponse struct {
	ID              string                    `json:"id"`
	SantriID        string                    `json:"santri_id"`
	BillingSchemeID string                    `json:"billing_scheme_id"`
	BillingScheme   *BillingSchemeBriefResponse `json:"billing_scheme,omitempty"`
	EffectiveFrom   string                    `json:"effective_from"`
	EffectiveUntil  *string                   `json:"effective_until,omitempty"`
	AssignedBy      string                    `json:"assigned_by"`
	CreatedAt       string                    `json:"created_at"`
}

type AssignmentListQuery struct {
	SantriID *string `form:"santri_id"`
}

type UpdateAssignmentRequest struct {
	BillingSchemeID string  `json:"billing_scheme_id" binding:"required"`
	EffectiveFrom   string  `json:"effective_from" binding:"required"`
	EffectiveUntil  *string `json:"effective_until,omitempty"`
}
