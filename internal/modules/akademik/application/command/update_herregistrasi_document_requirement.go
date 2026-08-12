package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/constant"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateHerregistrasiDocumentRequirementUseCase struct {
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository
}

func NewUpdateHerregistrasiDocumentRequirementUseCase(requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository) *UpdateHerregistrasiDocumentRequirementUseCase {
	return &UpdateHerregistrasiDocumentRequirementUseCase{requirementRepo: requirementRepo}
}

func (uc *UpdateHerregistrasiDocumentRequirementUseCase) Execute(ctx context.Context, id string, req dto.UpdateHerregistrasiDocumentRequirementRequest) (*dto.HerregistrasiDocumentRequirementResponse, error) {
	requirement, err := uc.requirementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeHerregistrasiDocumentRequirementNotFound)
	}
	if err := requirement.Update(req.Label, req.IsRequired, req.Description); err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}
	if err := uc.requirementRepo.Update(ctx, requirement); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapHerregistrasiDocumentRequirementToResponse(requirement), nil
}
