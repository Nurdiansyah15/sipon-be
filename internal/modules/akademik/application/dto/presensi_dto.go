package dto

// --- Halaman presensi (check-in via NIS) ---

type PresensiSessionInfo struct {
	ID           string `json:"id"`
	ActivityName string `json:"activity_name"`
	ActivityCode string `json:"activity_code"`
	ScheduleType string `json:"schedule_type"`
	StartsAt     string `json:"starts_at"`
	EndsAt       string `json:"ends_at"`
	Status       string `json:"status"`
	PeriodName   string `json:"period_name"`

	TotalEligible int `json:"total_eligible"`
	TotalPresent  int `json:"total_present"`
}

type CheckinRequest struct {
	NIS string `json:"nis" binding:"required"`
}

type CheckinResponse struct {
	Attendance AttendanceResponse `json:"attendance"`
	Message    string             `json:"message"`
}

type PresensiAttendanceItem struct {
	SantriID   string  `json:"santri_id"`
	NIS        *string `json:"nis,omitempty"`
	Fullname   *string `json:"fullname,omitempty"`
	Status     string  `json:"status"`
	RecordedAt string  `json:"recorded_at"`
}
