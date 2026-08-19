package rabbitmq

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"sipon-be/internal/modules/messaging/application/ports"
	"sipon-be/internal/modules/messaging/domain/message/valueobject"
)

// TestPublisherConsumer_Integration memverifikasi publisher confirm + consumer
// delivery + ack terhadap broker nyata. Dilewati bila RABBITMQ_DSN tidak diset.
func TestPublisherConsumer_Integration(t *testing.T) {
	dsn := os.Getenv("RABBITMQ_DSN")
	if dsn == "" {
		t.Skip("RABBITMQ_DSN tidak diset; skip integration test")
	}

	exchange := "sipon.events.integration"
	dlx := "sipon.events.integration.dlx"
	queue := "sipon.worker.integration"
	routing := "integ.test"

	// 1. Topology
	topo, err := NewTopology(Options{
		DSN: dsn, Exchange: exchange, DLXExchange: dlx,
		RetryDelays: []time.Duration{time.Minute},
	})
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	if err := topo.Declare([]valueobject.Binding{{Queue: queue, RoutingKey: routing}}); err != nil {
		t.Fatalf("Declare: %v", err)
	}
	_ = topo.Close()

	// 2. Publisher (confirm)
	pub, err := NewPublisher(dsn, exchange, 5*time.Second)
	if err != nil {
		t.Fatalf("NewPublisher: %v", err)
	}
	defer pub.Close()

	msg, err := valueobject.NewMessage(routing, json.RawMessage(`{"x":1}`))
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if err := pub.Publish(context.Background(), msg); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// 3. Consumer (manual ack)
	cons, err := NewConsumer(dsn, 1)
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
	defer cons.Close()

	received := make(chan valueobject.Message, 1)
	handler := func(ctx context.Context, d ports.Delivery) error {
		var got valueobject.Message
		if err := json.Unmarshal(d.Body(), &got); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		_ = d.Ack()
		received <- got
		return nil
	}

	consCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
			t.Fatalf("type mismatch: got %s", got.Type)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout menerima message")
	}
}
