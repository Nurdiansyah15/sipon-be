package application

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	notifentity "sipon-be/internal/modules/notification/domain/notification/entity"
	notifRepo "sipon-be/internal/modules/notification/domain/notification/repository"
	deliveryEntity "sipon-be/internal/modules/notification/domain/delivery/entity"
	deliveryRepo "sipon-be/internal/modules/notification/domain/delivery/repository"
	prefRepo "sipon-be/internal/modules/notification/domain/preference/repository"
notifvo "sipon-be/internal/modules/notification/domain/valueobject"
)

// TargetMode menentukan mode pengiriman notifikasi.
type TargetMode string

const (
	TargetModeUnicast   TargetMode = "unicast"
	TargetModeMulticast TargetMode = "multicast"
	TargetModeBroadcast TargetMode = "broadcast"
)

const broadcastBatch = 50

// Target mendefinisikan penerima notifikasi.
type Target struct {
	Mode         TargetMode
	RecipientID  string
	RecipientIDs []string
}

// NotificationTemplate adalah template isi notifikasi.
type NotificationTemplate struct {
	Type          notifconstant.NotificationType
	Title         string
	Body          string
	Payload       notifvo.NotificationPayload
	Channels      []notifconstant.NotificationChannel
	ReferenceID   *string
	ReferenceType *string
	Bypass        bool
}

// Dispatcher adalah orchestrator utama pengiriman notifikasi.
type Dispatcher struct {
	notifRepo    notifRepo.NotificationRepository
	deliveryRepo deliveryRepo.DeliveryAttemptRepository
	prefRepo     prefRepo.NotificationPreferenceRepository
	logger       *slog.Logger
}

func NewDispatcher(
	notifRepo notifRepo.NotificationRepository,
	deliveryRepo deliveryRepo.DeliveryAttemptRepository,
	prefRepo prefRepo.NotificationPreferenceRepository,
	logger *slog.Logger,
) *Dispatcher {
	return &Dispatcher{
		notifRepo:    notifRepo,
		deliveryRepo: deliveryRepo,
		prefRepo:     prefRepo,
		logger:       logger,
	}
}

// Dispatch mendistribusikan notifikasi ke semua channel sesuai Target.
func (d *Dispatcher) Dispatch(ctx context.Context, tmpl NotificationTemplate, target Target) error {
	switch target.Mode {
	case TargetModeUnicast:
		return d.dispatchToOne(ctx, tmpl, target.RecipientID)
	case TargetModeMulticast:
		return d.dispatchToMany(ctx, tmpl, target.RecipientIDs)
	case TargetModeBroadcast:
		return d.dispatchToMany(ctx, tmpl, target.RecipientIDs)
	default:
		return nil
	}
}

func (d *Dispatcher) dispatchToOne(ctx context.Context, tmpl NotificationTemplate, recipientID string) error {
	if recipientID == "" {
		return nil
	}

	if !d.isRecipientEnabled(ctx, recipientID) {
		return nil
	}

	channels := d.resolveChannels(tmpl.Channels)

	notif, err := notifentity.NewNotification(notifentity.NotificationParams{
		ID:            uuid.NewString(),
		AudienceType:  notifconstant.AudienceTypeUnicast,
		AudienceData:  map[string]any{"user_ids": []string{recipientID}},
		Type:          tmpl.Type,
		Title:         tmpl.Title,
		Body:          tmpl.Body,
		Payload:       tmpl.Payload,
		ReferenceID:   tmpl.ReferenceID,
		ReferenceType: tmpl.ReferenceType,
	})
	if err != nil {
		d.logger.WarnContext(ctx, "gagal buat blueprint unicast", slog.Any("error", err))
		return nil
	}

	for _, ch := range channels {
		if ch == notifconstant.NotificationChannelInApp {
			da, err := deliveryEntity.NewDeliveryAttempt(
				uuid.NewString(), notif.ID, recipientID, notifconstant.NotificationChannelInApp,
			)
			if err != nil {
				d.logger.WarnContext(ctx, "gagal buat in_app delivery attempt", slog.Any("error", err))
				continue
			}
			da.MarkSuccess()
			if err := d.notifRepo.Save(ctx, notif); err != nil {
				d.logger.WarnContext(ctx, "gagal simpan blueprint", slog.Any("error", err))
				return nil
			}
			if err := d.deliveryRepo.Save(ctx, da); err != nil {
				d.logger.WarnContext(ctx, "gagal simpan delivery attempt", slog.Any("error", err))
			}
		}
	}

	return nil
}

func (d *Dispatcher) dispatchToMany(ctx context.Context, tmpl NotificationTemplate, recipientIDs []string) error {
	if len(recipientIDs) == 0 {
		return nil
	}

	notif, err := notifentity.NewNotification(notifentity.NotificationParams{
		ID:            uuid.NewString(),
		AudienceType:  notifconstant.AudienceTypeMulticast,
		AudienceData:  map[string]any{"user_ids": recipientIDs},
		Type:          tmpl.Type,
		Title:         tmpl.Title,
		Body:          tmpl.Body,
		Payload:       tmpl.Payload,
		ReferenceID:   tmpl.ReferenceID,
		ReferenceType: tmpl.ReferenceType,
	})
	if err != nil {
		d.logger.WarnContext(ctx, "gagal buat blueprint multicast", slog.Any("error", err))
		return nil
	}

	if err := d.notifRepo.Save(ctx, notif); err != nil {
		d.logger.WarnContext(ctx, "gagal simpan blueprint", slog.Any("error", err))
		return nil
	}

	channels := d.resolveChannels(tmpl.Channels)

	for i := 0; i < len(recipientIDs); i += broadcastBatch {
		end := i + broadcastBatch
		if end > len(recipientIDs) {
			end = len(recipientIDs)
		}
		batch := recipientIDs[i:end]
		d.dispatchBatch(ctx, notif, tmpl, batch, channels)
	}

	return nil
}

func (d *Dispatcher) dispatchBatch(
	ctx context.Context,
	notif *notifentity.Notification,
	tmpl NotificationTemplate,
	recipientIDs []string,
	channels []notifconstant.NotificationChannel,
) {
	for _, rid := range recipientIDs {
		if !d.isRecipientEnabled(ctx, rid) {
			continue
		}

		for _, ch := range channels {
			if ch == notifconstant.NotificationChannelInApp {
				da, err := deliveryEntity.NewDeliveryAttempt(
					uuid.NewString(), notif.ID, rid, notifconstant.NotificationChannelInApp,
				)
				if err != nil {
					d.logger.WarnContext(ctx, "gagal buat in_app delivery attempt",
						slog.String("user_id", rid), slog.Any("error", err))
					continue
				}
				da.MarkSuccess()
				if err := d.deliveryRepo.Save(ctx, da); err != nil {
					d.logger.WarnContext(ctx, "gagal simpan in_app delivery attempt",
						slog.String("user_id", rid), slog.Any("error", err))
				}
			}
		}
	}
}

func (d *Dispatcher) isRecipientEnabled(ctx context.Context, userID string) bool {
	pref, err := d.prefRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return true // default allow on error
	}
	return pref.AllEnabled
}

func (d *Dispatcher) resolveChannels(channels []notifconstant.NotificationChannel) []notifconstant.NotificationChannel {
	if len(channels) == 0 {
		return []notifconstant.NotificationChannel{notifconstant.NotificationChannelInApp}
	}
	return channels
}

func resolveMediaURL(url string) string {
	return url
}
