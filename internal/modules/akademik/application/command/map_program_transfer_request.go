package command

import (
	"sipon-be/internal/modules/akademik/application/dto"
	progEntity "sipon-be/internal/modules/akademik/domain/program/entity"
	ptrEntity "sipon-be/internal/modules/akademik/domain/program_transfer_request/entity"
	"sipon-be/internal/shared/timeutil"
)

// MapProgramTransferRequestToResponse memetakan entity request ke response DTO,
// melengkapi nama program asal/tujuan dan nama santri bila tersedia.
func MapProgramTransferRequestToResponse(
	req *ptrEntity.ProgramTransferRequest,
	fromProgram, toProgram *progEntity.Program,
	santriName *string,
) *dto.ProgramTransferRequestResponse {
	resp := &dto.ProgramTransferRequestResponse{
		ID:            req.ID,
		SantriID:      req.SantriID,
		SantriName:    santriName,
		FromProgramID: req.FromProgramID,
		ToProgramID:   req.ToProgramID,
		Status:        string(req.Status),
		Notes:         req.Notes,
		AdminNotes:    req.AdminNotes,
		ReviewedBy:    req.ReviewedBy,
		ReviewedAt:    timeutil.ToPlatformPtr(req.ReviewedAt),
		CreatedAt:     timeutil.ToPlatform(req.CreatedAt),
	}
	if fromProgram != nil {
		resp.FromProgram = &dto.ProgramBrief{ID: fromProgram.ID, Code: fromProgram.Code, Name: fromProgram.Name}
	}
	if toProgram != nil {
		resp.ToProgram = &dto.ProgramBrief{ID: toProgram.ID, Code: toProgram.Code, Name: toProgram.Name}
	}
	return resp
}
