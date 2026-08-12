package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	actEntity "sipon-be/internal/modules/akademik/domain/activity/entity"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	apEntity "sipon-be/internal/modules/akademik/domain/activity_period/entity"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

type GetMyAttendanceUseCase struct {
	kesantrianReader   ports.KesantrianReader
	periodRepo         periodRepo.AcademicPeriodRepository
	santriProgramRepo  spRepo.SantriProgramRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo       actRepo.ActivityRepository
	scheduleRepo       schRepo.ActivityScheduleRepository
	sessionRepo        sesRepo.ActivitySessionRepository
	attendanceRepo     attRepo.AttendanceRepository
}

func NewGetMyAttendanceUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
	sessionRepo sesRepo.ActivitySessionRepository,
	attendanceRepo attRepo.AttendanceRepository,
) *GetMyAttendanceUseCase {
	return &GetMyAttendanceUseCase{
		kesantrianReader:   kesantrianReader,
		periodRepo:         periodRepo,
		santriProgramRepo:  santriProgramRepo,
		activityPeriodRepo: activityPeriodRepo,
		activityRepo:       activityRepo,
		scheduleRepo:       scheduleRepo,
		sessionRepo:        sessionRepo,
		attendanceRepo:     attendanceRepo,
	}
}

func (uc *GetMyAttendanceUseCase) Execute(ctx context.Context, userID string, q dto.MyAttendanceListQuery) (*dto.MyAttendanceResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	resp := &dto.MyAttendanceResponse{
		Sessions: []dto.MyAttendanceSessionItem{},
	}

	// Resolve periode akademik.
	var periodID string
	if q.AcademicPeriodID != "" {
		periodID = q.AcademicPeriodID
		if period, err := uc.periodRepo.FindByID(ctx, periodID); err == nil {
			resp.AcademicPeriod = command.MapAcademicPeriodToResponse(period)
		}
	} else {
		period, err := uc.periodRepo.FindOpen(ctx)
		if err != nil {
			if application.IsNotFoundErr(err, application.PeriodNotFoundCode) {
				return resp, nil
			}
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		periodID = period.ID
		resp.AcademicPeriod = command.MapAcademicPeriodToResponse(period)
	}

	sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, info.SantriID)
	if err != nil {
		// Tanpa program aktif → tidak ada kegiatan yang applicable.
		return resp, nil
	}

	periods, err := uc.activityPeriodRepo.ListByPeriodAndProgram(ctx, periodID, sp.ProgramID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if len(periods) == 0 {
		return resp, nil
	}

	apIDs := make([]string, 0, len(periods))
	apMap := map[string]*apEntity.ActivityPeriod{}
	activityIDs := []string{}
	for _, p := range periods {
		apIDs = append(apIDs, p.ID)
		apMap[p.ID] = p
		activityIDs = append(activityIDs, p.ActivityID)
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
	scheduleMap := map[string]*schEntity.ActivitySchedule{}
	scheduleIDs := []string{}
	for _, s := range schedules {
		if q.ActivityScheduleID != "" && s.ID != q.ActivityScheduleID {
			continue
		}
		scheduleMap[s.ID] = s
		scheduleIDs = append(scheduleIDs, s.ID)
	}

	sessions, err := uc.sessionRepo.ListByScheduleIDs(ctx, scheduleIDs)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	// Absensi yang sudah tercatat untuk santri di periode ini.
	recordedBySession := map[string]*attRepo.AttendanceWithSession{}
	if records, err := uc.attendanceRepo.ListBySantriAndPeriod(ctx, info.SantriID, periodID); err == nil {
		for _, r := range records {
			recordedBySession[r.SessionID] = r
		}
	}

	summary := dto.MyAttendanceSummary{}
	for _, s := range sessions {
		item := dto.MyAttendanceSessionItem{
			SessionID:    s.ID,
			StartsAt:     s.StartsAt.Format("2006-01-02T15:04:05Z07:00"),
			EndsAt:       s.EndsAt.Format("2006-01-02T15:04:05Z07:00"),
			Status:       "unrecorded",
			ScheduleType: scheduleTypeOf(scheduleMap[s.ActivityScheduleID]),
		}
		if sch, ok := scheduleMap[s.ActivityScheduleID]; ok {
			if ap, ok2 := apMap[sch.ActivityPeriodID]; ok2 {
				if a, ok3 := activityMap[ap.ActivityID]; ok3 {
					item.ActivityName = a.Name
					item.ActivityCode = a.Code
				}
			}
		}

		if rec, ok := recordedBySession[s.ID]; ok {
			item.Status = string(rec.Attendance.Status)
			recTime := rec.Attendance.RecordedAt.Format("2006-01-02T15:04:05Z07:00")
			item.RecordedAt = &recTime
		}

		switch item.Status {
		case "present":
			summary.Present++
		case "absent":
			summary.Absent++
		case "excused":
			summary.Excused++
		default:
			summary.Unrecorded++
		}
		summary.TotalSessions++
		resp.Sessions = append(resp.Sessions, item)
	}
	resp.Summary = summary
	return resp, nil
}

func scheduleTypeOf(s *schEntity.ActivitySchedule) string {
	if s == nil {
		return ""
	}
	return string(s.Type)
}
