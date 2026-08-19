package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/helper"
	"sipon-be/internal/modules/akademik/application/ports"
	schConst "sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

// GenerateSessionsFromScheduleUseCase membuat banyak sesi sekaligus dari satu
// jadwal berdasarkan recurrence pattern-nya untuk rentang tanggal tertentu.
// Waktu sesi diambil dari start_time/end_time jadwal. Sesi yang sudah ada pada
// tanggal+waktu yang sama di-skip (idempotent).
type GenerateSessionsFromScheduleUseCase struct {
	scheduleRepo schRepo.ActivityScheduleRepository
	sessionRepo  sesRepo.ActivitySessionRepository
	transactor   ports.Transactor
}

func NewGenerateSessionsFromScheduleUseCase(
	scheduleRepo schRepo.ActivityScheduleRepository,
	sessionRepo sesRepo.ActivitySessionRepository,
	transactor ports.Transactor,
) *GenerateSessionsFromScheduleUseCase {
	return &GenerateSessionsFromScheduleUseCase{
		scheduleRepo: scheduleRepo,
		sessionRepo:  sessionRepo,
		transactor:   transactor,
	}
}

func (uc *GenerateSessionsFromScheduleUseCase) Execute(ctx context.Context, scheduleID string, req dto.GenerateSessionsRequest) (*dto.GenerateSessionsResponse, error) {
	schedule, err := uc.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return nil, application.WrapRepoErr(err, schConst.CodeActivityScheduleNotFound)
	}

	fromDate, err := timeutil.ParseDate(req.FromDate)
	if err != nil {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	toDate := fromDate
	if req.ToDate != "" {
		toDate, err = timeutil.ParseDate(req.ToDate)
		if err != nil {
			return nil, kernel.New(application.ErrCodeBadRequest)
		}
	}
	if toDate.Before(fromDate) {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	weeklies, err := uc.scheduleRepo.ListWeeklies(ctx, schedule.ID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	monthlies, err := uc.scheduleRepo.ListMonthlies(ctx, schedule.ID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	yearlies, err := uc.scheduleRepo.ListYearlies(ctx, schedule.ID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	weekDays := make([]schConst.DayOfWeek, 0, len(weeklies))
	for _, w := range weeklies {
		weekDays = append(weekDays, w.DayOfWeek)
	}
	monthDays := make([]int, 0, len(monthlies))
	for _, m := range monthlies {
		monthDays = append(monthDays, m.DayOfMonth)
	}
	yearDates := make([]schEntity.YearlyDate, 0, len(yearlies))
	for _, y := range yearlies {
		yearDates = append(yearDates, schEntity.YearlyDate{Month: y.Month, Day: y.Day})
	}

	dates := helper.ExpandScheduleDates(schedule, weekDays, monthDays, yearDates, fromDate, toDate)

	existing, err := uc.sessionRepo.ListByScheduleIDs(ctx, []string{schedule.ID})
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	existingStarts := make(map[time.Time]struct{}, len(existing))
	for _, s := range existing {
		existingStarts[s.StartsAt.UTC()] = struct{}{}
	}

	startClock, err := parseClock(schedule.StartTime)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}
	endClock, err := parseClock(schedule.EndTime)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	created := make([]*sesEntity.ActivitySession, 0, len(dates))
	skipped := 0
	err = uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		for _, d := range dates {
			startsAt := clockOn(d, startClock)
			endsAt := clockOn(d, endClock)
			if _, ok := existingStarts[startsAt.UTC()]; ok {
				skipped++
				continue
			}
			session, err := sesEntity.NewActivitySession(uuid.NewString(), schedule.ID, startsAt, endsAt)
			if err != nil {
				return application.WrapBadRequestErr(err, sesConst.CodeActivitySessionInvalidTime)
			}
			if err := uc.sessionRepo.Save(txCtx, session); err != nil {
				return kernel.Wrap(application.ErrCodeInternal, err)
			}
			created = append(created, session)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	resp := &dto.GenerateSessionsResponse{
		TotalDatesExpanded: len(dates),
		TotalCreated:       len(created),
		TotalSkipped:       skipped,
		Sessions:           make([]dto.ActivitySessionResponse, 0, len(created)),
	}
	for _, s := range created {
		resp.Sessions = append(resp.Sessions, *MapSessionToResponse(s))
	}
	return resp, nil
}

type clock struct {
	hour, minute, second int
}

func parseClock(v string) (clock, error) {
	if len(v) != 8 || v[2] != ':' || v[5] != ':' {
		return clock{}, kernel.New(schConst.CodeActivityScheduleInvalid)
	}
	t, err := time.Parse("15:04:05", v)
	if err != nil {
		return clock{}, kernel.New(schConst.CodeActivityScheduleInvalid)
	}
	return clock{hour: t.Hour(), minute: t.Minute(), second: t.Second()}, nil
}

// clockOn menggabungkan tanggal (date-only, platform timezone) dengan jam
// wall-clock jadwal, menghasilkan timestamp dalam platform timezone.
func clockOn(date time.Time, c clock) time.Time {
	loc := timeutil.Loc()
	return time.Date(date.Year(), date.Month(), date.Day(), c.hour, c.minute, c.second, 0, loc)
}
