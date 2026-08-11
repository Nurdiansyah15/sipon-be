package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/santri_program/constant"
	"sipon-be/internal/shared/kernel"
)

// SantriProgram memetakan santri ke program akademik. Santri bisa punya
// banyak record historis, tapi hanya satu yang aktif pada satu waktu.
type SantriProgram struct {
	ID        string
	SantriID  string
	ProgramID string
	IsActive  bool
	StartedAt time.Time
	EndedAt   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewSantriProgram(id, santriID, programID string) (*SantriProgram, error) {
	if id == "" || santriID == "" || programID == "" {
		return nil, kernel.New(constant.CodeSantriProgramNotFound)
	}
	now := time.Now()
	return &SantriProgram{
		ID:        id,
		SantriID:  santriID,
		ProgramID: programID,
		IsActive:  true,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Deactivate menandai program tidak aktif lagi dan mencatat waktu berakhir.
func (sp *SantriProgram) Deactivate() error {
	if !sp.IsActive {
		return kernel.New(constant.CodeSantriProgramAlreadyActive)
	}
	now := time.Now()
	sp.IsActive = false
	sp.EndedAt = &now
	sp.UpdatedAt = now
	return nil
}

func (sp *SantriProgram) SoftDelete() {
	now := time.Now()
	sp.DeletedAt = &now
	sp.UpdatedAt = now
}
