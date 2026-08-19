package entity

import (
	"strings"
	"time"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
notifvo "sipon-be/internal/modules/notification/domain/valueobject"
"sipon-be/internal/shared/kernel"
)

// Notification adalah blueprint notifikasi sebelum didistribusikan ke penerima.
// AudienceType menentukan strategi fanout (unicast/multicast/broadcast).
// Delivery state per user per channel ada di DeliveryAttempt.
type Notification struct {
	ID            string
	AudienceType  notifconstant.AudienceType
	AudienceData  map[string]any
	Type          notifconstant.NotificationType
	Title         string
	Body          string
	Payload       notifvo.NotificationPayload
	ReferenceID   *string
	ReferenceType *string
	CreatedAt     time.Time
}

type NotificationParams struct {
	ID            string
	AudienceType  notifconstant.AudienceType
	AudienceData  map[string]any
	Type          notifconstant.NotificationType
	Title         string
	Body          string
	Payload       notifvo.NotificationPayload
	ReferenceID   *string
	ReferenceType *string
}

func NewNotification(params NotificationParams) (*Notification, error) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, kernel.New(notifconstant.CodeNotificationIDRequired)
	}
	if !params.AudienceType.IsValid() {
		if params.AudienceType == "" {
			return nil, kernel.New(notifconstant.CodeNotificationAudienceTypeRequired)
		}
		return nil, kernel.New(notifconstant.CodeNotificationAudienceTypeInvalid)
	}
	if strings.TrimSpace(params.Title) == "" {
		return nil, kernel.New(notifconstant.CodeNotificationTitleRequired)
	}
	if strings.TrimSpace(params.Body) == "" {
		return nil, kernel.New(notifconstant.CodeNotificationBodyRequired)
	}
	if !params.Type.IsValid() {
		return nil, kernel.New(notifconstant.CodeNotificationTypeInvalid)
	}

	audienceData := params.AudienceData
	if audienceData == nil {
		audienceData = map[string]any{}
	}

	return &Notification{
		ID:            strings.TrimSpace(params.ID),
		AudienceType:  params.AudienceType,
		AudienceData:  audienceData,
		Type:          params.Type,
		Title:         strings.TrimSpace(params.Title),
		Body:          strings.TrimSpace(params.Body),
		Payload:       params.Payload,
		ReferenceID:   normalizeOptionalText(params.ReferenceID),
		ReferenceType: normalizeOptionalText(params.ReferenceType),
		CreatedAt:     time.Now(),
	}, nil
}

func normalizeOptionalText(v *string) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
