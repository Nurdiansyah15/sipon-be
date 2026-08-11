package akademik

import (
	"context"
)

// SantriProgramInfo is a contract-level DTO for a santri's active program.
type SantriProgramInfo struct {
	SantriID    string
	ProgramID   string
	ProgramCode string
	ProgramName string
	IsActive    bool
}

// Contract is the outward-facing surface of the akademik module for other
// modules. Kesantrian & PSB use it to (a) resolve the default program and
// (b) persist the santri→program mapping.
type Contract interface {
	GetDefaultProgramID(ctx context.Context) (*string, error)
	AssignSantriProgram(ctx context.Context, santriID, programID string) error
	GetSantriProgram(ctx context.Context, santriID string) (*SantriProgramInfo, error)
}

var _ Contract = (*Module)(nil)
