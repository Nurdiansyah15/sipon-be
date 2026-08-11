package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/modules/akademik/domain/attendance/constant"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	"sipon-be/internal/shared/kernel"
)

type ListAttendanceUseCase struct {
	attendanceRepo   attRepo.AttendanceRepository
	sessionRepo      sesRepo.ActivitySessionRepository
	kesantrianReader ports.KesantrianReader
}

func NewListAttendanceUseCase(
	attendanceRepo attRepo.AttendanceRepository,
	sessionRepo sesRepo.ActivitySessionRepository,
	kesantrianReader ports.KesantrianReader,
) *ListAttendanceUseCase {
	return &ListAttendanceUseCase{attendanceRepo: attendanceRepo, sessionRepo: sessionRepo, kesantrianReader: kesantrianReader}
}

func (uc *ListAttendanceUseCase) Execute(ctx context.Context, sessionID string) ([]dto.AttendanceResponse, error) {
	if _, err := uc.sessionRepo.FindByID(ctx, sessionID); err != nil {
		return nil, application.WrapRepoErr(err, constant.CodeAttendanceNotFound)
	}

	attendances, err := uc.attendanceRepo.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	responses := make([]dto.AttendanceResponse, 0, len(attendances))
	for _, a := range attendances {
		resp := command.MapAttendanceToResponse(a)
		if info, err := uc.kesantrianReader.GetSantriByID(ctx, a.SantriID); err == nil {
			resp.SantriNIS = info.NIS
		}
		responses = append(responses, *resp)
	}
	return responses, nil
}
