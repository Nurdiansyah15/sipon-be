package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/kesantrian/application"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	suratconstant "sipon-be/internal/modules/kesantrian/domain/surat/constant"
	"sipon-be/internal/modules/kesantrian/domain/surat/entity"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AddSuratDokumenUseCase struct {
	suratRepo   suratrepo.SuratRepository
	scopeReader ports.ScopeReader
}

func NewAddSuratDokumenUseCase(suratRepo suratrepo.SuratRepository, scopeReader ports.ScopeReader) *AddSuratDokumenUseCase {
	return &AddSuratDokumenUseCase{suratRepo: suratRepo, scopeReader: scopeReader}
}

func (uc *AddSuratDokumenUseCase) Execute(ctx context.Context, userID, suratID, dokumenAsetID string) error {
	s, err := uc.suratRepo.FindByID(ctx, suratID)
	if err != nil {
		return application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	if err := uc.ensureScopeAccess(ctx, userID, s.ScopeID); err != nil {
		return err
	}

	link := &entity.SuratDokumenAset{
		ID:            uuid.NewString(),
		SuratID:       suratID,
		DokumenAsetID: dokumenAsetID,
		CreatedAt:     time.Now(),
	}

	if err := uc.suratRepo.SaveDokumenLink(ctx, link); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == suratconstant.CodeSuratDokumenExists {
			return kernel.Wrap(application.ErrCodeConflict, err)
		}
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	return nil
}

func (uc *AddSuratDokumenUseCase) ensureScopeAccess(ctx context.Context, userID string, scopeID *string) error {
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

type RemoveSuratDokumenUseCase struct {
	suratRepo   suratrepo.SuratRepository
	scopeReader ports.ScopeReader
}

func NewRemoveSuratDokumenUseCase(suratRepo suratrepo.SuratRepository, scopeReader ports.ScopeReader) *RemoveSuratDokumenUseCase {
	return &RemoveSuratDokumenUseCase{suratRepo: suratRepo, scopeReader: scopeReader}
}

func (uc *RemoveSuratDokumenUseCase) Execute(ctx context.Context, userID, suratID, dokumenAsetID string) error {
	s, err := uc.suratRepo.FindByID(ctx, suratID)
	if err != nil {
		return application.WrapRepoErr(err, suratconstant.CodeSuratNotFound)
	}

	if err := uc.ensureScopeAccess(ctx, userID, s.ScopeID); err != nil {
		return err
	}

	if err := uc.suratRepo.DeleteDokumenLink(ctx, suratID, dokumenAsetID); err != nil {
		return application.WrapRepoErr(err, suratconstant.CodeSuratDokumenNotFound)
	}
	return nil
}

func (uc *RemoveSuratDokumenUseCase) ensureScopeAccess(ctx context.Context, userID string, scopeID *string) error {
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
