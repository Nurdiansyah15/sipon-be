package dto

import "time"

type AttendanceRecordInput struct {
	SantriID string `json:"santri_id" binding:"required"`
	Status   string `json:"status" binding:"required"`
}

type RecordAttendanceRequest struct {
	Records []AttendanceRecordInput `json:"records" binding:"required,min=1,dive"`
}

type UpdateAttendanceRequest struct {
	Status string `json:"status" binding:"required"`
}

type AttendanceResponse struct {
	ID                string    `json:"id"`
	ActivitySessionID string    `json:"activity_session_id"`
	SantriID          string    `json:"santri_id"`
	SantriNIS         *string   `json:"santri_nis,omitempty"`
	SantriName        *string   `json:"santri_name,omitempty"`
	Status            string    `json:"status"`
	RecordedAt        time.Time `json:"recorded_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// EligibleSantriResponse is a santri who has completed herregistrasi for the
// session's academic period and can therefore be recorded for attendance.
type EligibleSantriResponse struct {
	SantriID string  `json:"santri_id"`
	NIS      *string `json:"nis,omitempty"`
	Fullname *string `json:"fullname,omitempty"`
}
