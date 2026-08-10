package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
)

type GetSuratUseCase struct {
	suratRepo     suratrepo.SuratRepository
	tipeSuratRepo tiperepo.TipeSuratRepository
}

func NewGetSuratUseCase(suratRepo suratrepo.SuratRepository, tipeSuratRepo tiperepo.TipeSuratRepository) *GetSuratUseCase {
	return &GetSuratUseCase{suratRepo: suratRepo, tipeSuratRepo: tipeSuratRepo}
}

func (uc *GetSuratUseCase) Execute(ctx context.Context, id string) (*dto.SuratDetailResponse, error) {
	detail, err := uc.suratRepo.FindDetail(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	tipe, err := uc.tipeSuratRepo.FindByID(ctx, detail.Surat.TipeSuratID)
	if err != nil {
		return nil, application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	s := detail.Surat
	dokIDs := detail.DokumenAsetIDs
	if dokIDs == nil {
		dokIDs = []string{}
	}

	return &dto.SuratDetailResponse{
		ID:             s.ID,
		Nomor:          s.Nomor,
		TipeSuratID:    s.TipeSuratID,
		TipeSuratNama:  tipe.Nama,
		TipeSuratKode:  tipe.Kode,
		Keterangan:     s.Keterangan,
		Tanggal:        s.Tanggal.Format("2006-01-02"),
		CreatedBy:      s.CreatedBy,
		CreatedAt:      s.CreatedAt,
		DokumenAsetIDs: dokIDs,
	}, nil
}
