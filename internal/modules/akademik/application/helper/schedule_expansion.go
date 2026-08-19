package helper

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/activity_schedule/constant"
	schEntity "sipon-be/internal/modules/akademik/domain/activity_schedule/entity"
	"sipon-be/internal/shared/timeutil"
)

// yearlyKey digunakan untuk mencocokkan tanggal pada recurrence tahunan.
type yearlyKey struct {
	month int
	day   int
}

// ExpandScheduleDates menghitung daftar tanggal (date-only) dalam rentang
// [from, to] di mana jadwal terjadi, sesuai recurrence pattern-nya. Rentang
// dibatasi oleh start_date/end_date jadwal bila ada.
//
// weeklyDays/monthlyDays/yearlyDates adalah detail recurrence yang dimuat
// dari repository. Mengembalikan daftar tanggal unik terurut naik dalam
// platform timezone (timeutil).
func ExpandScheduleDates(
	s *schEntity.ActivitySchedule,
	weeklyDays []constant.DayOfWeek,
	monthlyDays []int,
	yearlyDates []schEntity.YearlyDate,
	from, to time.Time,
) []time.Time {
	from = timeutil.DateOnly(from)
	to = timeutil.DateOnly(to)

	start := from
	if s.StartDate != nil && s.StartDate.After(start) {
		start = timeutil.DateOnly(*s.StartDate)
	}
	end := to
	if s.EndDate != nil && s.EndDate.Before(end) {
		end = timeutil.DateOnly(*s.EndDate)
	}
	if start.After(end) {
		return nil
	}

	weekSet := make(map[constant.DayOfWeek]struct{}, len(weeklyDays))
	for _, d := range weeklyDays {
		weekSet[d] = struct{}{}
	}
	monthSet := make(map[int]struct{}, len(monthlyDays))
	for _, d := range monthlyDays {
		monthSet[d] = struct{}{}
	}
	yearSet := make(map[yearlyKey]struct{}, len(yearlyDates))
	for _, yd := range yearlyDates {
		yearSet[yearlyKey{month: yd.Month, day: yd.Day}] = struct{}{}
	}

	dates := make([]time.Time, 0)
	switch s.Type {
	case constant.ActivityScheduleTypeOnce:
		if s.StartDate != nil {
			once := timeutil.DateOnly(*s.StartDate)
			if !once.Before(start) && !once.After(end) {
				dates = append(dates, once)
			}
		}
		return dates
	case constant.ActivityScheduleTypeDaily:
		for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
			dates = append(dates, d)
		}
		return dates
	}

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		switch s.Type {
		case constant.ActivityScheduleTypeWeekly:
			dow := constant.DayOfWeek(toDayOfWeek(d))
			if _, ok := weekSet[dow]; ok {
				dates = append(dates, d)
			}
		case constant.ActivityScheduleTypeMonthly:
			if _, ok := monthSet[d.Day()]; ok {
				dates = append(dates, d)
			}
		case constant.ActivityScheduleTypeYearly:
			if _, ok := yearSet[yearlyKey{month: int(d.Month()), day: d.Day()}]; ok {
				dates = append(dates, d)
			}
		}
	}
	return dates
}

func toDayOfWeek(t time.Time) string {
	names := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	return names[int(t.Weekday())]
}
