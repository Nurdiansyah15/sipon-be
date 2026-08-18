package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"sipon-be/internal/shared/messaging"
)

type amqpDelivery struct {
	d amqp.Delivery
}

func (a *amqpDelivery) Body() []byte { return a.d.Body }

func (a *amqpDelivery) Ack() error { return a.d.Ack(false) }

func (a *amqpDelivery) Nack(requeue bool) error { return a.d.Nack(false, requeue) }

// RabbitMQConsumer mengonsumsi message dari sebuah queue dengan manual
// acknowledgement dan prefetch configurable. Handler bertanggung jawab memanggil
// Ack/Nack pada delivery.
type RabbitMQConsumer struct {
	mu       sync.Mutex
	dsn      string
	prefetch int
	conn     *amqp.Connection
	ch       *amqp.Channel
}

func NewConsumer(dsn string, prefetch int) (*RabbitMQConsumer, error) {
	if dsn == "" {
		return nil, fmt.Errorf("rabbitmq: DSN kosong")
	}
	if prefetch <= 0 {
		prefetch = 10
	}
	c := &RabbitMQConsumer{dsn: dsn, prefetch: prefetch}
	if err := c.ensureChannel(); err != nil {
		return nil, err
	}
	return c, nil
}

// Consume menjalankan loop konsumsi sampai context selesai, dengan reconnect bila
// connection/channel terputus.
func (c *RabbitMQConsumer) Consume(ctx context.Context, queue string, handler messaging.ConsumerHandler) error {
	for {
		if err := c.ensureChannel(); err != nil {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second):
				continue
			}
		}

		msgs, err := c.ch.Consume(queue, "", false, false, false, false, nil)
		if err != nil {
			c.close()
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(time.Second):
				continue
			}
		}

		delivered := false
		for {
			select {
			case <-ctx.Done():
				_ = c.ch.Cancel(queue, false)
				return nil
			case d, ok := <-msgs:
				if !ok {
					delivered = true
					goto reconnect
				}
				delivery := &amqpDelivery{d: d}
				_ = handler(ctx, delivery)
			}
		}
	reconnect:
		_ = delivered
		c.close()
	}
}

func (c *RabbitMQConsumer) ensureChannel() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ch != nil && !c.ch.IsClosed() && c.conn != nil && !c.conn.IsClosed() {
		return nil
	}
	c.closeLocked()

	conn, err := amqp.Dial(c.dsn)
	if err != nil {
		return fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("rabbitmq channel: %w", err)
	}
	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return fmt.Errorf("rabbitmq qos: %w", err)
	}
	c.conn = conn
	c.ch = ch
	return nil
}

func (c *RabbitMQConsumer) close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closeLocked()
}

func (c *RabbitMQConsumer) closeLocked() {
	if c.ch != nil {
		_ = c.ch.Close()
		c.ch = nil
	}
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
}

func (c *RabbitMQConsumer) Close() error {
	c.close()
	return nil
}
