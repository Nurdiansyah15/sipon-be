package dto

import "time"

type CreateScheduleRequest struct {
	ActivityPeriodID string  `json:"activity_period_id" binding:"required"`
	Type             string  `json:"type" binding:"required"`
	StartDate        *string `json:"start_date,omitempty"`
	EndDate          *string `json:"end_date,omitempty"`
	StartTime        string  `json:"start_time" binding:"required"`
	EndTime          string  `json:"end_time" binding:"required"`

	WeeklyDays  []string       `json:"weekly_days,omitempty"`
	MonthlyDays []int          `json:"monthly_days,omitempty"`
	YearlyDates []YearlyDateIn `json:"yearly_dates,omitempty"`
}

type YearlyDateIn struct {
	Month int `json:"month"`
	Day   int `json:"day"`
}

type UpdateScheduleRequest struct {
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`

	WeeklyDays  []string       `json:"weekly_days,omitempty"`
	MonthlyDays []int          `json:"monthly_days,omitempty"`
	YearlyDates []YearlyDateIn `json:"yearly_dates,omitempty"`
}

type ActivityScheduleResponse struct {
	ID               string    `json:"id"`
	ActivityPeriodID string    `json:"activity_period_id"`
	ActivityName     string    `json:"activity_name,omitempty"`
	ActivityCode     string    `json:"activity_code,omitempty"`
	Type             string    `json:"type"`
	StartDate        *string   `json:"start_date,omitempty"`
	EndDate          *string   `json:"end_date,omitempty"`
	StartTime        string    `json:"start_time"`
	EndTime          string    `json:"end_time"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ActivityScheduleDetailResponse struct {
	ID               string         `json:"id"`
	ActivityPeriodID string         `json:"activity_period_id"`
	ActivityName     string         `json:"activity_name,omitempty"`
	ActivityCode     string         `json:"activity_code,omitempty"`
	Type             string         `json:"type"`
	StartDate        *string        `json:"start_date,omitempty"`
	EndDate          *string        `json:"end_date,omitempty"`
	StartTime        string         `json:"start_time"`
	EndTime          string         `json:"end_time"`
	WeeklyDays       []string       `json:"weekly_days,omitempty"`
	MonthlyDays      []int          `json:"monthly_days,omitempty"`
	YearlyDates      []YearlyDateIn `json:"yearly_dates,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// --- Calendar ---

type ScheduleCalendarQuery struct {
	From             string `form:"from" binding:"required"`
	To               string `form:"to" binding:"required"`
	AcademicPeriodID string `form:"academic_period_id"`
	Types            string `form:"types"`
}

type ScheduleCalendarItem struct {
	ID               string `json:"id"`
	ActivityPeriodID string `json:"activity_period_id"`
	ActivityName     string `json:"activity_name,omitempty"`
	ActivityCode     string `json:"activity_code,omitempty"`
	Type             string `json:"type"`
	StartTime        string `json:"start_time"`
	EndTime          string `json:"end_time"`
}

type ScheduleCalendarDay struct {
	Date  string                 `json:"date"`
	Items []ScheduleCalendarItem `json:"items"`
}

type ScheduleCalendarResponse struct {
	From string               `json:"from"`
	To   string               `json:"to"`
	Days []ScheduleCalendarDay `json:"days"`
}
