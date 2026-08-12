package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	regConst "sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type SubmitHerregistrasiUseCase struct {
	kesantrianReader ports.KesantrianReader
	periodRepo       periodRepo.AcademicPeriodRepository
	registrationRepo regRepo.SantriRegistrationRepository
	requirementRepo  reqRepo.HerregistrasiDocumentRequirementRepository
	documentRepo     docRepo.HerregistrasiDocumentRepository
}

func NewSubmitHerregistrasiUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	registrationRepo regRepo.SantriRegistrationRepository,
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository,
	documentRepo docRepo.HerregistrasiDocumentRepository,
) *SubmitHerregistrasiUseCase {
	return &SubmitHerregistrasiUseCase{
		kesantrianReader: kesantrianReader,
		periodRepo:       periodRepo,
		registrationRepo: registrationRepo,
		requirementRepo:  requirementRepo,
		documentRepo:     documentRepo,
	}
}

func (uc *SubmitHerregistrasiUseCase) Execute(ctx context.Context, userID string) (*dto.SantriRegistrationResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	period, err := uc.periodRepo.FindOpen(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, application.PeriodNotFoundCode)
	}

	reg, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, info.SantriID, period.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, regConst.CodeSantriRegistrationNotFound)
	}
	if reg.Status != regConst.SantriRegistrationStatusDraft {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	// Validasi: semua dokumen wajib sudah di-upload.
	if err := uc.validateRequiredDocuments(ctx, period.ID, reg.ID); err != nil {
		return nil, err
	}

	if err := reg.Submit(); err != nil {
		return nil, application.WrapBadRequestErr(err, regConst.CodeSantriRegistrationInvalidStatus)
	}
	if err := uc.registrationRepo.Update(ctx, reg); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	resp := MapSantriRegistrationToResponse(reg)
	resp.PeriodName = period.Name
	resp.SantriNIS = info.NIS
	resp.SantriName = info.Fullname
	return resp, nil
}

func (uc *SubmitHerregistrasiUseCase) validateRequiredDocuments(ctx context.Context, periodID, registrationID string) error {
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
	uploaded := map[string]bool{}
	for _, d := range docs {
		uploaded[d.Kind] = true
	}

	for _, req := range requirements {
		if req.IsRequired && !uploaded[req.Kind] {
			return kernel.New(regConst.CodeSantriRegistrationMissingDocuments)
		}
	}
	return nil
}
