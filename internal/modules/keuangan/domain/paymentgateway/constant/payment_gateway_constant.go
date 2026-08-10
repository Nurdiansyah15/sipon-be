package constant

import "sipon-be/internal/shared/kernel"

const (
	CodePaymentGatewayNotFound          kernel.Code = "PAYMENT_GATEWAY_NOT_FOUND"
	CodePaymentGatewayInvalidStatus     kernel.Code = "PAYMENT_GATEWAY_INVALID_STATUS"
	CodePaymentGatewayAlreadyPending    kernel.Code = "PAYMENT_GATEWAY_ALREADY_PENDING"
	CodePaymentGatewaySignatureInvalid  kernel.Code = "PAYMENT_GATEWAY_SIGNATURE_INVALID"
	CodePaymentGatewayPersistenceFailed kernel.Code = "PAYMENT_GATEWAY_PERSISTENCE_FAILED"
	CodePaymentGatewayQueryFailed       kernel.Code = "PAYMENT_GATEWAY_QUERY_FAILED"
	CodePaymentGatewayAPIFailed         kernel.Code = "PAYMENT_GATEWAY_API_FAILED"
)

// PaymentGatewayStatus mencerminkan status transaksi dari Midtrans Snap,
// diperkaya dengan "pending" sebagai status awal sebelum pembayaran dilakukan.
type PaymentGatewayStatus string

const (
	GatewayStatusPending          PaymentGatewayStatus = "pending"
	GatewayStatusPendingChallenge PaymentGatewayStatus = "pending_challenge"
	GatewayStatusCapture          PaymentGatewayStatus = "capture"
	GatewayStatusSettlement       PaymentGatewayStatus = "settlement"
	GatewayStatusDeny             PaymentGatewayStatus = "deny"
	GatewayStatusFailure          PaymentGatewayStatus = "failure"
	GatewayStatusExpire           PaymentGatewayStatus = "expire"
	GatewayStatusCancel           PaymentGatewayStatus = "cancel"
)

// IsFinal mengembalikan true bila status sudah terminal dan tidak boleh
// berubah lagi (terlepas dari kedatangan notifikasi webhook ganda).
func (s PaymentGatewayStatus) IsFinal() bool {
	switch s {
	case GatewayStatusDeny, GatewayStatusFailure, GatewayStatusExpire, GatewayStatusCancel:
		return true
	default:
		return false
	}
}

// IsSuccess mengembalikan true bila dana sudah benar-benar diterima
// (settlement / capture non-challenge).
func (s PaymentGatewayStatus) IsSuccess() bool {
	return s == GatewayStatusSettlement || s == GatewayStatusCapture
}

// ValidStatus menyediakan whitelist status untuk validasi input & database.
var ValidStatus = map[PaymentGatewayStatus]bool{
	GatewayStatusPending:          true,
	GatewayStatusPendingChallenge: true,
	GatewayStatusCapture:          true,
	GatewayStatusSettlement:       true,
	GatewayStatusDeny:             true,
	GatewayStatusFailure:          true,
	GatewayStatusExpire:           true,
	GatewayStatusCancel:           true,
}
