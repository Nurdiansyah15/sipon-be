package command

import (
	"context"
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
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	existing, _ := uc.assignmentRepo.FindActiveBySantriID(ctx, cmd.SantriID)
	if existing != nil {
		return nil, kernel.New(bsConst.CodeSchemeAssignmentExists)
	}

	effectiveFrom, err := time.Parse("2006-01-02", cmd.EffectiveFrom)
	if err != nil {
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	var effectiveUntil *time.Time
	if cmd.EffectiveUntil != nil {
		t, err := time.Parse("2006-01-02", *cmd.EffectiveUntil)
		if err != nil {
			return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
		}
		effectiveUntil = &t
	}

	assignment, err := bsEntity.NewSantriBillingAssignment(
		uuid.New().String(), cmd.SantriID, cmd.BillingSchemeID,
		cmd.AssignedBy, effectiveFrom, effectiveUntil,
	)
	if err != nil {
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	if err := uc.assignmentRepo.Save(ctx, assignment); err != nil {
		return nil, application.WrapRepoErr(err, bsConst.CodeBillingSchemeNotFound)
	}

	return &dto.MessageResponse{Message: "Skema berhasil di-assign ke santri"}, nil
}
