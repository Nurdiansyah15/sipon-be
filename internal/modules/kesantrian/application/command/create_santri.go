package command

import (
	"context"
	"errors"

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
	transactor  ports.Transactor
}

func NewCreateSantriUseCase(santriRepo santrirepo.SantriRepository, provisioner ports.AccountProvisioner, transactor ports.Transactor) *CreateSantriUseCase {
	return &CreateSantriUseCase{santriRepo: santriRepo, provisioner: provisioner, transactor: transactor}
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

	return &dto.CreateSantriResponse{
		UserID:            acc.UserID,
		SantriID:          santri.ID,
		NIS:               nis.String(),
		GeneratedPassword: acc.GeneratedPassword,
	}, nil
}
