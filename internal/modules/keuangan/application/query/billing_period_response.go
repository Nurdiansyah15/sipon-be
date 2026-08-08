package query

import (
	"sipon-be/internal/modules/keuangan/application/dto"
	bpEntity "sipon-be/internal/modules/keuangan/domain/billingperiod/entity"
)

func buildBillingPeriodResponse(p *bpEntity.BillingPeriod) dto.BillingPeriodResponse {
	return dto.BillingPeriodResponse{
		ID:         p.ID,
		Name:       p.Name,
		PeriodType: string(p.PeriodType),
		StartDate:  p.StartDate.Format("2006-01-02"),
		EndDate:    p.EndDate.Format("2006-01-02"),
		Status:     string(p.Status),
		CreatedAt:  p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func buildBillingPeriodBrief(p *bpEntity.BillingPeriod) dto.BillingPeriodBriefResponse {
	return dto.BillingPeriodBriefResponse{
		ID:     p.ID,
		Name:   p.Name,
		Status: string(p.Status),
	}
}
