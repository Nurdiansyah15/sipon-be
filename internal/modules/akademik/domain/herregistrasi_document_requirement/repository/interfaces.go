package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/entity"
)

type HerregistrasiDocumentRequirementRepository interface {
	Save(ctx context.Context, req *entity.HerregistrasiDocumentRequirement) error
	Update(ctx context.Context, req *entity.HerregistrasiDocumentRequirement) error
	FindByID(ctx context.Context, id string) (*entity.HerregistrasiDocumentRequirement, error)
	FindByAcademicPeriod(ctx context.Context, academicPeriodID string) ([]*entity.HerregistrasiDocumentRequirement, error)
	Delete(ctx context.Context, id string) error
}
