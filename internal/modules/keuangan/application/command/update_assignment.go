package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateAssignmentUseCase struct {
	assignmentRepo bsRepo.SantriBillingAssignmentRepository
	schemeRepo     bsRepo.BillingSchemeRepository
}

func NewUpdateAssignmentUseCase(assignmentRepo bsRepo.SantriBillingAssignmentRepository, schemeRepo bsRepo.BillingSchemeRepository) *UpdateAssignmentUseCase {
	return &UpdateAssignmentUseCase{assignmentRepo: assignmentRepo, schemeRepo: schemeRepo}
}

func (uc *UpdateAssignmentUseCase) Execute(ctx context.Context, id string, req dto.UpdateAssignmentRequest) (*dto.MessageResponse, error) {
	assignment, err := uc.assignmentRepo.FindByID(ctx, id)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bsConst.CodeBillingSchemeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	_, err = uc.schemeRepo.FindByID(ctx, req.BillingSchemeID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bsConst.CodeBillingSchemeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	effectiveFrom, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal berlaku tidak valid", err)
	}

	var effectiveUntil *time.Time
	if req.EffectiveUntil != nil {
		t, err := time.Parse("2006-01-02", *req.EffectiveUntil)
		if err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal berakhir tidak valid", err)
		}
		effectiveUntil = &t
	}
	if effectiveUntil != nil && !effectiveUntil.After(effectiveFrom) {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "tanggal berakhir harus setelah tanggal berlaku", nil)
	}

	overlap, err := uc.assignmentRepo.HasOverlappingAssignment(ctx, assignment.SantriID, effectiveFrom, effectiveUntil, assignment.ID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if overlap {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Santri sudah memiliki skema aktif pada rentang tanggal ini", nil)
	}

	assignment.BillingSchemeID = req.BillingSchemeID
	assignment.EffectiveFrom = effectiveFrom
	assignment.EffectiveUntil = effectiveUntil

	if err := uc.assignmentRepo.Update(ctx, assignment); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bsConst.CodeBillingSchemeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return &dto.MessageResponse{Message: "Skema santri berhasil diperbarui"}, nil
}
