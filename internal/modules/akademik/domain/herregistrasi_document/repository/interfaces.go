package repository

import (
	"context"

	"sipon-be/internal/modules/akademik/domain/herregistrasi_document/entity"
)

type HerregistrasiDocumentRepository interface {
	Save(ctx context.Context, doc *entity.HerregistrasiDocument) error
	Update(ctx context.Context, doc *entity.HerregistrasiDocument) error
	FindByID(ctx context.Context, id string) (*entity.HerregistrasiDocument, error)
	FindByRegistration(ctx context.Context, registrationID string) ([]*entity.HerregistrasiDocument, error)
	FindByRegistrationAndKind(ctx context.Context, registrationID, kind string) (*entity.HerregistrasiDocument, error)
	Delete(ctx context.Context, id string) error
}
