package dto

// CreateMidtransPaymentRequest dipakai santri untuk memulai pembayaran online
// atas sebuah invoice.
type CreateMidtransPaymentRequest struct {
	InvoiceID string `json:"invoice_id" binding:"required,uuid"`
}

// MidtransPaymentResponse dikembalikan setelah transaksi Snap berhasil dibuat.
// Frontend memakai SnapToken untuk membuka popup Snap atau melempar user ke
// RedirectURL.
type MidtransPaymentResponse struct {
	TransactionID string  `json:"transaction_id"`
	InvoiceID     string  `json:"invoice_id"`
	Amount        float64 `json:"amount"`
	SnapToken     string  `json:"snap_token"`
	RedirectURL   string  `json:"redirect_url"`
	Status        string  `json:"status"`
	ExpiresAt     string  `json:"expires_at"`
}

// MidtransWebhookNotification adalah subset payload notifikasi yang dikirim
// Midtrans ke endpoint webhook. Field signature_key dipakai untuk verifikasi.
type MidtransWebhookNotification struct {
	TransactionTime   string `json:"transaction_time"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	OrderID           string `json:"order_id"`
	TransactionID     string `json:"transaction_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
}

// PaymentGatewayStatusResponse adalah status transaksi pembayaran online
// sebuah invoice, dipakai frontend untuk polling setelah Snap ditutup.
type PaymentGatewayStatusResponse struct {
	TransactionID string  `json:"transaction_id"`
	InvoiceID     string  `json:"invoice_id"`
	PaymentID     *string `json:"payment_id,omitempty"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	PaymentMethod *string `json:"payment_method,omitempty"`
	SnapToken     string  `json:"snap_token,omitempty"`
	RedirectURL   string  `json:"redirect_url,omitempty"`
	ExpiresAt     string  `json:"expires_at"`
}
