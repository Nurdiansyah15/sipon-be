package query

import (
	"strings"

	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	dokumenconstant "sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	dokumenentity "sipon-be/internal/modules/kesantrian/domain/dokumen/entity"
	dokumenrepo "sipon-be/internal/modules/kesantrian/domain/dokumen/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

// AdminDokumenListUseCase lists a specific santri's documents by santriID —
// unlike DokumenListUseCase (self-service, resolves the santri from the
// caller's own userID), this is for the admin document-review UI, which
// needs to browse any santri's documents regardless of who's logged in.
type AdminDokumenListUseCase struct {
	santriRepo  santrirepo.SantriRepository
	dokumenRepo dokumenrepo.SantriDokumenRepository
}

func NewAdminDokumenListUseCase(santriRepo santrirepo.SantriRepository, dokumenRepo dokumenrepo.SantriDokumenRepository) *AdminDokumenListUseCase {
	return &AdminDokumenListUseCase{santriRepo: santriRepo, dokumenRepo: dokumenRepo}
}

func (uc *AdminDokumenListUseCase) Execute(ctx context.Context, santriID, kindFilter string) ([]dto.DokumenItem, error) {
	if _, err := uc.santriRepo.FindByID(ctx, santriID); err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	var dokumens []*dokumenentity.SantriDokumen
	if trimmed := strings.TrimSpace(kindFilter); trimmed != "" {
		d, err := uc.dokumenRepo.FindBySantriIDAndKind(ctx, santriID, dokumenconstant.DokumenKind(trimmed))
		if err != nil {
			return []dto.DokumenItem{}, nil
		}
		dokumens = []*dokumenentity.SantriDokumen{d}
	} else {
		all, err := uc.dokumenRepo.FindBySantriID(ctx, santriID)
		if err != nil {
			return nil, application.WrapRepoErr(err, dokumenconstant.CodeDokumenNotFound)
		}
		dokumens = all
	}

	return mapDokumenItems(dokumens), nil
}
