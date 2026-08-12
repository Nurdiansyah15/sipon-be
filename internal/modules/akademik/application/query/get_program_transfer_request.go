package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	ptrConst "sipon-be/internal/modules/akademik/domain/program_transfer_request/constant"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
)

// GetProgramTransferRequestUseCase menampilkan detail request transfer program.
type GetProgramTransferRequestUseCase struct {
	ptrRepo          ptrRepo.ProgramTransferRequestRepository
	programRepo      progRepo.ProgramRepository
	kesantrianReader ports.KesantrianReader
}

func NewGetProgramTransferRequestUseCase(
	ptrRepo ptrRepo.ProgramTransferRequestRepository,
	programRepo progRepo.ProgramRepository,
	kesantrianReader ports.KesantrianReader,
) *GetProgramTransferRequestUseCase {
	return &GetProgramTransferRequestUseCase{ptrRepo: ptrRepo, programRepo: programRepo, kesantrianReader: kesantrianReader}
}

func (uc *GetProgramTransferRequestUseCase) Execute(ctx context.Context, id string) (*dto.ProgramTransferRequestResponse, error) {
	req, err := uc.ptrRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, ptrConst.CodeProgramTransferRequestNotFound)
	}

	fromProg, _ := uc.programRepo.FindByID(ctx, req.FromProgramID)
	toProg, _ := uc.programRepo.FindByID(ctx, req.ToProgramID)
	var santriName *string
	if info, err := uc.kesantrianReader.GetSantriByID(ctx, req.SantriID); err == nil && info != nil {
		santriName = info.Fullname
	}
	return command.MapProgramTransferRequestToResponse(req, fromProg, toProg, santriName), nil
}
