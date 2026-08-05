package dto

type CreatePeriodRequest struct {
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type PeriodListQuery struct {
	Status *string `form:"status"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type PeriodResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Status    string  `json:"status"`
	ClosedBy  *string `json:"closed_by,omitempty"`
	ClosedAt  *string `json:"closed_at,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}
