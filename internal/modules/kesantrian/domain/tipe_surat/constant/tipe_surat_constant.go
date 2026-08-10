package constant

import "sipon-be/internal/shared/kernel"

const (
	CodeTipeSuratNotFound          kernel.Code = "TIPE_SURAT_NOT_FOUND"
	CodeTipeSuratPersistenceFailed kernel.Code = "TIPE_SURAT_PERSISTENCE_FAILED"
	CodeTipeSuratQueryFailed       kernel.Code = "TIPE_SURAT_QUERY_FAILED"
	CodeTipeSuratKodeDuplicate     kernel.Code = "TIPE_SURAT_KODE_DUPLICATE"
	CodeTipeSuratInUse             kernel.Code = "TIPE_SURAT_IN_USE"
)
