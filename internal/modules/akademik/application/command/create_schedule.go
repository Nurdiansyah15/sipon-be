package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

type CreateScheduleUseCase struct {
	scheduleRepo       schRepo.ActivityScheduleRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	transactor         ports.Transactor
}

func NewCreateScheduleUseCase(
	scheduleRepo schRepo.ActivityScheduleRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	transactor ports.Transactor,
) *CreateScheduleUseCase {
	return &CreateScheduleUseCase{scheduleRepo: scheduleRepo, activityPeriodRepo: activityPeriodRepo, transactor: transactor}
}

func (uc *CreateScheduleUseCase) Execute(ctx context.Context, req dto.CreateScheduleRequest) (*dto.ActivityScheduleDetailResponse, error) {
	if _, err := uc.activityPeriodRepo.FindByID(ctx, req.ActivityPeriodID); err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivityScheduleInvalid)
	}

	startDate, err := parseDatePtr(req.StartDate)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	endDate, err := parseDatePtr(req.EndDate)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	if err := validateScheduleDetail(req.Type, req.WeeklyDays, req.MonthlyDays, req.YearlyDates); err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	schedule, err := entity.NewActivitySchedule(uuid.NewString(), req.ActivityPeriodID,
		constant.ActivityScheduleType(req.Type), req.StartTime, req.EndTime, startDate, endDate)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeActivityScheduleInvalid)
	}

	var resp *dto.ActivityScheduleDetailResponse
	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.scheduleRepo.Save(txCtx, schedule); err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
		if err := replaceScheduleDetails(txCtx, uc.scheduleRepo, schedule.ID, req); err != nil {
			return err
		}
		resp = MapScheduleToDetailResponse(schedule)
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func validateScheduleDetail(typ string, weekly []string, monthly []int, yearly []dto.YearlyDateIn) error {
	switch constant.ActivityScheduleType(typ) {
	case constant.ActivityScheduleTypeWeekly:
		if len(weekly) == 0 {
			return kernel.New(constant.CodeActivityScheduleInvalid)
		}
		for _, d := range weekly {
			if !constant.IsValidDayOfWeek(d) {
				return kernel.New(constant.CodeActivityScheduleInvalid)
			}
		}
	case constant.ActivityScheduleTypeMonthly:
		if len(monthly) == 0 {
			return kernel.New(constant.CodeActivityScheduleInvalid)
		}
		for _, d := range monthly {
			if d < 1 || d > 31 {
				return kernel.New(constant.CodeActivityScheduleInvalid)
			}
		}
	case constant.ActivityScheduleTypeYearly:
		if len(yearly) == 0 {
			return kernel.New(constant.CodeActivityScheduleInvalid)
		}
		for _, d := range yearly {
			if d.Month < 1 || d.Month > 12 || d.Day < 1 || d.Day > 31 {
				return kernel.New(constant.CodeActivityScheduleInvalid)
			}
		}
	}
	return nil
}

func replaceScheduleDetails(ctx context.Context, scheduleRepo schRepo.ActivityScheduleRepository, scheduleID string, req dto.CreateScheduleRequest) error {
	switch constant.ActivityScheduleType(req.Type) {
	case constant.ActivityScheduleTypeWeekly:
		days := make([]constant.DayOfWeek, len(req.WeeklyDays))
		for i, d := range req.WeeklyDays {
			days[i] = constant.DayOfWeek(d)
		}
		return scheduleRepo.ReplaceWeeklies(ctx, scheduleID, days)
	case constant.ActivityScheduleTypeMonthly:
		return scheduleRepo.ReplaceMonthlies(ctx, scheduleID, req.MonthlyDays)
	case constant.ActivityScheduleTypeYearly:
		dates := make([]entity.YearlyDate, len(req.YearlyDates))
		for i, d := range req.YearlyDates {
			dates[i] = entity.YearlyDate{Month: d.Month, Day: d.Day}
		}
		return scheduleRepo.ReplaceYearlies(ctx, scheduleID, dates)
	}
	return nil
}

func parseDatePtr(v *string) (*time.Time, error) {
	return timeutil.ParseDatePtr(v)
}

func MapScheduleToDetailResponse(s *entity.ActivitySchedule) *dto.ActivityScheduleDetailResponse {
	resp := &dto.ActivityScheduleDetailResponse{
		ID:               s.ID,
		ActivityPeriodID: s.ActivityPeriodID,
		Type:             string(s.Type),
		StartTime:        s.StartTime,
		EndTime:          s.EndTime,
		CreatedAt:        timeutil.ToPlatform(s.CreatedAt),
		UpdatedAt:        timeutil.ToPlatform(s.UpdatedAt),
	}
	if s.StartDate != nil {
		v := timeutil.FormatDate(*s.StartDate)
		resp.StartDate = &v
	}
	if s.EndDate != nil {
		v := timeutil.FormatDate(*s.EndDate)
		resp.EndDate = &v
	}
	return resp
}
