package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	ptrConst "sipon-be/internal/modules/akademik/domain/program_transfer_request/constant"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
	spConst "sipon-be/internal/modules/akademik/domain/santri_program/constant"
	spEntity "sipon-be/internal/modules/akademik/domain/santri_program/entity"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

// ApproveProgramTransferUseCase menyetujui permintaan pindah program santri.
// Program lama di-deactivate dan program baru dibuat (atomic).
type ApproveProgramTransferUseCase struct {
	ptrRepo           ptrRepo.ProgramTransferRequestRepository
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
	transactor        ports.Transactor
}

func NewApproveProgramTransferUseCase(
	ptrRepo ptrRepo.ProgramTransferRequestRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	programRepo progRepo.ProgramRepository,
	transactor ports.Transactor,
) *ApproveProgramTransferUseCase {
	return &ApproveProgramTransferUseCase{
		ptrRepo:           ptrRepo,
		santriProgramRepo: santriProgramRepo,
		programRepo:       programRepo,
		transactor:        transactor,
	}
}

func (uc *ApproveProgramTransferUseCase) Execute(ctx context.Context, requestID, adminID string) (*dto.ProgramTransferRequestResponse, error) {
	transfer, err := uc.ptrRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, application.WrapRepoErr(err, ptrConst.CodeProgramTransferRequestNotFound)
	}
	if err := transfer.Approve(adminID); err != nil {
		return nil, application.WrapBadRequestErr(err, ptrConst.CodeProgramTransferRequestInvalidStatus)
	}

	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.ptrRepo.Update(txCtx, transfer); err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		if err := uc.santriProgramRepo.DeactivateAllBySantriID(txCtx, transfer.SantriID); err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		sp, err := spEntity.NewSantriProgram(uuid.NewString(), transfer.SantriID, transfer.ToProgramID)
		if err != nil {
			return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
		if err := uc.santriProgramRepo.Save(txCtx, sp); err != nil {
			return application.WrapConflictErr(err, spConst.CodeSantriProgramDuplicate)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	fromProg, _ := uc.programRepo.FindByID(ctx, transfer.FromProgramID)
	toProg, _ := uc.programRepo.FindByID(ctx, transfer.ToProgramID)
	return MapProgramTransferRequestToResponse(transfer, fromProg, toProg, nil), nil
}
