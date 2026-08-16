package repository

import (
	"context"

	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
	"sipon-be/internal/modules/kesantrian/domain/surat/entity"
)

type SuratListQuery struct {
	TipeSuratID *string
	Bulan       *int
	Tahun       *int
	Search      *string
	Page        int
	Limit       int
	SortBy      string
	SortType    string
	// Scope membatasi daftar surat sesuai akses scope user. Nilai allowedOptions
	// di sini adalah master scope ID (bukan kode gender). Zero value diperlakukan
	// sebagai tanpa filter.
	Scope santriscope.ScopeSet
}

type SuratListResult struct {
	Items []*entity.Surat
	Total int64
}

type SuratDetail struct {
	Surat          *entity.Surat
	DokumenAsetIDs []string
}

type SuratRepository interface {
	Save(ctx context.Context, s *entity.Surat) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*entity.Surat, error)
	List(ctx context.Context, q SuratListQuery) (*SuratListResult, error)
	FindMaxSeqByMonthYear(ctx context.Context, bulan, tahun int) (int, error)
	SaveDokumenLink(ctx context.Context, link *entity.SuratDokumenAset) error
	DeleteDokumenLink(ctx context.Context, suratID, dokumenAsetID string) error
	FindDokumenAsetIDsBySuratID(ctx context.Context, suratID string) ([]string, error)
	FindDetail(ctx context.Context, id string) (*SuratDetail, error)
}
