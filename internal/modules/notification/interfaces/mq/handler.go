package mq

import (
	"context"
	"encoding/json"
	"fmt"

	"sipon-be/internal/modules/messaging"
	"sipon-be/internal/modules/notification/application"
	notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	notifvo "sipon-be/internal/modules/notification/domain/valueobject"
)

type Dependencies struct {
	Dispatcher *application.Dispatcher
}

type handlers struct {
	deps Dependencies
}

var pushChannels = []notifconstant.NotificationChannel{notifconstant.NotificationChannelInApp, notifconstant.NotificationChannelPush}

// psbRiwayatClickAction adalah halaman status pendaftaran milik pendaftar
// sendiri (self-scoped, tanpa parameter id) — semua event PSB saat ini
// menotifikasi pendaftar tentang statusnya sendiri, bukan admin.
const psbRiwayatClickAction = "/psb/riwayat"

// dispatch mengirim tmpl secara unicast ke userID, membungkus error dispatcher
// sebagai retryable (RabbitMQ akan redeliver sesuai kebijakan retry consumer).
func (h handlers) dispatch(ctx context.Context, tmpl application.NotificationTemplate, userID string) error {
	target := application.Target{Mode: application.TargetModeUnicast, RecipientID: userID}
	if err := h.deps.Dispatcher.Dispatch(ctx, tmpl, target); err != nil {
		return messaging.NewRetryableError(err)
	}
	return nil
}

func (h handlers) handleLoginSucceeded(ctx context.Context, msg messaging.Message) error {
	var p LoginSucceededPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingLoginSucceeded, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Login Berhasil",
		Body:  "Anda telah berhasil masuk ke sistem.",
		Payload: notifvo.NotificationPayload{
			Module:    "identity",
			EventType: "login_succeeded",
			EntityID:  p.UserID,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbPendaftaranSubmitted(ctx context.Context, msg messaging.Message) error {
	var p PsbEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbPendaftaranSubmitted, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Pendaftaran Terkirim",
		Body:  "Pendaftaran Anda telah berhasil diajukan dan sedang menunggu verifikasi.",
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "pendaftaran_submitted",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbDaftarUlangSubmitted(ctx context.Context, msg messaging.Message) error {
	var p PsbEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbDaftarUlangSubmitted, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Daftar Ulang Terkirim",
		Body:  "Daftar ulang Anda telah berhasil diajukan dan sedang menunggu verifikasi.",
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "daftar_ulang_submitted",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbDokumenVerified(ctx context.Context, msg messaging.Message) error {
	var p PsbDokumenEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbDokumenVerified, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Dokumen Terverifikasi",
		Body:  fmt.Sprintf("Dokumen %s Anda telah diverifikasi.", dokumenKindLabel(p.DokumenKind)),
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "dokumen_verified",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbDokumenRejected(ctx context.Context, msg messaging.Message) error {
	var p PsbDokumenEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbDokumenRejected, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	body := fmt.Sprintf("Dokumen %s Anda ditolak, silakan upload ulang.", dokumenKindLabel(p.DokumenKind))
	if p.Notes != nil && *p.Notes != "" {
		body = fmt.Sprintf("%s Catatan: %s", body, *p.Notes)
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Dokumen Ditolak",
		Body:  body,
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "dokumen_rejected",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbRevisionRequested(ctx context.Context, msg messaging.Message) error {
	var p PsbNotesEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbRevisionRequested, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	body := "Pendaftaran Anda perlu direvisi."
	if p.Notes != nil && *p.Notes != "" {
		body = fmt.Sprintf("%s Catatan: %s", body, *p.Notes)
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Perlu Revisi",
		Body:  body,
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "revision_requested",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbRevisionRequestedDaftarUlang(ctx context.Context, msg messaging.Message) error {
	var p PsbNotesEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbRevisionRequestedDaftarUlang, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	body := "Daftar ulang Anda perlu direvisi."
	if p.Notes != nil && *p.Notes != "" {
		body = fmt.Sprintf("%s Catatan: %s", body, *p.Notes)
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Daftar Ulang Perlu Direvisi",
		Body:  body,
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "revision_requested_daftar_ulang",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbPendaftaranAccepted(ctx context.Context, msg messaging.Message) error {
	var p PsbEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbPendaftaranAccepted, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Selamat! Pendaftaran Diterima",
		Body:  "Pendaftaran Anda telah diterima. Silakan lanjutkan ke tahap daftar ulang.",
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "pendaftaran_accepted",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbPendaftaranRejected(ctx context.Context, msg messaging.Message) error {
	var p PsbNotesEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbPendaftaranRejected, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	body := "Mohon maaf, pendaftaran Anda tidak dapat diterima."
	if p.Notes != nil && *p.Notes != "" {
		body = fmt.Sprintf("%s Catatan: %s", body, *p.Notes)
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Pendaftaran Ditolak",
		Body:  body,
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "pendaftaran_rejected",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handlePsbNISGenerated(ctx context.Context, msg messaging.Message) error {
	var p PsbNISGeneratedPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingPsbNISGenerated, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "NIS Terbit",
		Body:  fmt.Sprintf("Selamat! NIS Anda telah terbit: %s. Proses pendaftaran santri baru telah selesai.", p.NIS),
		Payload: notifvo.NotificationPayload{
			Module:      "psb",
			EventType:   "nis_generated",
			EntityID:    p.PendaftarID,
			ClickAction: psbRiwayatClickAction,
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}
