package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeBillingSchemeNotFound         kernel.Code = "BILLING_SCHEME_NOT_FOUND"
	CodeBillingSchemeDuplicate        kernel.Code = "BILLING_SCHEME_DUPLICATE"
	CodeBillingSchemePersistenceFailed kernel.Code = "BILLING_SCHEME_PERSISTENCE_FAILED"
	CodeBillingSchemeQueryFailed      kernel.Code = "BILLING_SCHEME_QUERY_FAILED"
	CodeSchemeItemNotFound            kernel.Code = "SCHEME_ITEM_NOT_FOUND"
	CodeSchemeItemDuplicate           kernel.Code = "SCHEME_ITEM_DUPLICATE"
	CodeSchemeAssignmentExists        kernel.Code = "SCHEME_ASSIGNMENT_EXISTS"
)
