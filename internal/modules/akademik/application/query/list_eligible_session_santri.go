package query

import (
	"context"
	"log/slog"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

// ListEligibleSessionSantriUseCase returns santri who may be recorded for
// attendance on a session — i.e. active santri who have completed herregistrasi
// for the session's academic period (regardless of program assignment).
type ListEligibleSessionSantriUseCase struct {
	registrationRepo regRepo.SantriRegistrationRepository
	periodResolver   *application.SessionPeriodResolver
	kesantrianReader ports.KesantrianReader
}

func NewListEligibleSessionSantriUseCase(
	registrationRepo regRepo.SantriRegistrationRepository,
	periodResolver *application.SessionPeriodResolver,
	kesantrianReader ports.KesantrianReader,
) *ListEligibleSessionSantriUseCase {
	return &ListEligibleSessionSantriUseCase{
		registrationRepo: registrationRepo,
		periodResolver:   periodResolver,
		kesantrianReader: kesantrianReader,
	}
}

func (uc *ListEligibleSessionSantriUseCase) Execute(ctx context.Context, sessionID string) ([]dto.EligibleSantriResponse, error) {
	academicPeriodID, err := uc.periodResolver.Resolve(ctx, sessionID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	registrations, err := uc.registrationRepo.ListCompletedByAcademicPeriod(ctx, academicPeriodID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	items := make([]dto.EligibleSantriResponse, 0, len(registrations))
	for _, reg := range registrations {
		info, err := uc.kesantrianReader.GetSantriByID(ctx, reg.SantriID)
		if err != nil || info == nil || info.Status != "SANTRI" {
			slog.Warn("akademik: eligible santri enrichment failed", "santri_id", reg.SantriID, "error", err)
			continue
		}
		items = append(items, dto.EligibleSantriResponse{
			SantriID: reg.SantriID,
			NIS:      info.NIS,
			Fullname: info.Fullname,
		})
	}
	return items, nil
}
