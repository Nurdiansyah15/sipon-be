package dto

type InvoiceSummaryResponse struct {
	BillingPeriodID   string  `json:"billing_period_id"`
	BillingPeriodName string  `json:"billing_period_name"`
	TotalTagihan      float64 `json:"total_tagihan"`
	TotalTerbayar     float64 `json:"total_terbayar"`
	TotalTunggakan    float64 `json:"total_tunggakan"`
	JumlahInvoice     int64   `json:"jumlah_invoice"`
	JumlahLunas       int64   `json:"jumlah_lunas"`
	JumlahBelum       int64   `json:"jumlah_belum"`
}

type InvoiceSummaryQuery struct {
	BillingPeriodID *string `form:"billing_period_id"`
}

type OutstandingSantriResponse struct {
	SantriID         string  `json:"santri_id"`
	TotalOutstanding float64 `json:"total_outstanding"`
	JumlahInvoice    int     `json:"jumlah_invoice"`
}

type OutstandingListQuery struct {
	BillingPeriodID *string `form:"billing_period_id"`
	Page            int     `form:"page"`
	Limit           int     `form:"limit"`
}

type LedgerLineResponse struct {
	Date          string  `json:"date"`
	JournalNumber string  `json:"journal_number"`
	Description   string  `json:"description"`
	Debit         float64 `json:"debit"`
	Credit        float64 `json:"credit"`
	Balance       float64 `json:"balance"`
}

type LedgerResponse struct {
	AccountID   string               `json:"account_id"`
	AccountCode string               `json:"account_code"`
	AccountName string               `json:"account_name"`
	Lines       []LedgerLineResponse `json:"lines"`
}

type LedgerQuery struct {
	AccountID string `form:"account_id" binding:"required"`
	PeriodID  string `form:"period_id" binding:"required"`
}

type TrialBalanceLine struct {
	AccountID   string  `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	AccountType string  `json:"account_type"`
	Debit       float64 `json:"debit"`
	Credit      float64 `json:"credit"`
}

type TrialBalanceResponse struct {
	PeriodID    string             `json:"period_id"`
	PeriodName  string             `json:"period_name"`
	Lines       []TrialBalanceLine `json:"lines"`
	TotalDebit  float64            `json:"total_debit"`
	TotalCredit float64            `json:"total_credit"`
}

type TrialBalanceQuery struct {
	PeriodID string `form:"period_id" binding:"required"`
}

type BalanceSheetLine struct {
	AccountID   string  `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	Amount      float64 `json:"amount"`
}

type BalanceSheetResponse struct {
	AsOfDate         string             `json:"as_of_date"`
	Assets           []BalanceSheetLine `json:"assets"`
	TotalAssets      float64            `json:"total_assets"`
	Liabilities      []BalanceSheetLine `json:"liabilities"`
	TotalLiabilities float64            `json:"total_liabilities"`
	Equities         []BalanceSheetLine `json:"equities"`
	TotalEquities    float64            `json:"total_equities"`
}

type BalanceSheetQuery struct {
	PeriodID *string `form:"period_id"`
	AsOfDate *string `form:"as_of_date"`
}

type IncomeStatementLine struct {
	AccountID   string  `json:"account_id"`
	AccountCode string  `json:"account_code"`
	AccountName string  `json:"account_name"`
	Amount      float64 `json:"amount"`
}

type IncomeStatementResponse struct {
	PeriodID     string                `json:"period_id"`
	PeriodName   string                `json:"period_name"`
	Revenues     []IncomeStatementLine `json:"revenues"`
	TotalRevenue float64               `json:"total_revenue"`
	Expenses     []IncomeStatementLine `json:"expenses"`
	TotalExpense float64               `json:"total_expense"`
	NetIncome    float64               `json:"net_income"`
}

type IncomeStatementQuery struct {
	PeriodID string `form:"period_id" binding:"required"`
}
