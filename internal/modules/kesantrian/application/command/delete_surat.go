package command

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"
)

type DeleteSuratUseCase struct {
	suratRepo   suratrepo.SuratRepository
	scopeReader ports.ScopeReader
}

func NewDeleteSuratUseCase(suratRepo suratrepo.SuratRepository, scopeReader ports.ScopeReader) *DeleteSuratUseCase {
	return &DeleteSuratUseCase{suratRepo: suratRepo, scopeReader: scopeReader}
}

func (uc *DeleteSuratUseCase) Execute(ctx context.Context, userID, id string) error {
	s, err := uc.suratRepo.FindByID(ctx, id)
	if err != nil {
		return application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	if err := uc.ensureScopeAccess(ctx, userID, s.ScopeID); err != nil {
		return err
	}

	if err := uc.suratRepo.Delete(ctx, id); err != nil {
		return application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}
	return nil
}

// ensureScopeAccess menolak akses bila surat membawa scope_id di luar scope
// yang boleh diakses user. Surat global (scope_id nil) publik.
func (uc *DeleteSuratUseCase) ensureScopeAccess(ctx context.Context, userID string, scopeID *string) error {
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
