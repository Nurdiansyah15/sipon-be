package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type ListHerregistrasiDocumentsUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
	requirementRepo  reqRepo.HerregistrasiDocumentRequirementRepository
	documentRepo     docRepo.HerregistrasiDocumentRepository
}

func NewListHerregistrasiDocumentsUseCase(
	registrationRepo regRepo.SantriRegistrationRepository,
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository,
	documentRepo docRepo.HerregistrasiDocumentRepository,
) *ListHerregistrasiDocumentsUseCase {
	return &ListHerregistrasiDocumentsUseCase{
		registrationRepo: registrationRepo,
		requirementRepo:  requirementRepo,
		documentRepo:     documentRepo,
	}
}

func (uc *ListHerregistrasiDocumentsUseCase) Execute(ctx context.Context, registrationID string) ([]dto.HerregistrasiDocumentResponse, error) {
	reg, err := uc.registrationRepo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, application.WrapRepoErr(err, "SANTRI_REGISTRATION_NOT_FOUND")
	}

	labels := map[string]string{}
	if requirements, err := uc.requirementRepo.FindByAcademicPeriod(ctx, reg.AcademicPeriodID); err == nil {
		for _, req := range requirements {
			labels[req.Kind] = req.Label
		}
	}

	docs, err := uc.documentRepo.FindByRegistration(ctx, registrationID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	items := make([]dto.HerregistrasiDocumentResponse, 0, len(docs))
	for _, doc := range docs {
		items = append(items, *command.MapHerregistrasiDocumentToResponse(doc, labels[doc.Kind]))
	}
	return items, nil
}
