package repository

import (
	"context"
	"time"

	"sipon-be/internal/modules/akademik/domain/attendance/entity"
)

// AttendanceWithSession adalah record absensi yang sudah di-enrich dengan
// informasi sesi & kegiatan tempat absensi itu tercatat.
type AttendanceWithSession struct {
	Attendance   entity.Attendance
	SessionID    string
	SessionStartsAt time.Time
	SessionEndsAt   time.Time
	ActivityName    string
	ActivityCode    string
	ScheduleType    string
}

type AttendanceRepository interface {
	Save(ctx context.Context, attendance *entity.Attendance) error
	Update(ctx context.Context, attendance *entity.Attendance) error
	FindByID(ctx context.Context, id string) (*entity.Attendance, error)
	FindBySessionAndSantri(ctx context.Context, sessionID, santriID string) (*entity.Attendance, error)
	ListBySession(ctx context.Context, sessionID string) ([]*entity.Attendance, error)
	// ListSantriIDsBySession returns santri IDs that already have attendance
	// records in the given session.
	ListSantriIDsBySession(ctx context.Context, sessionID string) ([]string, error)
	// ListBySantriAndPeriod returns attendance records for a santri within an
	// academic period (resolved through session → schedule → activity_period).
	ListBySantriAndPeriod(ctx context.Context, santriID, academicPeriodID string) ([]*AttendanceWithSession, error)
}
