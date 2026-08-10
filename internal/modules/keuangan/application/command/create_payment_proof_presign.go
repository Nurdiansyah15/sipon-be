package command

import (
	"context"
	"path"
	"strings"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

var paymentProofExtByContentType = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"application/pdf": ".pdf",
}

const paymentProofPresignTTL = 15 * time.Minute

type CreatePaymentProofPresignUseCase struct {
	fileUploader ports.FileUploader
}

func NewCreatePaymentProofPresignUseCase(fileUploader ports.FileUploader) *CreatePaymentProofPresignUseCase {
	return &CreatePaymentProofPresignUseCase{fileUploader: fileUploader}
}

func (uc *CreatePaymentProofPresignUseCase) Execute(ctx context.Context, userID string, req dto.PresignPaymentProofRequest) (*dto.PresignPaymentProofResponse, error) {
	ct := strings.ToLower(strings.TrimSpace(req.ContentType))
	ext, ok := paymentProofExtByContentType[ct]
	if !ok {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Format bukti transfer tidak didukung (jpg/png/webp/pdf)", nil)
	}

	objectName := path.Join("payment-proofs", userID, uuid.NewString()+ext)

	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, ct, paymentProofPresignTTL, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat URL presign bukti transfer", err)
	}

	return &dto.PresignPaymentProofResponse{
		PresignURL: presignURL,
		Key:        key,
		ExpiresIn:  int(paymentProofPresignTTL.Seconds()),
	}, nil
}
