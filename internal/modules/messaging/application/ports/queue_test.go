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

func TestRetryDelayFor(t *testing.T) {
	delays := []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute}

	if RetryDelayFor(1, delays) != time.Minute {
		t.Fatal("attempt 1 harus tier pertama")
	}
	if RetryDelayFor(2, delays) != 5*time.Minute {
		t.Fatal("attempt 2 harus tier kedua")
	}
	if RetryDelayFor(3, delays) != 30*time.Minute {
		t.Fatal("attempt 3 harus tier ketiga")
	}
	if RetryDelayFor(10, delays) != 30*time.Minute {
		t.Fatal("attempt > len harus tier terakhir")
	}
	if RetryDelayFor(0, delays) != time.Minute {
		t.Fatal("attempt 0 harus tier pertama")
	}
	if RetryDelayFor(1, nil) != time.Minute {
		t.Fatal("tanpa delays harus default 1m")
	}
}
