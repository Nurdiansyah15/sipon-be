package application

import (
	"context"
	"log/slog"
	"time"

	"sipon-be/internal/modules/messaging/application/ports"
	outboxEntity "sipon-be/internal/modules/messaging/domain/event_outbox/entity"
	outboxRepo "sipon-be/internal/modules/messaging/domain/event_outbox/repository"
	"sipon-be/internal/modules/messaging/domain/message/valueobject"
	messagingpolicy "sipon-be/internal/modules/messaging/domain/message_job/policy"
)

// OutboxRelayOptions parameter relay outbox.
type OutboxRelayOptions struct {
	Interval  time.Duration
	Lease     time.Duration
	BatchSize int
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

// OutboxRelay mengambil event_outbox yang sudah committed lalu mempublish ke
// RabbitMQ secara reliable (publisher confirm). Row hanya ditandai PUBLISHED
// setelah broker mengonfirmasi; kegagalan di-retry dengan backoff.
type OutboxRelay struct {
	repo      outboxRepo.Repository
	publisher ports.Publisher
	opts      OutboxRelayOptions
	metrics   *Metrics
	logger    *slog.Logger
}

func NewOutboxRelay(
	repo outboxRepo.Repository,
	publisher ports.Publisher,
	opts OutboxRelayOptions,
	logger *slog.Logger,
) *OutboxRelay {
	if opts.Interval <= 0 {
		opts.Interval = 2 * time.Second
	}
	if opts.Lease <= 0 {
		opts.Lease = 30 * time.Second
	}
	if opts.BatchSize <= 0 {
		opts.BatchSize = 50
	}
	if opts.BaseDelay <= 0 {
		opts.BaseDelay = 30 * time.Second
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = 30 * time.Minute
	}
	return &OutboxRelay{repo: repo, publisher: publisher, opts: opts, logger: logger}
}

// WithMetrics memasang collector metrics (nil-safe).
func (r *OutboxRelay) WithMetrics(m *Metrics) *OutboxRelay {
	r.metrics = m
	return r
}

func (r *OutboxRelay) Start(ctx context.Context) {
	ticker := time.NewTicker(r.opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.processBatch(ctx)
		}
	}
}

func (r *OutboxRelay) processBatch(ctx context.Context) {
	now := time.Now()

	// Lease recovery: row PUBLISHING yang lease-nya sudah lewat dikembalikan ke
	// PENDING agar tidak hilang ketika worker crash di tengah publish.
	if n, err := r.repo.RecoverStuckPublishing(ctx, now.Add(-r.opts.Lease)); err != nil {
		r.logger.Warn("outbox: gagal recover stuck publishing", slog.Any("error", err))
	} else if n > 0 {
		if r.metrics != nil {
			r.metrics.OutboxRecovered.Add(n)
		}
		r.logger.Info("outbox: recover stuck publishing", slog.Int64("count", n))
	}

	entries, err := r.repo.ClaimDue(ctx, now, r.opts.BatchSize)
	if err != nil {
		r.logger.Error("outbox: gagal klaim due", slog.Any("error", err))
		return
	}
	for _, e := range entries {
		r.processEntry(ctx, e, now)
	}
}

func (r *OutboxRelay) processEntry(ctx context.Context, e *outboxEntity.OutboxEntry, now time.Time) {
	msg := valueobject.Message{
		ID:            e.ID,
		Type:          e.RoutingKey,
		Version:       e.Version,
		OccurredAt:    e.CreatedAt,
		Payload:       e.Payload,
		CorrelationID: e.CorrelationID,
		CausationID:   e.CausationID,
	}

	if err := r.publisher.Publish(ctx, msg); err != nil {
		if r.metrics != nil {
			r.metrics.OutboxPublishFail.Add(1)
		}
		nextAttempt := now.Add(messagingpolicy.CalculateRetryDelay(e.AttemptCount, r.opts.BaseDelay, r.opts.MaxDelay))
		if merr := r.repo.MarkFailed(ctx, e.ID, err.Error(), nextAttempt); merr != nil {
			r.logger.Error("outbox: gagal mark failed",
				slog.String("id", e.ID.String()), slog.Any("error", merr))
		}
		r.logger.Warn("outbox: publish gagal",
			slog.String("id", e.ID.String()),
			slog.String("routing_key", e.RoutingKey),
			slog.String("correlation_id", msg.CorrelationID),
			slog.Int("attempt", e.AttemptCount),
			slog.Any("error", err),
		)
		return
	}

	if r.metrics != nil {
		r.metrics.OutboxPublished.Add(1)
	}
	if err := r.repo.MarkPublished(ctx, e.ID, time.Now()); err != nil {
		r.logger.Error("outbox: gagal mark published",
			slog.String("id", e.ID.String()), slog.Any("error", err))
		return
	}

	r.logger.Info("outbox: published",
		slog.String("id", e.ID.String()),
		slog.String("routing_key", e.RoutingKey),
		slog.String("correlation_id", msg.CorrelationID),
		slog.Int("attempt", e.AttemptCount),
	)
}
