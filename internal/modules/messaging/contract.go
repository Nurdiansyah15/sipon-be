package messaging

import (
	"context"

	"sipon-be/internal/modules/messaging/application/ports"
)

// Contract adalah permukaan outward-facing modul messaging yang dikonsumsi oleh
// modul bisnis (untuk mendaftarkan handler asynchronous) dan proses worker
// (untuk menjalankan pipeline publish/consume). Dibuat konsisten dengan pola
// Contract pada modul bisnis (internal/modules/*).
type Contract interface {
	// Registry mengembalikan registry handler pesan.
	Registry() *ports.Registry
	// Register mendaftarkan handler untuk sebuah routing key.
	Register(routingKey string, handler ports.HandlerFunc) error
	// Dispatch mengirim message ke handler yang terdaftar (mode direct).
	Dispatch(ctx context.Context, msg ports.Message) error
	// Publish mengirim message ke broker bila publisher telah dipasang.
	Publish(ctx context.Context, msg ports.Message) error
}

var _ Contract = (*Module)(nil)
