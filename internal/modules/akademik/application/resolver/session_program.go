package resolver

import (
	"context"

	appRepo "sipon-be/internal/modules/akademik/domain/activity_period_program/repository"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
)

// SessionProgramResolver walks a session's schedule up to its activity period
// to resolve the program(s) the session applies to.
type SessionProgramResolver struct {
	sessionRepo   sesRepo.ActivitySessionRepository
	scheduleRepo  schRepo.ActivityScheduleRepository
	apProgramRepo appRepo.ActivityPeriodProgramRepository
	programRepo   progRepo.ProgramRepository
}

func NewSessionProgramResolver(
	sessionRepo sesRepo.ActivitySessionRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
	apProgramRepo appRepo.ActivityPeriodProgramRepository,
	programRepo progRepo.ProgramRepository,
) *SessionProgramResolver {
	return &SessionProgramResolver{
		sessionRepo:   sessionRepo,
		scheduleRepo:  scheduleRepo,
		apProgramRepo: apProgramRepo,
		programRepo:   programRepo,
	}
}

// Resolve returns the program IDs the given session applies to. When the
// activity period has no program scope (no activity_period_programs rows), the
// session applies to all active programs.
func (r *SessionProgramResolver) Resolve(ctx context.Context, sessionID string) ([]string, error) {
	session, err := r.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	schedule, err := r.scheduleRepo.FindByID(ctx, session.ActivityScheduleID)
	if err != nil {
		return nil, err
	}
	links, err := r.apProgramRepo.ListByActivityPeriod(ctx, schedule.ActivityPeriodID)
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return r.programRepo.ListActiveIDs(ctx)
	}
	programIDs := make([]string, 0, len(links))
	for _, l := range links {
		programIDs = append(programIDs, l.ProgramID)
	}
	return programIDs, nil
}
