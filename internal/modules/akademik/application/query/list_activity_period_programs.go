package query

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	appRepo "sipon-be/internal/modules/akademik/domain/activity_period_program/repository"
	progEntity "sipon-be/internal/modules/akademik/domain/program/entity"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

type ListActivityPeriodProgramsUseCase struct {
	programRepo        progRepo.ProgramRepository
	activityPeriodProg appRepo.ActivityPeriodProgramRepository
}

func NewListActivityPeriodProgramsUseCase(
	programRepo progRepo.ProgramRepository,
	activityPeriodProg appRepo.ActivityPeriodProgramRepository,
) *ListActivityPeriodProgramsUseCase {
	return &ListActivityPeriodProgramsUseCase{programRepo: programRepo, activityPeriodProg: activityPeriodProg}
}

func (uc *ListActivityPeriodProgramsUseCase) Execute(ctx context.Context, activityPeriodID string) ([]dto.ActivityPeriodProgramResponse, error) {
	items, err := uc.activityPeriodProg.ListByActivityPeriod(ctx, activityPeriodID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	programIDs := make([]string, 0, len(items))
	for _, p := range items {
		programIDs = append(programIDs, p.ProgramID)
	}
	programs, _ := uc.programRepo.FindByIDs(ctx, programIDs)
	programMap := make(map[string]*progEntity.Program, len(programs))
	for _, p := range programs {
		programMap[p.ID] = p
	}

	responses := make([]dto.ActivityPeriodProgramResponse, 0, len(items))
	for _, p := range items {
		resp := dto.ActivityPeriodProgramResponse{
			ID:               p.ID,
			ActivityPeriodID: p.ActivityPeriodID,
			ProgramID:        p.ProgramID,
		}
		if prog, ok := programMap[p.ProgramID]; ok {
			resp.ProgramCode = prog.Code
			resp.ProgramName = prog.Name
		}
		responses = append(responses, resp)
	}
	return responses, nil
}
