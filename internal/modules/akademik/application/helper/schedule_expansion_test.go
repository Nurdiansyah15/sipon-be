package helper

import (
	"testing"
	"time"

	schConst "sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
)

func mustParseDate(t *testing.T, v string) time.Time {
	t.Helper()
	tt, err := time.Parse("2006-01-02", v)
	if err != nil {
		t.Fatalf("parse %q: %v", v, err)
	}
	return tt
}

func mustDatePtr(t *testing.T, v string) *time.Time {
	t.Helper()
	tt := mustParseDate(t, v)
	return &tt
}

func dateStrings(dates []time.Time) []string {
	out := make([]string, 0, len(dates))
	for _, d := range dates {
		out = append(out, d.Format("2006-01-02"))
	}
	return out
}

func TestExpandScheduleDatesDaily(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		Type:      schConst.ActivityScheduleTypeDaily,
		StartDate: mustDatePtr(t, "2026-08-10"),
		EndDate:   mustDatePtr(t, "2026-08-12"),
	}
	dates := ExpandScheduleDates(s, nil, nil, nil, mustParseDate(t, "2026-08-01"), mustParseDate(t, "2026-08-31"))
	got := dateStrings(dates)
	want := []string{"2026-08-10", "2026-08-11", "2026-08-12"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dates mismatch: want %v, got %v", want, got)
		}
	}
}

func TestExpandScheduleDatesWeekly(t *testing.T) {
	s := &schEntity.ActivitySchedule{Type: schConst.ActivityScheduleTypeWeekly}
	// 2026-08-03 is Monday, 2026-08-09 is Sunday.
	dates := ExpandScheduleDates(s,
		[]schConst.DayOfWeek{schConst.DayOfWeekMonday, schConst.DayOfWeekFriday},
		nil, nil,
		mustParseDate(t, "2026-08-01"), mustParseDate(t, "2026-08-31"))
	want := []string{"2026-08-03", "2026-08-07", "2026-08-10", "2026-08-14", "2026-08-17", "2026-08-21", "2026-08-24", "2026-08-28", "2026-08-31"}
	got := dateStrings(dates)
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("dates mismatch: want %v, got %v", want, got)
		}
	}
}

func TestExpandScheduleDatesMonthly(t *testing.T) {
	s := &schEntity.ActivitySchedule{Type: schConst.ActivityScheduleTypeMonthly}
	dates := ExpandScheduleDates(s, nil, []int{1, 15}, nil,
		mustParseDate(t, "2026-08-01"), mustParseDate(t, "2026-08-31"))
	if got := dateStrings(dates); len(got) != 2 || got[0] != "2026-08-01" || got[1] != "2026-08-15" {
		t.Fatalf("expected [2026-08-01 2026-08-15], got %v", got)
	}
}

func TestExpandScheduleDatesYearly(t *testing.T) {
	s := &schEntity.ActivitySchedule{Type: schConst.ActivityScheduleTypeYearly}
	dates := ExpandScheduleDates(s, nil, nil, []schEntity.YearlyDate{{Month: 8, Day: 17}},
		mustParseDate(t, "2026-08-01"), mustParseDate(t, "2026-08-31"))
	if got := dateStrings(dates); len(got) != 1 || got[0] != "2026-08-17" {
		t.Fatalf("expected [2026-08-17], got %v", got)
	}
}

func TestExpandScheduleDatesOnce(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		Type:      schConst.ActivityScheduleTypeOnce,
		StartDate: mustDatePtr(t, "2026-08-20"),
	}
	dates := ExpandScheduleDates(s, nil, nil, nil,
		mustParseDate(t, "2026-08-01"), mustParseDate(t, "2026-08-31"))
	if got := dateStrings(dates); len(got) != 1 || got[0] != "2026-08-20" {
		t.Fatalf("expected [2026-08-20], got %v", got)
	}
}

func TestExpandScheduleDatesRespectsValidity(t *testing.T) {
	s := &schEntity.ActivitySchedule{
		Type:      schConst.ActivityScheduleTypeDaily,
		StartDate: mustDatePtr(t, "2026-09-01"),
		EndDate:   mustDatePtr(t, "2026-09-30"),
	}
	dates := ExpandScheduleDates(s, nil, nil, nil,
		mustParseDate(t, "2026-08-01"), mustParseDate(t, "2026-08-31"))
	if len(dates) != 0 {
		t.Fatalf("expected 0 dates outside validity, got %v", dateStrings(dates))
	}
}

func TestExpandScheduleDatesNilStartDateOnce(t *testing.T) {
	// Jadwal "once" tanpa start_date → tidak menghasilkan tanggal.
	s := &schEntity.ActivitySchedule{Type: schConst.ActivityScheduleTypeOnce}
	dates := ExpandScheduleDates(s, nil, nil, nil,
		mustParseDate(t, "2026-08-01"), mustParseDate(t, "2026-08-31"))
	if len(dates) != 0 {
		t.Fatalf("expected 0 dates, got %v", dateStrings(dates))
	}
}

func TestExpandScheduleDatesEmptyRange(t *testing.T) {
	s := &schEntity.ActivitySchedule{Type: schConst.ActivityScheduleTypeDaily}
	// from setelah to → rentang kosong.
	dates := ExpandScheduleDates(s, nil, nil, nil,
		mustParseDate(t, "2026-08-31"), mustParseDate(t, "2026-08-01"))
	if len(dates) != 0 {
		t.Fatalf("expected 0 dates for empty range, got %v", dateStrings(dates))
	}
}
