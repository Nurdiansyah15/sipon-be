package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/shared/messaging"
	messagejobEntity "sipon-be/internal/shared/messaging/domain/message_job/entity"
	messagejobRepo "sipon-be/internal/shared/messaging/domain/message_job/repository"
)

type fakeMsgJobRepo struct {
	byID    map[uuid.UUID]*messagejobEntity.MessageJob
	updates []*messagejobEntity.MessageJob
}

func (f *fakeMsgJobRepo) Save(ctx context.Context, job *messagejobEntity.MessageJob) error {
	if _, ok := f.byID[job.ID]; !ok {
		f.byID[job.ID] = job
	}
	return nil
}
func (f *fakeMsgJobRepo) FindByID(ctx context.Context, id uuid.UUID) (*messagejobEntity.MessageJob, error) {
	return f.byID[id], nil
}
func (f *fakeMsgJobRepo) ClaimPending(ctx context.Context, now time.Time, limit int, leaseUntil time.Time) ([]*messagejobEntity.MessageJob, error) {
	return nil, nil
}
func (f *fakeMsgJobRepo) ClaimByID(ctx context.Context, id uuid.UUID, now time.Time, leaseUntil time.Time) (*messagejobEntity.MessageJob, bool, error) {
	job, ok := f.byID[id]
	if !ok || job.IsTerminal() || job.Status == messagejobEntity.StatusRunning {
		return nil, false, nil
	}
	job.StartRun(now, leaseUntil)
	return job, true, nil
}
func (f *fakeMsgJobRepo) Update(ctx context.Context, job *messagejobEntity.MessageJob) error {
	f.byID[job.ID] = job
	f.updates = append(f.updates, job)
	return nil
}
func (f *fakeMsgJobRepo) RecoverStuckRunning(ctx context.Context, leaseBefore time.Time) (int64, error) {
	return 0, nil
}

var _ messagejobRepo.Repository = (*fakeMsgJobRepo)(nil)

type fakeDelivery struct {
	body    []byte
	acked   bool
	nacked  bool
	requeue bool
}

func (f *fakeDelivery) Body() []byte { return f.body }
func (f *fakeDelivery) Ack() error   { f.acked = true; return nil }
func (f *fakeDelivery) Nack(requeue bool) error {
	f.nacked = true
	f.requeue = requeue
	return nil
}

type fakeQueuePublisher struct {
	queue string
	msgs  []messaging.Message
}

func (f *fakeQueuePublisher) PublishToQueue(ctx context.Context, queue string, msg messaging.Message) error {
	f.queue = queue
	f.msgs = append(f.msgs, msg)
	return nil
}

var _ messaging.QueuePublisher = (*fakeQueuePublisher)(nil)

const routingKey = "akademik.fingerprint.sync"

func newConsumer(t *testing.T, repo messagejobRepo.Repository, pub *fakeQueuePublisher, policy *messaging.RetryPolicy, handler messaging.HandlerFunc) *MessageConsumer {
	t.Helper()
	reg := messaging.NewRegistry()
	if err := reg.Register(routingKey, handler); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return NewMessageConsumer(
		nil,
		repo,
		reg,
		policy,
		pub,
		MessageConsumerOptions{
			Lease:       time.Minute,
			BaseDelay:   30 * time.Second,
			MaxDelay:    30 * time.Minute,
			RetryDelays: []time.Duration{time.Minute, 5 * time.Minute},
		},
		slog.New(slog.DiscardHandler),
	)
}

func makeDelivery(msg messaging.Message) *fakeDelivery {
	body, _ := json.Marshal(msg)
	return &fakeDelivery{body: body}
}

func TestConsumer_SuccessAcks(t *testing.T) {
	repo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}
	delivered := false
	c := newConsumer(t, repo, &fakeQueuePublisher{}, messaging.NewRetryPolicy(5), func(ctx context.Context, msg messaging.Message) error {
		delivered = true
		return nil
	})

	msg, _ := messaging.NewMessage(routingKey, json.RawMessage(`{}`))
	d := makeDelivery(msg)

	c.handle(context.Background(), d, "sipon.worker.scheduler")

	if !delivered {
		t.Fatal("handler tidak dipanggil")
	}
	if !d.acked || d.nacked {
		t.Fatalf("harus ack, acked=%v nacked=%v", d.acked, d.nacked)
	}
	job := repo.byID[msg.ID]
	if job.Status != messagejobEntity.StatusSucceeded {
		t.Fatalf("status harus SUCCEEDED, got %s", job.Status)
	}
}

func TestConsumer_FatalNacksToDLQ(t *testing.T) {
	repo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}
	c := newConsumer(t, repo, &fakeQueuePublisher{}, messaging.NewRetryPolicy(5), func(ctx context.Context, msg messaging.Message) error {
		return messaging.NewFatalError(errors.New("permanent"))
	})

	msg, _ := messaging.NewMessage(routingKey, json.RawMessage(`{}`))
	d := makeDelivery(msg)
	c.handle(context.Background(), d, "sipon.worker.scheduler")

	if !d.nacked || d.requeue {
		t.Fatalf("fatal harus nack tanpa requeue (ke DLQ), nacked=%v requeue=%v", d.nacked, d.requeue)
	}
	if repo.byID[msg.ID].Status != messagejobEntity.StatusFailed {
		t.Fatalf("status harus FAILED, got %s", repo.byID[msg.ID].Status)
	}
}

func TestConsumer_RetryableSchedulesRetry(t *testing.T) {
	repo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}
	pub := &fakeQueuePublisher{}
	c := newConsumer(t, repo, pub, messaging.NewRetryPolicy(5), func(ctx context.Context, msg messaging.Message) error {
		return messaging.NewRetryableError(errors.New("transient"))
	})

	msg, _ := messaging.NewMessage(routingKey, json.RawMessage(`{}`))
	d := makeDelivery(msg)
	c.handle(context.Background(), d, "sipon.worker.scheduler")

	if !d.acked {
		t.Fatal("retryable harus ack pesan asli")
	}
	job := repo.byID[msg.ID]
	if job.Status != messagejobEntity.StatusRetryWait {
		t.Fatalf("status harus RETRY_WAIT, got %s", job.Status)
	}
	if job.AttemptCount != 1 {
		t.Fatalf("attempt harus 1, got %d", job.AttemptCount)
	}
	if pub.queue != "sipon.worker.scheduler.retry.akademik.fingerprint.sync.60" {
		t.Fatalf("retry queue: got %q", pub.queue)
	}
}

func TestConsumer_RetryExhaustedNacksToDLQ(t *testing.T) {
	repo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}
	c := newConsumer(t, repo, &fakeQueuePublisher{}, messaging.NewRetryPolicy(1), func(ctx context.Context, msg messaging.Message) error {
		return messaging.NewRetryableError(errors.New("transient"))
	})

	msg, _ := messaging.NewMessage(routingKey, json.RawMessage(`{}`))
	d := makeDelivery(msg)
	c.handle(context.Background(), d, "sipon.worker.scheduler")

	if !d.nacked || d.requeue {
		t.Fatalf("retry habis harus nack ke DLQ, nacked=%v requeue=%v", d.nacked, d.requeue)
	}
	if repo.byID[msg.ID].Status != messagejobEntity.StatusFailed {
		t.Fatalf("status harus FAILED, got %s", repo.byID[msg.ID].Status)
	}
}

func TestConsumer_DuplicateSucceededNotRedelivered(t *testing.T) {
	repo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}
	msg, _ := messaging.NewMessage(routingKey, json.RawMessage(`{}`))
	// Simulasikan message yang sudah sukses.
	job := messagejobEntity.NewMessageJob(msg.ID, routingKey, msg.Payload, msg.Version, msg.CorrelationID, 5)
	job.Succeed(time.Now())
	repo.byID[msg.ID] = job

	delivered := false
	c := newConsumer(t, repo, &fakeQueuePublisher{}, messaging.NewRetryPolicy(5), func(ctx context.Context, m messaging.Message) error {
		delivered = true
		return nil
	})

	d := makeDelivery(msg)
	c.handle(context.Background(), d, "sipon.worker.scheduler")

	if delivered {
		t.Fatal("handler tidak boleh dipanggil untuk message yang sudah SUCCEEDED")
	}
	if !d.acked {
		t.Fatal("harus ack")
	}
}

func TestConsumer_InvalidEnvelopeNacksToDLQ(t *testing.T) {
	repo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}
	c := newConsumer(t, repo, &fakeQueuePublisher{}, messaging.NewRetryPolicy(5), func(ctx context.Context, msg messaging.Message) error {
		t.Fatal("handler tidak boleh dipanggil untuk envelope invalid")
		return nil
	})

	d := &fakeDelivery{body: []byte(`not json`)}
	c.handle(context.Background(), d, "sipon.worker.scheduler")

	if !d.nacked || d.requeue {
		t.Fatalf("envelope invalid harus nack ke DLQ, nacked=%v requeue=%v", d.nacked, d.requeue)
	}
}

// TestConsumer_PerRoutingKeyPolicy membuktikan retry policy per routing key: default
// 5, tetapi override untuk routingKey=1 memaksa message gagal pada attempt pertama.
func TestConsumer_PerRoutingKeyPolicy(t *testing.T) {
	repo := &fakeMsgJobRepo{byID: map[uuid.UUID]*messagejobEntity.MessageJob{}}
	policy := messaging.NewRetryPolicy(5)
	policy.Register(routingKey, 1)

	c := newConsumer(t, repo, &fakeQueuePublisher{}, policy, func(ctx context.Context, msg messaging.Message) error {
		return messaging.NewRetryableError(errors.New("transient"))
	})

	msg, _ := messaging.NewMessage(routingKey, json.RawMessage(`{}`))
	d := makeDelivery(msg)
	c.handle(context.Background(), d, "sipon.worker.scheduler")

	if !d.nacked || d.requeue {
		t.Fatalf("override max=1 harus nack ke DLQ, nacked=%v requeue=%v", d.nacked, d.requeue)
	}
	if repo.byID[msg.ID].Status != messagejobEntity.StatusFailed {
		t.Fatalf("status harus FAILED, got %s", repo.byID[msg.ID].Status)
	}
}
