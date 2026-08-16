package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/kesantrian/application"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	"sipon-be/internal/modules/kesantrian/domain/surat/entity"
	suratrepo "sipon-be/internal/modules/kesantrian/domain/surat/repository"
	"sipon-be/internal/modules/kesantrian/domain/surat/service"
	tiperepo "sipon-be/internal/modules/kesantrian/domain/tipe_surat/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type CreateSuratUseCase struct {
	suratRepo      suratrepo.SuratRepository
	tipeSuratRepo  tiperepo.TipeSuratRepository
	nomorGenerator *service.NomorGenerator
	transactor     ports.Transactor
	scopeReader    ports.ScopeReader
}

func NewCreateSuratUseCase(
	suratRepo suratrepo.SuratRepository,
	tipeSuratRepo tiperepo.TipeSuratRepository,
	nomorGenerator *service.NomorGenerator,
	transactor ports.Transactor,
	scopeReader ports.ScopeReader,
) *CreateSuratUseCase {
	return &CreateSuratUseCase{
		suratRepo:      suratRepo,
		tipeSuratRepo:  tipeSuratRepo,
		nomorGenerator: nomorGenerator,
		transactor:     transactor,
		scopeReader:    scopeReader,
	}
}

func (uc *CreateSuratUseCase) Execute(ctx context.Context, createdBy string, tipeSuratID string, keterangan *string, tanggal string, dokumenAsetIDs []string) (*entity.Surat, error) {
	tipe, err := uc.tipeSuratRepo.FindByID(ctx, tipeSuratID)
	if err != nil {
		return nil, application.WrapRepoErr(err, "TIPE_SURAT_NOT_FOUND")
	}

	tgl, err := time.Parse("2006-01-02", tanggal)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	// Auto-fill scope_id dari role scope user pembuat. Nilai scope_id diambil
	// dari master scope yang boleh diakses user; bila user punya akses penuh
	// atau tanpa akses scope, scope_id dibiarkan kosong (global/publik).
	scopeID, err := uc.resolveScopeID(ctx, createdBy)
	if err != nil {
		return nil, err
	}

	var created *entity.Surat

	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		nomor, seq, err := uc.nomorGenerator.Generate(txCtx, tipe.Kode, scopeID, int(tgl.Month()), tgl.Year())
		if err != nil {
			return err
		}

		s, err := entity.NewSurat(uuid.NewString(), nomor, seq, tipeSuratID, keterangan, tgl, createdBy, scopeID)
		if err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}

		if err := uc.suratRepo.Save(txCtx, s); err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}

		for _, dokID := range dokumenAsetIDs {
			link := &entity.SuratDokumenAset{
				ID:            uuid.NewString(),
				SuratID:       s.ID,
				DokumenAsetID: dokID,
				CreatedAt:     time.Now(),
			}
			if err := uc.suratRepo.SaveDokumenLink(txCtx, link); err != nil {
				return kernel.Wrap(application.ErrCodeInternal, err)
			}
		}

		created = s
		return nil
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

// resolveScopeID menetapkan scope_id surat berdasarkan akses scope user.
// - Tidak ada akses scope -> nil (global/publik).
// - Akses penuh (HasFullAccess) -> nil (global/publik).
// - Akses terbatas ke tepat satu scope -> scope_id master tersebut.
func (uc *CreateSuratUseCase) resolveScopeID(ctx context.Context, userID string) (*string, error) {
	if uc.scopeReader == nil {
		return nil, nil
	}
	access, err := uc.scopeReader.GetSuratScopeAccess(ctx, userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if !access.HasAccess || access.HasFullAccess {
		return nil, nil
	}
	if len(access.AllowedScopeIDs) == 0 {
		return nil, nil
	}
	scopeID := access.AllowedScopeIDs[0]
	return &scopeID, nil
}
