package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeAccountNotFound        kernel.Code = "ACCOUNT_NOT_FOUND"
	CodeAccountDuplicate       kernel.Code = "ACCOUNT_DUPLICATE"
	CodeAccountNotPostable     kernel.Code = "ACCOUNT_NOT_POSTABLE"
	CodeAccountIsSystem        kernel.Code = "ACCOUNT_IS_SYSTEM"
	CodeAccountHasJournal      kernel.Code = "ACCOUNT_HAS_JOURNAL"
	CodeAccountPersistenceFailed kernel.Code = "ACCOUNT_PERSISTENCE_FAILED"
	CodeAccountQueryFailed     kernel.Code = "ACCOUNT_QUERY_FAILED"
)

type AccountType string

const (
	TypeAsset     AccountType = "asset"
	TypeLiability AccountType = "liability"
	TypeEquity    AccountType = "equity"
	TypeRevenue   AccountType = "revenue"
	TypeExpense   AccountType = "expense"
)

type NormalBalance string

const (
	BalanceDebit  NormalBalance = "debit"
	BalanceCredit NormalBalance = "credit"
)
