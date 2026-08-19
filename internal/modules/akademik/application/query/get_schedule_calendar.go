package query

import (
	"context"
	"time"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/helper"
	actEntity "sipon-be/internal/modules/akademik/domain/activity/entity"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

const maxCalendarSpanDays = 366 * 5

type GetScheduleCalendarUseCase struct {
	scheduleRepo       schRepo.ActivityScheduleRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
}

func NewGetScheduleCalendarUseCase(
	scheduleRepo schRepo.ActivityScheduleRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
) *GetScheduleCalendarUseCase {
	return &GetScheduleCalendarUseCase{scheduleRepo: scheduleRepo, activityPeriodRepo: activityPeriodRepo, activityRepo: activityRepo}
}

type scheduleRecurrence struct {
	weeklyDays  []constant.DayOfWeek
	monthlyDays []int
	yearlyDates []schEntity.YearlyDate
}

func (uc *GetScheduleCalendarUseCase) Execute(ctx context.Context, from, to time.Time, academicPeriodID string, types []constant.ActivityScheduleType) (*dto.ScheduleCalendarResponse, error) {
	from = dateOnly(from)
	to = dateOnly(to)
	if from.After(to) {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	if to.Sub(from) > maxCalendarSpanDays*24*time.Hour {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	periods, err := uc.listPeriods(ctx, academicPeriodID)
	if err != nil {
		return nil, err
	}

	schedules, activityBySchedule, err := uc.listSchedules(ctx, periods)
	if err != nil {
		return nil, err
	}

	byDate := map[string][]dto.ScheduleCalendarItem{}
	for _, s := range schedules {
		if len(types) > 0 && !typeAllowed(s.Type, types) {
			continue
		}
		rec, err := uc.loadRecurrence(ctx, s)
		if err != nil {
			continue
		}
		item := dto.ScheduleCalendarItem{
			ID:               s.ID,
			ActivityPeriodID: s.ActivityPeriodID,
			Type:             string(s.Type),
			StartTime:        s.StartTime,
			EndTime:          s.EndTime,
		}
		if a, ok := activityBySchedule[s.ID]; ok {
			item.ActivityName = a.Name
			item.ActivityCode = a.Code
		}
		uc.expand(s, rec, from, to, byDate, item)
	}

	return &dto.ScheduleCalendarResponse{
		From: timeutil.FormatDate(from),
		To:   timeutil.FormatDate(to),
		Days: buildCalendarDays(byDate),
	}, nil
}

func typeAllowed(t constant.ActivityScheduleType, types []constant.ActivityScheduleType) bool {
	for _, v := range types {
		if v == t {
			return true
		}
	}
	return false
}

func (uc *GetScheduleCalendarUseCase) listPeriods(ctx context.Context, academicPeriodID string) ([]*apEntity.ActivityPeriod, error) {
	query := apRepo.ActivityPeriodListQuery{
		Page:  1,
		Limit: 500,
	}
	if academicPeriodID != "" {
		query.AcademicPeriodID = &academicPeriodID
	} else {
		active := "active"
		query.Status = &active
	}
	result, err := uc.activityPeriodRepo.List(ctx, query)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return result.Items, nil
}

func (uc *GetScheduleCalendarUseCase) listSchedules(ctx context.Context, periods []*apEntity.ActivityPeriod) ([]*schEntity.ActivitySchedule, map[string]*actEntity.Activity, error) {
	var schedules []*schEntity.ActivitySchedule
	activityIDs := map[string]struct{}{}
	for _, ap := range periods {
		items, err := uc.scheduleRepo.ListByActivityPeriod(ctx, ap.ID)
		if err != nil {
			return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		schedules = append(schedules, items...)
		activityIDs[ap.ActivityID] = struct{}{}
	}

	activityMap := map[string]*actEntity.Activity{}
	if acts, err := uc.activityRepo.FindByIDs(ctx, idsFromSet(activityIDs)); err == nil {
		for _, a := range acts {
			activityMap[a.ID] = a
		}
	}

	apByID := map[string]*apEntity.ActivityPeriod{}
	for _, ap := range periods {
		apByID[ap.ID] = ap
	}

	// Map schedule -> activity via activity period.
	activityBySchedule := map[string]*actEntity.Activity{}
	for _, s := range schedules {
		if ap, ok := apByID[s.ActivityPeriodID]; ok {
			if a, ok2 := activityMap[ap.ActivityID]; ok2 {
				activityBySchedule[s.ID] = a
			}
		}
	}

	return schedules, activityBySchedule, nil
}

func (uc *GetScheduleCalendarUseCase) loadRecurrence(ctx context.Context, s *schEntity.ActivitySchedule) (*scheduleRecurrence, error) {
	rec := &scheduleRecurrence{}
	switch s.Type {
	case constant.ActivityScheduleTypeWeekly:
		weeklies, err := uc.scheduleRepo.ListWeeklies(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		rec.weeklyDays = make([]constant.DayOfWeek, 0, len(weeklies))
		for _, w := range weeklies {
			rec.weeklyDays = append(rec.weeklyDays, w.DayOfWeek)
		}
	case constant.ActivityScheduleTypeMonthly:
		monthlies, err := uc.scheduleRepo.ListMonthlies(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		rec.monthlyDays = make([]int, 0, len(monthlies))
		for _, m := range monthlies {
			rec.monthlyDays = append(rec.monthlyDays, m.DayOfMonth)
		}
	case constant.ActivityScheduleTypeYearly:
		yearlies, err := uc.scheduleRepo.ListYearlies(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		rec.yearlyDates = make([]schEntity.YearlyDate, 0, len(yearlies))
		for _, y := range yearlies {
			rec.yearlyDates = append(rec.yearlyDates, schEntity.YearlyDate{Month: y.Month, Day: y.Day})
		}
	}
	return rec, nil
}

func (uc *GetScheduleCalendarUseCase) expand(
	s *schEntity.ActivitySchedule,
	rec *scheduleRecurrence,
	from, to time.Time,
	byDate map[string][]dto.ScheduleCalendarItem,
	item dto.ScheduleCalendarItem,
) {
	dates := helper.ExpandScheduleDates(s, rec.weeklyDays, rec.monthlyDays, rec.yearlyDates, from, to)
	for _, d := range dates {
		appendItem(byDate, d, item)
	}
}

func appendItem(byDate map[string][]dto.ScheduleCalendarItem, d time.Time, item dto.ScheduleCalendarItem) {
	key := timeutil.FormatDate(d)
	byDate[key] = append(byDate[key], item)
}

func buildCalendarDays(byDate map[string][]dto.ScheduleCalendarItem) []dto.ScheduleCalendarDay {
	days := make([]dto.ScheduleCalendarDay, 0, len(byDate))
	for date, items := range byDate {
		days = append(days, dto.ScheduleCalendarDay{Date: date, Items: items})
	}
	return days
}

func dateOnly(t time.Time) time.Time {
	return timeutil.DateOnly(t)
}
