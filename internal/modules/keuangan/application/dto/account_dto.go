package dto

type CreateAccountRequest struct {
	Code          string  `json:"code" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	ParentID      *string `json:"parent_id,omitempty"`
	NormalBalance string  `json:"normal_balance" binding:"required"`
	Description   *string `json:"description,omitempty"`
	IsPostable    bool    `json:"is_postable"`
}

type UpdateAccountRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description,omitempty"`
	IsPostable  bool    `json:"is_postable"`
}

type AccountListQuery struct {
	Type   *string `form:"type"`
	Active *bool   `form:"active"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type AccountResponse struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	ParentID      *string `json:"parent_id,omitempty"`
	Level         int     `json:"level"`
	IsPostable    bool    `json:"is_postable"`
	NormalBalance string  `json:"normal_balance"`
	Description   *string `json:"description,omitempty"`
	IsActive      bool    `json:"is_active"`
	IsSystem      bool    `json:"is_system"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}
