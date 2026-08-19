package entity

import (
	"strings"
	"time"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
"sipon-be/internal/shared/kernel"
)

const maxRetryCount = 3

// DeliveryAttempt adalah notifikasi nyata milik seorang penerima pada channel tertentu.
// Setiap DeliveryAttempt = 1 user × 1 channel dari sebuah Notification blueprint.
type DeliveryAttempt struct {
	ID             string
	NotificationID string
	UserID         string
	Channel        notifconstant.NotificationChannel
	Status         notifconstant.DeliveryStatus
	ProviderCode   *string
	RetryCount     int
	NextRetryAt    *time.Time
	AttemptedAt    time.Time
	ReadAt         *time.Time
}

func NewDeliveryAttempt(id, notificationID, userID string, channel notifconstant.NotificationChannel) (*DeliveryAttempt, error) {
	if strings.TrimSpace(id) == "" {
		return nil, kernel.New(notifconstant.CodeDeliveryAttemptIDRequired)
	}
	if strings.TrimSpace(userID) == "" {
		return nil, kernel.New(notifconstant.CodeDeliveryAttemptUserIDRequired)
	}
	if !channel.IsValid() {
		return nil, kernel.New(notifconstant.CodeDeliveryAttemptChannelInvalid)
	}

	return &DeliveryAttempt{
		ID:             strings.TrimSpace(id),
		NotificationID: strings.TrimSpace(notificationID),
		UserID:         strings.TrimSpace(userID),
		Channel:        channel,
		Status:         notifconstant.DeliveryStatusPending,
		RetryCount:     0,
		AttemptedAt:    time.Now(),
	}, nil
}

func (da *DeliveryAttempt) MarkSuccess() {
	da.Status = notifconstant.DeliveryStatusSuccess
	da.ProviderCode = nil
}

func (da *DeliveryAttempt) MarkFailed(providerCode string) {
	da.Status = notifconstant.DeliveryStatusFailed
	if strings.TrimSpace(providerCode) != "" {
		code := strings.TrimSpace(providerCode)
		da.ProviderCode = &code
	}
}

func (da *DeliveryAttempt) ScheduleRetry(at time.Time) {
	da.Status = notifconstant.DeliveryStatusRetrying
	da.NextRetryAt = &at
	da.RetryCount++
}

func (da *DeliveryAttempt) IsRetryable() bool {
	return da.RetryCount < maxRetryCount
}

func (da *DeliveryAttempt) MarkRead() {
	now := time.Now()
	da.ReadAt = &now
}

func (da *DeliveryAttempt) IsUnread() bool {
	return da.ReadAt == nil
}
