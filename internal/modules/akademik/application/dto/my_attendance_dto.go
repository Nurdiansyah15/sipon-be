package dto

// --- History absensi santri (portal santri) ---

type MyAttendanceSessionItem struct {
	SessionID    string `json:"session_id"`
	ActivityName string `json:"activity_name"`
	ActivityCode string `json:"activity_code"`
	ScheduleType string `json:"schedule_type"`
	StartsAt     string `json:"starts_at"`
	EndsAt       string `json:"ends_at"`
	Status       string `json:"status"` // present / absent / excused / unrecorded
	RecordedAt   *string `json:"recorded_at,omitempty"`
}

type MyAttendanceSummary struct {
	TotalSessions int `json:"total_sessions"`
	Present       int `json:"present"`
	Absent        int `json:"absent"`
	Excused       int `json:"excused"`
	Unrecorded    int `json:"unrecorded"`
}

type MyAttendanceResponse struct {
	AcademicPeriod *AcademicPeriodResponse   `json:"academic_period"`
	Summary        MyAttendanceSummary       `json:"summary"`
	Sessions       []MyAttendanceSessionItem `json:"sessions"`
}

type MyAttendanceListQuery struct {
	AcademicPeriodID   string `form:"academic_period_id"`
	ActivityScheduleID string `form:"activity_schedule_id"`
}
