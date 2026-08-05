package entity_test

import (
	"testing"
	"time"

	"sipon-be/internal/modules/keuangan/domain/journal/constant"
	"sipon-be/internal/modules/keuangan/domain/journal/entity"
)

func createTestJournal() *entity.JournalEntry {
	j, _ := entity.NewJournalEntry(
		"je-1", "JRN/2026/08/000001", time.Now(),
		"Penerimaan SPP", "per-1", "user-1",
	)
	return j
}

func TestNewJournalEntry(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		journalNumber string
		description   string
		periodID      string
		postedBy      string
		wantErr       bool
	}{
		{"valid entry", "je-1", "JRN/001", "desc", "per-1", "user-1", false},
		{"empty id", "", "JRN/001", "desc", "per-1", "user-1", true},
		{"empty journalNumber", "je-2", "", "desc", "per-1", "user-1", true},
		{"empty description", "je-3", "JRN/001", "", "per-1", "user-1", true},
		{"empty periodID", "je-4", "JRN/001", "desc", "", "user-1", true},
		{"empty postedBy", "je-5", "JRN/001", "desc", "per-1", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j, err := entity.NewJournalEntry(tt.id, tt.journalNumber, time.Now(), tt.description, tt.periodID, tt.postedBy)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if j.Status != constant.JournalDraft {
				t.Errorf("expected draft status, got %s", j.Status)
			}
		})
	}
}

func TestJournalAddLine(t *testing.T) {
	j := createTestJournal()
	line1 := entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 500000, 0, nil)
	line2 := entity.NewJournalEntryLine("line-2", "je-1", "acc-2", "4101", 0, 500000, nil)

	j.AddLine(line1)
	j.AddLine(line2)

	if len(j.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(j.Lines))
	}
	if j.TotalDebit != 500000 {
		t.Errorf("expected totalDebit 500000, got %f", j.TotalDebit)
	}
	if j.TotalCredit != 500000 {
		t.Errorf("expected totalCredit 500000, got %f", j.TotalCredit)
	}
}

func TestJournalValidateBalanced(t *testing.T) {
	j := createTestJournal()
	j.AddLine(entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 500000, 0, nil))
	j.AddLine(entity.NewJournalEntryLine("line-2", "je-1", "acc-2", "4101", 0, 500000, nil))

	if err := j.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJournalValidateUnbalanced(t *testing.T) {
	j := createTestJournal()
	j.AddLine(entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 500000, 0, nil))
	j.AddLine(entity.NewJournalEntryLine("line-2", "je-1", "acc-2", "4101", 0, 300000, nil))

	if err := j.Validate(); err == nil {
		t.Error("expected error for unbalanced journal")
	}
}

func TestJournalValidateMinLines(t *testing.T) {
	j := createTestJournal()
	j.AddLine(entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 100000, 0, nil))

	if err := j.Validate(); err == nil {
		t.Error("expected error for less than 2 lines")
	}
}

func TestJournalPost(t *testing.T) {
	j := createTestJournal()
	j.AddLine(entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 500000, 0, nil))
	j.AddLine(entity.NewJournalEntryLine("line-2", "je-1", "acc-2", "4101", 0, 500000, nil))

	if err := j.Post(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.Status != constant.JournalPosted {
		t.Errorf("expected posted status, got %s", j.Status)
	}
	if j.PostedAt == nil {
		t.Error("postedAt should be set after posting")
	}

	if err := j.Post(); err == nil {
		t.Error("expected error when posting already posted journal")
	}
}

func TestJournalPostUnbalanced(t *testing.T) {
	j := createTestJournal()
	j.AddLine(entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 500000, 0, nil))
	j.AddLine(entity.NewJournalEntryLine("line-2", "je-1", "acc-2", "4101", 0, 300000, nil))

	if err := j.Post(); err == nil {
		t.Error("expected error when posting unbalanced journal")
	}
}

func TestJournalCancelManual(t *testing.T) {
	j := createTestJournal()
	j.AddLine(entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 500000, 0, nil))
	j.AddLine(entity.NewJournalEntryLine("line-2", "je-1", "acc-2", "4101", 0, 500000, nil))
	j.SetSource(constant.SourceManual, "ref-1")
	if err := j.Post(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := j.Cancel(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if j.Status != constant.JournalCancelled {
		t.Errorf("expected cancelled status, got %s", j.Status)
	}
}

func TestJournalCancelAuto(t *testing.T) {
	j := createTestJournal()
	j.AddLine(entity.NewJournalEntryLine("line-1", "je-1", "acc-1", "1101", 500000, 0, nil))
	j.AddLine(entity.NewJournalEntryLine("line-2", "je-1", "acc-2", "4101", 0, 500000, nil))
	j.SetSource(constant.SourcePaymentVerified, "pay-1")
	if err := j.Post(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := j.Cancel(); err == nil {
		t.Error("expected error when cancelling auto-generated journal")
	}
}

func TestJournalCancelDraft(t *testing.T) {
	j := createTestJournal()

	if err := j.Cancel(); err == nil {
		t.Error("expected error when cancelling draft journal")
	}
}

func TestJournalSetSource(t *testing.T) {
	j := createTestJournal()
	j.SetSource(constant.SourceInvoiceIssued, "inv-1")

	if j.SourceType == nil || *j.SourceType != constant.SourceInvoiceIssued {
		t.Error("sourceType should be set")
	}
	if j.SourceID == nil || *j.SourceID != "inv-1" {
		t.Error("sourceID should be set")
	}
}
