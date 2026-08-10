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
}

func NewCreateSuratUseCase(
	suratRepo suratrepo.SuratRepository,
	tipeSuratRepo tiperepo.TipeSuratRepository,
	nomorGenerator *service.NomorGenerator,
	transactor ports.Transactor,
) *CreateSuratUseCase {
	return &CreateSuratUseCase{
		suratRepo:      suratRepo,
		tipeSuratRepo:  tipeSuratRepo,
		nomorGenerator: nomorGenerator,
		transactor:     transactor,
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

	var created *entity.Surat

	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		nomor, seq, err := uc.nomorGenerator.Generate(txCtx, tipe.Kode, int(tgl.Month()), tgl.Year())
		if err != nil {
			return err
		}

		s, err := entity.NewSurat(uuid.NewString(), nomor, seq, tipeSuratID, keterangan, tgl, createdBy)
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
