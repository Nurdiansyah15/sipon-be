package dto

type CreateFeeComponentRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	IsPeriodic  bool    `json:"is_periodic"`
	PeriodType  *string `json:"period_type,omitempty"`
	Description *string `json:"description,omitempty"`
}

type UpdateFeeComponentRequest struct {
	Name        string  `json:"name" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	IsPeriodic  bool    `json:"is_periodic"`
	PeriodType  *string `json:"period_type,omitempty"`
	Description *string `json:"description,omitempty"`
}

type FeeComponentListQuery struct {
	Type   *string `form:"type"`
	Active *bool   `form:"is_active"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type FeeComponentResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	IsPeriodic  bool    `json:"is_periodic"`
	PeriodType  *string `json:"period_type,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    bool    `json:"is_active"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
