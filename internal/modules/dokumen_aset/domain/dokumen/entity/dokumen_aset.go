package entity

import (
	"time"

	"sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	"sipon-be/internal/shared/kernel"
)

type DokumenAset struct {
	ID        string
	Judul     string
	Deskripsi *string
	Kategori  constant.Kategori
	Key       string
	Filename  string
	MimeType  string
	Size      int64
	IsPublic  bool
	CreatedBy string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewDokumenAset(id, judul string, kategori constant.Kategori, key, filename, mimeType string, size int64, isPublic bool, createdBy string) (*DokumenAset, error) {
	if !constant.ValidKategoris[kategori] {
		return nil, kernel.New(constant.CodeDokumenInvalidKategori)
	}
	now := time.Now()
	return &DokumenAset{
		ID:        id,
		Judul:     judul,
		Kategori:  kategori,
		Key:       key,
		Filename:  filename,
		MimeType:  mimeType,
		Size:      size,
		IsPublic:  isPublic,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (d *DokumenAset) UpdateMetadata(judul, deskripsi *string, kategori *constant.Kategori, isPublic *bool) error {
	if judul != nil {
		d.Judul = *judul
	}
	if deskripsi != nil {
		d.Deskripsi = deskripsi
	}
	if kategori != nil {
		if !constant.ValidKategoris[*kategori] {
			return kernel.New(constant.CodeDokumenInvalidKategori)
		}
		d.Kategori = *kategori
	}
	if isPublic != nil {
		d.IsPublic = *isPublic
	}
	d.UpdatedAt = time.Now()
	return nil
}

func (d *DokumenAset) SoftDelete() {
	now := time.Now()
	d.DeletedAt = &now
	d.UpdatedAt = now
}
