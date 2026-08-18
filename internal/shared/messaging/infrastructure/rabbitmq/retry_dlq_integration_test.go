package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"sipon-be/internal/shared/messaging"
)

// TestRetryTTLRedelivery_Integration membuktikan message yang masuk retry queue
// TTL, setelah TTL habis, kembali ke main queue dengan routing key asli dan
// message ID yang sama.
func TestRetryTTLRedelivery_Integration(t *testing.T) {
	dsn := os.Getenv("RABBITMQ_DSN")
	if dsn == "" {
		t.Skip("RABBITMQ_DSN tidak diset; skip integration test")
	}

	ns := time.Now().UnixNano()
	exchange := fmt.Sprintf("sipon.events.retry.%d", ns)
	dlx := exchange + ".dlx"
	queue := fmt.Sprintf("sipon.worker.retry.%d", ns)
	routing := "retry.test"
	delay := time.Second

	topo, err := NewTopology(Options{
		DSN: dsn, Exchange: exchange, DLXExchange: dlx,
		RetryDelays: []time.Duration{delay},
	})
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	if err := topo.Declare([]messaging.Binding{{Queue: queue, RoutingKey: routing}}); err != nil {
		t.Fatalf("Declare: %v", err)
	}
	_ = topo.Close()

	pub, err := NewPublisher(dsn, exchange, 5*time.Second)
	if err != nil {
		t.Fatalf("NewPublisher: %v", err)
	}
	defer pub.Close()

	msg, _ := messaging.NewMessage(routing, json.RawMessage(`{"x":1}`))
	retryQ := messaging.RetryQueueName(queue, routing, delay)
	if err := pub.PublishToQueue(context.Background(), retryQ, msg); err != nil {
		t.Fatalf("PublishToQueue: %v", err)
	}

	cons, err := NewConsumer(dsn, 1)
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer cons.Close()

	received := make(chan messaging.Message, 1)
	handler := func(ctx context.Context, d messaging.Delivery) error {
		var got messaging.Message
		if err := json.Unmarshal(d.Body(), &got); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		_ = d.Ack()
		received <- got
		return nil
	}

	consCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	go func() {
		_ = cons.Consume(consCtx, queue, handler)
	}()

	select {
	case got := <-received:
		if got.ID != msg.ID {
			t.Fatalf("id mismatch: got %s want %s", got.ID, msg.ID)
		}
		if got.Type != routing {
			t.Fatalf("routing key hilang: got %q want %q", got.Type, routing)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("timeout: message tidak kembali ke main queue setelah TTL")
	}
}

// TestDLQ_Integration membuktikan message yang di-nack tanpa requeue masuk ke DLQ
// per consumer role.
func TestDLQ_Integration(t *testing.T) {
	dsn := os.Getenv("RABBITMQ_DSN")
	if dsn == "" {
		t.Skip("RABBITMQ_DSN tidak diset; skip integration test")
	}

	ns := time.Now().UnixNano()
	exchange := fmt.Sprintf("sipon.events.dlqtest.%d", ns)
	dlx := exchange + ".dlx"
	queue := fmt.Sprintf("sipon.worker.dlqtest.%d", ns)
	routing := "dlq.test"

	topo, err := NewTopology(Options{
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

	pub, err := NewPublisher(dsn, exchange, 5*time.Second)
	if err != nil {
		t.Fatalf("NewPublisher: %v", err)
	}
	defer pub.Close()

	msg, _ := messaging.NewMessage(routing, json.RawMessage(`{"x":1}`))
	if err := pub.Publish(context.Background(), msg); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// 1. Konsumen utama men-nack tanpa requeue (simulasi fatal/max retry).
	cons, err := NewConsumer(dsn, 1)
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer cons.Close()

	nacked := make(chan struct{}, 1)
	consCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	go func() {
		_ = cons.Consume(consCtx, queue, func(ctx context.Context, d messaging.Delivery) error {
			_ = d.Nack(false)
			nacked <- struct{}{}
			return nil
		})
	}()

	select {
	case <-nacked:
	case <-time.After(15 * time.Second):
		t.Fatal("timeout: main queue tidak menerima message")
	}

	// 2. Konsumen DLQ harus menerima message yang sama.
	consDLQ, err := NewConsumer(dsn, 1)
	if err != nil {
		t.Fatalf("NewConsumer dlq: %v", err)
	}
	defer consDLQ.Close()

	dlqReceived := make(chan messaging.Message, 1)
	dlqCtx, dCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer dCancel()
	go func() {
		_ = consDLQ.Consume(dlqCtx, queue+".dlq", func(ctx context.Context, d messaging.Delivery) error {
			var got messaging.Message
			_ = json.Unmarshal(d.Body(), &got)
			_ = d.Ack()
			dlqReceived <- got
			return nil
		})
	}()

	select {
	case got := <-dlqReceived:
		if got.ID != msg.ID {
			t.Fatalf("id mismatch di DLQ: got %s want %s", got.ID, msg.ID)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("timeout: message tidak sampai ke DLQ")
	}
}
