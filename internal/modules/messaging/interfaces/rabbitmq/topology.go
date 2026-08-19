package rabbitmq

import (
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	messaging "sipon-be/internal/modules/messaging/application/ports"
)

// Options konfigurasi koneksi dan topology RabbitMQ.
type Options struct {
	DSN         string
	Exchange    string
	DLXExchange string
	RetryDelays []time.Duration
}

// Topology mendeklarasikan exchange, queue per consumer role, retry queue dengan
// TTL, dan DLQ. Topology bersifat durable dan idempotent terhadap restart broker.
type Topology struct {
	opts Options
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewTopology(opts Options) (*Topology, error) {
	if opts.DSN == "" {
		return nil, fmt.Errorf("rabbitmq: DSN kosong")
	}
	if opts.Exchange == "" {
		opts.Exchange = "sipon.events"
	}
	if opts.DLXExchange == "" {
		opts.DLXExchange = "sipon.events.dlx"
	}
	if len(opts.RetryDelays) == 0 {
		opts.RetryDelays = []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute}
	}

	conn, err := amqp.Dial(opts.DSN)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	return &Topology{opts: opts, conn: conn, ch: ch}, nil
}

func (t *Topology) Close() error {
	var errs []error
	if t.ch != nil {
		if err := t.ch.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if t.conn != nil {
		if err := t.conn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// Declare membuat topology. Retry queue dan DLQ adalah bagian dari RabbitMQ
// consumer topology, bukan tanggung jawab business handler atau scheduler worker.
func (t *Topology) Declare(bindings []messaging.Binding) error {
	if err := t.ch.ExchangeDeclare(t.opts.Exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", t.opts.Exchange, err)
	}
	if err := t.ch.ExchangeDeclare(t.opts.DLXExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx %s: %w", t.opts.DLXExchange, err)
	}

	queues := groupBindings(bindings)
	if len(queues) == 0 {
		return nil
	}

	for queue, routingKeys := range queues {
		if err := t.declareRoleQueue(queue, routingKeys); err != nil {
			return err
		}
	}
	return nil
}

func (t *Topology) declareRoleQueue(queue string, routingKeys []string) error {
	// Queue utama dengan dead-letter ke DLX. Kegagalan fatal/max-retry masuk DLQ.
	if _, err := t.ch.QueueDeclare(queue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange":    t.opts.DLXExchange,
		"x-dead-letter-routing-key": queue + ".dlq",
	}); err != nil {
		return fmt.Errorf("declare queue %s: %w", queue, err)
	}
	// Bind main queue ke exchange + deklarasikan retry queue per routing key.
	// Retry queue memakai x-dead-letter-routing-key = routing key asli supaya
	// setelah TTL habis message kembali ke main exchange dengan key yang benar.
	for _, key := range routingKeys {
		if err := t.ch.QueueBind(queue, key, t.opts.Exchange, false, nil); err != nil {
			return fmt.Errorf("bind %s -> %s: %w", queue, key, err)
		}
		for _, delay := range t.opts.RetryDelays {
			if delay <= 0 {
				continue
			}
			retryQ := messaging.RetryQueueName(queue, key, delay)
			if _, err := t.ch.QueueDeclare(retryQ, true, false, false, false, amqp.Table{
				"x-message-ttl":             delay.Milliseconds(),
				"x-dead-letter-exchange":    t.opts.Exchange,
				"x-dead-letter-routing-key": key,
			}); err != nil {
				return fmt.Errorf("declare retry queue %s: %w", retryQ, err)
			}
		}
	}

	// DLQ per consumer role.
	dlq := queue + ".dlq"
	if _, err := t.ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlq %s: %w", dlq, err)
	}
	if err := t.ch.QueueBind(dlq, dlq, t.opts.DLXExchange, false, nil); err != nil {
		return fmt.Errorf("bind dlq %s: %w", dlq, err)
	}
	return nil
}

func groupBindings(bindings []messaging.Binding) map[string][]string {
	queues := make(map[string][]string)
	for _, b := range bindings {
		if err := b.Validate(); err != nil {
			continue
		}
		queues[b.Queue] = append(queues[b.Queue], b.RoutingKey)
	}
	return queues
}
