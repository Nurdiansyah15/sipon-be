package ports

import (
	"errors"
	"testing"
)

func TestErrorClassification(t *testing.T) {
	retry := NewRetryableError(errors.New("transient"))
	fatal := NewFatalError(errors.New("bad"))

	if !IsRetryable(retry) {
		t.Fatal("harus retryable")
	}
	if IsRetryable(fatal) {
		t.Fatal("fatal tidak boleh retryable")
	}
	if !IsFatal(fatal) {
		t.Fatal("harus fatal")
	}
	if IsFatal(retry) {
		t.Fatal("retryable tidak boleh fatal")
	}
	if IsFatal(errors.New("plain")) {
		t.Fatal("plain error tidak boleh fatal")
	}
	if IsRetryable(errors.New("plain")) {
		t.Fatal("plain error tidak boleh retryable")
	}
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("root")
	if !errors.Is(NewRetryableError(inner), inner) {
		t.Fatal("RetryableError harus unwrap inner")
	}
	if !errors.Is(NewFatalError(inner), inner) {
		t.Fatal("FatalError harus unwrap inner")
	}
}

func TestError_EmptyMessage(t *testing.T) {
	if NewRetryableError(nil).Error() == "" {
		t.Fatal("retryable error message tidak boleh kosong")
	}
	if NewFatalError(nil).Error() == "" {
		t.Fatal("fatal error message tidak boleh kosong")
	}
}
