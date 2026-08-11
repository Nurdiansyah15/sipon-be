package command

import (
	"context"
	"errors"
	"log/slog"

	"sipon-be/internal/modules/identity"
	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrientity "sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	santrivo "sipon-be/internal/modules/kesantrian/domain/santri/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type CreateSantriUseCase struct {
	santriRepo  santrirepo.SantriRepository
	provisioner ports.AccountProvisioner
	akademik    ports.AkademikProvisioner
	transactor  ports.Transactor
}

func NewCreateSantriUseCase(santriRepo santrirepo.SantriRepository, provisioner ports.AccountProvisioner, transactor ports.Transactor) *CreateSantriUseCase {
	return &CreateSantriUseCase{santriRepo: santriRepo, provisioner: provisioner, transactor: transactor}
}

// SetAkademikProvisioner late-binds the akademik port (see Module).
func (uc *CreateSantriUseCase) SetAkademikProvisioner(p ports.AkademikProvisioner) {
	uc.akademik = p
}

func (uc *CreateSantriUseCase) Execute(ctx context.Context, req dto.CreateSantriRequest) (*dto.CreateSantriResponse, error) {
	nis, err := santrivo.NewNIS(req.NIS)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	_, err = uc.santriRepo.FindByNIS(ctx, nis.String())
	if err == nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}
	var ke *kernel.AppError
	if !errors.As(err, &ke) || ke.Code != santriconstant.CodeSantriNotFound {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	email := nis.String() + "@santri.sipon"
	acc, err := uc.provisioner.CreateAccountWithNIS(ctx, identity.CreateAccountInput{
		Username: nis.String(),
		Email:    email,
		NISValue: nis.String(),
	})
	if err != nil {
		var ake *kernel.AppError
		if errors.As(err, &ake) {
			return nil, err
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	santri, err := santrientity.NewSantri(uuid.NewString(), acc.UserID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	santri.SetNIS(nis)

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.santriRepo.Save(txCtx, santri)
	}); err != nil {
		return nil, application.WrapConflictErr(err, santriconstant.CodeSantriDuplicate)
	}

	programID := req.ProgramID
	if programID == nil && uc.akademik != nil {
		defaultID, err := uc.akademik.GetDefaultProgramID(ctx)
		if err != nil {
			slog.Warn("kesantrian: gagal ambil default program dari akademik", "error", err)
		} else {
			programID = defaultID
		}
	}

	if programID != nil && uc.akademik != nil {
		if err := uc.akademik.AssignSantriProgram(ctx, santri.ID, *programID); err != nil {
			slog.Warn("kesantrian: best-effort assign santri program ke akademik gagal",
				"santri_id", santri.ID, "program_id", *programID, "error", err)
		}
	}

	return &dto.CreateSantriResponse{
		UserID:            acc.UserID,
		SantriID:          santri.ID,
		NIS:               nis.String(),
		GeneratedPassword: acc.GeneratedPassword,
	}, nil
}
