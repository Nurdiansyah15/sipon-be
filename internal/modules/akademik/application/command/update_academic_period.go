package command

import (
	"context"
	"time"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/academic_period/constant"
	repo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateAcademicPeriodUseCase struct {
	periodRepo repo.AcademicPeriodRepository
}

func NewUpdateAcademicPeriodUseCase(periodRepo repo.AcademicPeriodRepository) *UpdateAcademicPeriodUseCase {
	return &UpdateAcademicPeriodUseCase{periodRepo: periodRepo}
}

func (uc *UpdateAcademicPeriodUseCase) Execute(ctx context.Context, id string, req dto.UpdateAcademicPeriodRequest) (*dto.AcademicPeriodResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAcademicPeriodNotFound)
	}

	if req.Code != nil && *req.Code != "" && *req.Code != period.Code {
		existing, _ := uc.periodRepo.FindByCode(ctx, *req.Code)
		if existing != nil && existing.ID != id {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		period.Code = *req.Code
	}

	name := ""
	if req.Name != nil {
		name = *req.Name
	}
	var startDate, endDate *time.Time
	if req.StartDate != nil {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, kernel.New(application.ErrCodeBadRequest)
		}
		startDate = &t
	}
	if req.EndDate != nil {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, kernel.New(application.ErrCodeBadRequest)
		}
		endDate = &t
	}

	if err := period.Update(name, startDate, endDate); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeAcademicPeriodInvalidRange)
	}

	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAcademicPeriodNotFound)
	}
	return MapAcademicPeriodToResponse(period), nil
}
