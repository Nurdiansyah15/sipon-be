package entity

import (
	"time"

	"sipon-be/internal/shared/kernel"
)

type TipeSurat struct {
	ID        string
	Nama      string
	Kode      string
	CreatedBy *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewTipeSurat(id, nama, kode string, createdBy *string) (*TipeSurat, error) {
	if nama == "" || kode == "" {
		return nil, kernel.New("TIPE_SURAT_INVALID")
	}
	now := time.Now()
	return &TipeSurat{
		ID:        id,
		Nama:      nama,
		Kode:      kode,
		CreatedBy: createdBy,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (t *TipeSurat) Update(nama, kode string) {
	t.Nama = nama
	t.Kode = kode
	t.UpdatedAt = time.Now()
}

func (t *TipeSurat) UpdateNama(nama string) {
	t.Nama = nama
	t.UpdatedAt = time.Now()
}
