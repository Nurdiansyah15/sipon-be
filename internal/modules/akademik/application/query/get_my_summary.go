package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/akademik/application/resolver"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	regConst "sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
	"sipon-be/internal/shared/timeutil"
)

type GetMySummaryUseCase struct {
	kesantrianReader  ports.KesantrianReader
	periodRepo        periodRepo.AcademicPeriodRepository
	registrationRepo  regRepo.SantriRegistrationRepository
	santriProgramRepo spRepo.SantriProgramRepository
	programRepo       progRepo.ProgramRepository
}

func NewGetMySummaryUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	registrationRepo regRepo.SantriRegistrationRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	programRepo progRepo.ProgramRepository,
) *GetMySummaryUseCase {
	return &GetMySummaryUseCase{
		kesantrianReader:  kesantrianReader,
		periodRepo:        periodRepo,
		registrationRepo:  registrationRepo,
		santriProgramRepo: santriProgramRepo,
		programRepo:       programRepo,
	}
}

func (uc *GetMySummaryUseCase) Execute(ctx context.Context, userID string) (*dto.MySummaryResponse, error) {
	info, err := resolver.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	resp := &dto.MySummaryResponse{Herregistrasi: &dto.HerregistrasiStatus{Status: "none"}}

	if sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, info.SantriID); err == nil {
		if prog, err := uc.programRepo.FindByID(ctx, sp.ProgramID); err == nil {
			started := timeutil.ToPlatform(sp.StartedAt)
			resp.Program = &dto.ProgramInfo{
				ID:        prog.ID,
				Code:      prog.Code,
				Name:      prog.Name,
				Status:    string(prog.Status),
				StartedAt: &started,
			}
		}
	}

	period, err := uc.periodRepo.FindOpen(ctx)
	if err != nil {
		if application.IsNotFoundErr(err, resolver.PeriodNotFoundCode) {
			return resp, nil
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	resp.AcademicPeriod = command.MapAcademicPeriodToResponse(period)

	reg, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, info.SantriID, period.ID)
	if err != nil && !application.IsNotFoundErr(err, regConst.CodeSantriRegistrationNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if reg != nil {
		regID := reg.ID
		resp.Herregistrasi = &dto.HerregistrasiStatus{
			Status:         string(reg.Status),
			RegistrationID: &regID,
			RegisteredAt:   timeutil.ToPlatformPtr(reg.RegisteredAt),
			RevisionNotes:  reg.RevisionNotes,
		}
	}
	return resp, nil
}
