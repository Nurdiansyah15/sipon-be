package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/timeutil"
)

type GetPresensiSessionInfoUseCase struct {
	sessionRepo      sesRepo.ActivitySessionRepository
	scheduleRepo     schRepo.ActivityScheduleRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo     actRepo.ActivityRepository
	periodRepo       periodRepo.AcademicPeriodRepository
	registrationRepo regRepo.SantriRegistrationRepository
	attendanceRepo   attRepo.AttendanceRepository
}

func NewGetPresensiSessionInfoUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	periodRepo periodRepo.AcademicPeriodRepository,
	registrationRepo regRepo.SantriRegistrationRepository,
	attendanceRepo attRepo.AttendanceRepository,
) *GetPresensiSessionInfoUseCase {
	return &GetPresensiSessionInfoUseCase{
		sessionRepo:       sessionRepo,
		scheduleRepo:      scheduleRepo,
		activityPeriodRepo: activityPeriodRepo,
		activityRepo:      activityRepo,
		periodRepo:        periodRepo,
		registrationRepo:  registrationRepo,
		attendanceRepo:    attendanceRepo,
	}
}

func (uc *GetPresensiSessionInfoUseCase) Execute(ctx context.Context, sessionID string) (*dto.PresensiSessionInfo, error) {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, application.WrapRepoErr(err, sesConst.CodeActivitySessionNotFound)
	}

	resp := &dto.PresensiSessionInfo{
		ID:       session.ID,
		Status:   string(session.Status),
		StartsAt: timeutil.FormatDateTime(session.StartsAt),
		EndsAt:   timeutil.FormatDateTime(session.EndsAt),
	}

	academicPeriodID := ""
	if sch, err := uc.scheduleRepo.FindByID(ctx, session.ActivityScheduleID); err == nil {
		resp.ScheduleType = string(sch.Type)
		if ap, err := uc.activityPeriodRepo.FindByID(ctx, sch.ActivityPeriodID); err == nil {
			academicPeriodID = ap.AcademicPeriodID
			if activity, err := uc.activityRepo.FindByID(ctx, ap.ActivityID); err == nil {
				resp.ActivityName = activity.Name
				resp.ActivityCode = activity.Code
			}
			if period, err := uc.periodRepo.FindByID(ctx, ap.AcademicPeriodID); err == nil {
				resp.PeriodName = period.Name
			}
		}
	}

	if academicPeriodID != "" {
		if registrations, err := uc.registrationRepo.ListCompletedByAcademicPeriod(ctx, academicPeriodID); err == nil {
			resp.TotalEligible = len(registrations)
		}
	}
	if attendances, err := uc.attendanceRepo.ListBySession(ctx, session.ID); err == nil {
		for _, a := range attendances {
			if a.Status == "present" {
				resp.TotalPresent++
			}
		}
	}
	return resp, nil
}
