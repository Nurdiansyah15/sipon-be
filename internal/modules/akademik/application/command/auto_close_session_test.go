package command

import (
	"context"
	"errors"
	"testing"

	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
)

type stubSessionCompleter struct {
	err error
}

func (s stubSessionCompleter) Execute(ctx context.Context, id string) (*dto.ActivitySessionResponse, error) {
	return nil, s.err
}

type stubScheduler struct {
	paused  []string
	pauseFn func(ctx context.Context, jobType, referenceID string) error
}

func (s *stubScheduler) ScheduleRecurring(ctx context.Context, in ports.ScheduleRecurringInput) error {
	return nil
}

func (s *stubScheduler) ScheduleOneOff(ctx context.Context, in ports.ScheduleOneOffInput) error {
	return nil
}

func (s *stubScheduler) PauseByTypeAndReferenceID(ctx context.Context, jobType, referenceID string) error {
	s.paused = append(s.paused, referenceID)
	if s.pauseFn != nil {
		return s.pauseFn(ctx, jobType, referenceID)
	}
	return nil
}

const fingerprintSyncType = "akademik.fingerprint.sync"

func TestAutoCloseSessionUseCase_CompleteError(t *testing.T) {
	uc := NewAutoCloseSessionUseCase(
		stubSessionCompleter{err: errors.New("boom")},
		&stubScheduler{},
		fingerprintSyncType,
	)
	if err := uc.Execute(context.Background(), "session-1"); err == nil {
		t.Fatal("harus mengembalikan error complete session")
	}
}

func TestAutoCloseSessionUseCase_PausesActiveSyncJob(t *testing.T) {
	ref := "session-1"
	sc := &stubScheduler{}

	uc := NewAutoCloseSessionUseCase(stubSessionCompleter{}, sc, fingerprintSyncType)
	if err := uc.Execute(context.Background(), ref); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(sc.paused) != 1 || sc.paused[0] != ref {
		t.Fatalf("harus pause job milik %q, got %v", ref, sc.paused)
	}
}
