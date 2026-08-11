package dto

import "time"

type CreateSessionRequest struct {
	ActivityScheduleID string `json:"activity_schedule_id" binding:"required"`
	StartsAt           string `json:"starts_at" binding:"required"`
	EndsAt             string `json:"ends_at" binding:"required"`
}

type ActivitySessionListQuery struct {
	ActivityScheduleID *string `form:"activity_schedule_id"`
	AcademicPeriodID   *string `form:"academic_period_id"`
	Status             *string `form:"status"`
	StartDate          *string `form:"start_date"`
	EndDate            *string `form:"end_date"`
	Page               int     `form:"page"`
	Limit              int     `form:"limit"`
}

type AttendanceSummary struct {
	Total   int64 `json:"total"`
	Present int64 `json:"present"`
	Absent  int64 `json:"absent"`
	Excused int64 `json:"excused"`
}

type ActivitySessionResponse struct {
	ID                 string    `json:"id"`
	ActivityScheduleID string    `json:"activity_schedule_id"`
	ActivityName       string    `json:"activity_name,omitempty"`
	ActivityCode       string    `json:"activity_code,omitempty"`
	ScheduleType       string    `json:"schedule_type,omitempty"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ActivitySessionDetailResponse struct {
	ActivitySessionResponse
	AttendanceSummary *AttendanceSummary `json:"attendance_summary,omitempty"`
}
