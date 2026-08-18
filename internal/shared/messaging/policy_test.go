package messaging

import (
	"testing"
	"time"
)

func TestRetryPolicy_DefaultAndOverride(t *testing.T) {
	p := NewRetryPolicy(3)
	if p.MaxRetryFor("a.b") != 3 {
		t.Fatalf("default: got %d", p.MaxRetryFor("a.b"))
	}
	p.Register("a.b", 7)
	if p.MaxRetryFor("a.b") != 7 {
		t.Fatalf("override: got %d", p.MaxRetryFor("a.b"))
	}
	if p.MaxRetryFor("x.y") != 3 {
		t.Fatalf("default setelah override: got %d", p.MaxRetryFor("x.y"))
	}
}

func TestRetryPolicy_IsRetryable(t *testing.T) {
	p := NewRetryPolicy(3)
	max := p.MaxRetryFor("a.b")
	if !p.IsRetryable(0, max) {
		t.Fatal("attempt 0 harus retryable")
	}
	if !p.IsRetryable(2, max) {
		t.Fatal("attempt 2 harus retryable")
	}
	if p.IsRetryable(3, max) {
		t.Fatal("attempt 3 (== max) tidak boleh retryable")
	}
}

func TestCalculateRetryDelay_BoundsAndMonotonic(t *testing.T) {
	base := 30 * time.Second
	max := 30 * time.Minute
	prev := time.Duration(0)
	for attempt := 0; attempt < 10; attempt++ {
		d := CalculateRetryDelay(attempt, base, max)
		if d < 0 || d > max {
			t.Fatalf("attempt %d di luar bounds: %v", attempt, d)
		}
		if d < prev {
			t.Fatalf("delay tidak monoton: %v < %v", d, prev)
		}
		prev = d
	}
}

func TestCalculateRetryDelay_InvalidInput(t *testing.T) {
	d := CalculateRetryDelay(-1, 0, 30*time.Minute)
	if d <= 0 {
		t.Fatalf("delay harus positif: %v", d)
	}
}
