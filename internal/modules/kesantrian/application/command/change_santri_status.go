package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	"sipon-be/internal/shared/kernel"
)

type ChangeSantriStatusUseCase struct {
	santriRepo santrirepo.SantriRepository
}

func NewChangeSantriStatusUseCase(santriRepo santrirepo.SantriRepository) *ChangeSantriStatusUseCase {
	return &ChangeSantriStatusUseCase{santriRepo: santriRepo}
}

type ChangeSantriStatusRequest struct {
	Status string  `json:"status" binding:"required"`
	Notes  *string `json:"notes,omitempty"`
}

func (uc *ChangeSantriStatusUseCase) Execute(ctx context.Context, santriID, changedBy string, req ChangeSantriStatusRequest) (*dto.MessageResponse, error) {
	santri, err := uc.santriRepo.FindByID(ctx, santriID)
	if err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	switch santriconstant.SantriStatus(req.Status) {
	case santriconstant.SantriStatusAlumni:
		if err := santri.MarkAlumni(changedBy); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) && ke.Code == santriconstant.CodeSantriInvalidStatus {
				return nil, kernel.New(application.ErrCodeConflict)
			}
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	case santriconstant.SantriStatusDropOut:
		if err := santri.MarkDropOut(changedBy, req.Notes); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) && ke.Code == santriconstant.CodeSantriInvalidStatus {
				return nil, kernel.New(application.ErrCodeConflict)
			}
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	default:
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.santriRepo.Update(ctx, santri); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "status santri berhasil diubah"}, nil
}
