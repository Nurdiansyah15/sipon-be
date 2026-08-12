package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

type ListPresensiAttendanceUseCase struct {
	attendanceRepo   attRepo.AttendanceRepository
	kesantrianReader ports.KesantrianReader
}

func NewListPresensiAttendanceUseCase(
	attendanceRepo attRepo.AttendanceRepository,
	kesantrianReader ports.KesantrianReader,
) *ListPresensiAttendanceUseCase {
	return &ListPresensiAttendanceUseCase{attendanceRepo: attendanceRepo, kesantrianReader: kesantrianReader}
}

func (uc *ListPresensiAttendanceUseCase) Execute(ctx context.Context, sessionID string) ([]dto.PresensiAttendanceItem, error) {
	attendances, err := uc.attendanceRepo.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.PresensiAttendanceItem, 0, len(attendances))
	for _, a := range attendances {
		item := dto.PresensiAttendanceItem{
			SantriID:   a.SantriID,
			Status:     string(a.Status),
			RecordedAt: timeutil.FormatDateTime(a.RecordedAt),
		}
		if info, err := uc.kesantrianReader.GetSantriByID(ctx, a.SantriID); err == nil && info != nil {
			item.NIS = info.NIS
			item.Fullname = info.Fullname
		}
		items = append(items, item)
	}
	return items, nil
}
