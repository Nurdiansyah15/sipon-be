package dto

type CreateManualPaymentRequest struct {
	InvoiceID       string  `json:"invoice_id" binding:"required"`
	DebitAccountID  string  `json:"debit_account_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required"`
	Method          string  `json:"method" binding:"required"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	PaymentDate     string  `json:"payment_date" binding:"required"`
	Notes           *string `json:"notes,omitempty"`
	ProofKey        *string `json:"proof_key,omitempty"`
}

type PresignPaymentProofRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type PresignPaymentProofResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	ExpiresIn  int    `json:"expires_in"`
}

type SubmitPaymentRequest struct {
	InvoiceID       string  `json:"invoice_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Method          string  `json:"method" binding:"required,oneof=transfer"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	PaymentDate     string  `json:"payment_date" binding:"required"`
	ProofKey        string  `json:"proof_key" binding:"required"`
	Notes           *string `json:"notes,omitempty"`
}

type VerifyPaymentRequest struct {
	DebitAccountID string `json:"debit_account_id" binding:"required"`
}

type PaymentProofResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
}

type PaymentListQuery struct {
	InvoiceID *string `form:"invoice_id"`
	Status    *string `form:"status"`
	PeriodID  *string `form:"period_id"`
	Page      int     `form:"page"`
	Limit     int     `form:"limit"`
}

type PaymentResponse struct {
	ID              string                `json:"id"`
	PaymentNumber   string                `json:"payment_number"`
	InvoiceID       string                `json:"invoice_id"`
	Invoice         *InvoiceResponse      `json:"invoice,omitempty"`
	DebitAccountID  *string               `json:"debit_account_id,omitempty"`
	DebitAccount    *AccountBriefResponse `json:"debit_account,omitempty"`
	Amount          float64               `json:"amount"`
	Method          string                `json:"method"`
	ReferenceNumber *string               `json:"reference_number,omitempty"`
	PaymentDate     string                `json:"payment_date"`
	Status          string                `json:"status"`
	VerifiedBy      *string               `json:"verified_by,omitempty"`
	VerifiedAt      *string               `json:"verified_at,omitempty"`
	Notes           *string               `json:"notes,omitempty"`
	ProofKey        *string               `json:"proof_key,omitempty"`
	CreatedAt       string                `json:"created_at"`
	UpdatedAt       string                `json:"updated_at"`
}
