package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	regConst "sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type GetMyHerregistrasiDetailUseCase struct {
	kesantrianReader  ports.KesantrianReader
	periodRepo        periodRepo.AcademicPeriodRepository
	registrationRepo  regRepo.SantriRegistrationRepository
	requirementRepo   reqRepo.HerregistrasiDocumentRequirementRepository
	documentRepo      docRepo.HerregistrasiDocumentRepository
}

func NewGetMyHerregistrasiDetailUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	registrationRepo regRepo.SantriRegistrationRepository,
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository,
	documentRepo docRepo.HerregistrasiDocumentRepository,
) *GetMyHerregistrasiDetailUseCase {
	return &GetMyHerregistrasiDetailUseCase{
		kesantrianReader: kesantrianReader,
		periodRepo:       periodRepo,
		registrationRepo: registrationRepo,
		requirementRepo:  requirementRepo,
		documentRepo:     documentRepo,
	}
}

func (uc *GetMyHerregistrasiDetailUseCase) Execute(ctx context.Context, userID string) (*dto.MyHerregistrasiDetailResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	resp := &dto.MyHerregistrasiDetailResponse{
		Requirements: []dto.HerregistrasiDocumentRequirementResponse{},
		Documents:    []dto.HerregistrasiDocumentResponse{},
	}

	period, err := uc.periodRepo.FindOpen(ctx)
	if err != nil {
		if application.IsNotFoundErr(err, application.PeriodNotFoundCode) {
			return resp, nil
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	resp.AcademicPeriod = command.MapAcademicPeriodToResponse(period)

	labels := map[string]string{}
	if requirements, err := uc.requirementRepo.FindByAcademicPeriod(ctx, period.ID); err == nil {
		for _, req := range requirements {
			resp.Requirements = append(resp.Requirements, *command.MapHerregistrasiDocumentRequirementToResponse(req))
			labels[req.Kind] = req.Label
		}
	}

	reg, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, info.SantriID, period.ID)
	if err != nil && !application.IsNotFoundErr(err, regConst.CodeSantriRegistrationNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if reg != nil {
		regResp := command.MapSantriRegistrationToResponse(reg)
		regResp.PeriodName = period.Name
		regResp.SantriNIS = info.NIS
		regResp.SantriName = info.Fullname
		resp.Registration = regResp

		if docs, err := uc.documentRepo.FindByRegistration(ctx, reg.ID); err == nil {
			for _, doc := range docs {
				resp.Documents = append(resp.Documents, *command.MapHerregistrasiDocumentToResponse(doc, labels[doc.Kind]))
			}
		}
	}
	return resp, nil
}
