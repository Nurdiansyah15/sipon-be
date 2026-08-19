package ports

import (
	"testing"
	"time"
)

func TestRetryQueueName(t *testing.T) {
	got := RetryQueueName("sipon.worker.scheduler", "akademik.fingerprint.sync", 60*time.Second)
	want := "sipon.worker.scheduler.retry.akademik.fingerprint.sync.60"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
