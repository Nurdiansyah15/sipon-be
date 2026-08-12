package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	docConst "sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	docEntity "sipon-be/internal/modules/akademik/domain/herregistrasi_document/entity"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	"sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type CompleteRegistrationUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
	requirementRepo  reqRepo.HerregistrasiDocumentRequirementRepository
	documentRepo     docRepo.HerregistrasiDocumentRepository
}

func NewCompleteRegistrationUseCase(
	registrationRepo regRepo.SantriRegistrationRepository,
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository,
	documentRepo docRepo.HerregistrasiDocumentRepository,
) *CompleteRegistrationUseCase {
	return &CompleteRegistrationUseCase{registrationRepo: registrationRepo, requirementRepo: requirementRepo, documentRepo: documentRepo}
}

func (uc *CompleteRegistrationUseCase) Execute(ctx context.Context, id string) (*dto.SantriRegistrationResponse, error) {
	registration, err := uc.registrationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeSantriRegistrationNotFound)
	}

	// Validasi: semua dokumen wajib harus sudah terverifikasi.
	if err := uc.validateRequiredDocuments(ctx, registration.AcademicPeriodID, registration.ID); err != nil {
		return nil, err
	}

	if err := registration.Complete(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeSantriRegistrationInvalidStatus)
	}
	if err := uc.registrationRepo.Update(ctx, registration); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapSantriRegistrationToResponse(registration), nil
}

func (uc *CompleteRegistrationUseCase) validateRequiredDocuments(ctx context.Context, periodID, registrationID string) error {
	requirements, err := uc.requirementRepo.FindByAcademicPeriod(ctx, periodID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	if len(requirements) == 0 {
		return nil
	}

	docs, err := uc.documentRepo.FindByRegistration(ctx, registrationID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	docByKind := map[string]*docEntity.HerregistrasiDocument{}
	for _, d := range docs {
		docByKind[d.Kind] = d
	}

	for _, req := range requirements {
		if !req.IsRequired {
			continue
		}
		doc, ok := docByKind[req.Kind]
		if !ok {
			return kernel.New(constant.CodeSantriRegistrationMissingDocuments)
		}
		if doc.Status != docConst.HerregistrasiDocumentStatusVerified {
			return kernel.New(constant.CodeSantriRegistrationMissingDocuments)
		}
	}
	return nil
}
