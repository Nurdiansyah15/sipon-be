package ports

import (
	"context"

	pendaftarscope "sipon-be/internal/modules/psb/domain/pendaftar/scope"
)

// ScopeReader membaca akses scope user terhadap data pendaftar PSB.
// Implementasinya (scopegateway) memanggil identity.Contract — psb tidak
// pernah menyentuh domain identity secara langsung. Hasilnya sudah
// diterjemahkan ke kosakata domain pendaftar (nilai gender '1'/'2'), bukan
// kode scope master identity ("male"/"female").
type ScopeReader interface {
	// GetPendaftarScopeSet mengembalikan scope pendaftar yang boleh diakses
	// user berdasarkan scope type "gender" milik module identity.
	GetPendaftarScopeSet(ctx context.Context, userID string) (pendaftarscope.ScopeSet, error)
}