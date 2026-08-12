package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	ptrRepo "sipon-be/internal/modules/akademik/domain/program_transfer_request/repository"
	"sipon-be/internal/shared/kernel"
)

// ListMyProgramTransferRequestsUseCase menampilkan daftar request transfer
// program milik santri yang sedang login.
type ListMyProgramTransferRequestsUseCase struct {
	ptrRepo          ptrRepo.ProgramTransferRequestRepository
	programRepo      progRepo.ProgramRepository
	kesantrianReader ports.KesantrianReader
}

func NewListMyProgramTransferRequestsUseCase(
	ptrRepo ptrRepo.ProgramTransferRequestRepository,
	programRepo progRepo.ProgramRepository,
	kesantrianReader ports.KesantrianReader,
) *ListMyProgramTransferRequestsUseCase {
	return &ListMyProgramTransferRequestsUseCase{ptrRepo: ptrRepo, programRepo: programRepo, kesantrianReader: kesantrianReader}
}

func (uc *ListMyProgramTransferRequestsUseCase) Execute(ctx context.Context, userID string) ([]dto.ProgramTransferRequestResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	result, err := uc.ptrRepo.List(ctx, ptrRepo.ProgramTransferRequestListQuery{
		SantriID: &info.SantriID,
		Page:     1,
		Limit:    50,
	})
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.ProgramTransferRequestResponse, 0, len(result.Items))
	for _, req := range result.Items {
		fromProg, _ := uc.programRepo.FindByID(ctx, req.FromProgramID)
		toProg, _ := uc.programRepo.FindByID(ctx, req.ToProgramID)
		items = append(items, *command.MapProgramTransferRequestToResponse(req, fromProg, toProg, info.Fullname))
	}
	return items, nil
}
