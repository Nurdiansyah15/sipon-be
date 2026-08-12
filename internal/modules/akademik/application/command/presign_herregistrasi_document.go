package command

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	docConst "sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	"sipon-be/internal/shared/kernel"
)

const herregUploadExpiry = 15 * time.Minute

type PresignHerregistrasiDocumentUseCase struct {
	fileUploader ports.FileUploader
}

func NewPresignHerregistrasiDocumentUseCase(fileUploader ports.FileUploader) *PresignHerregistrasiDocumentUseCase {
	return &PresignHerregistrasiDocumentUseCase{fileUploader: fileUploader}
}

func (uc *PresignHerregistrasiDocumentUseCase) Execute(ctx context.Context, req dto.HerregistrasiDocumentPresignRequest) (*dto.HerregistrasiDocumentPresignResponse, error) {
	if req.Kind == "" {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}
	if !docConst.AllowedContentTypes[req.ContentType] {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	ext := filepath.Ext(req.Filename)
	objectName := fmt.Sprintf("pending/akademik/herregistrasi/%s/%s%s", req.Kind, uuid.NewString(), ext)

	presignURL, key, publicURL, err := uc.fileUploader.RequestUpload(ctx, objectName, req.ContentType, herregUploadExpiry, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.HerregistrasiDocumentPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		PublicURL:  publicURL,
		ExpiresIn:  int(herregUploadExpiry.Seconds()),
	}, nil
}
