package dto

type CreateAccountRequest struct {
	Code          string  `json:"code" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	Type          string  `json:"type" binding:"required"`
	SubType       *string `json:"sub_type,omitempty"`
	ParentID      *string `json:"parent_id,omitempty"`
	NormalBalance string  `json:"normal_balance" binding:"required"`
	Description   *string `json:"description,omitempty"`
	IsPostable    bool    `json:"is_postable"`
}

type UpdateAccountRequest struct {
	Name        string  `json:"name" binding:"required"`
	SubType     *string `json:"sub_type,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPostable  bool    `json:"is_postable"`
}

type AccountListQuery struct {
	Type    *string `form:"type"`
	SubType *string `form:"sub_type"`
	Active  *bool   `form:"active"`
	Page    int     `form:"page"`
	Limit   int     `form:"limit"`
}

type AccountBriefResponse struct {
	ID      string  `json:"id"`
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Type    string  `json:"type"`
	SubType *string `json:"sub_type,omitempty"`
}

type AccountResponse struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	SubType       *string `json:"sub_type,omitempty"`
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
