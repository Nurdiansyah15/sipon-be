package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	dokumenconstant "sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	dokumenrepo "sipon-be/internal/modules/kesantrian/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"
)

type DokumenVerifyUseCase struct {
	dokumenRepo dokumenrepo.SantriDokumenRepository
	transactor  ports.Transactor
}

func NewDokumenVerifyUseCase(dokumenRepo dokumenrepo.SantriDokumenRepository, transactor ports.Transactor) *DokumenVerifyUseCase {
	return &DokumenVerifyUseCase{dokumenRepo: dokumenRepo, transactor: transactor}
}

func (uc *DokumenVerifyUseCase) Execute(ctx context.Context, verifierID, dokumenID string) (*dto.MessageResponse, error) {
	dokumen, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dokumenconstant.CodeDokumenNotFound)
	}

	if err := dokumen.Verify(verifierID); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == dokumenconstant.CodeDokumenInvalidStatus {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.dokumenRepo.Update(txCtx, dokumen)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "dokumen berhasil diverifikasi"}, nil
}

type DokumenRejectUseCase struct {
	dokumenRepo dokumenrepo.SantriDokumenRepository
	transactor  ports.Transactor
}

func NewDokumenRejectUseCase(dokumenRepo dokumenrepo.SantriDokumenRepository, transactor ports.Transactor) *DokumenRejectUseCase {
	return &DokumenRejectUseCase{dokumenRepo: dokumenRepo, transactor: transactor}
}

func (uc *DokumenRejectUseCase) Execute(ctx context.Context, verifierID, dokumenID string, req dto.RejectDokumenRequest) (*dto.MessageResponse, error) {
	dokumen, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dokumenconstant.CodeDokumenNotFound)
	}

	if err := dokumen.Reject(verifierID, req.Notes); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.dokumenRepo.Update(txCtx, dokumen)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "dokumen berhasil ditolak"}, nil
}
