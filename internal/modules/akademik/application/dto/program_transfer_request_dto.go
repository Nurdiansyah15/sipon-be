package dto

import "time"

type ProgramBrief struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type SantriProgramAdminResponse struct {
	SantriID  string       `json:"santri_id"`
	ProgramID string       `json:"program_id"`
	Program   ProgramBrief `json:"program"`
	IsActive  bool         `json:"is_active"`
}

type SantriProgramListItem struct {
	SantriID  string        `json:"santri_id"`
	NIS       *string       `json:"nis,omitempty"`
	Fullname  *string       `json:"fullname,omitempty"`
	ProgramID string        `json:"program_id"`
	Program   *ProgramBrief `json:"program,omitempty"`
}

type SantriProgramListQuery struct {
	Search *string `form:"search"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}

type ProgramTransferRequestResponse struct {
	ID            string        `json:"id"`
	SantriID      string        `json:"santri_id"`
	SantriName    *string       `json:"santri_name,omitempty"`
	FromProgramID string        `json:"from_program_id"`
	FromProgram   *ProgramBrief `json:"from_program,omitempty"`
	ToProgramID   string        `json:"to_program_id"`
	ToProgram     *ProgramBrief `json:"to_program,omitempty"`
	Status        string        `json:"status"`
	Notes         *string       `json:"notes,omitempty"`
	AdminNotes    *string       `json:"admin_notes,omitempty"`
	ReviewedBy    *string       `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time    `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

type RequestProgramTransferRequest struct {
	ToProgramID string  `json:"to_program_id" binding:"required"`
	Notes       *string `json:"notes"`
}

type RejectProgramTransferRequest struct {
	AdminNotes *string `json:"admin_notes"`
}

type ProgramTransferRequestListQuery struct {
	Status *string `form:"status"`
	Page   int     `form:"page"`
	Limit  int     `form:"limit"`
}
