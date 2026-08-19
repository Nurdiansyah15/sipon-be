package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"

	"github.com/google/uuid"

notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	notifentity "sipon-be/internal/modules/notification/domain/notification/entity"
	notifRepo "sipon-be/internal/modules/notification/domain/notification/repository"
	deliveryEntity "sipon-be/internal/modules/notification/domain/delivery/entity"
	deliveryRepo "sipon-be/internal/modules/notification/domain/delivery/repository"
	deviceEntity "sipon-be/internal/modules/notification/domain/device/entity"
	deviceRepo "sipon-be/internal/modules/notification/domain/device/repository"
	prefRepo "sipon-be/internal/modules/notification/domain/preference/repository"
notifvo "sipon-be/internal/modules/notification/domain/valueobject"
	"sipon-be/internal/modules/notification/infrastructure/external"
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
	deviceRepo   deviceRepo.DeviceRegistrationRepository
	pushSender   external.PushSender
	logger       *slog.Logger
}

func NewDispatcher(
	notifRepo notifRepo.NotificationRepository,
	deliveryRepo deliveryRepo.DeliveryAttemptRepository,
	prefRepo prefRepo.NotificationPreferenceRepository,
	deviceRepo deviceRepo.DeviceRegistrationRepository,
	pushSender external.PushSender,
	logger *slog.Logger,
) *Dispatcher {
	return &Dispatcher{
		notifRepo:    notifRepo,
		deliveryRepo: deliveryRepo,
		prefRepo:     prefRepo,
		deviceRepo:   deviceRepo,
		pushSender:   pushSender,
		logger:       logger,
	}
}

// Dispatch mendistribusikan notifikasi ke semua channel sesuai Target.
func (d *Dispatcher) Dispatch(ctx context.Context, tmpl NotificationTemplate, target Target) error {
	switch target.Mode {
	case TargetModeUnicast:
		return d.dispatchToOne(ctx, tmpl, target.RecipientID)
	case TargetModeMulticast, TargetModeBroadcast:
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

	notifSaved := false
	for _, ch := range channels {
		switch ch {
		case notifconstant.NotificationChannelInApp:
			if !notifSaved {
				if err := d.notifRepo.Save(ctx, notif); err != nil {
					d.logger.WarnContext(ctx, "gagal simpan blueprint", slog.Any("error", err))
					return nil
				}
				notifSaved = true
			}
			da, err := deliveryEntity.NewDeliveryAttempt(
				uuid.NewString(), notif.ID, recipientID, notifconstant.NotificationChannelInApp,
			)
			if err != nil {
				d.logger.WarnContext(ctx, "gagal buat in_app delivery attempt", slog.Any("error", err))
				continue
			}
			da.MarkSuccess()
			if err := d.deliveryRepo.Save(ctx, da); err != nil {
				d.logger.WarnContext(ctx, "gagal simpan delivery attempt", slog.Any("error", err))
			}

		case notifconstant.NotificationChannelPush:
			if !notifSaved {
				if err := d.notifRepo.Save(ctx, notif); err != nil {
					d.logger.WarnContext(ctx, "gagal simpan blueprint", slog.Any("error", err))
					return nil
				}
				notifSaved = true
			}
			d.sendPush(ctx, notif, recipientID, tmpl)
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
	hasPush := slices.Contains(channels, notifconstant.NotificationChannelPush)

	var allPushMsgs []external.PushMessage
	type pushEntry struct {
		attempt  *deliveryEntity.DeliveryAttempt
		messages []external.PushMessage
	}
	var pushEntries []pushEntry

	for i := 0; i < len(recipientIDs); i += broadcastBatch {
		end := i + broadcastBatch
		if end > len(recipientIDs) {
			end = len(recipientIDs)
		}
		batch := recipientIDs[i:end]

		for _, rid := range batch {
			if !d.isRecipientEnabled(ctx, rid) {
				continue
			}

			if slices.Contains(channels, notifconstant.NotificationChannelInApp) {
				da, err := deliveryEntity.NewDeliveryAttempt(
					uuid.NewString(), notif.ID, rid, notifconstant.NotificationChannelInApp,
				)
				if err == nil {
					da.MarkSuccess()
					_ = d.deliveryRepo.Save(ctx, da)
				}
			}

			if hasPush && d.pushSender != nil {
				da, err := deliveryEntity.NewDeliveryAttempt(
					uuid.NewString(), notif.ID, rid, notifconstant.NotificationChannelPush,
				)
				if err != nil {
					continue
				}
				devices, _ := d.deviceRepo.FindActiveByUserID(ctx, rid)
				if len(devices) == 0 {
					da.MarkFailed("NO_ACTIVE_DEVICE")
					_ = d.deliveryRepo.Save(ctx, da)
				} else {
					msgs := d.buildPushMessages(devices, tmpl)
					allPushMsgs = append(allPushMsgs, msgs...)
					pushEntries = append(pushEntries, pushEntry{attempt: da, messages: msgs})
				}
			}
		}
	}

	if len(allPushMsgs) > 0 && d.pushSender != nil {
		results, batchErr := d.pushSender.SendBatch(ctx, allPushMsgs)
		d.deactivateInvalidTokens(ctx, results)

		for _, entry := range pushEntries {
			if batchErr != nil {
				if entry.attempt.IsRetryable() {
					entry.attempt.ScheduleRetry(entry.attempt.AttemptedAt.Add(5 * 60 * 1e9)) // 5 min
				} else {
					entry.attempt.MarkFailed("FCM_BATCH_FAILED")
				}
			} else {
				entry.attempt.MarkSuccess()
			}
			_ = d.deliveryRepo.Save(ctx, entry.attempt)
		}
	}

	return nil
}

func (d *Dispatcher) sendPush(ctx context.Context, notif *notifentity.Notification, recipientID string, tmpl NotificationTemplate) {
	da, err := deliveryEntity.NewDeliveryAttempt(
		uuid.NewString(), notif.ID, recipientID, notifconstant.NotificationChannelPush,
	)
	if err != nil {
		d.logger.WarnContext(ctx, "gagal buat push delivery attempt", slog.Any("error", err))
		return
	}

	devices, err := d.deviceRepo.FindActiveByUserID(ctx, recipientID)
	if err != nil || len(devices) == 0 {
		da.MarkFailed("NO_ACTIVE_DEVICE")
		_ = d.deliveryRepo.Save(ctx, da)
		return
	}

	msgs := d.buildPushMessages(devices, tmpl)
	results, err := d.pushSender.SendBatch(ctx, msgs)
	d.deactivateInvalidTokens(ctx, results)

	if err != nil {
		d.logger.WarnContext(ctx, "gagal kirim push", slog.String("user_id", recipientID), slog.Any("error", err))
		if da.IsRetryable() {
			da.ScheduleRetry(da.AttemptedAt.Add(5 * 60 * 1e9))
		} else {
			da.MarkFailed("FCM_SEND_FAILED")
		}
	} else {
		da.MarkSuccess()
	}

	if err := d.deliveryRepo.Save(ctx, da); err != nil {
		d.logger.WarnContext(ctx, "gagal simpan push delivery attempt", slog.Any("error", err))
	}
}

func (d *Dispatcher) buildPushMessages(devices []*deviceEntity.DeviceRegistration, tmpl NotificationTemplate) []external.PushMessage {
	msgs := make([]external.PushMessage, 0, len(devices))
	data := map[string]string{
		"module":       tmpl.Payload.Module,
		"event_type":   tmpl.Payload.EventType,
		"entity_id":    tmpl.Payload.EntityID,
		"click_action": tmpl.Payload.ClickAction,
	}
	if tmpl.Payload.Extra != nil {
		b, _ := json.Marshal(tmpl.Payload.Extra)
		data["extra"] = string(b)
	}

	for _, dev := range devices {
		msgs = append(msgs, external.PushMessage{
			Token:   dev.ProviderToken,
			Title:   tmpl.Title,
			Body:    tmpl.Body,
			Payload: data,
			Priority: func() string {
				if tmpl.Bypass {
					return "high"
				}
				return "normal"
			}(),
		})
	}
	return msgs
}

func (d *Dispatcher) deactivateInvalidTokens(ctx context.Context, results []external.PushResult) {
	if len(results) == 0 {
		return
	}
	for _, r := range results {
		if r.TokenInvalid {
			if err := d.deviceRepo.DeactivateByToken(ctx, r.Token); err != nil {
				d.logger.WarnContext(ctx, "gagal nonaktifkan token invalid", slog.String("token_prefix", r.Token[:min(20, len(r.Token))]), slog.Any("error", err))
			}
		}
	}
}

func (d *Dispatcher) isRecipientEnabled(ctx context.Context, userID string) bool {
	pref, err := d.prefRepo.FindOrCreateByUserID(ctx, userID)
	if err != nil {
		return true
	}
	return pref.AllEnabled
}

func (d *Dispatcher) resolveChannels(channels []notifconstant.NotificationChannel) []notifconstant.NotificationChannel {
	if len(channels) == 0 {
		return []notifconstant.NotificationChannel{notifconstant.NotificationChannelInApp}
	}
	return channels
}
