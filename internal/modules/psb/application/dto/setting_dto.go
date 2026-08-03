package dto

import (
	"encoding/json"
	"time"
)

type CreateSettingRequest struct {
	Name         string           `json:"name" binding:"required"`
	StartPeriod  string           `json:"start_period" binding:"required"`
	EndPeriod    string           `json:"end_period" binding:"required"`
	Quota        json.RawMessage  `json:"quota"`
	RegFee       float64          `json:"reg_fee"`
	BankAccounts json.RawMessage  `json:"bank_accounts"`
}

type UpdateSettingRequest struct {
	Name         *string          `json:"name,omitempty"`
	StartPeriod  *string          `json:"start_period,omitempty"`
	EndPeriod    *string          `json:"end_period,omitempty"`
	Status       *string          `json:"status,omitempty"`
	Quota        json.RawMessage  `json:"quota,omitempty"`
	RegFee       *float64         `json:"reg_fee,omitempty"`
	BankAccounts json.RawMessage  `json:"bank_accounts,omitempty"`
}

type SettingResponse struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	StartPeriod  string           `json:"start_period"`
	EndPeriod    string           `json:"end_period"`
	Status       string           `json:"status"`
	Quota        json.RawMessage  `json:"quota"`
	RegFee       float64          `json:"reg_fee"`
	BankAccounts json.RawMessage  `json:"bank_accounts"`
	DataPurgedAt *time.Time       `json:"data_purged_at,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}
