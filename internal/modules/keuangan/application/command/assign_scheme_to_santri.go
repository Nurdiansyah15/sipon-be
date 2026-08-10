package command

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	bsConst "sipon-be/internal/modules/keuangan/domain/billingscheme/constant"
	bsEntity "sipon-be/internal/modules/keuangan/domain/billingscheme/entity"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	"sipon-be/internal/shared/kernel"
)

type AssignSchemeToSantriUseCase struct {
	assignmentRepo bsRepo.SantriBillingAssignmentRepository
	schemeRepo     bsRepo.BillingSchemeRepository
}

func NewAssignSchemeToSantriUseCase(assignmentRepo bsRepo.SantriBillingAssignmentRepository, schemeRepo bsRepo.BillingSchemeRepository) *AssignSchemeToSantriUseCase {
	return &AssignSchemeToSantriUseCase{assignmentRepo: assignmentRepo, schemeRepo: schemeRepo}
}

type AssignSchemeCmd struct {
	SantriID        string
	BillingSchemeID string
	EffectiveFrom   string
	EffectiveUntil  *string
	AssignedBy      string
}

func (uc *AssignSchemeToSantriUseCase) Execute(ctx context.Context, cmd AssignSchemeCmd) (*dto.MessageResponse, error) {
	_, err := uc.schemeRepo.FindByID(ctx, cmd.BillingSchemeID)
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

	effectiveFrom, err := time.Parse("2006-01-02", cmd.EffectiveFrom)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal berlaku tidak valid", err)
	}

	var effectiveUntil *time.Time
	if cmd.EffectiveUntil != nil {
		t, err := time.Parse("2006-01-02", *cmd.EffectiveUntil)
		if err != nil {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal berakhir tidak valid", err)
		}
		effectiveUntil = &t
	}
	if effectiveUntil != nil && !effectiveUntil.After(effectiveFrom) {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "tanggal berakhir harus setelah tanggal berlaku", nil)
	}

	existing, _ := uc.assignmentRepo.FindActiveBySantriID(ctx, cmd.SantriID)
	excludeID := ""
	if existing != nil {
		if !existing.EffectiveFrom.Before(effectiveFrom) {
			return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "tanggal berlaku skema baru harus setelah tanggal mulai skema yang sedang aktif", nil)
		}
		excludeID = existing.ID
	}

	overlap, err := uc.assignmentRepo.HasOverlappingAssignment(ctx, cmd.SantriID, effectiveFrom, effectiveUntil, excludeID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if overlap {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Santri sudah memiliki skema aktif pada rentang tanggal ini", nil)
	}

	if existing != nil {
		if err := uc.assignmentRepo.EndAssignment(ctx, existing.ID, effectiveFrom.AddDate(0, 0, -1)); err != nil {
			var ke *kernel.AppError
			if errors.As(err, &ke) {
				switch ke.Code {
				case bsConst.CodeBillingSchemeNotFound:
					return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
				}
			}
			return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
		}
	}

	assignment, err := bsEntity.NewSantriBillingAssignment(
		uuid.New().String(), cmd.SantriID, cmd.BillingSchemeID,
		cmd.AssignedBy, effectiveFrom, effectiveUntil,
	)
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

	if err := uc.assignmentRepo.Save(ctx, assignment); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case bsConst.CodeSchemeAssignmentExists:
				return nil, kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	message := "Skema berhasil ditetapkan ke santri"
	if existing != nil {
		message = "Skema santri berhasil diganti"
	}
	return &dto.MessageResponse{Message: message}, nil
}
