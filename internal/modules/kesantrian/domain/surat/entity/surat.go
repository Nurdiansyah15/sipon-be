package entity

import (
	"time"

	"sipon-be/internal/shared/kernel"
)

type Surat struct {
	ID          string
	Nomor       string
	Seq         int
	TipeSuratID string
	Keterangan  *string
	Tanggal     time.Time
	CreatedBy   string
	ScopeID     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewSurat(id, nomor string, seq int, tipeSuratID string, keterangan *string, tanggal time.Time, createdBy string, scopeID *string) (*Surat, error) {
	if nomor == "" || tipeSuratID == "" || createdBy == "" {
		return nil, kernel.New("SURAT_INVALID")
	}
	now := time.Now()
	return &Surat{
		ID:          id,
		Nomor:       nomor,
		Seq:         seq,
		TipeSuratID: tipeSuratID,
		Keterangan:  keterangan,
		Tanggal:     tanggal,
		CreatedBy:   createdBy,
		ScopeID:     scopeID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
