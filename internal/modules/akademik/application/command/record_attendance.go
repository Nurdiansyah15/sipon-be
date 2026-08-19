package command

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/akademik/application/resolver"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/modules/akademik/domain/attendance/constant"
	"sipon-be/internal/modules/akademik/domain/attendance/entity"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

type RecordAttendanceUseCase struct {
	attendanceRepo   attRepo.AttendanceRepository
	sessionRepo      sesRepo.ActivitySessionRepository
	registrationRepo regRepo.SantriRegistrationRepository
	periodResolver   *resolver.SessionPeriodResolver
	kesantrianReader ports.KesantrianReader
}

func NewRecordAttendanceUseCase(
	attendanceRepo attRepo.AttendanceRepository,
	sessionRepo sesRepo.ActivitySessionRepository,
	registrationRepo regRepo.SantriRegistrationRepository,
	periodResolver *resolver.SessionPeriodResolver,
	kesantrianReader ports.KesantrianReader,
) *RecordAttendanceUseCase {
	return &RecordAttendanceUseCase{
		attendanceRepo:   attendanceRepo,
		sessionRepo:      sessionRepo,
		registrationRepo: registrationRepo,
		periodResolver:   periodResolver,
		kesantrianReader: kesantrianReader,
	}
}

func (uc *RecordAttendanceUseCase) Execute(ctx context.Context, sessionID string, req dto.RecordAttendanceRequest) ([]dto.AttendanceResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAttendanceNotFound)
	}
	if session.Status != "open" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "sesi harus dibuka terlebih dahulu untuk mencatat absensi", nil)
	}

	academicPeriodID, err := uc.periodResolver.Resolve(ctx, sessionID)
	if err != nil {
		slog.Warn("akademik: resolve session academic period failed", "session_id", sessionID, "error", err)
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "periode akademik sesi tidak ditemukan", nil)
	}

	responses := make([]dto.AttendanceResponse, 0, len(req.Records))
	for _, rec := range req.Records {
		info, err := uc.kesantrianReader.GetSantriByID(ctx, rec.SantriID)
		if err != nil || info == nil || info.Status != "SANTRI" {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri tidak valid", nil)
		}

		if err := uc.ensureHerreg(ctx, rec.SantriID, academicPeriodID); err != nil {
			return nil, err
		}

		existing, _ := uc.attendanceRepo.FindBySessionAndSantri(ctx, sessionID, rec.SantriID)
		if existing != nil {
			return nil, kernel.New(application.ErrCodeConflict)
		}

		attendance, err := entity.NewAttendance(uuid.NewString(), sessionID, rec.SantriID, constant.AttendanceStatus(rec.Status))
		if err != nil {
			return nil, application.WrapBadRequestErr(err, constant.CodeAttendanceInvalidStatus)
		}
		if err := uc.attendanceRepo.Save(ctx, attendance); err != nil {
			return nil, application.WrapConflictErr(err, constant.CodeAttendanceDuplicate)
		}

		resp := MapAttendanceToResponse(attendance)
		resp.SantriNIS = info.NIS
		resp.SantriName = info.Fullname
		responses = append(responses, *resp)
	}
	return responses, nil
}

// ensureHerreg rejects the attendance record when the santri has not completed
// herregistrasi for the session's academic period.
func (uc *RecordAttendanceUseCase) ensureHerreg(ctx context.Context, santriID, academicPeriodID string) error {
	reg, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, santriID, academicPeriodID)
	if err != nil || reg == nil || reg.Status != "completed" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri belum herregistrasi pada periode ini", nil)
	}
	return nil
}

func MapAttendanceToResponse(a *entity.Attendance) *dto.AttendanceResponse {
	return &dto.AttendanceResponse{
		ID:                a.ID,
		ActivitySessionID: a.ActivitySessionID,
		SantriID:          a.SantriID,
		Status:            string(a.Status),
		RecordedAt:        timeutil.ToPlatform(a.RecordedAt),
		CreatedAt:         timeutil.ToPlatform(a.CreatedAt),
		UpdatedAt:         timeutil.ToPlatform(a.UpdatedAt),
	}
}
