package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/santri_registration/entity"
)

type SantriRegistrationListQuery struct {
	AcademicPeriodID *string
	SantriID         *string
	Status           *string
	Page             int
	Limit            int
}

type SantriRegistrationListResult struct {
	Items []*entity.SantriRegistration
	Total int64
}

type SantriRegistrationRepository interface {
	Save(ctx context.Context, registration *entity.SantriRegistration) error
	Update(ctx context.Context, registration *entity.SantriRegistration) error
	FindByID(ctx context.Context, id string) (*entity.SantriRegistration, error)
	FindBySantriAndPeriod(ctx context.Context, santriID, academicPeriodID string) (*entity.SantriRegistration, error)
	List(ctx context.Context, query SantriRegistrationListQuery) (*SantriRegistrationListResult, error)
	// ListCompletedByAcademicPeriod returns registrations with status
	// 'completed' for the given academic period (herregistrasi done).
	ListCompletedByAcademicPeriod(ctx context.Context, academicPeriodID string) ([]*entity.SantriRegistration, error)
}
