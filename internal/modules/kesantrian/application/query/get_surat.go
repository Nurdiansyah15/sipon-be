package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
	"sipon-be/internal/shared/kernel"
)

type GetSuratUseCase struct {
	suratRepo     suratrepo.SuratRepository
	tipeSuratRepo tiperepo.TipeSuratRepository
	scopeReader   ports.ScopeReader
}

func NewGetSuratUseCase(suratRepo suratrepo.SuratRepository, tipeSuratRepo tiperepo.TipeSuratRepository, scopeReader ports.ScopeReader) *GetSuratUseCase {
	return &GetSuratUseCase{suratRepo: suratRepo, tipeSuratRepo: tipeSuratRepo, scopeReader: scopeReader}
}

func (uc *GetSuratUseCase) Execute(ctx context.Context, userID, id string) (*dto.SuratDetailResponse, error) {
	detail, err := uc.suratRepo.FindDetail(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	if err := uc.ensureScopeAccess(ctx, userID, detail.Surat.ScopeID); err != nil {
		return nil, err
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
		ScopeID:        s.ScopeID,
		CreatedAt:      s.CreatedAt,
		DokumenAsetIDs: dokIDs,
	}, nil
}

// ensureScopeAccess menolak akses bila surat membawa scope_id di luar scope
// yang boleh diakses user. Surat global (scope_id nil) publik.
func (uc *GetSuratUseCase) ensureScopeAccess(ctx context.Context, userID string, scopeID *string) error {
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
