package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	"sipon-be/internal/shared/kernel"
)

type ListHerregistrasiDocumentRequirementsUseCase struct {
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository
}

func NewListHerregistrasiDocumentRequirementsUseCase(requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository) *ListHerregistrasiDocumentRequirementsUseCase {
	return &ListHerregistrasiDocumentRequirementsUseCase{requirementRepo: requirementRepo}
}

func (uc *ListHerregistrasiDocumentRequirementsUseCase) Execute(ctx context.Context, periodID string) ([]dto.HerregistrasiDocumentRequirementResponse, error) {
	requirements, err := uc.requirementRepo.FindByAcademicPeriod(ctx, periodID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	items := make([]dto.HerregistrasiDocumentRequirementResponse, 0, len(requirements))
	for _, req := range requirements {
		items = append(items, *command.MapHerregistrasiDocumentRequirementToResponse(req))
	}
	return items, nil
}
