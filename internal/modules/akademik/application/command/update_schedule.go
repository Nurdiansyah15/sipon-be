package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateScheduleUseCase struct {
	scheduleRepo schRepo.ActivityScheduleRepository
	transactor   ports.Transactor
}

func NewUpdateScheduleUseCase(scheduleRepo schRepo.ActivityScheduleRepository, transactor ports.Transactor) *UpdateScheduleUseCase {
	return &UpdateScheduleUseCase{scheduleRepo: scheduleRepo, transactor: transactor}
}

func (uc *UpdateScheduleUseCase) Execute(ctx context.Context, id string, req dto.UpdateScheduleRequest) (*dto.ActivityScheduleDetailResponse, error) {
	schedule, err := uc.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivityScheduleNotFound)
	}

	startDate, err := parseDatePtr(req.StartDate)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	endDate, err := parseDatePtr(req.EndDate)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	startTime := ""
	if req.StartTime != nil {
		startTime = *req.StartTime
	}
	endTime := ""
	if req.EndTime != nil {
		endTime = *req.EndTime
	}
	if err := schedule.Update(startTime, endTime, startDate, endDate, req.EarlyMinutes, req.LateMinutes); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivityScheduleInvalid)
	}

	var resp *dto.ActivityScheduleDetailResponse
	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.scheduleRepo.Update(txCtx, schedule); err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
		if err := validateScheduleDetail(string(schedule.Type), req.WeeklyDays, req.MonthlyDays, req.YearlyDates); err != nil {
			return kernel.New(application.ErrCodeUnprocessableEntity)
		}
		if err := replaceScheduleDetails(txCtx, uc.scheduleRepo, schedule.ID, dto.CreateScheduleRequest{
			Type:        string(schedule.Type),
			WeeklyDays:  req.WeeklyDays,
			MonthlyDays: req.MonthlyDays,
			YearlyDates: req.YearlyDates,
		}); err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
		resp = MapScheduleToDetailResponse(schedule)
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}
