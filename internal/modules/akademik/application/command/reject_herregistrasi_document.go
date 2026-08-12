package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	docConst "sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	"sipon-be/internal/shared/kernel"
)

type RejectHerregistrasiDocumentUseCase struct {
	documentRepo docRepo.HerregistrasiDocumentRepository
}

func NewRejectHerregistrasiDocumentUseCase(documentRepo docRepo.HerregistrasiDocumentRepository) *RejectHerregistrasiDocumentUseCase {
	return &RejectHerregistrasiDocumentUseCase{documentRepo: documentRepo}
}

func (uc *RejectHerregistrasiDocumentUseCase) Execute(ctx context.Context, verifierID, documentID string, notes string) (*dto.HerregistrasiDocumentResponse, error) {
	doc, err := uc.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, docConst.CodeHerregistrasiDocumentNotFound)
	}
	if notes == "" {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	doc.Reject(verifierID, &notes)
	if err := uc.documentRepo.Update(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapHerregistrasiDocumentToResponse(doc, ""), nil
}
