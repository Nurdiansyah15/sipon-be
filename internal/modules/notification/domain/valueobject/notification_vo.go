package valueobject

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NotificationPayload membawa data kontekstual tambahan pada sebuah notifikasi. Immutable.
type NotificationPayload struct {
	Module    string            `json:"module"`
	EventType string            `json:"event_type"`
	EntityID  string            `json:"entity_id"`
	ClickAction string          `json:"click_action"`
	ImageURL  string            `json:"image_url"`
	Bypass    bool              `json:"bypass"`
	Extra     map[string]string `json:"extra,omitempty"`
}

// DoNotDisturbWindow menyimpan rentang waktu DND dalam format "HH:MM". Immutable.
type DoNotDisturbWindow struct {
	StartTime string
	EndTime   string
}

// NewDoNotDisturbWindow membuat DoNotDisturbWindow baru dengan validasi format "HH:MM".
func NewDoNotDisturbWindow(start, end string) (DoNotDisturbWindow, error) {
	if err := validateTimeFormat(start); err != nil {
		return DoNotDisturbWindow{}, fmt.Errorf("dnd_start_time: %w", err)
	}
	if err := validateTimeFormat(end); err != nil {
		return DoNotDisturbWindow{}, fmt.Errorf("dnd_end_time: %w", err)
	}
	return DoNotDisturbWindow{StartTime: start, EndTime: end}, nil
}

// IsActive mengecek apakah waktu sekarang berada dalam rentang DND.
// Mendukung rentang yang melewati tengah malam (contoh: 22:00 - 07:00).
func (w DoNotDisturbWindow) IsActive(now time.Time) bool {
	startH, startM := parseTime(w.StartTime)
	endH, endM := parseTime(w.EndTime)

	nowMinutes := now.Hour()*60 + now.Minute()
	startMinutes := startH*60 + startM
	endMinutes := endH*60 + endM

	if startMinutes <= endMinutes {
		return nowMinutes >= startMinutes && nowMinutes < endMinutes
	}
	// Melewati tengah malam
	return nowMinutes >= startMinutes || nowMinutes < endMinutes
}

func validateTimeFormat(t string) error {
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return fmt.Errorf("format harus HH:MM")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return fmt.Errorf("jam tidak valid")
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return fmt.Errorf("menit tidak valid")
	}
	return nil
}

func parseTime(t string) (int, int) {
	parts := strings.Split(t, ":")
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h, m
}
