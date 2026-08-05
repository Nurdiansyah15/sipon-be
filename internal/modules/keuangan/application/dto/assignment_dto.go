package dto

type AssignmentResponse struct {
	ID              string  `json:"id"`
	SantriID        string  `json:"santri_id"`
	BillingSchemeID string  `json:"billing_scheme_id"`
	EffectiveFrom   string  `json:"effective_from"`
	EffectiveUntil  *string `json:"effective_until,omitempty"`
	AssignedBy      string  `json:"assigned_by"`
	CreatedAt       string  `json:"created_at"`
}
