package ports

import (
	"context"

	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
)

// SuratScopeAccess adalah akses scope user untuk resource surat — disimpan
// sebagai master scope ID (bukan kode) karena surat menyimpan scope_id.
type SuratScopeAccess struct {
	HasAccess       bool
	HasFullAccess   bool
	AllowedScopeIDs []string
}

// ScopeReader membaca akses scope user terhadap data santri. Implementasinya
// (scopegateway) memanggil identity.Contract — kesantrian tidak pernah menyentuh
// domain identity secara langsung. Hasilnya sudah diterjemahkan ke kosakata
// domain santri (nilai gender '1'/'2'), bukan kode scope master identity
// ("male"/"female").
type ScopeReader interface {
	// GetSantriScopeSet mengembalikan scope santri yang boleh diakses user
	// berdasarkan scope type "gender" milik module identity.
	GetSantriScopeSet(ctx context.Context, userID string) (santriscope.ScopeSet, error)

	// GetSuratScopeAccess mengembalikan akses scope user untuk resource surat —
	// AllowedScopeIDs berisi master scope ID yang boleh diakses.
	GetSuratScopeAccess(ctx context.Context, userID string) (*SuratScopeAccess, error)
}
