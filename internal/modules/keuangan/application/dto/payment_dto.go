package dto

type CreateManualPaymentRequest struct {
	InvoiceID       string  `json:"invoice_id" binding:"required"`
	DebitAccountID  *string `json:"debit_account_id,omitempty"`
	Amount          float64 `json:"amount" binding:"required"`
	Method          string  `json:"method" binding:"required"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	PaymentDate     string  `json:"payment_date" binding:"required"`
	Notes           *string `json:"notes,omitempty"`
	ProofKey        *string `json:"proof_key,omitempty"`
}

type PaymentListQuery struct {
	InvoiceID *string `form:"invoice_id"`
	Status    *string `form:"status"`
	Page      int     `form:"page"`
	Limit     int     `form:"limit"`
}

type PaymentResponse struct {
	ID              string  `json:"id"`
	PaymentNumber   string  `json:"payment_number"`
	InvoiceID       string  `json:"invoice_id"`
	DebitAccountID  *string `json:"debit_account_id,omitempty"`
	Amount          float64 `json:"amount"`
	Method          string  `json:"method"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	PaymentDate     string  `json:"payment_date"`
	Status          string  `json:"status"`
	VerifiedBy      *string `json:"verified_by,omitempty"`
	VerifiedAt      *string `json:"verified_at,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	ProofKey        *string `json:"proof_key,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}
