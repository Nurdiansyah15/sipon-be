package entity

import (
	"time"

	"sipon-be/internal/modules/akademik/domain/academic_period/constant"
	"sipon-be/internal/shared/kernel"
)

type AcademicPeriod struct {
	ID        string
	Code      string
	Name      string
	StartDate time.Time
	EndDate   time.Time
	Status    constant.AcademicPeriodStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func NewAcademicPeriod(id, code, name string, startDate, endDate time.Time) (*AcademicPeriod, error) {
	if id == "" || code == "" || name == "" {
		return nil, kernel.New(constant.CodeAcademicPeriodNotFound)
	}
	if endDate.Before(startDate) {
		return nil, kernel.New(constant.CodeAcademicPeriodInvalidRange)
	}
	now := time.Now()
	return &AcademicPeriod{
		ID:        id,
		Code:      code,
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
		Status:    constant.AcademicPeriodStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (p *AcademicPeriod) Update(name string, startDate, endDate *time.Time) error {
	if startDate != nil && endDate != nil && endDate.Before(*startDate) {
		return kernel.New(constant.CodeAcademicPeriodInvalidRange)
	}
	if name != "" {
		p.Name = name
	}
	if startDate != nil {
		p.StartDate = *startDate
	}
	if endDate != nil {
		p.EndDate = *endDate
	}
	p.UpdatedAt = time.Now()
	return nil
}

func (p *AcademicPeriod) Open() error {
	if p.Status != constant.AcademicPeriodStatusDraft {
		return kernel.New(constant.CodeAcademicPeriodInvalidStatus)
	}
	p.Status = constant.AcademicPeriodStatusOpen
	p.UpdatedAt = time.Now()
	return nil
}

func (p *AcademicPeriod) Close() error {
	if p.Status != constant.AcademicPeriodStatusOpen {
		return kernel.New(constant.CodeAcademicPeriodInvalidStatus)
	}
	p.Status = constant.AcademicPeriodStatusClosed
	p.UpdatedAt = time.Now()
	return nil
}

func (p *AcademicPeriod) Archive() error {
	if p.Status != constant.AcademicPeriodStatusClosed {
		return kernel.New(constant.CodeAcademicPeriodInvalidStatus)
	}
	p.Status = constant.AcademicPeriodStatusArchived
	p.UpdatedAt = time.Now()
	return nil
}

func (p *AcademicPeriod) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}
