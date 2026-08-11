package dto

import "time"

type CreateSantriRegistrationRequest struct {
	SantriID         string `json:"santri_id" binding:"required"`
	AcademicPeriodID string `json:"academic_period_id" binding:"required"`
}

type SantriRegistrationListQuery struct {
	AcademicPeriodID *string `form:"academic_period_id"`
	SantriID         *string `form:"santri_id"`
	Status           *string `form:"status"`
	Page             int     `form:"page"`
	Limit            int     `form:"limit"`
}

type SantriRegistrationResponse struct {
	ID               string     `json:"id"`
	SantriID         string     `json:"santri_id"`
	SantriNIS        *string    `json:"santri_nis,omitempty"`
	SantriName       *string    `json:"santri_name,omitempty"`
	AcademicPeriodID string     `json:"academic_period_id"`
	PeriodName       string     `json:"period_name,omitempty"`
	Status           string     `json:"status"`
	RegisteredAt     *time.Time `json:"registered_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
