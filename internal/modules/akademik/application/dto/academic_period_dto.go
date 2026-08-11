package dto

import "time"

type CreateAcademicPeriodRequest struct {
	Code      string `json:"code" binding:"required"`
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type UpdateAcademicPeriodRequest struct {
	Code      *string `json:"code,omitempty"`
	Name      *string `json:"name,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
}

type AcademicPeriodListQuery struct {
	Status *string `form:"status"`
	Search *string `form:"search"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type AcademicPeriodResponse struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
