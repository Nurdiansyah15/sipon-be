package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/attendance/entity"
)

type AttendanceRepository interface {
	Save(ctx context.Context, attendance *entity.Attendance) error
	Update(ctx context.Context, attendance *entity.Attendance) error
	FindByID(ctx context.Context, id string) (*entity.Attendance, error)
	FindBySessionAndSantri(ctx context.Context, sessionID, santriID string) (*entity.Attendance, error)
	ListBySession(ctx context.Context, sessionID string) ([]*entity.Attendance, error)
}
