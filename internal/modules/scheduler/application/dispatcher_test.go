package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	outboxEntity "sipon-be/internal/modules/messaging/domain/event_outbox/entity"
	outboxPersistence "sipon-be/internal/modules/messaging/infrastructure/persistence"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/constant"
	"sipon-be/internal/modules/scheduler/domain/scheduled_job/entity"
	schedulerPersistence "sipon-be/internal/modules/scheduler/infrastructure/persistence"
	"sipon-be/internal/shared/database"
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

// outboxWriterAdapter mengadaptasi messaging outbox repository ke port
// scheduler.OutboxWriter (dependency inversion untuk keperluan test).
type outboxWriterAdapter struct {
	repo *outboxPersistence.PostgresOutboxRepository
}

func (a *outboxWriterAdapter) Save(ctx context.Context, routingKey string, payload json.RawMessage) error {
	return a.repo.Save(ctx, outboxEntity.NewOutboxEntry(routingKey, payload, ""))
}

// Dispatcher selalu memakai pola outbox: menulis event_outbox DAN memajukan
// state scheduled job dalam satu DB transaction, tanpa memanggil handler.
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
		WithOutbox(outboxRepo, transactor)

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

func TestDispatcher_WithoutOutbox_ReturnsError(t *testing.T) {
	repo := &fakeJobRepo{}
	dispatcher := NewDispatcher(repo, time.Second, time.Minute, slog.New(slog.DiscardHandler))

	if err := dispatcher.dispatchJob(context.Background(), newOneOffJob(), time.Now()); err == nil {
		t.Fatal("dispatchJob tanpa outbox harus error")
	}
}
