package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_period/constant"
	"sipon-be/internal/shared/kernel"
)

type ActivityPeriod struct {
	ID               string
	ActivityID       string
	AcademicPeriodID string
	Status           constant.ActivityPeriodStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewActivityPeriod(id, activityID, academicPeriodID string) (*ActivityPeriod, error) {
	if id == "" || activityID == "" || academicPeriodID == "" {
		return nil, kernel.New(constant.CodeActivityPeriodNotFound)
	}
	now := time.Now()
	return &ActivityPeriod{
		ID:               id,
		ActivityID:       activityID,
		AcademicPeriodID: academicPeriodID,
		Status:           constant.ActivityPeriodStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (p *ActivityPeriod) Activate() error {
	if p.Status != constant.ActivityPeriodStatusInactive {
		return kernel.New(constant.CodeActivityPeriodInvalidStatus)
	}
	p.Status = constant.ActivityPeriodStatusActive
	p.UpdatedAt = time.Now()
	return nil
}

func (p *ActivityPeriod) Deactivate() error {
	if p.Status != constant.ActivityPeriodStatusActive {
		return kernel.New(constant.CodeActivityPeriodInvalidStatus)
	}
	p.Status = constant.ActivityPeriodStatusInactive
	p.UpdatedAt = time.Now()
	return nil
}

func (p *ActivityPeriod) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}
