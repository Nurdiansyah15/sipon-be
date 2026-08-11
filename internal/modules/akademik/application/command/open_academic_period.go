package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/academic_period/constant"
	repo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/shared/kernel"
)

type OpenAcademicPeriodUseCase struct {
	periodRepo repo.AcademicPeriodRepository
}

func NewOpenAcademicPeriodUseCase(periodRepo repo.AcademicPeriodRepository) *OpenAcademicPeriodUseCase {
	return &OpenAcademicPeriodUseCase{periodRepo: periodRepo}
}

func (uc *OpenAcademicPeriodUseCase) Execute(ctx context.Context, id string) (*dto.AcademicPeriodResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAcademicPeriodNotFound)
	}
	if err := period.Open(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeAcademicPeriodInvalidStatus)
	}
	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapAcademicPeriodToResponse(period), nil
}
