package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/constant"
	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/entity"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	"sipon-be/internal/shared/timeutil"
)

type CreateHerregistrasiDocumentRequirementUseCase struct {
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository
	periodRepo      periodRepo.AcademicPeriodRepository
}

func NewCreateHerregistrasiDocumentRequirementUseCase(
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository,
	periodRepo periodRepo.AcademicPeriodRepository,
) *CreateHerregistrasiDocumentRequirementUseCase {
	return &CreateHerregistrasiDocumentRequirementUseCase{requirementRepo: requirementRepo, periodRepo: periodRepo}
}

func (uc *CreateHerregistrasiDocumentRequirementUseCase) Execute(ctx context.Context, periodID string, req dto.CreateHerregistrasiDocumentRequirementRequest) (*dto.HerregistrasiDocumentRequirementResponse, error) {
	if _, err := uc.periodRepo.FindByID(ctx, periodID); err != nil {
		return nil, application.WrapRepoErr(err, application.PeriodNotFoundCode)
	}

	isRequired := true
	if req.IsRequired != nil {
		isRequired = *req.IsRequired
	}

	requirement, err := entity.NewHerregistrasiDocumentRequirement(
		uuid.NewString(), periodID, req.Kind, req.Label, isRequired, req.Description,
	)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeHerregistrasiDocumentRequirementInvalidKind)
	}
	if err := uc.requirementRepo.Save(ctx, requirement); err != nil {
		return nil, application.WrapConflictErr(err, constant.CodeHerregistrasiDocumentRequirementDuplicate)
	}
	return MapHerregistrasiDocumentRequirementToResponse(requirement), nil
}

func MapHerregistrasiDocumentRequirementToResponse(r *entity.HerregistrasiDocumentRequirement) *dto.HerregistrasiDocumentRequirementResponse {
	return &dto.HerregistrasiDocumentRequirementResponse{
		ID:               r.ID,
		AcademicPeriodID: r.AcademicPeriodID,
		Kind:             r.Kind,
		Label:            r.Label,
		IsRequired:       r.IsRequired,
		Description:      r.Description,
		CreatedAt:        timeutil.ToPlatform(r.CreatedAt),
		UpdatedAt:        timeutil.ToPlatform(r.UpdatedAt),
	}
}
