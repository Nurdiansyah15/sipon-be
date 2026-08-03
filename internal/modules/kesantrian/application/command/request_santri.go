package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	requestconstant "sipon-be/internal/modules/kesantrian/domain/request/constant"
	requestentity "sipon-be/internal/modules/kesantrian/domain/request/entity"
	requestrepo "sipon-be/internal/modules/kesantrian/domain/request/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type RequestSantriUseCase struct {
	santriRepo  santrirepo.SantriRepository
	requestRepo requestrepo.SantriRequestRepository
	transactor  ports.Transactor
}

func NewRequestSantriUseCase(santriRepo santrirepo.SantriRepository, requestRepo requestrepo.SantriRequestRepository, transactor ports.Transactor) *RequestSantriUseCase {
	return &RequestSantriUseCase{santriRepo: santriRepo, requestRepo: requestRepo, transactor: transactor}
}

func (uc *RequestSantriUseCase) Execute(ctx context.Context, userID string) (*dto.RequestSantriResponse, error) {
	if _, err := uc.santriRepo.FindByUserID(ctx, userID); err == nil {
		return nil, kernel.New(application.ErrCodeConflict)
	} else {
		var ke *kernel.AppError
		if !errors.As(err, &ke) || ke.Code != santriconstant.CodeSantriNotFound {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	if _, err := uc.requestRepo.FindPendingByUserID(ctx, userID); err == nil {
		return nil, kernel.New(application.ErrCodeConflict)
	} else {
		var ke *kernel.AppError
		if !errors.As(err, &ke) || ke.Code != requestconstant.CodeSantriRequestNotFound {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	req, err := requestentity.NewSantriRequest(uuid.NewString(), userID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.requestRepo.Save(txCtx, req)
	}); err != nil {
		return nil, application.WrapConflictErr(err, requestconstant.CodeSantriRequestAlreadyExists)
	}

	return &dto.RequestSantriResponse{ID: req.ID, Message: "permintaan menjadi santri berhasil diajukan"}, nil
}
