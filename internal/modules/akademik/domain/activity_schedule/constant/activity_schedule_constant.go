package constant

import "sipon-be/internal/shared/kernel"

type ActivityScheduleType string

const (
	ActivityScheduleTypeOnce    ActivityScheduleType = "once"
	ActivityScheduleTypeDaily   ActivityScheduleType = "daily"
	ActivityScheduleTypeWeekly  ActivityScheduleType = "weekly"
	ActivityScheduleTypeMonthly ActivityScheduleType = "monthly"
	ActivityScheduleTypeYearly  ActivityScheduleType = "yearly"
)

type DayOfWeek string

const (
	DayOfWeekMonday    DayOfWeek = "monday"
	DayOfWeekTuesday   DayOfWeek = "tuesday"
	DayOfWeekWednesday DayOfWeek = "wednesday"
	DayOfWeekThursday  DayOfWeek = "thursday"
	DayOfWeekFriday    DayOfWeek = "friday"
	DayOfWeekSaturday  DayOfWeek = "saturday"
	DayOfWeekSunday    DayOfWeek = "sunday"
)

func IsValidDayOfWeek(v string) bool {
	switch DayOfWeek(v) {
	case DayOfWeekMonday, DayOfWeekTuesday, DayOfWeekWednesday, DayOfWeekThursday,
		DayOfWeekFriday, DayOfWeekSaturday, DayOfWeekSunday:
		return true
	}
	return false
}

const (
	CodeActivityScheduleNotFound          kernel.Code = "ACTIVITY_SCHEDULE_NOT_FOUND"
	CodeActivityScheduleInvalid           kernel.Code = "ACTIVITY_SCHEDULE_INVALID"
	CodeActivitySchedulePersistenceFailed kernel.Code = "ACTIVITY_SCHEDULE_PERSISTENCE_FAILED"
	CodeActivityScheduleQueryFailed       kernel.Code = "ACTIVITY_SCHEDULE_QUERY_FAILED"
)
