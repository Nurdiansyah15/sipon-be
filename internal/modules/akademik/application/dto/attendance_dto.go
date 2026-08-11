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
	Status            string    `json:"status"`
	RecordedAt        time.Time `json:"recorded_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
