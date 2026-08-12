package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	actEntity "sipon-be/internal/modules/akademik/domain/activity/entity"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

type ListMyActivitiesUseCase struct {
	kesantrianReader  ports.KesantrianReader
	periodRepo        periodRepo.AcademicPeriodRepository
	santriProgramRepo spRepo.SantriProgramRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
	scheduleRepo       schRepo.ActivityScheduleRepository
}

func NewListMyActivitiesUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
) *ListMyActivitiesUseCase {
	return &ListMyActivitiesUseCase{
		kesantrianReader:   kesantrianReader,
		periodRepo:         periodRepo,
		santriProgramRepo:  santriProgramRepo,
		activityPeriodRepo: activityPeriodRepo,
		activityRepo:       activityRepo,
		scheduleRepo:       scheduleRepo,
	}
}

func (uc *ListMyActivitiesUseCase) Execute(ctx context.Context, userID string) ([]dto.MyActivityResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	periodID, programID, err := uc.resolveContext(ctx, info.SantriID)
	if err != nil {
		return nil, err
	}

	periods, err := uc.activityPeriodRepo.ListByPeriodAndProgram(ctx, periodID, programID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if len(periods) == 0 {
		return []dto.MyActivityResponse{}, nil
	}

	apIDs := make([]string, 0, len(periods))
	activityIDs := make([]string, 0, len(periods))
	for _, p := range periods {
		apIDs = append(apIDs, p.ID)
		activityIDs = append(activityIDs, p.ActivityID)
	}

	activityMap := map[string]*actEntity.Activity{}
	if acts, err := uc.activityRepo.FindByIDs(ctx, activityIDs); err == nil {
		for _, a := range acts {
			activityMap[a.ID] = a
		}
	}

	counts := map[string]int{}
	if scheds, err := uc.scheduleRepo.ListByActivityPeriodIDs(ctx, apIDs); err == nil {
		for _, s := range scheds {
			counts[s.ActivityPeriodID]++
		}
	}

	items := make([]dto.MyActivityResponse, 0, len(periods))
	for _, p := range periods {
		item := dto.MyActivityResponse{
			ID:               p.ID,
			ActivityID:       p.ActivityID,
			ActivityPeriodID: p.ID,
			Status:           string(p.Status),
			ScheduleCount:    counts[p.ID],
		}
		if a, ok := activityMap[p.ActivityID]; ok {
			item.ActivityCode = a.Code
			item.ActivityName = a.Name
		}
		items = append(items, item)
	}
	return items, nil
}

// resolveContext resolves the santri's active program and the current open
// period. It returns 404 when no open period exists and 422 when the santri
// has no active program.
func (uc *ListMyActivitiesUseCase) resolveContext(ctx context.Context, santriID string) (periodID, programID string, err error) {
	period, err := uc.periodRepo.FindOpen(ctx)
	if err != nil {
		if application.IsNotFoundErr(err, application.PeriodNotFoundCode) {
			return "", "", kernel.New(application.ErrCodeNotFound)
		}
		return "", "", kernel.Wrap(application.ErrCodeInternal, err)
	}

	sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, santriID)
	if err != nil {
		return "", "", kernel.New(application.ErrCodeUnprocessableEntity)
	}
	return period.ID, sp.ProgramID, nil
}
