package resolver

import (
	"context"

	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/shared/kernel"
)

// SessionPeriodResolver walks a session's schedule up to its activity period
// to find the academic period the session belongs to.
type SessionPeriodResolver struct {
	sessionRepo  sesRepo.ActivitySessionRepository
	scheduleRepo schRepo.ActivityScheduleRepository
	apRepo       apRepo.ActivityPeriodRepository
}

func NewSessionPeriodResolver(
	sessionRepo sesRepo.ActivitySessionRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
	apRepo apRepo.ActivityPeriodRepository,
) *SessionPeriodResolver {
	return &SessionPeriodResolver{sessionRepo: sessionRepo, scheduleRepo: scheduleRepo, apRepo: apRepo}
}

// Resolve returns the academic period ID for the given session, or the wrapped
// underlying error when any link in the chain is missing.
func (r *SessionPeriodResolver) Resolve(ctx context.Context, sessionID string) (string, error) {
	session, err := r.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return "", err
	}
	schedule, err := r.scheduleRepo.FindByID(ctx, session.ActivityScheduleID)
	if err != nil {
		return "", err
	}
	period, err := r.apRepo.FindByID(ctx, schedule.ActivityPeriodID)
	if err != nil {
		return "", err
	}
	if period.AcademicPeriodID == "" {
		return "", kernel.New(application.ErrCodeUnprocessableEntity)
	}
	return period.AcademicPeriodID, nil
}
