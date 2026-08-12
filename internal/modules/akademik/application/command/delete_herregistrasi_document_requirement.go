package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/constant"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
)

type DeleteHerregistrasiDocumentRequirementUseCase struct {
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository
}

func NewDeleteHerregistrasiDocumentRequirementUseCase(requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository) *DeleteHerregistrasiDocumentRequirementUseCase {
	return &DeleteHerregistrasiDocumentRequirementUseCase{requirementRepo: requirementRepo}
}

func (uc *DeleteHerregistrasiDocumentRequirementUseCase) Execute(ctx context.Context, id string) error {
	if err := uc.requirementRepo.Delete(ctx, id); err != nil {
		return application.WrapRepoErr(err, constant.CodeHerregistrasiDocumentRequirementNotFound)
	}
	return nil
}
