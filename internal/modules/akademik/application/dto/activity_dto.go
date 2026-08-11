package dto

import "time"

type CreateActivityRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type UpdateActivityRequest struct {
	Code   *string `json:"code,omitempty"`
	Name   *string `json:"name,omitempty"`
	Status *string `json:"status,omitempty"`
}

type ActivityListQuery struct {
	Status *string `form:"status"`
	Search *string `form:"search"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type ActivityResponse struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
