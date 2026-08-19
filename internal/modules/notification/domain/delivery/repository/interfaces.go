package repository

import (
	"context"
	"time"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	"sipon-be/internal/modules/notification/domain/delivery/entity"
)

// InboxReadItem adalah read model untuk satu item inbox notifikasi.
type InboxReadItem struct {
	DeliveryAttemptID string
	Type              string
	Title             string
	Body              string
	ImageURL          string
	Module            string
	EventType         string
	EntityID          string
	ClickAction       string
	Bypass            bool
	Extra             map[string]string
	ReferenceID       *string
	ReferenceType     *string
	AttemptedAt       time.Time
	ReadAt            *time.Time
}

// ListInAppQuery adalah input query untuk listing inbox notifikasi.
type ListInAppQuery struct {
	UserID     string
	UnreadOnly bool
	Page       int
	Limit      int
}

// Meta adalah metadata pagination (format standar sipon-be).
type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

// DeliveryAttemptRepository menyimpan dan mengambil entity DeliveryAttempt.
type DeliveryAttemptRepository interface {
	Save(ctx context.Context, da *entity.DeliveryAttempt) error
	FindByID(ctx context.Context, id string) (*entity.DeliveryAttempt, error)

	// Inbox queries — hanya untuk channel in_app.
	ListInApp(ctx context.Context, q ListInAppQuery) ([]InboxReadItem, Meta, error)
	CountUnreadInApp(ctx context.Context, userID string) (int64, error)
	MarkRead(ctx context.Context, id, userID string) error
	MarkAllRead(ctx context.Context, userID string) (int, error)

	// FindPendingByChannel digunakan fanout relay untuk memproses batch pending.
	FindPendingByChannel(ctx context.Context, channel notifconstant.NotificationChannel) ([]*entity.DeliveryAttempt, error)
}
