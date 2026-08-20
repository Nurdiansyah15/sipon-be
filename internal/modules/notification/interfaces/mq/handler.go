package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sipon-be/internal/modules/messaging"
	"sipon-be/internal/modules/notification/application"
	notifconstant "sipon-be/internal/modules/notification/domain/notification/constant"
	notifvo "sipon-be/internal/modules/notification/domain/valueobject"
)

// UserProvider menyediakan daftar user ID aktif untuk broadcast notifikasi.
type UserProvider interface {
	ListActiveUserIDs(ctx context.Context) ([]string, error)
}

type Dependencies struct {
	Dispatcher   *application.Dispatcher
	UserProvider UserProvider
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

func (h handlers) broadcast(ctx context.Context, tmpl application.NotificationTemplate) error {
	if h.deps.UserProvider == nil {
		return nil
	}
	userIDs, err := h.deps.UserProvider.ListActiveUserIDs(ctx)
	if err != nil {
		return messaging.NewRetryableError(fmt.Errorf("gagal ambil daftar user aktif: %w", err))
	}
	if len(userIDs) == 0 {
		return nil
	}
	target := application.Target{Mode: application.TargetModeBroadcast, RecipientIDs: userIDs}
	if err := h.deps.Dispatcher.Dispatch(ctx, tmpl, target); err != nil {
		return messaging.NewRetryableError(err)
	}
	return nil
}

func (h handlers) handleArticlePublished(ctx context.Context, msg messaging.Message) error {
	var p ArticlePublishedPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingArticlePublished, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeContent,
		Title: "Artikel Baru",
		Body:  p.Title,
		Payload: notifvo.NotificationPayload{
			Module:      "article",
			EventType:   "article_published",
			EntityID:    p.ArticleID,
			ClickAction: fmt.Sprintf("/artikel/%s", p.ArticleID),
		},
		Channels: pushChannels,
	}

	return h.broadcast(ctx, tmpl)
}

func (h handlers) handleArticlesScraped(ctx context.Context, msg messaging.Message) error {
	var p ArticlesScrapedPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingArticlesScraped, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	body := fmt.Sprintf("%d artikel baru telah ditambahkan", p.Count)
	if len(p.Titles) > 0 {
		shown := p.Titles
		if len(shown) > 3 {
			shown = shown[:3]
			body = fmt.Sprintf("%s: %s, dan lainnya", body, strings.Join(shown, ", "))
		} else {
			body = fmt.Sprintf("%s: %s", body, strings.Join(shown, ", "))
		}
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeContent,
		Title: fmt.Sprintf("Artikel Baru dari %s", p.SourceName),
		Body:  body,
		Payload: notifvo.NotificationPayload{
			Module:      "article",
			EventType:   "articles_scraped",
			EntityID:    p.SourceID,
			ClickAction: "/artikel",
		},
		Channels: pushChannels,
	}

	return h.broadcast(ctx, tmpl)
}

// ─── Keuangan handlers ────────────────────────────────────────────────────

const keuanganTagihanClickAction = "/keuangan/tagihan"

func (h handlers) handleKeuanganInvoiceIssued(ctx context.Context, msg messaging.Message) error {
	var p KeuanganInvoiceEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingKeuanganInvoiceIssued, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Tagihan Baru",
		Body:  fmt.Sprintf("Tagihan baru dengan nomor %s telah diterbitkan untuk Anda.", p.InvoiceNumber),
		Payload: notifvo.NotificationPayload{
			Module:      "keuangan",
			EventType:   "invoice_issued",
			EntityID:    p.InvoiceID,
			ClickAction: fmt.Sprintf("%s/%s", keuanganTagihanClickAction, p.InvoiceID),
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handleKeuanganInvoiceCancelled(ctx context.Context, msg messaging.Message) error {
	var p KeuanganInvoiceEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingKeuanganInvoiceCancelled, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Tagihan Dibatalkan",
		Body:  fmt.Sprintf("Tagihan dengan nomor %s telah dibatalkan.", p.InvoiceNumber),
		Payload: notifvo.NotificationPayload{
			Module:      "keuangan",
			EventType:   "invoice_cancelled",
			EntityID:    p.InvoiceID,
			ClickAction: fmt.Sprintf("%s/%s", keuanganTagihanClickAction, p.InvoiceID),
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handleKeuanganPaymentSubmitted(ctx context.Context, msg messaging.Message) error {
	var p KeuanganPaymentEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingKeuanganPaymentSubmitted, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Pembayaran Diajukan",
		Body:  "Pembayaran Anda telah diajukan dan sedang menunggu verifikasi oleh admin.",
		Payload: notifvo.NotificationPayload{
			Module:      "keuangan",
			EventType:   "payment_submitted",
			EntityID:    p.InvoiceID,
			ClickAction: fmt.Sprintf("%s/%s", keuanganTagihanClickAction, p.InvoiceID),
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handleKeuanganPaymentVerified(ctx context.Context, msg messaging.Message) error {
	var p KeuanganPaymentEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingKeuanganPaymentVerified, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Pembayaran Terverifikasi",
		Body:  "Pembayaran Anda telah diverifikasi oleh admin. Terima kasih!",
		Payload: notifvo.NotificationPayload{
			Module:      "keuangan",
			EventType:   "payment_verified",
			EntityID:    p.InvoiceID,
			ClickAction: fmt.Sprintf("%s/%s", keuanganTagihanClickAction, p.InvoiceID),
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}

func (h handlers) handleKeuanganPaymentRejected(ctx context.Context, msg messaging.Message) error {
	var p KeuanganPaymentEventPayload
	if err := json.Unmarshal(msg.Payload, &p); err != nil {
		return messaging.NewFatalError(fmt.Errorf("decode %s payload: %w", RoutingKeuanganPaymentRejected, err))
	}
	if err := p.Validate(); err != nil {
		return messaging.NewFatalError(fmt.Errorf("payload invalid: %w", err))
	}

	tmpl := application.NotificationTemplate{
		Type:  notifconstant.NotificationTypeSystem,
		Title: "Pembayaran Ditolak",
		Body:  "Pembayaran Anda ditolak oleh admin. Silakan hubungi bagian keuangan untuk informasi lebih lanjut.",
		Payload: notifvo.NotificationPayload{
			Module:      "keuangan",
			EventType:   "payment_rejected",
			EntityID:    p.InvoiceID,
			ClickAction: fmt.Sprintf("%s/%s", keuanganTagihanClickAction, p.InvoiceID),
		},
		Channels: pushChannels,
		Bypass:   true,
	}

	return h.dispatch(ctx, tmpl, p.UserID)
}
