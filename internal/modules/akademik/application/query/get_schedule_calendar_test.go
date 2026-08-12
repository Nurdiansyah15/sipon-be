package query

import (
	"testing"
	"time"

	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
)

func mustTime(v string) time.Time {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		panic(err)
	}
	return t
}

func mustTimePtr(v string) *time.Time {
	t := mustTime(v)
	return &t
}

func buildCalendar(byDate map[string][]dto.ScheduleCalendarItem) []dto.ScheduleCalendarDay {
	return buildCalendarDays(byDate)
}

func TestExpandDaily(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		ID:        "s1",
		Type:      constant.ActivityScheduleTypeDaily,
		StartDate: mustTimePtr("2026-08-10"),
		EndDate:   mustTimePtr("2026-08-12"),
	}
	item := dto.ScheduleCalendarItem{ID: "s1", Type: "daily"}
	byDate := map[string][]dto.ScheduleCalendarItem{}

	uc := &GetScheduleCalendarUseCase{}
	uc.expand(s, &scheduleRecurrence{}, mustTime("2026-08-01"), mustTime("2026-08-31"), byDate, item)

	if len(byDate) != 3 {
		t.Fatalf("expected 3 days, got %d: %v", len(byDate), byDate)
	}
	for _, want := range []string{"2026-08-10", "2026-08-11", "2026-08-12"} {
		if _, ok := byDate[want]; !ok {
			t.Fatalf("missing %s in %v", want, byDate)
		}
	}
}

func TestExpandWeekly(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		ID:   "s1",
		Type: constant.ActivityScheduleTypeWeekly,
	}
	// 2026-08-03 is Monday, 2026-08-09 is Sunday.
	rec := &scheduleRecurrence{weeklyDays: []constant.DayOfWeek{
		constant.DayOfWeekMonday,
		constant.DayOfWeekFriday,
	}}
	item := dto.ScheduleCalendarItem{ID: "s1", Type: "weekly"}
	byDate := map[string][]dto.ScheduleCalendarItem{}

	uc := &GetScheduleCalendarUseCase{}
	uc.expand(s, rec, mustTime("2026-08-01"), mustTime("2026-08-31"), byDate, item)

	for _, want := range []string{"2026-08-03", "2026-08-07", "2026-08-10", "2026-08-14", "2026-08-17", "2026-08-21", "2026-08-24", "2026-08-28", "2026-08-31"} {
		if _, ok := byDate[want]; !ok {
			t.Fatalf("missing %s in %v", want, byDate)
		}
	}
	if len(byDate) != 9 {
		t.Fatalf("expected 9 days, got %d", len(byDate))
	}
}

func TestExpandMonthly(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		ID:   "s1",
		Type: constant.ActivityScheduleTypeMonthly,
	}
	rec := &scheduleRecurrence{monthlyDays: []int{1, 15}}
	item := dto.ScheduleCalendarItem{ID: "s1", Type: "monthly"}
	byDate := map[string][]dto.ScheduleCalendarItem{}

	uc := &GetScheduleCalendarUseCase{}
	uc.expand(s, rec, mustTime("2026-08-01"), mustTime("2026-08-31"), byDate, item)

	for _, want := range []string{"2026-08-01", "2026-08-15"} {
		if _, ok := byDate[want]; !ok {
			t.Fatalf("missing %s in %v", want, byDate)
		}
	}
}

func TestExpandYearly(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		ID:   "s1",
		Type: constant.ActivityScheduleTypeYearly,
	}
	rec := &scheduleRecurrence{yearlyDates: []schEntity.YearlyDate{
		{Month: 8, Day: 17},
	}}
	item := dto.ScheduleCalendarItem{ID: "s1", Type: "yearly"}
	byDate := map[string][]dto.ScheduleCalendarItem{}

	uc := &GetScheduleCalendarUseCase{}
	uc.expand(s, rec, mustTime("2026-08-01"), mustTime("2026-08-31"), byDate, item)

	if _, ok := byDate["2026-08-17"]; !ok {
		t.Fatalf("missing 2026-08-17 in %v", byDate)
	}
}

func TestExpandOnce(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		ID:        "s1",
		Type:      constant.ActivityScheduleTypeOnce,
		StartDate: mustTimePtr("2026-08-20"),
	}
	item := dto.ScheduleCalendarItem{ID: "s1", Type: "once"}
	byDate := map[string][]dto.ScheduleCalendarItem{}

	uc := &GetScheduleCalendarUseCase{}
	uc.expand(s, &scheduleRecurrence{}, mustTime("2026-08-01"), mustTime("2026-08-31"), byDate, item)

	if len(byDate) != 1 {
		t.Fatalf("expected 1 day, got %d: %v", len(byDate), byDate)
	}
	if _, ok := byDate["2026-08-20"]; !ok {
		t.Fatalf("missing 2026-08-20 in %v", byDate)
	}
}

func TestExpandRespectsValidity(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		ID:        "s1",
		Type:      constant.ActivityScheduleTypeDaily,
		StartDate: mustTimePtr("2026-09-01"),
		EndDate:   mustTimePtr("2026-09-30"),
	}
	item := dto.ScheduleCalendarItem{ID: "s1", Type: "daily"}
	byDate := map[string][]dto.ScheduleCalendarItem{}

	uc := &GetScheduleCalendarUseCase{}
	uc.expand(s, &scheduleRecurrence{}, mustTime("2026-08-01"), mustTime("2026-08-31"), byDate, item)

	if len(byDate) != 0 {
		t.Fatalf("expected 0 days outside validity, got %d: %v", len(byDate), byDate)
	}
}

func TestBuildCalendarDays(t *testing.T) {
	byDate := map[string][]dto.ScheduleCalendarItem{
		"2026-08-03": {{ID: "a", Type: "daily"}},
		"2026-08-05": {{ID: "b", Type: "weekly"}},
	}
	days := buildCalendar(byDate)
	if len(days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(days))
	}
}
