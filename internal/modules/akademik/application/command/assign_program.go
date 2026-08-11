package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	appRepo "sipon-be/internal/modules/akademik/domain/activity_period_program/constant"
	"sipon-be/internal/modules/akademik/domain/activity_period_program/entity"
	appRepoIface "sipon-be/internal/modules/akademik/domain/activity_period_program/repository"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

type AssignProgramUseCase struct {
	programRepo        progRepo.ProgramRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityPeriodProg appRepoIface.ActivityPeriodProgramRepository
}

func NewAssignProgramUseCase(
	programRepo progRepo.ProgramRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityPeriodProg appRepoIface.ActivityPeriodProgramRepository,
) *AssignProgramUseCase {
	return &AssignProgramUseCase{programRepo: programRepo, activityPeriodRepo: activityPeriodRepo, activityPeriodProg: activityPeriodProg}
}

func (uc *AssignProgramUseCase) Execute(ctx context.Context, activityPeriodID string, req dto.AssignProgramRequest) (*dto.ActivityPeriodProgramResponse, error) {
	if _, err := uc.activityPeriodRepo.FindByID(ctx, activityPeriodID); err != nil {
		return nil, application.WrapRepoErr(err, appRepo.CodeActivityPeriodProgramInvalid)
	}
	program, err := uc.programRepo.FindByID(ctx, req.ProgramID)
	if err != nil {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	existing, err := uc.activityPeriodProg.FindByActivityPeriodAndProgram(ctx, activityPeriodID, req.ProgramID)
	if err != nil && !application.IsNotFoundErr(err, appRepo.CodeActivityPeriodProgramNotFound) {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if existing != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	app, err := entity.NewActivityPeriodProgram(uuid.NewString(), activityPeriodID, req.ProgramID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	if err := uc.activityPeriodProg.Save(ctx, app); err != nil {
		return nil, application.WrapConflictErr(err, appRepo.CodeActivityPeriodProgramDuplicate)
	}

	return &dto.ActivityPeriodProgramResponse{
		ID:               app.ID,
		ActivityPeriodID: app.ActivityPeriodID,
		ProgramID:        app.ProgramID,
		ProgramCode:      program.Code,
		ProgramName:      program.Name,
	}, nil
}
