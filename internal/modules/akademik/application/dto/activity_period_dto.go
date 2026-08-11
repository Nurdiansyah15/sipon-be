package dto

import "time"

type CreateActivityPeriodRequest struct {
	ActivityID       string `json:"activity_id" binding:"required"`
	AcademicPeriodID string `json:"academic_period_id" binding:"required"`
}

type ActivityPeriodListQuery struct {
	ActivityID       *string `form:"activity_id"`
	AcademicPeriodID *string `form:"academic_period_id"`
	Status           *string `form:"status"`
	Page             int     `form:"page"`
	Limit            int     `form:"limit"`
}

type ActivityPeriodResponse struct {
	ID               string    `json:"id"`
	ActivityID       string    `json:"activity_id"`
	ActivityCode     string    `json:"activity_code,omitempty"`
	ActivityName     string    `json:"activity_name,omitempty"`
	AcademicPeriodID string    `json:"academic_period_id"`
	PeriodName       string    `json:"period_name,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AssignProgramRequest struct {
	ProgramID string `json:"program_id" binding:"required"`
}

type ActivityPeriodProgramResponse struct {
	ID               string `json:"id"`
	ActivityPeriodID string `json:"activity_period_id"`
	ProgramID        string `json:"program_id"`
	ProgramCode      string `json:"program_code,omitempty"`
	ProgramName      string `json:"program_name,omitempty"`
}
