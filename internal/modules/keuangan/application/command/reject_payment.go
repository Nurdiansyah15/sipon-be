package command

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	payConst "sipon-be/internal/modules/keuangan/domain/payment/constant"
	payRepo "sipon-be/internal/modules/keuangan/domain/payment/repository"
)

type RejectPaymentUseCase struct {
	paymentRepo payRepo.PaymentRepository
}

func NewRejectPaymentUseCase(paymentRepo payRepo.PaymentRepository) *RejectPaymentUseCase {
	return &RejectPaymentUseCase{paymentRepo: paymentRepo}
}

func (uc *RejectPaymentUseCase) Execute(ctx context.Context, paymentID string) (*dto.MessageResponse, error) {
	payment, err := uc.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	if err := payment.Reject(); err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentInvalidStatus)
	}

	if err := uc.paymentRepo.Update(ctx, payment); err != nil {
		return nil, application.WrapRepoErr(err, payConst.CodePaymentNotFound)
	}

	return &dto.MessageResponse{Message: "Pembayaran berhasil ditolak"}, nil
}
