package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	requestconstant "sipon-be/internal/modules/kesantrian/domain/request/constant"
	requestrepo "sipon-be/internal/modules/kesantrian/domain/request/repository"
	"sipon-be/internal/shared/kernel"
)

type RejectSantriRequestUseCase struct {
	requestRepo requestrepo.SantriRequestRepository
	transactor  ports.Transactor
}

func NewRejectSantriRequestUseCase(requestRepo requestrepo.SantriRequestRepository, transactor ports.Transactor) *RejectSantriRequestUseCase {
	return &RejectSantriRequestUseCase{requestRepo: requestRepo, transactor: transactor}
}

func (uc *RejectSantriRequestUseCase) Execute(ctx context.Context, reviewerID, requestID string, req dto.RejectSantriRequestRequest) (*dto.MessageResponse, error) {
	sr, err := uc.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, application.WrapRepoErr(err, requestconstant.CodeSantriRequestNotFound)
	}

	if err := sr.Reject(reviewerID, req.Notes); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == requestconstant.CodeSantriRequestInvalidStatus {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.requestRepo.Update(txCtx, sr)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "permintaan santri berhasil ditolak"}, nil
}
