package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeBillingBatchNotFound          kernel.Code = "BILLING_BATCH_NOT_FOUND"
	CodeBillingBatchPersistenceFailed kernel.Code = "BILLING_BATCH_PERSISTENCE_FAILED"
	CodeBillingBatchQueryFailed       kernel.Code = "BILLING_BATCH_QUERY_FAILED"
)

type BillingBatchStatus string

const (
	BillingBatchProcessing BillingBatchStatus = "processing"
	BillingBatchCompleted  BillingBatchStatus = "completed"
	BillingBatchFailed     BillingBatchStatus = "failed"
)

type BillingBatchTargetStatus string

const (
	TargetPending                  BillingBatchTargetStatus = "pending"
	TargetCreated                  BillingBatchTargetStatus = "created"
	TargetSkippedNoAssignment      BillingBatchTargetStatus = "skipped_no_assignment"
	TargetSkippedWrongScheme       BillingBatchTargetStatus = "skipped_wrong_scheme"
	TargetSkippedAlreadyInvoiced   BillingBatchTargetStatus = "skipped_already_invoiced"
	TargetSkippedComponentMissing  BillingBatchTargetStatus = "skipped_component_missing"
	TargetError                    BillingBatchTargetStatus = "error"
)
