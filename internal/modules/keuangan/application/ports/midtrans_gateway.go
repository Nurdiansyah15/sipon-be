package ports

import "context"

// SnapItem adalah detail item yang dikirim ke Midtrans untuk ditampilkan di
// halaman pembayaran Snap.
type SnapItem struct {
	ID       string
	Price    float64
	Quantity int
	Name     string
}

// SnapTransactionRequest adalah payload pembuatan transaksi Snap.
type SnapTransactionRequest struct {
	OrderID       string
	GrossAmount   float64
	CustomerName  string
	CustomerEmail string
	CustomerPhone string
	ExpiryMinutes int
	Items         []SnapItem
}

// SnapTransactionResponse berisi token & redirect URL hasil pembuatan
// transaksi Snap. Token dipakai frontend untuk membuka popup Snap, redirect
// URL untuk mode pembayaran via halaman (non-popup).
type SnapTransactionResponse struct {
	Token       string
	RedirectURL string
}

// MidtransGateway adalah outbound port untuk integrasi API Midtrans (Snap &
// verifikasi signature webhook). Implementasi berada di infrastructure/external.
type MidtransGateway interface {
	CreateSnapTransaction(ctx context.Context, req SnapTransactionRequest) (*SnapTransactionResponse, error)
	VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool
}
