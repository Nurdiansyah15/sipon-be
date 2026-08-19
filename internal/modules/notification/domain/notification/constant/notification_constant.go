package constant

import "sipon-be/internal/shared/kernel"

// NotificationType adalah kategori notifikasi.
type NotificationType string

const (
	NotificationTypeSystem   NotificationType = "system"
	NotificationTypeSocial   NotificationType = "social"
	NotificationTypeContent  NotificationType = "content"
	NotificationTypeReminder NotificationType = "reminder"
	NotificationTypeSecurity NotificationType = "security"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case NotificationTypeSystem, NotificationTypeSocial, NotificationTypeContent,
		NotificationTypeReminder, NotificationTypeSecurity:
		return true
	}
	return false
}

// AudienceType menentukan strategi distribusi notifikasi.
type AudienceType string

const (
	AudienceTypeUnicast   AudienceType = "unicast"
	AudienceTypeMulticast AudienceType = "multicast"
	AudienceTypeBroadcast AudienceType = "broadcast"
)

func (a AudienceType) IsValid() bool {
	switch a {
	case AudienceTypeUnicast, AudienceTypeMulticast, AudienceTypeBroadcast:
		return true
	}
	return false
}

// NotificationChannel adalah saluran pengiriman notifikasi.
type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelEmail NotificationChannel = "email"
	NotificationChannelSMS   NotificationChannel = "sms"
	NotificationChannelPush  NotificationChannel = "push"
)

func (c NotificationChannel) IsValid() bool {
	switch c {
	case NotificationChannelInApp, NotificationChannelEmail, NotificationChannelSMS, NotificationChannelPush:
		return true
	}
	return false
}

// DeliveryStatus adalah status pengiriman per user per channel.
type DeliveryStatus string

const (
	DeliveryStatusPending  DeliveryStatus = "pending"
	DeliveryStatusSuccess  DeliveryStatus = "success"
	DeliveryStatusFailed   DeliveryStatus = "failed"
	DeliveryStatusRetrying DeliveryStatus = "retrying"
)

func (s DeliveryStatus) IsValid() bool {
	switch s {
	case DeliveryStatusPending, DeliveryStatusSuccess, DeliveryStatusFailed, DeliveryStatusRetrying:
		return true
	}
	return false
}

// Domain error codes.
const (
	CodeNotificationIDRequired           kernel.Code = "NOTIFICATION_ID_REQUIRED"
	CodeNotificationAudienceTypeRequired kernel.Code = "NOTIFICATION_AUDIENCE_TYPE_REQUIRED"
	CodeNotificationAudienceTypeInvalid  kernel.Code = "NOTIFICATION_AUDIENCE_TYPE_INVALID"
	CodeNotificationTitleRequired        kernel.Code = "NOTIFICATION_TITLE_REQUIRED"
	CodeNotificationBodyRequired         kernel.Code = "NOTIFICATION_BODY_REQUIRED"
	CodeNotificationTypeInvalid          kernel.Code = "NOTIFICATION_TYPE_INVALID"
	CodeNotificationNotFound             kernel.Code = "NOTIFICATION_NOT_FOUND"
	CodeNotificationPersistenceFailed    kernel.Code = "NOTIFICATION_PERSISTENCE_FAILED"
	CodeNotificationQueryFailed          kernel.Code = "NOTIFICATION_QUERY_FAILED"

	CodeDeliveryAttemptIDRequired        kernel.Code = "DELIVERY_ID_REQUIRED"
	CodeDeliveryAttemptUserIDRequired    kernel.Code = "DELIVERY_USER_ID_REQUIRED"
	CodeDeliveryAttemptChannelInvalid    kernel.Code = "DELIVERY_CHANNEL_INVALID"
	CodeDeliveryAttemptNotFound          kernel.Code = "DELIVERY_NOT_FOUND"
	CodeDeliveryAttemptPersistenceFailed kernel.Code = "DELIVERY_PERSISTENCE_FAILED"

	CodePreferenceUserIDRequired    kernel.Code = "PREFERENCE_USER_ID_REQUIRED"
	CodePreferenceNotFound          kernel.Code = "PREFERENCE_NOT_FOUND"
	CodePreferencePersistenceFailed kernel.Code = "PREFERENCE_PERSISTENCE_FAILED"

	CodeDNDInvalidTimeFormat kernel.Code = "DND_INVALID_TIME_FORMAT"
)
