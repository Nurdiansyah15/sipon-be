package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	appRepo "sipon-be/internal/modules/akademik/domain/activity_period_program/constant"
	appRepoIface "sipon-be/internal/modules/akademik/domain/activity_period_program/repository"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	"sipon-be/internal/shared/kernel"
)

type RemoveProgramUseCase struct {
	programRepo        progRepo.ProgramRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityPeriodProg appRepoIface.ActivityPeriodProgramRepository
}

func NewRemoveProgramUseCase(
	programRepo progRepo.ProgramRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityPeriodProg appRepoIface.ActivityPeriodProgramRepository,
) *RemoveProgramUseCase {
	return &RemoveProgramUseCase{programRepo: programRepo, activityPeriodRepo: activityPeriodRepo, activityPeriodProg: activityPeriodProg}
}

func (uc *RemoveProgramUseCase) Execute(ctx context.Context, activityPeriodID, programID string) error {
	if _, err := uc.activityPeriodRepo.FindByID(ctx, activityPeriodID); err != nil {
		return application.WrapRepoErr(err, appRepo.CodeActivityPeriodProgramNotFound)
	}
	if _, err := uc.programRepo.FindByID(ctx, programID); err != nil {
		return kernel.New(application.ErrCodeUnprocessableEntity)
	}
	app, err := uc.activityPeriodProg.FindByActivityPeriodAndProgram(ctx, activityPeriodID, programID)
	if err != nil {
		return application.WrapRepoErr(err, appRepo.CodeActivityPeriodProgramNotFound)
	}
	if err := uc.activityPeriodProg.Delete(ctx, app.ID); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	return nil
}
