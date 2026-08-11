package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/modules/akademik/domain/attendance/constant"
	"sipon-be/internal/modules/akademik/domain/attendance/entity"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	"sipon-be/internal/shared/kernel"
)

type RecordAttendanceUseCase struct {
	attendanceRepo   attRepo.AttendanceRepository
	sessionRepo      sesRepo.ActivitySessionRepository
	kesantrianReader ports.KesantrianReader
}

func NewRecordAttendanceUseCase(
	attendanceRepo attRepo.AttendanceRepository,
	sessionRepo sesRepo.ActivitySessionRepository,
	kesantrianReader ports.KesantrianReader,
) *RecordAttendanceUseCase {
	return &RecordAttendanceUseCase{attendanceRepo: attendanceRepo, sessionRepo: sessionRepo, kesantrianReader: kesantrianReader}
}

func (uc *RecordAttendanceUseCase) Execute(ctx context.Context, sessionID string, req dto.RecordAttendanceRequest) ([]dto.AttendanceResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAttendanceNotFound)
	}
	if session.Status == "cancelled" {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	responses := make([]dto.AttendanceResponse, 0, len(req.Records))
	for _, rec := range req.Records {
		info, err := uc.kesantrianReader.GetSantriByID(ctx, rec.SantriID)
		if err != nil || info == nil || info.Status != "SANTRI" {
			return nil, kernel.New(application.ErrCodeUnprocessableEntity)
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
		responses = append(responses, *resp)
	}
	return responses, nil
}

func MapAttendanceToResponse(a *entity.Attendance) *dto.AttendanceResponse {
	return &dto.AttendanceResponse{
		ID:                a.ID,
		ActivitySessionID: a.ActivitySessionID,
		SantriID:          a.SantriID,
		Status:            string(a.Status),
		RecordedAt:        a.RecordedAt,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
}
