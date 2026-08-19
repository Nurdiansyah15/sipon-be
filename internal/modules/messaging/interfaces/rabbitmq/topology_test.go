package rabbitmq

import (
	"os"
	"testing"
	"time"

	messaging "sipon-be/internal/modules/messaging/application/ports"
)

// TestTopology_Declare_Integration memverifikasi deklarasi topology idempotent
// dan durable terhadap broker nyata. Dilewati bila RABBITMQ_DSN tidak diset.
func TestTopology_Declare_Integration(t *testing.T) {
	dsn := os.Getenv("RABBITMQ_DSN")
	if dsn == "" {
		t.Skip("RABBITMQ_DSN tidak diset; skip integration test")
	}

	opts := Options{
		DSN:         dsn,
		Exchange:    "sipon.events.test",
		DLXExchange: "sipon.events.test.dlx",
		RetryDelays: []time.Duration{time.Minute, 5 * time.Minute},
	}
	bindings := []messaging.Binding{
		{Queue: "sipon.worker.scheduler.test", RoutingKey: "akademik.fingerprint.sync"},
		{Queue: "sipon.worker.scheduler.test", RoutingKey: "akademik.session.auto_close"},
	}

	topo, err := NewTopology(opts)
	if err != nil {
		t.Fatalf("NewTopology: %v", err)
	}
	defer topo.Close()

	// Idempotent: declare dua kali seharusnya tidak error.
	if err := topo.Declare(bindings); err != nil {
		t.Fatalf("Declare #1: %v", err)
	}
	if err := topo.Declare(bindings); err != nil {
		t.Fatalf("Declare #2 (idempotent): %v", err)
	}
}
