package constant

import "sipon-be/internal/shared/kernel"

const (
	CodePaymentNotFound         kernel.Code = "PAYMENT_NOT_FOUND"
	CodePaymentInvalidStatus    kernel.Code = "PAYMENT_INVALID_STATUS"
	CodePaymentAlreadyVerified  kernel.Code = "PAYMENT_ALREADY_VERIFIED"
	CodePaymentPersistenceFailed kernel.Code = "PAYMENT_PERSISTENCE_FAILED"
	CodePaymentQueryFailed      kernel.Code = "PAYMENT_QUERY_FAILED"
)

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentVerified PaymentStatus = "verified"
	PaymentRejected PaymentStatus = "rejected"
)

type PaymentMethod string

const (
	MethodTransfer PaymentMethod = "transfer"
	MethodCash     PaymentMethod = "cash"
	MethodCheck    PaymentMethod = "check"
)
