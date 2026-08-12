package command

import (
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/herregistrasi_document/entity"
	"sipon-be/internal/shared/timeutil"
)

func MapHerregistrasiDocumentToResponse(doc *entity.HerregistrasiDocument, kindLabel string) *dto.HerregistrasiDocumentResponse {
	return &dto.HerregistrasiDocumentResponse{
		ID:                   doc.ID,
		SantriRegistrationID: doc.SantriRegistrationID,
		Kind:                 doc.Kind,
		KindLabel:            kindLabel,
		Key:                  doc.Key,
		OriginalFilename:     doc.OriginalFilename,
		MimeType:             doc.MimeType,
		Size:                 doc.Size,
		Status:               string(doc.Status),
		Notes:                doc.Notes,
		VerifiedBy:           doc.VerifiedBy,
		VerifiedAt:           timeutil.ToPlatformPtr(doc.VerifiedAt),
		CreatedAt:            timeutil.ToPlatform(doc.CreatedAt),
		UpdatedAt:            timeutil.ToPlatform(doc.UpdatedAt),
	}
}
