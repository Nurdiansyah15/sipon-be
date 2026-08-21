package mq

import "errors"

type LoginSucceededPayload struct {
	UserID string `json:"user_id"`
}

func (p LoginSucceededPayload) Validate() error {
	if p.UserID == "" {
		return errors.New("user_id wajib diisi")
	}
	return nil
}

// PsbEventPayload dipakai untuk event PSB yang cuma butuh identitas
// user/pendaftar (submit, accepted).
type PsbEventPayload struct {
	UserID      string `json:"user_id"`
	PendaftarID string `json:"pendaftar_id"`
}

func (p PsbEventPayload) Validate() error {
	if p.UserID == "" || p.PendaftarID == "" {
		return errors.New("user_id dan pendaftar_id wajib diisi")
	}
	return nil
}

// PsbNotesEventPayload dipakai untuk event PSB yang membawa catatan admin
// (revisi, penolakan pendaftaran).
type PsbNotesEventPayload struct {
	UserID      string  `json:"user_id"`
	PendaftarID string  `json:"pendaftar_id"`
	Notes       *string `json:"notes,omitempty"`
}

func (p PsbNotesEventPayload) Validate() error {
	if p.UserID == "" || p.PendaftarID == "" {
		return errors.New("user_id dan pendaftar_id wajib diisi")
	}
	return nil
}

// PsbDokumenEventPayload dipakai untuk event verifikasi/penolakan dokumen PSB.
type PsbDokumenEventPayload struct {
	UserID      string  `json:"user_id"`
	PendaftarID string  `json:"pendaftar_id"`
	DokumenKind string  `json:"dokumen_kind"`
	Notes       *string `json:"notes,omitempty"`
}

func (p PsbDokumenEventPayload) Validate() error {
	if p.UserID == "" || p.PendaftarID == "" || p.DokumenKind == "" {
		return errors.New("user_id, pendaftar_id, dan dokumen_kind wajib diisi")
	}
	return nil
}

// PsbNISGeneratedPayload dipakai saat NIS santri baru diterbitkan.
type PsbNISGeneratedPayload struct {
	UserID      string `json:"user_id"`
	PendaftarID string `json:"pendaftar_id"`
	NIS         string `json:"nis"`
}

func (p PsbNISGeneratedPayload) Validate() error {
	if p.UserID == "" || p.PendaftarID == "" || p.NIS == "" {
		return errors.New("user_id, pendaftar_id, dan nis wajib diisi")
	}
	return nil
}

// dokumenKindLabels memetakan kode jenis dokumen PSB ke label yang enak
// dibaca di isi notifikasi.
var dokumenKindLabels = map[string]string{
	"surat_pernyataan": "Surat Pernyataan",
	"ktp":              "KTP",
	"kk":               "Kartu Keluarga",
	"mutasi":           "Surat Mutasi",
	"pembayaran":       "Bukti Pembayaran",
}

func dokumenKindLabel(kind string) string {
	if label, ok := dokumenKindLabels[kind]; ok {
		return label
	}
	return kind
}

// ArticlePublishedPayload dipakai saat artikel manual dipublikasikan.
type ArticlePublishedPayload struct {
	ArticleID string `json:"article_id"`
	Title     string `json:"title"`
}

func (p ArticlePublishedPayload) Validate() error {
	if p.ArticleID == "" || p.Title == "" {
		return errors.New("article_id dan title wajib diisi")
	}
	return nil
}

// ArticlesScrapedPayload dipakai saat scrape selesai dan ada artikel baru.
type ArticlesScrapedPayload struct {
	SourceID   string   `json:"source_id"`
	SourceName string   `json:"source_name"`
	Count      int      `json:"count"`
	Titles     []string `json:"titles"`
}

func (p ArticlesScrapedPayload) Validate() error {
	if p.SourceID == "" || p.Count == 0 {
		return errors.New("source_id dan count wajib diisi")
	}
	return nil
}

// KeuanganInvoiceEventPayload dipakai untuk event invoice (issued, cancelled).
type KeuanganInvoiceEventPayload struct {
	UserID        string `json:"user_id"`
	InvoiceID     string `json:"invoice_id"`
	InvoiceNumber string `json:"invoice_number"`
}

func (p KeuanganInvoiceEventPayload) Validate() error {
	if p.UserID == "" || p.InvoiceID == "" {
		return errors.New("user_id dan invoice_id wajib diisi")
	}
	return nil
}

// KeuanganPaymentEventPayload dipakai untuk event payment (submitted, verified, rejected).
type KeuanganPaymentEventPayload struct {
	UserID    string `json:"user_id"`
	InvoiceID string `json:"invoice_id"`
}

func (p KeuanganPaymentEventPayload) Validate() error {
	if p.UserID == "" || p.InvoiceID == "" {
		return errors.New("user_id dan invoice_id wajib diisi")
	}
	return nil
}

// AkademikSessionReminderPayload dipakai untuk reminder sesi kegiatan — dikirim
// ke banyak user (multicast) yang sudah herregistrasi pada periode terkait.
type AkademikSessionReminderPayload struct {
	SessionID    string   `json:"session_id"`
	UserIDs      []string `json:"user_ids"`
	ActivityName string   `json:"activity_name,omitempty"`
	StartsAt     string   `json:"starts_at,omitempty"`
}

func (p AkademikSessionReminderPayload) Validate() error {
	if p.SessionID == "" {
		return errors.New("session_id wajib diisi")
	}
	if len(p.UserIDs) == 0 {
		return errors.New("user_ids wajib diisi")
	}
	return nil
}

// AkademikAttendanceRecordedPayload dipakai saat kehadiran santri tercatat —
// baik lewat check-in manual via NIS (source "nis") maupun sinkronisasi
// fingerprint (source "fingerprint").
type AkademikAttendanceRecordedPayload struct {
	UserID       string `json:"user_id"`
	AttendanceID string `json:"attendance_id"`
	SantriID     string `json:"santri_id"`
	NIS          string `json:"nis"`
	Name         string `json:"name"`
	SessionID    string `json:"session_id"`
	Source       string `json:"source"`
}

func (p AkademikAttendanceRecordedPayload) Validate() error {
	if p.UserID == "" {
		return errors.New("user_id wajib diisi")
	}
	if p.AttendanceID == "" {
		return errors.New("attendance_id wajib diisi")
	}
	if p.SessionID == "" {
		return errors.New("session_id wajib diisi")
	}
	return nil
}
