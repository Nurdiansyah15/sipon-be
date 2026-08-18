package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/shared/messaging"
	outboxEntity "sipon-be/internal/shared/messaging/domain/event_outbox/entity"
	messagejobEntity "sipon-be/internal/shared/messaging/domain/message_job/entity"
	"sipon-be/internal/shared/messaging/infrastructure/rabbitmq"
)

// e2eOutboxRepo meniru event_outbox in-memory dengan claim-once, seolah-olah
// diisi oleh Scheduler Dispatcher.
type e2eOutboxRepo struct {
	mu        sync.Mutex
	pending   []*outboxEntity.OutboxEntry
	published []uuid.UUID
}

func (f *e2eOutboxRepo) Save(ctx context.Context, e *outboxEntity.OutboxEntry) error { return nil }
func (f *e2eOutboxRepo) ClaimDue(ctx context.Context, now time.Time, limit int) ([]*outboxEntity.OutboxEntry, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	es := f.pending
	f.pending = nil
	return es, nil
}
func (f *e2eOutboxRepo) MarkPublished(ctx context.Context, id uuid.UUID, publishedAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.published = append(f.published, id)
	return nil
}
func (f *e2eOutboxRepo) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, nextAttemptAt time.Time) error {
	return nil
}
func (f *e2eOutboxRepo) RecoverStuckPublishing(ctx context.Context, leaseBefore time.Time) (int64, error) {
	return 0, nil
}
func (f *e2eOutboxRepo) FindByID(ctx context.Context, id uuid.UUID) (*outboxEntity.OutboxEntry, error) {
	return nil, nil
}

// TestPilot_EndToEnd_Integration memvalidasi pilot akademik: event_outbox ->
// Outbox Relay -> RabbitMQ -> Message Consumer -> message_jobs -> handler.
// Menggunakan transport RabbitMQ nyata dan inbox/outbox in-memory.
func TestPilot_EndToEnd_Integration(t *testing.T) {
	dsn := os.Getenv("RABBITMQ_DSN")
	if dsn == "" {
		t.Skip("RABBITMQ_DSN tidak diset; skip integration test")
	}

	ns := time.Now().UnixNano()
	exchange := fmt.Sprintf("sipon.events.pilot.%d", ns)
	dlx := exchange + ".dlx"
	queue := fmt.Sprintf("sipon.worker.pilot.%d", ns)
	routing := "akademik.session.auto_close" // mewakili routing akademik

	// Topology
	topo, err := rabbitmq.NewTopology(rabbitmq.Options{
		DSN: dsn, Exchange: exchange, DLXExchange: dlx,
		RetryDelays: []time.Duration{time.Minute},
	})
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	if err := topo.Declare([]messaging.Binding{{Queue: queue, RoutingKey: routing}}); err != nil {
		t.Fatalf("Declare: %v", err)
	}
	_ = topo.Close()

	pub, err := rabbitmq.NewPublisher(dsn, exchange, 5*time.Second)
	if err != nil {
		t.Fatalf("NewPublisher: %v", err)
	}
	defer pub.Close()

	consumer, err := rabbitmq.NewConsumer(dsn, 1)
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Registry + handler akademik-like.
	reg := messaging.NewRegistry()
	handled := make(chan messaging.Message, 1)
	if err := reg.Register(routing, func(ctx context.Context, msg messaging.Message) error {
		handled <- msg
		return nil
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	outboxRepo := &e2eOutboxRepo{}
	msgJobRepo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}

	// Seed outbox (seolah-olah ditulis Scheduler Dispatcher dalam satu transaksi).
	entry := outboxEntity.NewOutboxEntry(routing, json.RawMessage(`{"session_id":"s1"}`), "corr")
	outboxRepo.pending = []*outboxEntity.OutboxEntry{entry}

	// Outbox Relay
	relay := NewOutboxRelay(outboxRepo, pub, OutboxRelayOptions{
		Interval: 50 * time.Millisecond,
		Lease:    time.Second,
	}, slog.New(slog.DiscardHandler))
	go relay.Start(ctx)

	// Message Consumer
	msgConsumer := NewMessageConsumer(
		consumer, msgJobRepo, reg, messaging.NewRetryPolicy(5), pub,
		MessageConsumerOptions{Lease: time.Minute, RetryDelays: []time.Duration{time.Minute}},
		slog.New(slog.DiscardHandler),
	)
	go func() {
		_ = msgConsumer.Start(ctx, queue)
	}()

	// 1. Happy path: handler terpanggil dengan message ID = outbox entry ID.
	select {
	case got := <-handled:
		if got.ID != entry.ID {
			t.Fatalf("message id mismatch: got %s want %s", got.ID, entry.ID)
		}
		if got.Type != routing {
			t.Fatalf("routing key: got %q", got.Type)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout: handler tidak terpanggil end-to-end")
	}

	// 2. Verifikasi side effects (relay mark published + inbox SUCCEEDED) dengan
	//    polling karena berjalan di goroutine terpisah.
	waitFor(t, 5*time.Second, func() bool {
		outboxRepo.mu.Lock()
		published := len(outboxRepo.published)
		outboxRepo.mu.Unlock()
		job := msgJobRepo.byID[entry.ID]
		return published == 1 && job != nil && job.Status == messagejobEntity.StatusSucceeded
	})

	// 3. Idempotency: duplicate delivery (ID sama) tidak memicu handler lagi.
	dup := messaging.Message{
		ID: entry.ID, Type: routing, Version: 1,
		OccurredAt: time.Now(), Payload: entry.Payload, CorrelationID: "corr",
	}
	if err := pub.Publish(ctx, dup); err != nil {
		t.Fatalf("Publish duplicate: %v", err)
	}

	select {
	case <-handled:
		t.Fatal("duplicate delivery tidak boleh memicu handler lagi")
	case <-time.After(2 * time.Second):
		// expected
	}
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("kondisi tidak terpenuhi dalam batas waktu")
}
