package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	docConst "sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	"sipon-be/internal/shared/kernel"
)

type VerifyHerregistrasiDocumentUseCase struct {
	documentRepo docRepo.HerregistrasiDocumentRepository
}

func NewVerifyHerregistrasiDocumentUseCase(documentRepo docRepo.HerregistrasiDocumentRepository) *VerifyHerregistrasiDocumentUseCase {
	return &VerifyHerregistrasiDocumentUseCase{documentRepo: documentRepo}
}

func (uc *VerifyHerregistrasiDocumentUseCase) Execute(ctx context.Context, verifierID, documentID string) (*dto.HerregistrasiDocumentResponse, error) {
	doc, err := uc.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, docConst.CodeHerregistrasiDocumentNotFound)
	}
	if err := doc.Verify(verifierID); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}
	if err := uc.documentRepo.Update(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapHerregistrasiDocumentToResponse(doc, ""), nil
}
