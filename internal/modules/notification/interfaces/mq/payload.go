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
