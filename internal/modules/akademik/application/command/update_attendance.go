package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/modules/akademik/domain/attendance/constant"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateAttendanceUseCase struct {
	attendanceRepo attRepo.AttendanceRepository
	sessionRepo    sesRepo.ActivitySessionRepository
}

func NewUpdateAttendanceUseCase(attendanceRepo attRepo.AttendanceRepository, sessionRepo sesRepo.ActivitySessionRepository) *UpdateAttendanceUseCase {
	return &UpdateAttendanceUseCase{attendanceRepo: attendanceRepo, sessionRepo: sessionRepo}
}

func (uc *UpdateAttendanceUseCase) Execute(ctx context.Context, sessionID, santriID string, req dto.UpdateAttendanceRequest) (*dto.AttendanceResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAttendanceNotFound)
	}
	if session.Status != "open" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "absensi terkunci pada sesi yang sudah selesai atau dibatalkan", nil)
	}

	attendance, err := uc.attendanceRepo.FindBySessionAndSantri(ctx, sessionID, santriID)
	if err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAttendanceNotFound)
	}
	if err := attendance.UpdateStatus(constant.AttendanceStatus(req.Status)); err != nil {
		return nil, application.WrapBadRequestErr(err, constant.CodeAttendanceInvalidStatus)
	}
	if err := uc.attendanceRepo.Update(ctx, attendance); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapAttendanceToResponse(attendance), nil
}
