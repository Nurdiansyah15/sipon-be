package dto

import "time"

// --- Santri Portal (non-admin) ---

type HerregistrasiStatus struct {
	// Status is one of: none, pending, completed, cancelled
	Status         string     `json:"status"`
	RegistrationID *string    `json:"registration_id,omitempty"`
	RegisteredAt   *time.Time `json:"registered_at,omitempty"`
	RevisionNotes  *string    `json:"revision_notes,omitempty"`
}

type MySummaryResponse struct {
	AcademicPeriod *AcademicPeriodResponse `json:"academic_period"`
	Herregistrasi  *HerregistrasiStatus    `json:"herregistrasi"`
	Program        *ProgramInfo            `json:"program"`
}

type ProgramInfo struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
}

type MyProgramResponse struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
}

type MyActivityResponse struct {
	ID               string `json:"id"`
	ActivityID       string `json:"activity_id"`
	ActivityCode     string `json:"activity_code"`
	ActivityName     string `json:"activity_name"`
	ActivityPeriodID string `json:"activity_period_id"`
	Status           string `json:"status"`
	ScheduleCount    int    `json:"schedule_count"`
}

type MyScheduleResponse struct {
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
}
