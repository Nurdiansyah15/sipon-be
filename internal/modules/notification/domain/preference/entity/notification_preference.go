package entity

import (
	"strings"
	"time"

	"sipon-be/internal/shared/kernel"
)

// NotificationPreference adalah aggregate root untuk setting notifikasi per user.
// Disimpan di tabel notification_preferences.
type NotificationPreference struct {
	ID                   string
	UserID               string
	AllEnabled           bool
	DoNotDisturbEnabled  bool
	DNDStartTime         *string
	DNDEndTime           *string
	ModulePreferences    map[string]any // JSONB — fase 2
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewNotificationPreference membuat preferensi baru dengan semua nilai default.
func NewNotificationPreference(id, userID string) *NotificationPreference {
	now := time.Now()
	return &NotificationPreference{
		ID:                strings.TrimSpace(id),
		UserID:            strings.TrimSpace(userID),
		AllEnabled:        true,
		DoNotDisturbEnabled: false,
		ModulePreferences: map[string]any{},
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// Validate memvalidasi field wajib.
func (p *NotificationPreference) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return kernel.New(kernel.Code("PREFERENCE_ID_REQUIRED"))
	}
	if strings.TrimSpace(p.UserID) == "" {
		return kernel.New(kernel.Code("PREFERENCE_USER_ID_REQUIRED"))
	}
	return nil
}
