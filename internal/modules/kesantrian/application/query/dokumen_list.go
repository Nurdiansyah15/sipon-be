package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	dokumenconstant "sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	dokumenentity "sipon-be/internal/modules/kesantrian/domain/dokumen/entity"
	dokumenrepo "sipon-be/internal/modules/kesantrian/domain/dokumen/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type DokumenListUseCase struct {
	santriRepo  santrirepo.SantriRepository
	dokumenRepo dokumenrepo.SantriDokumenRepository
}

func NewDokumenListUseCase(santriRepo santrirepo.SantriRepository, dokumenRepo dokumenrepo.SantriDokumenRepository) *DokumenListUseCase {
	return &DokumenListUseCase{santriRepo: santriRepo, dokumenRepo: dokumenRepo}
}

func (uc *DokumenListUseCase) Execute(ctx context.Context, userID, kindFilter string) ([]dto.DokumenItem, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	var dokumens []*dokumenentity.SantriDokumen
	if trimmed := strings.TrimSpace(kindFilter); trimmed != "" {
		d, err := uc.dokumenRepo.FindBySantriIDAndKind(ctx, santri.ID, dokumenconstant.DokumenKind(trimmed))
		if err != nil {
			return []dto.DokumenItem{}, nil
		}
		dokumens = []*dokumenentity.SantriDokumen{d}
	} else {
		all, err := uc.dokumenRepo.FindBySantriID(ctx, santri.ID)
		if err != nil {
			return nil, application.WrapRepoErr(err, dokumenconstant.CodeDokumenNotFound)
		}
		dokumens = all
	}

	return mapDokumenItems(dokumens), nil
}

func mapDokumenItems(dokumens []*dokumenentity.SantriDokumen) []dto.DokumenItem {
	items := make([]dto.DokumenItem, 0, len(dokumens))
	for _, d := range dokumens {
		items = append(items, dto.DokumenItem{
			ID:               d.ID,
			Kind:             string(d.Kind),
			Key:              d.Key,
			Status:           string(d.Status),
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			Size:             d.Size,
			Notes:            d.Notes,
			VerifiedBy:       d.VerifiedBy,
			VerifiedAt:       d.VerifiedAt,
			CreatedAt:        d.CreatedAt,
		})
	}
	return items
}
