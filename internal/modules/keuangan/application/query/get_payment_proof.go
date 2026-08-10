package query

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
	"sipon-be/internal/shared/kernel"
)

const paymentProofDownloadTTL = 15 * time.Minute

type GetPaymentProofUseCase struct {
	paymentRepo  payRepo.PaymentRepository
	fileUploader ports.FileUploader
}

func NewGetPaymentProofUseCase(paymentRepo payRepo.PaymentRepository, fileUploader ports.FileUploader) *GetPaymentProofUseCase {
	return &GetPaymentProofUseCase{paymentRepo: paymentRepo, fileUploader: fileUploader}
}

func (uc *GetPaymentProofUseCase) Execute(ctx context.Context, paymentID string) (*dto.PaymentProofResponse, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case payConst.CodePaymentNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if payment.ProofKey == nil || *payment.ProofKey == "" {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "Bukti transfer tidak tersedia", nil)
	}

	url, err := uc.fileUploader.GeneratePresignedDownloadURL(ctx, *payment.ProofKey, paymentProofDownloadTTL, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membuat URL unduhan bukti transfer", err)
	}

	return &dto.PaymentProofResponse{
		URL:       url,
		ExpiresIn: int(paymentProofDownloadTTL.Seconds()),
	}, nil
}
