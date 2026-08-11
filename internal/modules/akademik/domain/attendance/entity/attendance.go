package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/attendance/constant"
	"sipon-be/internal/shared/kernel"
)

type Attendance struct {
	ID                string
	ActivitySessionID string
	SantriID          string
	Status            constant.AttendanceStatus
	RecordedAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

func NewAttendance(id, activitySessionID, santriID string, status constant.AttendanceStatus) (*Attendance, error) {
	if id == "" || activitySessionID == "" || santriID == "" {
		return nil, kernel.New(constant.CodeAttendanceNotFound)
	}
	if status != constant.AttendanceStatusPresent &&
		status != constant.AttendanceStatusAbsent &&
		status != constant.AttendanceStatusExcused {
		return nil, kernel.New(constant.CodeAttendanceInvalidStatus)
	}
	now := time.Now()
	return &Attendance{
		ID:                id,
		ActivitySessionID: activitySessionID,
		SantriID:          santriID,
		Status:            status,
		RecordedAt:        now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (a *Attendance) UpdateStatus(status constant.AttendanceStatus) error {
	if status != constant.AttendanceStatusPresent &&
		status != constant.AttendanceStatusAbsent &&
		status != constant.AttendanceStatusExcused {
		return kernel.New(constant.CodeAttendanceInvalidStatus)
	}
	a.Status = status
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Attendance) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
}
