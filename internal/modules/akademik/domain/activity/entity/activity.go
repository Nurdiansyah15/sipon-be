package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/activity/constant"
	"sipon-be/internal/shared/kernel"
)

type Activity struct {
	ID        string
	Code      string
	Name      string
	Status    constant.ActivityStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewActivity(id, code, name string) (*Activity, error) {
	if id == "" || code == "" || name == "" {
		return nil, kernel.New(constant.CodeActivityNotFound)
	}
	now := time.Now()
	return &Activity{
		ID:        id,
		Code:      code,
		Name:      name,
		Status:    constant.ActivityStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (a *Activity) Update(name, status string) error {
	if status != "" {
		if status != string(constant.ActivityStatusActive) && status != string(constant.ActivityStatusInactive) {
			return kernel.New(constant.CodeActivityInvalidStatus)
		}
		a.Status = constant.ActivityStatus(status)
	}
	if name != "" {
		a.Name = name
	}
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Activity) SoftDelete() {
	now := time.Now()
	a.DeletedAt = &now
	a.UpdatedAt = now
}
