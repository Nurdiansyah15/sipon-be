package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/shared/kernel"
)

type GetAcademicPeriodUseCase struct {
	periodRepo repository.AcademicPeriodRepository
}

func NewGetAcademicPeriodUseCase(periodRepo repository.AcademicPeriodRepository) *GetAcademicPeriodUseCase {
	return &GetAcademicPeriodUseCase{periodRepo: periodRepo}
}

func (uc *GetAcademicPeriodUseCase) Execute(ctx context.Context, id string) (*dto.AcademicPeriodResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, kernel.Code("ACADEMIC_PERIOD_NOT_FOUND"))
	}
	return command.MapAcademicPeriodToResponse(period), nil
}
