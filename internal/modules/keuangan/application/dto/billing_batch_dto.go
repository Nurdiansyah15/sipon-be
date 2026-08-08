package dto

type CreateInvoiceBatchResponse struct {
	BatchID string `json:"batch_id"`
	Status  string `json:"status"`
}

type BillingBatchListQuery struct {
	Status *string `form:"status"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type BillingBatchResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	BillingSchemeID string  `json:"billing_scheme_id"`
	BillingPeriodID string  `json:"billing_period_id"`
	Status          string  `json:"status"`
	CreatedBy       string  `json:"created_by"`
	CreatedAt       string  `json:"created_at"`
	CompletedAt     *string `json:"completed_at,omitempty"`
	TotalCreated    int     `json:"total_created"`
	TotalSkipped    int     `json:"total_skipped"`
	TotalError      int     `json:"total_error"`
}

type BillingBatchTargetResponse struct {
	ID          string  `json:"id"`
	SantriID    string  `json:"santri_id"`
	Status      string  `json:"status"`
	InvoiceID   *string `json:"invoice_id,omitempty"`
	Reason      *string `json:"reason,omitempty"`
	ProcessedAt *string `json:"processed_at,omitempty"`
}

type BillingBatchDetailResponse struct {
	BillingBatchResponse
	Targets []BillingBatchTargetResponse `json:"targets"`
}
