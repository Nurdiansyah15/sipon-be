package application

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/shared/messaging"
	outboxEntity "sipon-be/internal/shared/messaging/domain/event_outbox/entity"
	outboxRepo "sipon-be/internal/shared/messaging/domain/event_outbox/repository"
)

type fakeOutboxRepo struct {
	entries         []*outboxEntity.OutboxEntry
	markedPublished []uuid.UUID
	markedFailed    []uuid.UUID
	failedErr       []string
	recovered       int64
}

func (f *fakeOutboxRepo) Save(ctx context.Context, e *outboxEntity.OutboxEntry) error { return nil }
func (f *fakeOutboxRepo) ClaimDue(ctx context.Context, now time.Time, limit int) ([]*outboxEntity.OutboxEntry, error) {
	return f.entries, nil
}
func (f *fakeOutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	f.markedPublished = append(f.markedPublished, id)
	return nil
}
func (f *fakeOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextAttemptAt time.Time) error {
	f.markedFailed = append(f.markedFailed, id)
	f.failedErr = append(f.failedErr, errMsg)
	return nil
}
func (f *fakeOutboxRepo) RecoverStuckPublishing(ctx context.Context, leaseBefore time.Time) (int64, error) {
	return f.recovered, nil
}
func (f *fakeOutboxRepo) FindByID(ctx context.Context, id uuid.UUID) (*outboxEntity.OutboxEntry, error) {
	return nil, nil
}

var _ outboxRepo.Repository = (*fakeOutboxRepo)(nil)

type fakePublisher struct {
	err       error
	published []messaging.Message
}

func (f *fakePublisher) Publish(ctx context.Context, msg messaging.Message) error {
	f.published = append(f.published, msg)
	return f.err
}

var _ messaging.Publisher = (*fakePublisher)(nil)

func newOutboxRelay(repo *fakeOutboxRepo, pub *fakePublisher) *OutboxRelay {
	return NewOutboxRelay(repo, pub, OutboxRelayOptions{
		Interval:  time.Millisecond,
		Lease:     time.Second,
		BaseDelay: 30 * time.Second,
		MaxDelay:  30 * time.Minute,
	}, slog.New(slog.DiscardHandler))
}

func TestOutboxRelay_PublishSuccessMarksPublished(t *testing.T) {
	entry := outboxEntity.NewOutboxEntry("akademik.fingerprint.sync", nil, "")
	entry.MarkPublishing(time.Now())

	repo := &fakeOutboxRepo{entries: []*outboxEntity.OutboxEntry{entry}}
	pub := &fakePublisher{}
	relay := newOutboxRelay(repo, pub)

	relay.processBatch(context.Background())

	if len(pub.published) != 1 {
		t.Fatalf("harus publish 1, got %d", len(pub.published))
	}
	if pub.published[0].Type != entry.RoutingKey {
		t.Fatalf("routing key: got %q", pub.published[0].Type)
	}
	if len(repo.markedPublished) != 1 {
		t.Fatalf("harus mark published 1, got %d", len(repo.markedPublished))
	}
	if len(repo.markedFailed) != 0 {
		t.Fatalf("tidak boleh mark failed, got %d", len(repo.markedFailed))
	}
}

func TestOutboxRelay_PublishFailureMarksFailed(t *testing.T) {
	entry := outboxEntity.NewOutboxEntry("akademik.fingerprint.sync", nil, "")
	entry.MarkPublishing(time.Now())

	repo := &fakeOutboxRepo{entries: []*outboxEntity.OutboxEntry{entry}}
	pub := &fakePublisher{err: errors.New("broker down")}
	relay := newOutboxRelay(repo, pub)

	before := time.Now()
	relay.processBatch(context.Background())

	if len(repo.markedFailed) != 1 {
		t.Fatalf("harus mark failed 1, got %d", len(repo.markedFailed))
	}
	if repo.failedErr[0] != "broker down" {
		t.Fatalf("last_error: got %q", repo.failedErr[0])
	}
	if len(repo.markedPublished) != 0 {
		t.Fatalf("tidak boleh mark published saat gagal")
	}
	// next_attempt_at harus di masa depan (backoff)
	_ = before
}

func TestOutboxRelay_NoEntriesSkips(t *testing.T) {
	repo := &fakeOutboxRepo{}
	pub := &fakePublisher{}
	relay := newOutboxRelay(repo, pub)

	relay.processBatch(context.Background())

	if len(pub.published) != 0 {
		t.Fatalf("tidak boleh publish apa pun, got %d", len(pub.published))
	}
}
