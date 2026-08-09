package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeAccountNotFound         kernel.Code = "ACCOUNT_NOT_FOUND"
	CodeAccountDuplicate        kernel.Code = "ACCOUNT_DUPLICATE"
	CodeAccountNotPostable      kernel.Code = "ACCOUNT_NOT_POSTABLE"
	CodeAccountIsSystem         kernel.Code = "ACCOUNT_IS_SYSTEM"
	CodeAccountHasJournal       kernel.Code = "ACCOUNT_HAS_JOURNAL"
	CodeAccountInvalidSubType   kernel.Code = "ACCOUNT_INVALID_SUB_TYPE"
	CodeAccountSubTypeRequired  kernel.Code = "ACCOUNT_SUB_TYPE_REQUIRED"
	CodeAccountPersistenceFailed kernel.Code = "ACCOUNT_PERSISTENCE_FAILED"
	CodeAccountQueryFailed      kernel.Code = "ACCOUNT_QUERY_FAILED"
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

type AccountSubType string

const (
	SubTypeCashBank                AccountSubType = "cash_bank"
	SubTypeReceivable              AccountSubType = "receivable"
	SubTypePrepaidExpense          AccountSubType = "prepaid_expense"
	SubTypeInventory               AccountSubType = "inventory"
	SubTypeFixedAsset              AccountSubType = "fixed_asset"
	SubTypeAccumulatedDepreciation AccountSubType = "accumulated_depreciation"
	SubTypeIntangibleAsset         AccountSubType = "intangible_asset"
	SubTypeInvestment              AccountSubType = "investment"
	SubTypeOtherAsset              AccountSubType = "other_asset"
	SubTypePayable                 AccountSubType = "payable"
	SubTypeCustomerAdvance         AccountSubType = "customer_advance"
	SubTypeUnearnedRevenue         AccountSubType = "unearned_revenue"
	SubTypeTaxPayable              AccountSubType = "tax_payable"
	SubTypeAccruedLiability        AccountSubType = "accrued_liability"
	SubTypeLongTermLiability       AccountSubType = "long_term_liability"
	SubTypeOtherLiability          AccountSubType = "other_liability"
	SubTypeCapital                 AccountSubType = "capital"
	SubTypeRetainedEarnings        AccountSubType = "retained_earnings"
	SubTypeCurrentYearEarnings     AccountSubType = "current_year_earnings"
	SubTypeWithdrawal              AccountSubType = "withdrawal"
	SubTypeOperatingRevenue        AccountSubType = "operating_revenue"
	SubTypeNonOperatingRevenue     AccountSubType = "non_operating_revenue"
	SubTypeCostOfGoodsSold         AccountSubType = "cost_of_goods_sold"
	SubTypeOperatingExpense        AccountSubType = "operating_expense"
	SubTypeDepreciationExpense     AccountSubType = "depreciation_expense"
	SubTypeNonOperatingExpense     AccountSubType = "non_operating_expense"
	SubTypeTaxExpense              AccountSubType = "tax_expense"
)

// SubTypesByType memetakan sub-tipe yang sah untuk setiap tipe akun,
// mengikuti taksonomi di docs/plan/coa-sub-type.md.
var SubTypesByType = map[AccountType][]AccountSubType{
	TypeAsset: {
		SubTypeCashBank, SubTypeReceivable, SubTypePrepaidExpense, SubTypeInventory,
		SubTypeFixedAsset, SubTypeAccumulatedDepreciation, SubTypeIntangibleAsset,
		SubTypeInvestment, SubTypeOtherAsset,
	},
	TypeLiability: {
		SubTypePayable, SubTypeCustomerAdvance, SubTypeUnearnedRevenue,
		SubTypeTaxPayable, SubTypeAccruedLiability, SubTypeLongTermLiability,
		SubTypeOtherLiability,
	},
	TypeEquity: {
		SubTypeCapital, SubTypeRetainedEarnings, SubTypeCurrentYearEarnings, SubTypeWithdrawal,
	},
	TypeRevenue: {
		SubTypeOperatingRevenue, SubTypeNonOperatingRevenue,
	},
	TypeExpense: {
		SubTypeCostOfGoodsSold, SubTypeOperatingExpense, SubTypeDepreciationExpense,
		SubTypeNonOperatingExpense, SubTypeTaxExpense,
	},
}

func IsValidSubTypeForType(t AccountType, st AccountSubType) bool {
	valid, ok := SubTypesByType[t]
	if !ok {
		return false
	}
	for _, v := range valid {
		if v == st {
			return true
		}
	}
	return false
}
