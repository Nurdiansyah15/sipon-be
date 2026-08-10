package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
)

type GetSuratDownloadUseCase struct {
	suratRepo       suratrepo.SuratRepository
	dokumenReader   ports.DokumenAsetReader
}

func NewGetSuratDownloadUseCase(suratRepo suratrepo.SuratRepository, dokumenReader ports.DokumenAsetReader) *GetSuratDownloadUseCase {
	return &GetSuratDownloadUseCase{suratRepo: suratRepo, dokumenReader: dokumenReader}
}

func (uc *GetSuratDownloadUseCase) Execute(ctx context.Context, suratID, dokumenAsetID string) (*dto.DownloadResponse, error) {
	_, err := uc.suratRepo.FindByID(ctx, suratID)
	if err != nil {
		return nil, application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	result, err := uc.dokumenReader.GetDownloadURL(ctx, dokumenAsetID, true)
	if err != nil {
		return nil, application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	return &dto.DownloadResponse{
		AccessURL: result.AccessURL,
		ExpiresIn: result.ExpiresIn,
	}, nil
}
