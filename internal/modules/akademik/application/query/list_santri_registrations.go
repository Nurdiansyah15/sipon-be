package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type ListSantriRegistrationsUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
	periodRepo       periodRepo.AcademicPeriodRepository
	kesantrianReader ports.KesantrianReader
}

func NewListSantriRegistrationsUseCase(
	registrationRepo regRepo.SantriRegistrationRepository,
	periodRepo periodRepo.AcademicPeriodRepository,
	kesantrianReader ports.KesantrianReader,
) *ListSantriRegistrationsUseCase {
	return &ListSantriRegistrationsUseCase{registrationRepo: registrationRepo, periodRepo: periodRepo, kesantrianReader: kesantrianReader}
}

func (uc *ListSantriRegistrationsUseCase) Execute(ctx context.Context, q dto.SantriRegistrationListQuery) ([]dto.SantriRegistrationResponse, *dto.Meta, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Limit == 0 {
		q.Limit = 20
	}

	result, err := uc.registrationRepo.List(ctx, regRepo.SantriRegistrationListQuery{
		AcademicPeriodID: q.AcademicPeriodID,
		SantriID:         q.SantriID,
		Status:           q.Status,
		Page:             q.Page,
		Limit:            q.Limit,
	})
	if err != nil {
		return nil, nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	periodIDs := make([]string, 0, len(result.Items))
	seen := map[string]struct{}{}
	for _, r := range result.Items {
		if _, ok := seen[r.AcademicPeriodID]; ok {
			continue
		}
		seen[r.AcademicPeriodID] = struct{}{}
		periodIDs = append(periodIDs, r.AcademicPeriodID)
	}
	periods, _ := uc.periodRepo.FindByIDs(ctx, periodIDs)
	periodMap := make(map[string]string, len(periods))
	for _, p := range periods {
		periodMap[p.ID] = p.Name
	}

	items := make([]dto.SantriRegistrationResponse, len(result.Items))
	for i, r := range result.Items {
		resp := command.MapSantriRegistrationToResponse(r)
		resp.PeriodName = periodMap[r.AcademicPeriodID]
		items[i] = *resp
	}
	return items, dto.NewMeta(q.Page, q.Limit, result.Total), nil
}
