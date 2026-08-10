package constant

import "sipon-be/internal/shared/kernel"

const OrgCode = "ORG"

const (
	CodeSuratNotFound          kernel.Code = "SURAT_NOT_FOUND"
	CodeSuratPersistenceFailed kernel.Code = "SURAT_PERSISTENCE_FAILED"
	CodeSuratQueryFailed       kernel.Code = "SURAT_QUERY_FAILED"
	CodeSuratNomorFailed       kernel.Code = "SURAT_NOMOR_FAILED"
	CodeSuratDokumenExists     kernel.Code = "SURAT_DOKUMEN_EXISTS"
	CodeSuratDokumenNotFound   kernel.Code = "SURAT_DOKUMEN_NOT_FOUND"
)
