package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/shared/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/shared/scheduler/domain/scheduled_job/entity"
)

type stubSessionCompleter struct {
	err error
}

func (s stubSessionCompleter) Execute(ctx context.Context, id string) (*dto.ActivitySessionResponse, error) {
	return nil, s.err
}

type stubScheduledJobRepo struct {
	jobs    []*entity.ScheduledJob
	updates []*entity.ScheduledJob
}

func (s *stubScheduledJobRepo) Save(ctx context.Context, j *entity.ScheduledJob) error { return nil }

func (s *stubScheduledJobRepo) FindDueAndClaim(ctx context.Context, now time.Time, limit int, leaseUntil time.Time) ([]*entity.ScheduledJob, error) {
	return nil, nil
}

func (s *stubScheduledJobRepo) Update(ctx context.Context, j *entity.ScheduledJob) error {
	s.updates = append(s.updates, j)
	return nil
}

func (s *stubScheduledJobRepo) FindByTypeAndReferenceID(ctx context.Context, jobType, referenceID string) (*entity.ScheduledJob, error) {
	for _, j := range s.jobs {
		if j.Type == jobType && j.ReferenceID != nil && *j.ReferenceID == referenceID {
			return j, nil
		}
	}
	return nil, nil
}

const fingerprintSyncType = "akademik.fingerprint.sync"

func TestAutoCloseSessionUseCase_CompleteError(t *testing.T) {
	uc := NewAutoCloseSessionUseCase(
		stubSessionCompleter{err: errors.New("boom")},
		&stubScheduledJobRepo{},
		fingerprintSyncType,
	)
	if err := uc.Execute(context.Background(), "session-1"); err == nil {
		t.Fatal("harus mengembalikan error complete session")
	}
}

func TestAutoCloseSessionUseCase_PausesActiveSyncJob(t *testing.T) {
	ref := "session-1"
	syncJob := &entity.ScheduledJob{
		ID:          uuid.New(),
		Type:        fingerprintSyncType,
		ReferenceID: &ref,
		Status:      constant.StatusActive,
	}
	repo := &stubScheduledJobRepo{jobs: []*entity.ScheduledJob{syncJob}}

	uc := NewAutoCloseSessionUseCase(stubSessionCompleter{}, repo, fingerprintSyncType)
	if err := uc.Execute(context.Background(), ref); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(repo.updates) != 1 {
		t.Fatalf("harus 1 update, got %d", len(repo.updates))
	}
	if repo.updates[0].Status != constant.StatusPaused {
		t.Fatalf("status harus PAUSED, got %s", repo.updates[0].Status)
	}
}

func TestAutoCloseSessionUseCase_NoPauseWhenNotActive(t *testing.T) {
	ref := "session-1"
	syncJob := &entity.ScheduledJob{
		ID:          uuid.New(),
		Type:        fingerprintSyncType,
		ReferenceID: &ref,
		Status:      constant.StatusCompleted,
	}
	repo := &stubScheduledJobRepo{jobs: []*entity.ScheduledJob{syncJob}}

	uc := NewAutoCloseSessionUseCase(stubSessionCompleter{}, repo, fingerprintSyncType)
	if err := uc.Execute(context.Background(), ref); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(repo.updates) != 0 {
		t.Fatalf("completed job tidak boleh di-update, got %d update", len(repo.updates))
	}
}

func TestAutoCloseSessionUseCase_NoJobFound(t *testing.T) {
	repo := &stubScheduledJobRepo{}

	uc := NewAutoCloseSessionUseCase(stubSessionCompleter{}, repo, fingerprintSyncType)
	if err := uc.Execute(context.Background(), "session-x"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(repo.updates) != 0 {
		t.Fatalf("tidak boleh ada update, got %d", len(repo.updates))
	}
}
