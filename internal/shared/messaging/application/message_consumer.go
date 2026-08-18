package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"sipon-be/internal/shared/messaging"
	messagejobEntity "sipon-be/internal/shared/messaging/domain/message_job/entity"
	messagejobRepo "sipon-be/internal/shared/messaging/domain/message_job/repository"
)

// MessageConsumerOptions parameter consumer.
type MessageConsumerOptions struct {
	Lease       time.Duration
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	RetryDelays []time.Duration
}

// MessageConsumer mengonsumsi message dari queue, menulis durable inbox
// (message_jobs) SEBELUM handler menjalankan side effect, lalu dispatch ke
// registry module. Alur: RabbitMQ -> message_jobs -> Module Handler.
type MessageConsumer struct {
	consumer       messaging.Consumer
	repo           messagejobRepo.Repository
	registry       *messaging.Registry
	policy         *messaging.RetryPolicy
	retryPublisher messaging.QueuePublisher
	opts           MessageConsumerOptions
	metrics        *Metrics
	logger         *slog.Logger
}

func NewMessageConsumer(
	consumer messaging.Consumer,
	repo messagejobRepo.Repository,
	registry *messaging.Registry,
	policy *messaging.RetryPolicy,
	retryPublisher messaging.QueuePublisher,
	opts MessageConsumerOptions,
	logger *slog.Logger,
) *MessageConsumer {
	if opts.Lease <= 0 {
		opts.Lease = 5 * time.Minute
	}
	if opts.BaseDelay <= 0 {
		opts.BaseDelay = 30 * time.Second
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = 30 * time.Minute
	}
	if policy == nil {
		policy = messaging.NewRetryPolicy(5)
	}
	return &MessageConsumer{
		consumer:       consumer,
		repo:           repo,
		registry:       registry,
		policy:         policy,
		retryPublisher: retryPublisher,
		opts:           opts,
		logger:         logger,
	}
}

// WithMetrics memasang collector metrics (nil-safe).
func (c *MessageConsumer) WithMetrics(m *Metrics) *MessageConsumer {
	c.metrics = m
	return c
}

// Start mengonsumsi queue hingga context selesai.
func (c *MessageConsumer) Start(ctx context.Context, queue string) error {
	return c.consumer.Consume(ctx, queue, func(ctx context.Context, d messaging.Delivery) error {
		c.handle(ctx, d, queue)
		return nil
	})
}

func (c *MessageConsumer) handle(ctx context.Context, d messaging.Delivery, queue string) {
	var msg messaging.Message
	if err := json.Unmarshal(d.Body(), &msg); err != nil {
		c.logger.Warn("consumer: payload tidak valid, kirim ke DLQ", slog.Any("error", err))
		_ = d.Nack(false)
		return
	}
	if err := msg.Validate(); err != nil {
		c.logger.Warn("consumer: envelope tidak valid, kirim ke DLQ", slog.Any("error", err))
		_ = d.Nack(false)
		return
	}

	// 1. Durable inbox: buat/pertahankan row PENDING (ON CONFLICT DO NOTHING) dan
	//    commit sebelum handler dipanggil.
	job := messagejobEntity.NewMessageJob(
		msg.ID, msg.Type, msg.Payload, msg.Version, msg.CorrelationID,
		c.policy.MaxRetryFor(msg.Type),
	)
	if err := c.repo.Save(ctx, job); err != nil {
		c.logger.Error("consumer: gagal save durable inbox, requeue", slog.String("id", msg.ID.String()), slog.Any("error", err))
		_ = d.Nack(true)
		return
	}

	existing, err := c.repo.FindByID(ctx, msg.ID)
	if err != nil {
		c.logger.Error("consumer: gagal find inbox, requeue", slog.String("id", msg.ID.String()), slog.Any("error", err))
		_ = d.Nack(true)
		return
	}

	// 2. Idempotency: message yang sudah terminal tidak dijalankan ulang.
	if existing.IsTerminal() {
		if c.metrics != nil {
			c.metrics.Duplicate.Add(1)
		}
		c.logger.Info("consumer: message sudah terminal, skip",
			slog.String("id", msg.ID.String()),
			slog.String("routing_key", msg.Type),
			slog.String("status", string(existing.Status)),
		)
		_ = d.Ack()
		return
	}

	// 3. Claim RUNNING + lease. Bila gagal (diproses worker lain), ack saja.
	now := time.Now()
	start := now
	claimed, ok, err := c.repo.ClaimByID(ctx, msg.ID, now, now.Add(c.opts.Lease))
	if err != nil {
		c.logger.Error("consumer: gagal claim inbox, requeue", slog.String("id", msg.ID.String()), slog.Any("error", err))
		_ = d.Nack(true)
		return
	}
	if !ok {
		_ = d.Ack()
		return
	}

	// 4. Dispatch ke handler module.
	dispatchErr := c.registry.Dispatch(ctx, msg)
	now = time.Now()
	duration := now.Sub(start)

	logBase := []slog.Attr{
		slog.String("id", msg.ID.String()),
		slog.String("routing_key", msg.Type),
		slog.String("correlation_id", msg.CorrelationID),
		slog.Int("attempt", claimed.AttemptCount),
		slog.Duration("duration", duration),
	}

	switch {
	case dispatchErr == nil:
		claimed.Succeed(now)
		if err := c.repo.Update(ctx, claimed); err != nil {
			c.logger.Error("consumer: gagal mark succeeded", slog.String("id", msg.ID.String()), slog.Any("error", err))
		}
		if c.metrics != nil {
			c.metrics.Handled.Add(1)
			c.metrics.Succeeded.Add(1)
		}
		c.logger.LogAttrs(context.Background(), slog.LevelInfo, "consumer: message sukses", logBase...)
		_ = d.Ack()

	case messaging.IsFatal(dispatchErr):
		claimed.Fail(dispatchErr.Error(), now)
		if err := c.repo.Update(ctx, claimed); err != nil {
			c.logger.Error("consumer: gagal mark failed (fatal)", slog.String("id", msg.ID.String()), slog.Any("error", err))
		}
		if c.metrics != nil {
			c.metrics.Handled.Add(1)
			c.metrics.Failed.Add(1)
		}
		c.logger.LogAttrs(context.Background(), slog.LevelError, "consumer: message fatal",
			append(logBase, slog.String("error_class", "fatal"), slog.String("status", string(claimed.Status)), slog.Any("error", dispatchErr))...)
		_ = d.Nack(false) // ke DLQ

	default:
		// Error retryable.
		if claimed.AttemptCount < claimed.MaxAttempts {
			delay := messaging.RetryDelayFor(claimed.AttemptCount, c.opts.RetryDelays)
			retryQ := messaging.RetryQueueName(queue, msg.Type, delay)

			// Publish ke retry queue TTL dulu; hanya setelah ter-confirm kita tandai
			// RETRY_WAIT dan ack pesan asli. Bila gagal, kirim ke DLQ agar tidak hilang.
			if err := c.retryPublisher.PublishToQueue(ctx, retryQ, msg); err != nil {
				c.logger.Error("consumer: gagal publish ke retry queue",
					slog.String("queue", retryQ), slog.String("id", msg.ID.String()), slog.Any("error", err))
				claimed.Fail("gagal schedule retry: "+err.Error(), now)
				if uerr := c.repo.Update(ctx, claimed); uerr != nil {
					c.logger.Error("consumer: gagal mark failed (retry schedule)", slog.String("id", msg.ID.String()), slog.Any("error", uerr))
				}
				if c.metrics != nil {
					c.metrics.Handled.Add(1)
					c.metrics.Failed.Add(1)
				}
				_ = d.Nack(false) // ke DLQ
				return
			}

			nextDelay := messaging.CalculateRetryDelay(claimed.AttemptCount, c.opts.BaseDelay, c.opts.MaxDelay)
			claimed.ScheduleRetry(dispatchErr.Error(), now.Add(nextDelay), now)
			if err := c.repo.Update(ctx, claimed); err != nil {
				c.logger.Error("consumer: gagal mark retry_wait", slog.String("id", msg.ID.String()), slog.Any("error", err))
			}
			if c.metrics != nil {
				c.metrics.Handled.Add(1)
				c.metrics.Retried.Add(1)
			}
			c.logger.LogAttrs(context.Background(), slog.LevelWarn, "consumer: message retry",
				append(logBase, slog.String("error_class", "retryable"), slog.String("status", string(claimed.Status)), slog.Any("error", dispatchErr))...)
			_ = d.Ack()
		} else {
			claimed.Fail(dispatchErr.Error(), now)
			if err := c.repo.Update(ctx, claimed); err != nil {
				c.logger.Error("consumer: gagal mark failed (retry habis)", slog.String("id", msg.ID.String()), slog.Any("error", err))
			}
			if c.metrics != nil {
				c.metrics.Handled.Add(1)
				c.metrics.Failed.Add(1)
			}
			c.logger.LogAttrs(context.Background(), slog.LevelError, "consumer: message gagal (retry habis)",
				append(logBase, slog.String("error_class", "retry_exhausted"), slog.String("status", string(claimed.Status)), slog.Any("error", dispatchErr))...)
			_ = d.Nack(false) // ke DLQ
		}
	}
}
