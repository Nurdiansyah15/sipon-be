package query

import (
	"context"
	"sort"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	actEntity "sipon-be/internal/modules/akademik/domain/activity/entity"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	schConst "sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

type ListMySchedulesUseCase struct {
	kesantrianReader   ports.KesantrianReader
	periodRepo         periodRepo.AcademicPeriodRepository
	santriProgramRepo  spRepo.SantriProgramRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
	scheduleRepo       schRepo.ActivityScheduleRepository
}

func NewListMySchedulesUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
) *ListMySchedulesUseCase {
	return &ListMySchedulesUseCase{
		kesantrianReader:   kesantrianReader,
		periodRepo:         periodRepo,
		santriProgramRepo:  santriProgramRepo,
		activityPeriodRepo: activityPeriodRepo,
		activityRepo:       activityRepo,
		scheduleRepo:       scheduleRepo,
	}
}

func (uc *ListMySchedulesUseCase) Execute(ctx context.Context, userID string) ([]dto.MyScheduleResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	period, err := uc.periodRepo.FindOpen(ctx)
	if err != nil {
		if application.IsNotFoundErr(err, application.PeriodNotFoundCode) {
			return nil, kernel.New(application.ErrCodeNotFound)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, info.SantriID)
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	periods, err := uc.activityPeriodRepo.ListByPeriodAndProgram(ctx, period.ID, sp.ProgramID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if len(periods) == 0 {
		return []dto.MyScheduleResponse{}, nil
	}

	apIDs := make([]string, 0, len(periods))
	activityIDs := make([]string, 0, len(periods))
	apMap := map[string]*apEntity.ActivityPeriod{}
	for _, p := range periods {
		apIDs = append(apIDs, p.ID)
		activityIDs = append(activityIDs, p.ActivityID)
		apMap[p.ID] = p
	}

	activityMap := map[string]*actEntity.Activity{}
	if acts, err := uc.activityRepo.FindByIDs(ctx, activityIDs); err == nil {
		for _, a := range acts {
			activityMap[a.ID] = a
		}
	}

	schedules, err := uc.scheduleRepo.ListByActivityPeriodIDs(ctx, apIDs)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.MyScheduleResponse, 0, len(schedules))
	for _, s := range schedules {
		item := dto.MyScheduleResponse{
			ID:               s.ID,
			ActivityPeriodID: s.ActivityPeriodID,
			Type:             string(s.Type),
			StartTime:        s.StartTime,
			EndTime:          s.EndTime,
		}
		if s.StartDate != nil {
			v := timeutil.FormatDate(*s.StartDate)
			item.StartDate = &v
		}
		if s.EndDate != nil {
			v := timeutil.FormatDate(*s.EndDate)
			item.EndDate = &v
		}
		if ap, ok := apMap[s.ActivityPeriodID]; ok {
			if a, ok2 := activityMap[ap.ActivityID]; ok2 {
				item.ActivityName = a.Name
				item.ActivityCode = a.Code
			}
		}
		switch s.Type {
		case schConst.ActivityScheduleTypeWeekly:
			if weeklies, err := uc.scheduleRepo.ListWeeklies(ctx, s.ID); err == nil {
				for _, w := range weeklies {
					item.WeeklyDays = append(item.WeeklyDays, string(w.DayOfWeek))
				}
			}
		case schConst.ActivityScheduleTypeMonthly:
			if monthlies, err := uc.scheduleRepo.ListMonthlies(ctx, s.ID); err == nil {
				for _, m := range monthlies {
					item.MonthlyDays = append(item.MonthlyDays, m.DayOfMonth)
				}
			}
		case schConst.ActivityScheduleTypeYearly:
			if yearlies, err := uc.scheduleRepo.ListYearlies(ctx, s.ID); err == nil {
				for _, y := range yearlies {
					item.YearlyDates = append(item.YearlyDates, dto.YearlyDateIn{Month: y.Month, Day: y.Day})
				}
			}
		}
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		oi, oj := scheduleTypeOrder(items[i].Type), scheduleTypeOrder(items[j].Type)
		if oi != oj {
			return oi < oj
		}
		return items[i].StartTime < items[j].StartTime
	})
	return items, nil
}

func scheduleTypeOrder(t string) int {
	switch schConst.ActivityScheduleType(t) {
	case schConst.ActivityScheduleTypeDaily:
		return 1
	case schConst.ActivityScheduleTypeWeekly:
		return 2
	case schConst.ActivityScheduleTypeMonthly:
		return 3
	case schConst.ActivityScheduleTypeYearly:
		return 4
	default:
		return 5
	}
}
