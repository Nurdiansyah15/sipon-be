package entity_test

import (
	"testing"
	"time"

	"sipon-be/internal/modules/akademik/domain/academic_period/constant"
	"sipon-be/internal/modules/akademik/domain/academic_period/entity"
)

func createTestPeriod() *entity.AcademicPeriod {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	p, _ := entity.NewAcademicPeriod("per-1", "2026/2027-P1", "Periode 1 2026/2027", start, end)
	return p
}

func TestNewAcademicPeriod(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        string
		code      string
		pName     string
		startDate time.Time
		endDate   time.Time
		wantErr   bool
	}{
		{"valid period", "per-1", "2026/2027-P1", "Periode 1", start, end, false},
		{"empty id", "", "2026/2027-P1", "Periode 1", start, end, true},
		{"empty code", "per-2", "", "Periode 1", start, end, true},
		{"empty name", "per-3", "2026/2027-P1", "", start, end, true},
		{"end before start", "per-4", "2026/2027-P1", "Periode 1", start, start.AddDate(0, -1, 0), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := entity.NewAcademicPeriod(tt.id, tt.code, tt.pName, tt.startDate, tt.endDate)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Status != constant.AcademicPeriodStatusDraft {
				t.Errorf("expected draft status, got %s", p.Status)
			}
		})
	}
}

func TestAcademicPeriodLifecycle(t *testing.T) {
	p := createTestPeriod()

	if err := p.Open(); err != nil {
		t.Fatalf("open failed: %v", err)
	}
	if p.Status != constant.AcademicPeriodStatusOpen {
		t.Errorf("expected open, got %s", p.Status)
	}
	if err := p.Open(); err == nil {
		t.Error("expected error when opening already-open period")
	}

	if err := p.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if p.Status != constant.AcademicPeriodStatusClosed {
		t.Errorf("expected closed, got %s", p.Status)
	}
	if err := p.Close(); err == nil {
		t.Error("expected error when closing already-closed period")
	}

	if err := p.Archive(); err != nil {
		t.Fatalf("archive failed: %v", err)
	}
	if p.Status != constant.AcademicPeriodStatusArchived {
		t.Errorf("expected archived, got %s", p.Status)
	}
}

func TestAcademicPeriodUpdate(t *testing.T) {
	p := createTestPeriod()
	newEnd := time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC)
	if err := p.Update("Periode 1 Revisi", nil, &newEnd); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if p.Name != "Periode 1 Revisi" {
		t.Errorf("expected updated name, got %s", p.Name)
	}

	badStart := time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC)
	if err := p.Update("", &badStart, &newEnd); err == nil {
		t.Error("expected error when end before start")
	}
}
