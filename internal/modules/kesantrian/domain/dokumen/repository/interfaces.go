package repository

import (
	"context"

	"sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	"sipon-be/internal/modules/kesantrian/domain/dokumen/entity"
)

// Deletion is a soft-delete via Update (DokumenDeleteUseCase calls
// dokumen.SoftDelete() then Update) — there is no hard-delete method here.
type SantriDokumenRepository interface {
	Save(ctx context.Context, dokumen *entity.SantriDokumen) error
	Update(ctx context.Context, dokumen *entity.SantriDokumen) error
	FindByID(ctx context.Context, id string) (*entity.SantriDokumen, error)
	FindBySantriID(ctx context.Context, santriID string) ([]*entity.SantriDokumen, error)
	FindBySantriIDAndKind(ctx context.Context, santriID string, kind constant.DokumenKind) (*entity.SantriDokumen, error)
}
