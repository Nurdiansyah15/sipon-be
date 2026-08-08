package dto

type CreateBillingPeriodRequest struct {
	Name       string `json:"name" binding:"required"`
	PeriodType string `json:"period_type" binding:"required"`
	StartDate  string `json:"start_date" binding:"required"`
	EndDate    string `json:"end_date" binding:"required"`
}

type BillingPeriodListQuery struct {
	Status *string `form:"status"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type BillingPeriodResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PeriodType string `json:"period_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type BillingPeriodBriefResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
