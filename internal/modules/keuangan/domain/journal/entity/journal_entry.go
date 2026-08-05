package entity

import (
	"time"

	"sipon-be/internal/modules/keuangan/domain/journal/constant"
	"sipon-be/internal/shared/kernel"
)

type JournalEntry struct {
	ID             string
	JournalNumber  string
	EntryDate      time.Time
	Description    string
	SourceType     *constant.SourceType
	SourceID       *string
	PeriodID       string
	TotalDebit     float64
	TotalCredit    float64
	PostedBy       string
	PostedAt       *time.Time
	Status         constant.JournalStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Lines          []*JournalEntryLine
}

func NewJournalEntry(id, journalNumber string, entryDate time.Time, description, periodID, postedBy string) (*JournalEntry, error) {
	if id == "" || journalNumber == "" || description == "" || periodID == "" || postedBy == "" {
		return nil, kernel.New(constant.CodeJournalNotFound)
	}
	now := time.Now()
	return &JournalEntry{
		ID:            id,
		JournalNumber: journalNumber,
		EntryDate:     entryDate,
		Description:   description,
		PeriodID:      periodID,
		PostedBy:      postedBy,
		Status:        constant.JournalDraft,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (j *JournalEntry) AddLine(line *JournalEntryLine) error {
	j.Lines = append(j.Lines, line)
	j.TotalDebit += line.Debit
	j.TotalCredit += line.Credit
	return nil
}

func (j *JournalEntry) Validate() error {
	if len(j.Lines) < 2 {
		return kernel.New(constant.CodeJournalMinLines)
	}
	if j.TotalDebit != j.TotalCredit {
		return kernel.New(constant.CodeJournalNotBalanced)
	}
	return nil
}

func (j *JournalEntry) Post() error {
	if j.Status != constant.JournalDraft {
		return kernel.New(constant.CodeJournalInvalidStatus)
	}
	if err := j.Validate(); err != nil {
		return err
	}
	now := time.Now()
	j.Status = constant.JournalPosted
	j.PostedAt = &now
	j.UpdatedAt = now
	return nil
}

func (j *JournalEntry) Cancel() error {
	if j.Status != constant.JournalPosted {
		return kernel.New(constant.CodeJournalInvalidStatus)
	}
	if j.SourceType != nil && *j.SourceType != constant.SourceManual {
		return kernel.New(constant.CodeJournalAutoCannotCancel)
	}
	j.Status = constant.JournalCancelled
	j.UpdatedAt = time.Now()
	return nil
}

func (j *JournalEntry) SetSource(sourceType constant.SourceType, sourceID string) {
	j.SourceType = &sourceType
	j.SourceID = &sourceID
}
