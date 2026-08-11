package entity

import (
	"sipon-be/internal/modules/akademik/domain/activity_period_program/constant"
	"sipon-be/internal/shared/kernel"
)

type ActivityPeriodProgram struct {
	ID               string
	ActivityPeriodID string
	ProgramID        string
}

func NewActivityPeriodProgram(id, activityPeriodID, programID string) (*ActivityPeriodProgram, error) {
	if id == "" || activityPeriodID == "" || programID == "" {
		return nil, kernel.New(constant.CodeActivityPeriodProgramInvalid)
	}
	return &ActivityPeriodProgram{
		ID:               id,
		ActivityPeriodID: activityPeriodID,
		ProgramID:        programID,
	}, nil
}
