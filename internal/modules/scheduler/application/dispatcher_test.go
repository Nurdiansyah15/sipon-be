package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"sipon-be/internal/shared/database"
	outboxEntity "sipon-be/internal/modules/messaging/domain/event_outbox/entity"
	outboxPersistence "sipon-be/internal/modules/messaging/infrastructure/persistence"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/entity"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/repository"
	schedulerPersistence "sipon-be/internal/modules/scheduler/infrastructure/persistence"
)

type fakeJobRepo struct {
	updates []*entity.ScheduledJob
}

func (f *fakeJobRepo) Save(ctx context.Context, job *entity.ScheduledJob) error { return nil }
func (f *fakeJobRepo) FindDueAndClaim(ctx context.Context, now time.Time, limit int, leaseUntil time.Time) ([]*entity.ScheduledJob, error) {
	return nil, nil
}
func (f *fakeJobRepo) Update(ctx context.Context, job *entity.ScheduledJob) error {
	f.updates = append(f.updates, job)
	return nil
}
func (f *fakeJobRepo) FindByTypeAndReferenceID(ctx context.Context, jobType, referenceID string) (*entity.ScheduledJob, error) {
	return nil, nil
}

func newDispatcher(repo repository.Repository) *SchedulerDispatcher {
	return NewDispatcher(repo, time.Second, time.Minute, slog.New(slog.DiscardHandler))
}

func newOneOffJob() *entity.ScheduledJob {
	return &entity.ScheduledJob{
		ID:           uuid.New(),
		Type:         "akademik.session.auto_close",
		Payload:      json.RawMessage(`{"session_id":"s1"}`),
		ScheduleType: constant.ScheduleTypeOneOff,
		NextRunAt:    time.Now(),
		Status:       constant.StatusProcessing,
		MaxRetry:     3,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func newRecurringJob() *entity.ScheduledJob {
	expr := "*/1 * * * *"
	return &entity.ScheduledJob{
		ID:           uuid.New(),
		Type:         "akademik.fingerprint.sync",
		Payload:      json.RawMessage(`{"session_id":"s1"}`),
		ScheduleType: constant.ScheduleTypeRecurring,
		CronExpr:     &expr,
		NextRunAt:    time.Now(),
		Status:       constant.StatusProcessing,
		MaxRetry:     3,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestDispatcher_DirectMode_OneOffCompletes(t *testing.T) {
	repo := &fakeJobRepo{}
	called := false
	dispatcher := newDispatcher(repo).WithDirectMode(func(ctx context.Context, jobType string, payload json.RawMessage) error {
		called = true
		if jobType != "akademik.session.auto_close" {
			t.Fatalf("jobType: got %q", jobType)
		}
		return nil
	})

	job := newOneOffJob()
	if err := dispatcher.dispatchJob(context.Background(), job, time.Now()); err != nil {
		t.Fatalf("dispatchJob: %v", err)
	}
	if !called {
		t.Fatal("handler tidak dipanggil")
	}
	if len(repo.updates) != 1 {
		t.Fatalf("harus 1 update, got %d", len(repo.updates))
	}
	if repo.updates[0].Status != constant.StatusCompleted {
		t.Fatalf("status harus COMPLETED, got %s", repo.updates[0].Status)
	}
	if repo.updates[0].LeaseUntil != nil {
		t.Fatal("lease harus di-reset setelah selesai")
	}
}

func TestDispatcher_DirectMode_RecurringAdvances(t *testing.T) {
	repo := &fakeJobRepo{}
	dispatcher := newDispatcher(repo).WithDirectMode(func(ctx context.Context, jobType string, payload json.RawMessage) error {
		return nil
	})

	job := newRecurringJob()
	now := time.Now()
	if err := dispatcher.dispatchJob(context.Background(), job, now); err != nil {
		t.Fatalf("dispatchJob: %v", err)
	}
	if repo.updates[0].Status != constant.StatusActive {
		t.Fatalf("status harus ACTIVE, got %s", repo.updates[0].Status)
	}
	if !repo.updates[0].NextRunAt.After(now) {
		t.Fatal("next_run_at harus maju untuk recurring job")
	}
}

func TestDispatcher_DirectMode_FatalMarksFailed(t *testing.T) {
	repo := &fakeJobRepo{}
	dispatcher := newDispatcher(repo).WithDirectMode(func(ctx context.Context, jobType string, payload json.RawMessage) error {
		return &FatalError{Err: errors.New("permanent")}
	})

	if err := dispatcher.dispatchJob(context.Background(), newOneOffJob(), time.Now()); err != nil {
		t.Fatalf("dispatchJob: %v", err)
	}
	if repo.updates[0].Status != constant.StatusFailed {
		t.Fatalf("status harus FAILED, got %s", repo.updates[0].Status)
	}
}

func TestDispatcher_DirectMode_RetryableIncrementsRetry(t *testing.T) {
	repo := &fakeJobRepo{}
	dispatcher := newDispatcher(repo).WithDirectMode(func(ctx context.Context, jobType string, payload json.RawMessage) error {
		return &RetryableError{Err: errors.New("transient")}
	})

	if err := dispatcher.dispatchJob(context.Background(), newOneOffJob(), time.Now()); err != nil {
		t.Fatalf("dispatchJob: %v", err)
	}
	if repo.updates[0].RetryCount != 1 {
		t.Fatalf("retry_count harus 1, got %d", repo.updates[0].RetryCount)
	}
	if repo.updates[0].Status != constant.StatusActive {
		t.Fatalf("status harus ACTIVE (utk retry), got %s", repo.updates[0].Status)
	}
}

// TestDispatcher_OutboxMode_OneOffWritesOutboxAndCompletes membuktikan Scheduler
// outboxWriterAdapter mengadaptasi messaging outbox repository ke port
// scheduler.OutboxWriter (dependency inversion untuk keperluan test).
type outboxWriterAdapter struct {
	repo *outboxPersistence.PostgresOutboxRepository
}

func (a *outboxWriterAdapter) Save(ctx context.Context, routingKey string, payload json.RawMessage) error {
	return a.repo.Save(ctx, outboxEntity.NewOutboxEntry(routingKey, payload, ""))
}

// Dispatcher pada mode outbox menulis event_outbox DAN memajukan state scheduled
// job dalam satu DB transaction, tanpa memanggil handler.
func TestDispatcher_OutboxMode_OneOffWritesOutboxAndCompletes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	schedRepo := schedulerPersistence.NewPostgresScheduledJobRepository(db)
	outboxRepo := &outboxWriterAdapter{repo: outboxPersistence.NewPostgresOutboxRepository(db)}
	transactor := database.NewTransactor(db)

	dispatcher := NewDispatcher(schedRepo, time.Second, time.Minute, slog.New(slog.DiscardHandler)).
		WithOutboxMode(outboxRepo, transactor)

	job := newOneOffJob()

	mock.ExpectBegin()
	mock.ExpectExec(`(?s)INSERT INTO event_outbox`).
		WithArgs(sqlmock.AnyArg(), job.Type, sqlmock.AnyArg(), 1, sqlmock.AnyArg(), nil, "PENDING", 0, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`(?s)UPDATE scheduled_jobs`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := dispatcher.dispatchJob(context.Background(), job, time.Now()); err != nil {
		t.Fatalf("dispatchJob (outbox): %v", err)
	}
	if job.Status != constant.StatusCompleted {
		t.Fatalf("job harus COMPLETED setelah outbox ditulis, got %s", job.Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
