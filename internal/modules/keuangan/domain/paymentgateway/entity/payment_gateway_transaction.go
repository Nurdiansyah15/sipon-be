package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/paymentgateway/constant"
	"sipon-be/internal/shared/kernel"
)

// PaymentGatewayTransaction merepresentasikan satu percobaan pembayaran
// online melalui Midtrans Snap terhadap sebuah invoice. Entity ini berdiri
// sendiri dari Payment (pembayaran) agar status transaksi gateway tidak
// tercampur dengan status pembayaran internal yang di-review admin.
type PaymentGatewayTransaction struct {
	ID              string
	TransactionID   string
	InvoiceID       string
	PaymentID       *string
	Amount          float64
	Status          constant.PaymentGatewayStatus
	PaymentMethod   *string
	SnapToken       string
	RedirectURL     string
	RawNotification []byte
	Metadata        []byte
	ExpiredAt       time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewPaymentGatewayTransaction(
	id, transactionID, invoiceID string,
	amount float64,
	snapToken, redirectURL string,
	metadata []byte,
	expiredAt time.Time,
) (*PaymentGatewayTransaction, error) {
	if id == "" || transactionID == "" || invoiceID == "" {
		return nil, kernel.WrapMsg(constant.CodePaymentGatewayNotFound, "Data transaksi payment gateway tidak lengkap", nil)
	}
	if amount <= 0 {
		return nil, kernel.WrapMsg(constant.CodePaymentGatewayInvalidStatus, "Nominal transaksi harus lebih dari nol", nil)
	}
	if snapToken == "" {
		return nil, kernel.WrapMsg(constant.CodePaymentGatewayNotFound, "Snap token tidak tersedia", nil)
	}
	now := time.Now()
	return &PaymentGatewayTransaction{
		ID:            id,
		TransactionID: transactionID,
		InvoiceID:     invoiceID,
		Amount:        amount,
		Status:        constant.GatewayStatusPending,
		SnapToken:     snapToken,
		RedirectURL:   redirectURL,
		Metadata:      metadata,
		ExpiredAt:     expiredAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// ApplyNotification memperbarui status transaksi berdasarkan notifikasi
// webhook Midtrans. Metode ini idempotent: status sukses tidak pernah
// diregresi, dan status final tidak berubah walaupun webhook datang berulang.
func (t *PaymentGatewayTransaction) ApplyNotification(status constant.PaymentGatewayStatus, paymentMethod *string, raw []byte) error {
	if !constant.ValidStatus[status] {
		return kernel.WrapMsg(constant.CodePaymentGatewayInvalidStatus, "Status notifikasi tidak dikenal", nil)
	}

	// Transaksi yang sudah sukses tidak boleh berubah statusnya lagi.
	if t.Status.IsSuccess() && !status.IsSuccess() {
		return nil
	}
	// Transaksi yang sudah berstatus final tidak boleh berubah lagi.
	if t.Status.IsFinal() && t.Status != status {
		return nil
	}

	t.Status = status
	if paymentMethod != nil && *paymentMethod != "" {
		t.PaymentMethod = paymentMethod
	}
	if raw != nil {
		t.RawNotification = raw
	}
	t.UpdatedAt = time.Now()
	return nil
}

// LinkPayment menautkan Payment internal yang terverifikasi dengan transaksi
// gateway. Hanya dapat dilakukan sekali.
func (t *PaymentGatewayTransaction) LinkPayment(paymentID string) error {
	if t.PaymentID != nil {
		return kernel.WrapMsg(constant.CodePaymentGatewayInvalidStatus, "Transaksi sudah tertaut dengan pembayaran", nil)
	}
	t.PaymentID = &paymentID
	t.UpdatedAt = time.Now()
	return nil
}

// MarkRejected memindahkan transaksi ke status gagal secara manual (misalnya
// saat webhook tidak pernah datang dan admin memutuskan menolak).
func (t *PaymentGatewayTransaction) MarkRejected() error {
	if t.Status.IsSuccess() || t.Status.IsFinal() {
		return nil
	}
	t.Status = constant.GatewayStatusCancel
	t.UpdatedAt = time.Now()
	return nil
}
