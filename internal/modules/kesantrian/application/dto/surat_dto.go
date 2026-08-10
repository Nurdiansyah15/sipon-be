package dto

import "time"

type CreateSuratRequest struct {
	TipeSuratID    string   `json:"tipe_surat_id" binding:"required"`
	Keterangan     *string  `json:"keterangan,omitempty"`
	Tanggal        string   `json:"tanggal" binding:"required"`
	DokumenAsetIDs []string `json:"dokumen_aset_ids,omitempty"`
}

type SuratResponse struct {
	ID          string    `json:"id"`
	Nomor       string    `json:"nomor"`
	TipeSuratID string    `json:"tipe_surat_id"`
	Keterangan  *string   `json:"keterangan,omitempty"`
	Tanggal     string    `json:"tanggal"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type SuratDetailResponse struct {
	ID              string    `json:"id"`
	Nomor           string    `json:"nomor"`
	TipeSuratID     string    `json:"tipe_surat_id"`
	TipeSuratNama   string    `json:"tipe_surat_nama"`
	TipeSuratKode   string    `json:"tipe_surat_kode"`
	Keterangan      *string   `json:"keterangan,omitempty"`
	Tanggal         string    `json:"tanggal"`
	CreatedBy       string    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	DokumenAsetIDs  []string  `json:"dokumen_aset_ids"`
}

type ListSuratQuery struct {
	TipeSuratID string `form:"tipe_surat_id"`
	Bulan       *int   `form:"bulan"`
	Tahun       *int   `form:"tahun"`
	Search      string `form:"search"`
	SortBy      string `form:"sort_by"`
	SortType    string `form:"sort_type"`
	PaginationParams
}

type AddDokumenRequest struct {
	DokumenAsetID string `json:"dokumen_aset_id" binding:"required"`
}

type TautDokumenResponse struct {
	SuratID       string `json:"surat_id"`
	DokumenAsetID string `json:"dokumen_aset_id"`
}

type DownloadResponse struct {
	AccessURL string `json:"access_url"`
	ExpiresIn int    `json:"expires_in"`
}
