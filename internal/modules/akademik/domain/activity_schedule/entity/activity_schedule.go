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
	EarlyMinutes     int
	LateMinutes      int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
}

// EffectiveStart returns StartTime offset backward by EarlyMinutes.
func (s *ActivitySchedule) EffectiveStart(baseDate time.Time) time.Time {
	loc := baseDate.Location()
	t, _ := time.Parse("15:04:05", s.StartTime)
	return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		t.Hour(), t.Minute()-s.EarlyMinutes, t.Second(), 0, loc)
}

// EffectiveEnd returns EndTime offset forward by LateMinutes.
func (s *ActivitySchedule) EffectiveEnd(baseDate time.Time) time.Time {
	loc := baseDate.Location()
	t, _ := time.Parse("15:04:05", s.EndTime)
	return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		t.Hour(), t.Minute()+s.LateMinutes, t.Second(), 0, loc)
}

func NewActivitySchedule(id, activityPeriodID string, typ constant.ActivityScheduleType, startTime, endTime string, startDate, endDate *time.Time, earlyMinutes, lateMinutes int) (*ActivitySchedule, error) {
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
	if earlyMinutes < 0 {
		earlyMinutes = 0
	}
	if lateMinutes < 0 {
		lateMinutes = 0
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
		EarlyMinutes:     earlyMinutes,
		LateMinutes:      lateMinutes,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func (s *ActivitySchedule) Update(startTime, endTime string, startDate, endDate *time.Time, earlyMinutes, lateMinutes *int) error {
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
	if earlyMinutes != nil {
		s.EarlyMinutes = *earlyMinutes
		if s.EarlyMinutes < 0 {
			s.EarlyMinutes = 0
		}
	}
	if lateMinutes != nil {
		s.LateMinutes = *lateMinutes
		if s.LateMinutes < 0 {
			s.LateMinutes = 0
		}
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
