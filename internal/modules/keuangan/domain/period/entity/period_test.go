package entity_test

import (
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/domain/period/constant"
	"sipon-be/internal/modules/keuangan/domain/period/entity"
)

func createTestPeriod() *entity.AccountingPeriod {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
	p, _ := entity.NewAccountingPeriod("per-1", "Periode 2026", start, end, "user-1")
	return p
}

func TestNewAccountingPeriod(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        string
		pName     string
		startDate time.Time
		endDate   time.Time
		createdBy string
		wantErr   bool
	}{
		{"valid period", "per-1", "Periode 2026", start, end, "user-1", false},
		{"empty id", "", "Periode 2026", start, end, "user-1", true},
		{"empty name", "per-2", "", start, end, "user-1", true},
		{"empty createdBy", "per-3", "Periode 2026", start, end, "", true},
		{"end before start", "per-4", "Periode 2026", start, start.AddDate(0, -1, 0), "user-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := entity.NewAccountingPeriod(tt.id, tt.pName, tt.startDate, tt.endDate, tt.createdBy)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Status != constant.PeriodOpen {
				t.Errorf("expected open status, got %s", p.Status)
			}
		})
	}
}

func TestPeriodClose(t *testing.T) {
	p := createTestPeriod()

	if err := p.Close("user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != constant.PeriodClosed {
		t.Errorf("expected closed status, got %s", p.Status)
	}

	if err := p.Close("user-1"); err == nil {
		t.Error("expected error when closing already closed period")
	}
}

func TestPeriodReopen(t *testing.T) {
	p := createTestPeriod()
	if err := p.Close("user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Reopen(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != constant.PeriodOpen {
		t.Errorf("expected open status, got %s", p.Status)
	}
	if p.ClosedBy != nil {
		t.Error("closedBy should be nil after reopen")
	}

	if err := p.Reopen(); err == nil {
		t.Error("expected error when reopening already open period")
	}
}

func TestPeriodLock(t *testing.T) {
	p := createTestPeriod()
	if err := p.Close("user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := p.Lock("user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != constant.PeriodLocked {
		t.Errorf("expected locked status, got %s", p.Status)
	}

	if err := p.Lock("user-1"); err == nil {
		t.Error("expected error when locking already locked period")
	}
}

func TestPeriodLockFromOpen(t *testing.T) {
	p := createTestPeriod()

	if err := p.Lock("user-1"); err == nil {
		t.Error("expected error when locking open period")
	}
}

func TestPeriodCanPost(t *testing.T) {
	t.Run("open", func(t *testing.T) {
		p := createTestPeriod()
		if !p.CanPost() {
			t.Error("open period should be postable")
		}
	})

	t.Run("closing", func(t *testing.T) {
		p := createTestPeriod()
		if err := p.StartClosing(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.CanPost() {
			t.Error("closing period should not be postable")
		}
	})

	t.Run("closed", func(t *testing.T) {
		p := createTestPeriod()
		if err := p.Close("user-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.CanPost() {
			t.Error("closed period should not be postable")
		}
	})

	t.Run("locked", func(t *testing.T) {
		p := createTestPeriod()
		if err := p.Close("user-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if err := p.Lock("user-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.CanPost() {
			t.Error("locked period should not be postable")
		}
	})
}

func TestPeriodStartClosing(t *testing.T) {
	p := createTestPeriod()

	if err := p.StartClosing(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != constant.PeriodClosing {
		t.Errorf("expected closing status, got %s", p.Status)
	}

	if err := p.StartClosing(); err == nil {
		t.Error("expected error when starting closing on non-open period")
	}
}

func TestPeriodIsOpen(t *testing.T) {
	p := createTestPeriod()
	if !p.IsOpen() {
		t.Error("new period should be open")
	}

	if err := p.Close("user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.IsOpen() {
		t.Error("closed period should not be open")
	}
}
