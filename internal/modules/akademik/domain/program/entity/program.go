package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/program/constant"
	"sipon-be/internal/shared/kernel"
)

type Program struct {
	ID        string
	Code      string
	Name      string
	Status    constant.ProgramStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewProgram(id, code, name string) (*Program, error) {
	if id == "" || code == "" || name == "" {
		return nil, kernel.New(constant.CodeProgramNotFound)
	}
	now := time.Now()
	return &Program{
		ID:        id,
		Code:      code,
		Name:      name,
		Status:    constant.ProgramStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *Program) Update(name, status string) error {
	if status != "" {
		if status != string(constant.ProgramStatusActive) && status != string(constant.ProgramStatusInactive) {
			return kernel.New(constant.CodeProgramInvalidStatus)
		}
		p.Status = constant.ProgramStatus(status)
	}
	if name != "" {
		p.Name = name
	}
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Program) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}
