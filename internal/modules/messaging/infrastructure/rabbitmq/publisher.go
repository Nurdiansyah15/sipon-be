package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"sipon-be/internal/modules/messaging/domain/message/valueobject"
)

// RabbitMQPublisher mempublish message ke exchange utama dengan publisher confirm
// dan mendukung reconnect otomatis bila connection/channel terputus.
type RabbitMQPublisher struct {
	mu             sync.Mutex
	dsn            string
	exchange       string
	confirmTimeout time.Duration
	conn           *amqp.Connection
	ch             *amqp.Channel
}

func NewPublisher(dsn, exchange string, confirmTimeout time.Duration) (*RabbitMQPublisher, error) {
	if dsn == "" {
		return nil, fmt.Errorf("rabbitmq: DSN kosong")
	}
	if exchange == "" {
		exchange = "sipon.events"
	}
	if confirmTimeout <= 0 {
		confirmTimeout = 10 * time.Second
	}
	p := &RabbitMQPublisher{dsn: dsn, exchange: exchange, confirmTimeout: confirmTimeout}
	if err := p.ensureChannel(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, msg valueobject.Message) error {
	return p.publish(ctx, msg, p.exchange, msg.Type)
}

func (p *RabbitMQPublisher) PublishToQueue(ctx context.Context, queue string, msg valueobject.Message) error {
	return p.publish(ctx, msg, "", queue)
}

// Ping memastikan koneksi/channel tersedia (dial ulang bila perlu).
func (p *RabbitMQPublisher) Ping(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ensureChannelLocked()
}

func (p *RabbitMQPublisher) publish(ctx context.Context, msg valueobject.Message, exchange, routingKey string) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ensureChannelLocked(); err != nil {
		return err
	}

	conf, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		MessageId:     msg.ID.String(),
		CorrelationId: msg.CorrelationID,
		Timestamp:     msg.OccurredAt,
		Body:          body,
	})
	if err != nil {
		p.closeLocked()
		return fmt.Errorf("rabbitmq publish: %w", err)
	}
	ok, err := conf.WaitContext(ctx)
	if err != nil {
		p.closeLocked()
		return fmt.Errorf("rabbitmq confirm: %w", err)
	}
	if !ok {
		p.closeLocked()
		return errors.New("rabbitmq: publish tidak terkonfirmasi oleh broker")
	}
	return nil
}

func (p *RabbitMQPublisher) ensureChannel() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.ensureChannelLocked()
}

func (p *RabbitMQPublisher) ensureChannelLocked() error {
	if p.ch != nil && !p.ch.IsClosed() && p.conn != nil && !p.conn.IsClosed() {
		return nil
	}
	p.closeLocked()

	conn, err := amqp.Dial(p.dsn)
	if err != nil {
		return fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("rabbitmq channel: %w", err)
	}
	if err := ch.Confirm(false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq confirm mode: %w", err)
	}
	if err := ch.ExchangeDeclare(p.exchange, "topic", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq declare exchange: %w", err)
	}
	p.conn = conn
	p.ch = ch
	return nil
}

func (p *RabbitMQPublisher) closeLocked() {
	if p.ch != nil {
		_ = p.ch.Close()
		p.ch = nil
	}
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
	}
}

func (p *RabbitMQPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closeLocked()
	return nil
}
