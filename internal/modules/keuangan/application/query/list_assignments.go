package query

import (
	"context"

	"sipon-be/internal/modules/keuangan/application"
	"sipon-be/internal/modules/keuangan/application/dto"
	"sipon-be/internal/modules/keuangan/application/ports"
	"sipon-be/internal/shared/kernel"
)

type ListAssignmentsUseCase struct {
	assignmentReader ports.AssignmentReader
}

func NewListAssignmentsUseCase(assignmentReader ports.AssignmentReader) *ListAssignmentsUseCase {
	return &ListAssignmentsUseCase{assignmentReader: assignmentReader}
}

func (uc *ListAssignmentsUseCase) Execute(ctx context.Context) ([]dto.AssignmentResponse, error) {
	items, err := uc.assignmentReader.ListActive(ctx)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
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
		results = append(results, result)
	}

	return results, nil
}
