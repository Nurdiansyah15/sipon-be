package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	"sipon-be/internal/shared/kernel"
)

type ActivitySchedule struct {
	ID               string
	ActivityPeriodID string
	Type             constant.ActivityScheduleType
	StartDate        *time.Time
	EndDate          *time.Time
	StartTime        string
	EndTime          string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

func NewActivitySchedule(id, activityPeriodID string, typ constant.ActivityScheduleType, startTime, endTime string, startDate, endDate *time.Time) (*ActivitySchedule, error) {
	if id == "" || activityPeriodID == "" {
		return nil, kernel.New(constant.CodeActivityScheduleNotFound)
	}
	if typ != constant.ActivityScheduleTypeOnce &&
		typ != constant.ActivityScheduleTypeDaily &&
		typ != constant.ActivityScheduleTypeWeekly &&
		typ != constant.ActivityScheduleTypeMonthly &&
		typ != constant.ActivityScheduleTypeYearly {
		return nil, kernel.New(constant.CodeActivityScheduleInvalid)
	}
	if err := validateTime(startTime); err != nil {
		return nil, err
	}
	if err := validateTime(endTime); err != nil {
		return nil, err
	}
	if endTime <= startTime {
		return nil, kernel.New(constant.CodeActivityScheduleInvalid)
	}
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return nil, kernel.New(constant.CodeActivityScheduleInvalid)
	}
	now := time.Now()
	return &ActivitySchedule{
		ID:               id,
		ActivityPeriodID: activityPeriodID,
		Type:             typ,
		StartDate:        startDate,
		EndDate:          endDate,
		StartTime:        startTime,
		EndTime:          endTime,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (s *ActivitySchedule) Update(startTime, endTime string, startDate, endDate *time.Time) error {
	if startTime != "" {
		if err := validateTime(startTime); err != nil {
			return err
		}
		s.StartTime = startTime
	}
	if endTime != "" {
		if err := validateTime(endTime); err != nil {
			return err
		}
		s.EndTime = endTime
	}
	if s.EndTime <= s.StartTime {
		return kernel.New(constant.CodeActivityScheduleInvalid)
	}
	if startDate != nil {
		s.StartDate = startDate
	}
	if endDate != nil {
		s.EndDate = endDate
	}
	if s.StartDate != nil && s.EndDate != nil && s.EndDate.Before(*s.StartDate) {
		return kernel.New(constant.CodeActivityScheduleInvalid)
	}
	s.UpdatedAt = time.Now()
	return nil
}

func (s *ActivitySchedule) SoftDelete() {
	now := time.Now()
	s.DeletedAt = &now
	s.UpdatedAt = now
}

func validateTime(v string) error {
	if len(v) != 8 || v[2] != ':' || v[5] != ':' {
		return kernel.New(constant.CodeActivityScheduleInvalid)
	}
	return nil
}

type ActivityScheduleWeekly struct {
	ID         string
	ScheduleID string
	DayOfWeek  constant.DayOfWeek
}

type ActivityScheduleMonthly struct {
	ID         string
	ScheduleID string
	DayOfMonth int
}

type YearlyDate struct {
	Month int
	Day   int
}

type ActivityScheduleYearly struct {
	ID         string
	ScheduleID string
	Month      int
	Day        int
}
