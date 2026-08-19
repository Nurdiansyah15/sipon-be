package dto

import "time"

// NotificationItem adalah representasi satu notifikasi untuk client.
type NotificationItem struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"`
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	ImageURL    string            `json:"image_url,omitempty"`
	Module      string            `json:"module"`
	EventType   string            `json:"event_type"`
	EntityID    string            `json:"entity_id"`
	ClickAction string            `json:"click_action"`
	Bypass      bool              `json:"bypass"`
	Extra       map[string]string `json:"extra,omitempty"`
	IsRead      bool              `json:"is_read"`
	CreatedAt   string            `json:"created_at"`
	ReadAt      *string           `json:"read_at,omitempty"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

type MarkAllReadResponse struct {
	Marked int `json:"marked"`
}

type ListNotificationsQuery struct {
	UnreadOnly bool `form:"unread_only"`
	Page       int  `form:"page"`
	Limit      int  `form:"limit"`
}

// --- Preferences ---

type NotificationPreferenceResponse struct {
	ID                      string  `json:"id"`
	UserID                  string  `json:"user_id"`
	AllNotificationsEnabled bool    `json:"all_notifications_enabled"`
	DoNotDisturbEnabled     bool    `json:"do_not_disturb_enabled"`
	DoNotDisturbStartTime   *string `json:"do_not_disturb_start_time,omitempty"`
	DoNotDisturbEndTime     *string `json:"do_not_disturb_end_time,omitempty"`
	CreatedAt               string  `json:"created_at"`
	UpdatedAt               string  `json:"updated_at"`
}

type UpdatePreferenceRequest struct {
	AllNotificationsEnabled *bool   `json:"all_notifications_enabled"`
	DoNotDisturbEnabled     *bool   `json:"do_not_disturb_enabled"`
	DoNotDisturbStartTime   *string `json:"do_not_disturb_start_time"`
	DoNotDisturbEndTime     *string `json:"do_not_disturb_end_time"`
}

// --- Broadcast ---

type BroadcastRequest struct {
	Type     string   `json:"type" binding:"required"`
	Title    string   `json:"title" binding:"required"`
	Body     string   `json:"body" binding:"required"`
	Channels []string `json:"channels"`
}

func FormatRFC3339(t time.Time) string {
	return t.Format("2006-01-02T15:04:05Z07:00")
}
