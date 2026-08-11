package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
)

type GetActivityScheduleUseCase struct {
	scheduleRepo       schRepo.ActivityScheduleRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
}

func NewGetActivityScheduleUseCase(
	scheduleRepo schRepo.ActivityScheduleRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
) *GetActivityScheduleUseCase {
	return &GetActivityScheduleUseCase{scheduleRepo: scheduleRepo, activityPeriodRepo: activityPeriodRepo, activityRepo: activityRepo}
}

func (uc *GetActivityScheduleUseCase) Execute(ctx context.Context, id string) (*dto.ActivityScheduleDetailResponse, error) {
	schedule, err := uc.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivityScheduleNotFound)
	}

	resp := command.MapScheduleToDetailResponse(schedule)

	ap, err := uc.activityPeriodRepo.FindByID(ctx, schedule.ActivityPeriodID)
	if err == nil {
		if activity, err := uc.activityRepo.FindByID(ctx, ap.ActivityID); err == nil {
			resp.ActivityName = activity.Name
			resp.ActivityCode = activity.Code
		}
	}

	switch schedule.Type {
	case constant.ActivityScheduleTypeWeekly:
		weeklies, _ := uc.scheduleRepo.ListWeeklies(ctx, schedule.ID)
		for _, w := range weeklies {
			resp.WeeklyDays = append(resp.WeeklyDays, string(w.DayOfWeek))
		}
	case constant.ActivityScheduleTypeMonthly:
		monthlies, _ := uc.scheduleRepo.ListMonthlies(ctx, schedule.ID)
		for _, m := range monthlies {
			resp.MonthlyDays = append(resp.MonthlyDays, m.DayOfMonth)
		}
	case constant.ActivityScheduleTypeYearly:
		yearlies, _ := uc.scheduleRepo.ListYearlies(ctx, schedule.ID)
		for _, y := range yearlies {
			resp.YearlyDates = append(resp.YearlyDates, dto.YearlyDateIn{Month: y.Month, Day: y.Day})
		}
	}
	return resp, nil
}
