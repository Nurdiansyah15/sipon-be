package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	"sipon-be/internal/shared/kernel"
)

type ActivitySession struct {
	ID                 string
	ActivityScheduleID string
	StartsAt           time.Time
	EndsAt             time.Time
	Status             constant.ActivitySessionStatus
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}

func NewActivitySession(id, activityScheduleID string, startsAt, endsAt time.Time) (*ActivitySession, error) {
	if id == "" || activityScheduleID == "" {
		return nil, kernel.New(constant.CodeActivitySessionNotFound)
	}
	if !endsAt.After(startsAt) {
		return nil, kernel.New(constant.CodeActivitySessionInvalidTime)
	}
	now := time.Now()
	return &ActivitySession{
		ID:                 id,
		ActivityScheduleID: activityScheduleID,
		StartsAt:           startsAt,
		EndsAt:             endsAt,
		Status:             constant.ActivitySessionStatusScheduled,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

func (s *ActivitySession) Open() error {
	if s.Status != constant.ActivitySessionStatusScheduled {
		return kernel.New(constant.CodeActivitySessionInvalidStatus)
	}
	s.Status = constant.ActivitySessionStatusOpen
	s.UpdatedAt = time.Now()
	return nil
}

func (s *ActivitySession) Complete() error {
	if s.Status != constant.ActivitySessionStatusScheduled && s.Status != constant.ActivitySessionStatusOpen {
		return kernel.New(constant.CodeActivitySessionInvalidStatus)
	}
	s.Status = constant.ActivitySessionStatusCompleted
	s.UpdatedAt = time.Now()
	return nil
}

func (s *ActivitySession) Cancel() error {
	if s.Status == constant.ActivitySessionStatusCompleted || s.Status == constant.ActivitySessionStatusCancelled {
		return kernel.New(constant.CodeActivitySessionInvalidStatus)
	}
	s.Status = constant.ActivitySessionStatusCancelled
	s.UpdatedAt = time.Now()
	return nil
}

func (s *ActivitySession) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now
	s.UpdatedAt = now
}
