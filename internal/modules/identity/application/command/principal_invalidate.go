package command

import (
	"context"
	"log/slog"

	ports "sipon-be/internal/modules/identity/application/ports"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
)

// invalidateRoleHolders menghapus principal cache untuk semua user aktif yang
// memegang role tertentu. Best-effort: kegagalan hanya dicatat, tidak
// menggagalkan operasi mutasi yang sudah berhasil ditulis ke DB.
func invalidateRoleHolders(ctx context.Context, invalidator ports.PrincipalCacheInvalidator, userRoleRepo rolerepo.UserRoleRepository, roleName roleconstant.RoleName) {
	userIDs, err := userRoleRepo.ListActiveUserIDsByRoleName(ctx, roleName)
	if err != nil {
		slog.Warn("identity: gagal mendaftar user pemegang role untuk invalidasi principal", "role", roleName, "error", err)
		return
	}
	invalidatePrincipalUsers(ctx, invalidator, userIDs)
}

func invalidatePrincipalUsers(ctx context.Context, invalidator ports.PrincipalCacheInvalidator, userIDs []string) {
	for _, userID := range userIDs {
		if err := invalidator.Delete(ctx, userID); err != nil {
			slog.Warn("identity: gagal menghapus principal cache user", "user_id", userID, "error", err)
		}
	}
}
