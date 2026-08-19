package messaging

import (
	"context"
	"encoding/json"

	"sipon-be/internal/modules/messaging/application"
	"sipon-be/internal/modules/messaging/domain/message/valueobject"
	messagingerrors "sipon-be/internal/modules/messaging/domain/message_job/errors"
)

// Message adalah alias untuk envelope transport milik domain messaging. Modul
// lain memakai tipe ini lewat package root messaging, bukan lewat subpackage
// domain secara langsung.
type Message = valueobject.Message

// Binding memetakan satu queue ke satu routing key. Modul bisnis mengembalikan
// daftar Binding ini dari RegisterMessageHandlers.
type Binding = valueobject.Binding

// HandlerFunc adalah tipe handler yang didaftarkan lewat Contract.Register.
type HandlerFunc = application.HandlerFunc

// NewMessage membuat envelope Message baru dengan ID dan correlation ID yang
// di-generate otomatis.
func NewMessage(msgType string, payload json.RawMessage) (Message, error) {
	return valueobject.NewMessage(msgType, payload)
}

// NewFatalError membungkus err sebagai kegagalan permanen (tidak akan membaik
// dengan retry) — dipakai handler MQ modul bisnis untuk klasifikasi error.
func NewFatalError(err error) error {
	return messagingerrors.NewFatalError(err)
}

// NewRetryableError membungkus err sebagai kegagalan sementara yang layak
// di-retry dengan backoff.
func NewRetryableError(err error) error {
	return messagingerrors.NewRetryableError(err)
}

// IsFatal melaporkan apakah err diklasifikasikan sebagai fatal.
func IsFatal(err error) bool {
	return messagingerrors.IsFatal(err)
}

// IsRetryable melaporkan apakah err diklasifikasikan sebagai retryable.
func IsRetryable(err error) bool {
	return messagingerrors.IsRetryable(err)
}

// Contract adalah permukaan outward-facing modul messaging yang dikonsumsi oleh
// modul bisnis (untuk mendaftarkan handler asynchronous) dan proses worker
// (untuk menjalankan pipeline publish/consume). Dibuat konsisten dengan pola
// Contract pada modul bisnis (internal/modules/*).
type Contract interface {
	// Registry mengembalikan registry handler pesan.
	Registry() *application.Registry
	// Register mendaftarkan handler untuk sebuah routing key.
	Register(routingKey string, handler HandlerFunc) error
	// Dispatch mengirim message ke handler yang terdaftar (mode direct).
	Dispatch(ctx context.Context, msg Message) error
	// Publish mengirim message ke broker bila publisher telah dipasang.
	Publish(ctx context.Context, msg Message) error
}

var _ Contract = (*Module)(nil)
