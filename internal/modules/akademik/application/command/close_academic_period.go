package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/academic_period/constant"
	repo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	"sipon-be/internal/shared/kernel"
)

type CloseAcademicPeriodUseCase struct {
	periodRepo repo.AcademicPeriodRepository
}

func NewCloseAcademicPeriodUseCase(periodRepo repo.AcademicPeriodRepository) *CloseAcademicPeriodUseCase {
	return &CloseAcademicPeriodUseCase{periodRepo: periodRepo}
}

func (uc *CloseAcademicPeriodUseCase) Execute(ctx context.Context, id string) (*dto.AcademicPeriodResponse, error) {
	period, err := uc.periodRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAcademicPeriodNotFound)
	}
	if err := period.Close(); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeAcademicPeriodInvalidStatus)
	}
	if err := uc.periodRepo.Update(ctx, period); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapAcademicPeriodToResponse(period), nil
}
