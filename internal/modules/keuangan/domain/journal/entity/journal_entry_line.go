package entity

type JournalEntryLine struct {
	ID              string
	JournalEntryID  string
	AccountID       string
	AccountCode     string
	AccountName     string
	Description     *string
	Debit           float64
	Credit          float64
}

func NewJournalEntryLine(id, journalEntryID, accountID, accountCode string, debit, credit float64, description *string) *JournalEntryLine {
	return &JournalEntryLine{
		ID:             id,
		JournalEntryID: journalEntryID,
		AccountID:      accountID,
		AccountCode:    accountCode,
		Description:    description,
		Debit:          debit,
		Credit:         credit,
	}
}
