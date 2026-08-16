package entity

import "time"

type SuratDokumenAset struct {
	ID            string
	SuratID       string
	DokumenAsetID string
	CreatedAt     time.Time
}
