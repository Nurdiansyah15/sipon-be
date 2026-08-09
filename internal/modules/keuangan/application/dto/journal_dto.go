package dto

type CreateJournalEntryRequest struct {
	EntryDate   string                     `json:"entry_date" binding:"required"`
	Description string                     `json:"description" binding:"required"`
	PeriodID    string                     `json:"period_id" binding:"required"`
	Lines       []CreateJournalLineRequest `json:"lines" binding:"required"`
}

type CreateJournalLineRequest struct {
	AccountID   string  `json:"account_id" binding:"required"`
	Description *string `json:"description,omitempty"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
}

type JournalListQuery struct {
	PeriodID   *string `form:"period_id"`
	Status     *string `form:"status"`
	SourceType *string `form:"source_type"`
	Page       int     `form:"page"`
	Limit      int     `form:"limit"`
}

type JournalEntryResponse struct {
	ID            string                `json:"id"`
	JournalNumber string                `json:"journal_number"`
	EntryDate     string                `json:"entry_date"`
	Description   string                `json:"description"`
	SourceType    *string               `json:"source_type,omitempty"`
	SourceID      *string               `json:"source_id,omitempty"`
	PeriodID      string                `json:"period_id"`
	TotalDebit    float64               `json:"total_debit"`
	TotalCredit   float64               `json:"total_credit"`
	Status        string                `json:"status"`
	PostedBy      string                `json:"posted_by"`
	PostedAt      *string               `json:"posted_at,omitempty"`
	Lines         []JournalLineResponse `json:"lines,omitempty"`
	CreatedAt     string                `json:"created_at"`
	UpdatedAt     string                `json:"updated_at"`
}

type JournalLineResponse struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name,omitempty"`
	Description *string `json:"description,omitempty"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
}
