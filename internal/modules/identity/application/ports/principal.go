package ports

import "context"

// PrincipalCacheInvalidator menghapus snapshot principal user dari cache
// (Redis). Dipanggil setelah mutasi role/permission/user-role supaya akses
// berikutnya di-build ulang dari DB, bukan memakai snapshot yang basi.
// Implementasi disediakan oleh cache.RedisPrincipalCache.
type PrincipalCacheInvalidator interface {
	Delete(ctx context.Context, userID string) error
}
