package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	bsRepo "sipon-be/internal/modules/keuangan/domain/billingscheme/repository"
	"sipon-be/internal/shared/kernel"
)

type ListAssignmentsUseCase struct {
	assignmentReader ports.AssignmentReader
	schemeRepo       bsRepo.BillingSchemeRepository
}

func NewListAssignmentsUseCase(assignmentReader ports.AssignmentReader, schemeRepo bsRepo.BillingSchemeRepository) *ListAssignmentsUseCase {
	return &ListAssignmentsUseCase{assignmentReader: assignmentReader, schemeRepo: schemeRepo}
}

func (uc *ListAssignmentsUseCase) Execute(ctx context.Context, q dto.AssignmentListQuery) ([]dto.AssignmentResponse, error) {
	items, err := uc.assignmentReader.ListAll(ctx, q.SantriID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	schemeNames := map[string]string{}
	schemes, err := uc.schemeRepo.List(ctx, bsRepo.BillingSchemeListQuery{Page: 1, Limit: 10000})
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	for _, s := range schemes.Items {
		schemeNames[s.ID] = s.Name
	}

	results := make([]dto.AssignmentResponse, 0, len(items))
	for _, a := range items {
		result := dto.AssignmentResponse{
			ID:              a.ID,
			SantriID:        a.SantriID,
			BillingSchemeID: a.BillingSchemeID,
			EffectiveFrom:   a.EffectiveFrom.Format("2006-01-02"),
			AssignedBy:      a.AssignedBy,
			CreatedAt:       a.CreatedAt.Format("2006-01-02"),
		}
		if a.EffectiveUntil != nil {
			s := a.EffectiveUntil.Format("2006-01-02")
			result.EffectiveUntil = &s
		}
		if name, ok := schemeNames[a.BillingSchemeID]; ok {
			result.BillingScheme = &dto.BillingSchemeBriefResponse{ID: a.BillingSchemeID, Name: name}
		}
		results = append(results, result)
	}

	return results, nil
}
