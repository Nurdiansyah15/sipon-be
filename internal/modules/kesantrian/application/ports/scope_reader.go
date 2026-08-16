package ports

import (
	"context"

	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
)

// ScopeReader membaca akses scope user terhadap data santri. Implementasinya
// (scopegateway) memanggil identity.Contract — kesantrian tidak pernah menyentuh
// domain identity secara langsung. Hasilnya sudah diterjemahkan ke kosakata
// domain santri (nilai gender '1'/'2'), bukan kode scope master identity
// ("male"/"female").
type ScopeReader interface {
	// GetSantriScopeSet mengembalikan scope santri yang boleh diakses user
	// berdasarkan scope type "gender" milik module identity.
	GetSantriScopeSet(ctx context.Context, userID string) (santriscope.ScopeSet, error)
}
