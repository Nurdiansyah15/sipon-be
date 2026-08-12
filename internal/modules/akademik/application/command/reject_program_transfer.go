package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	ptrConst "sipon-be/internal/modules/akademik/domain/program_transfer_request/constant"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
)

// RejectProgramTransferUseCase menolak permintaan pindah program santri.
// Program santri tidak berubah.
type RejectProgramTransferUseCase struct {
	ptrRepo     ptrRepo.ProgramTransferRequestRepository
	programRepo progRepo.ProgramRepository
}

func NewRejectProgramTransferUseCase(
	ptrRepo ptrRepo.ProgramTransferRequestRepository,
	programRepo progRepo.ProgramRepository,
) *RejectProgramTransferUseCase {
	return &RejectProgramTransferUseCase{ptrRepo: ptrRepo, programRepo: programRepo}
}

func (uc *RejectProgramTransferUseCase) Execute(ctx context.Context, requestID, adminID string, req dto.RejectProgramTransferRequest) (*dto.ProgramTransferRequestResponse, error) {
	transfer, err := uc.ptrRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, application.WrapRepoErr(err, ptrConst.CodeProgramTransferRequestNotFound)
	}
	if err := transfer.Reject(adminID, req.AdminNotes); err != nil {
		return nil, application.WrapBadRequestErr(err, ptrConst.CodeProgramTransferRequestInvalidStatus)
	}
	if err := uc.ptrRepo.Update(ctx, transfer); err != nil {
		return nil, application.WrapRepoErr(err, ptrConst.CodeProgramTransferRequestNotFound)
	}

	fromProg, _ := uc.programRepo.FindByID(ctx, transfer.FromProgramID)
	toProg, _ := uc.programRepo.FindByID(ctx, transfer.ToProgramID)
	return MapProgramTransferRequestToResponse(transfer, fromProg, toProg, nil), nil
}
