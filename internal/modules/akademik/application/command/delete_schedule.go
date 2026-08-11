package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
)

type DeleteScheduleUseCase struct {
	scheduleRepo schRepo.ActivityScheduleRepository
}

func NewDeleteScheduleUseCase(scheduleRepo schRepo.ActivityScheduleRepository) *DeleteScheduleUseCase {
	return &DeleteScheduleUseCase{scheduleRepo: scheduleRepo}
}

func (uc *DeleteScheduleUseCase) Execute(ctx context.Context, id string) error {
	schedule, err := uc.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return application.WrapRepoErr(err, constant.CodeActivityScheduleNotFound)
	}
	schedule.SoftDelete()
	if err := uc.scheduleRepo.Update(ctx, schedule); err != nil {
		return application.WrapRepoErr(err, constant.CodeActivityScheduleNotFound)
	}
	return nil
}
