package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
)

type GetActivitySessionUseCase struct {
	sessionRepo        sesRepo.ActivitySessionRepository
	scheduleRepo       schRepo.ActivityScheduleRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
	attendanceRepo     attRepo.AttendanceRepository
}

func NewGetActivitySessionUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	attendanceRepo attRepo.AttendanceRepository,
) *GetActivitySessionUseCase {
	return &GetActivitySessionUseCase{sessionRepo: sessionRepo, scheduleRepo: scheduleRepo, activityPeriodRepo: activityPeriodRepo, activityRepo: activityRepo, attendanceRepo: attendanceRepo}
}

func (uc *GetActivitySessionUseCase) Execute(ctx context.Context, id string) (*dto.ActivitySessionDetailResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeActivitySessionNotFound)
	}

	resp := &dto.ActivitySessionDetailResponse{
		ActivitySessionResponse: *command.MapSessionToResponse(session),
	}

	if sch, err := uc.scheduleRepo.FindByID(ctx, session.ActivityScheduleID); err == nil {
		resp.ScheduleType = string(sch.Type)
		if ap, err := uc.activityPeriodRepo.FindByID(ctx, sch.ActivityPeriodID); err == nil {
			if activity, err := uc.activityRepo.FindByID(ctx, ap.ActivityID); err == nil {
				resp.ActivityName = activity.Name
				resp.ActivityCode = activity.Code
			}
		}
	}

	summary := &dto.AttendanceSummary{}
	if attendances, err := uc.attendanceRepo.ListBySession(ctx, session.ID); err == nil {
		summary.Total = int64(len(attendances))
		for _, a := range attendances {
			switch a.Status {
			case "present":
				summary.Present++
			case "absent":
				summary.Absent++
			case "excused":
				summary.Excused++
			}
		}
		resp.AttendanceSummary = summary
	}
	return resp, nil
}
