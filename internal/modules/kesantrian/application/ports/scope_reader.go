package ports

import (
	"context"

	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
)

// ScopeReader membaca akses scope user terhadap data santri. Implementasinya
// (scopegateway) memanggil system.Contract — kesantrian tidak pernah menyentuh
// domain system. Hasilnya sudah diterjemahkan ke kosakata domain santri
// (nilai gender '1'/'2'), bukan kode scope sistem ("male"/"female").
type ScopeReader interface {
	// GetSantriScopeSet mengembalikan scope santri yang boleh diakses user
	// berdasarkan scope type "gender" milik module system.
	GetSantriScopeSet(ctx context.Context, userID string) (santriscope.ScopeSet, error)
}
