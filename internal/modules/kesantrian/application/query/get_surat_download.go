package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"
)

type GetSuratDownloadUseCase struct {
	suratRepo     suratrepo.SuratRepository
	dokumenReader ports.DokumenAsetReader
	scopeReader   ports.ScopeReader
}

func NewGetSuratDownloadUseCase(suratRepo suratrepo.SuratRepository, dokumenReader ports.DokumenAsetReader, scopeReader ports.ScopeReader) *GetSuratDownloadUseCase {
	return &GetSuratDownloadUseCase{suratRepo: suratRepo, dokumenReader: dokumenReader, scopeReader: scopeReader}
}

func (uc *GetSuratDownloadUseCase) Execute(ctx context.Context, userID, suratID, dokumenAsetID string) (*dto.DownloadResponse, error) {
	s, err := uc.suratRepo.FindByID(ctx, suratID)
	if err != nil {
		return nil, application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	if err := uc.ensureScopeAccess(ctx, userID, s.ScopeID); err != nil {
		return nil, err
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

// ensureScopeAccess menolak akses bila surat membawa scope_id di luar scope
// yang boleh diakses user. Surat global (scope_id nil) publik.
func (uc *GetSuratDownloadUseCase) ensureScopeAccess(ctx context.Context, userID string, scopeID *string) error {
	if uc.scopeReader == nil {
		return nil
	}
	access, err := uc.scopeReader.GetSuratScopeAccess(ctx, userID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	if access.HasFullAccess || scopeID == nil {
		return nil
	}
	allowed := make(map[string]struct{}, len(access.AllowedScopeIDs))
	for _, id := range access.AllowedScopeIDs {
		allowed[id] = struct{}{}
	}
	if _, ok := allowed[*scopeID]; !ok {
		return kernel.New(suratconstant.CodeSuratNotFound)
	}
	return nil
}
