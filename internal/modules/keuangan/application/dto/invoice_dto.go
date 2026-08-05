package dto

type CreateInvoiceRequest struct {
	SantriID       string  `json:"santri_id" binding:"required"`
	FeeComponentID string  `json:"fee_component_id" binding:"required"`
	Periode        string  `json:"periode" binding:"required"`
	TahunAjaran    string  `json:"tahun_ajaran" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	DueDate        string  `json:"due_date" binding:"required"`
	Notes          *string `json:"notes,omitempty"`
}

type CreateInvoiceBatchRequest struct {
	BillingSchemeID string `json:"billing_scheme_id" binding:"required"`
	Periode         string `json:"periode" binding:"required"`
	TahunAjaran     string `json:"tahun_ajaran" binding:"required"`
	DueDate         string `json:"due_date" binding:"required"`
}

type InvoiceListQuery struct {
	SantriID    *string `form:"santri_id"`
	UserID      *string `form:"user_id"`
	Status      *string `form:"status"`
	Periode     *string `form:"periode"`
	TahunAjaran *string `form:"tahun_ajaran"`
	Page        int     `form:"page"`
	Limit       int     `form:"limit"`
}

type InvoiceResponse struct {
	ID              string                     `json:"id"`
	InvoiceNumber   string                     `json:"invoice_number"`
	SantriID        string                     `json:"santri_id"`
	UserID          string                     `json:"user_id"`
	BillingSchemeID *string                    `json:"billing_scheme_id,omitempty"`
	FeeComponentID  string                     `json:"fee_component_id"`
	FeeComponent    *FeeComponentBriefResponse `json:"fee_component,omitempty"`
	Periode         string                     `json:"periode"`
	TahunAjaran     string                     `json:"tahun_ajaran"`
	Amount          float64                    `json:"amount"`
	DiscountAmount  float64                    `json:"discount_amount"`
	PaidAmount      float64                    `json:"paid_amount"`
	Status          string                     `json:"status"`
	DueDate         string                     `json:"due_date"`
	IssuedAt        *string                    `json:"issued_at,omitempty"`
	Notes           *string                    `json:"notes,omitempty"`
	CreatedAt       string                     `json:"created_at"`
	UpdatedAt       string                     `json:"updated_at"`
	Payments        []PaymentResponse            `json:"payments,omitempty"`
	Adjustments     []InvoiceAdjustmentResponse   `json:"adjustments,omitempty"`
}

type ApplyAdjustmentRequest struct {
	Type        string   `json:"type" binding:"required"`
	Amount      float64  `json:"amount"`
	Percentage  *float64 `json:"percentage,omitempty"`
	Description *string  `json:"description,omitempty"`
}

type InvoiceAdjustmentResponse struct {
	ID          string   `json:"id"`
	InvoiceID   string   `json:"invoice_id"`
	Type        string   `json:"type"`
	Amount      float64  `json:"amount"`
	Percentage  *float64 `json:"percentage,omitempty"`
	Description *string  `json:"description,omitempty"`
	AppliedBy   string   `json:"applied_by"`
	AppliedAt   string   `json:"applied_at"`
}
