package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeScanLogNotFound          kernel.Code = "SCAN_LOG_NOT_FOUND"
	CodeScanLogInvalid           kernel.Code = "SCAN_LOG_INVALID"
	CodeScanLogPersistenceFailed kernel.Code = "SCAN_LOG_PERSISTENCE_FAILED"
	CodeScanLogQueryFailed       kernel.Code = "SCAN_LOG_QUERY_FAILED"
)
