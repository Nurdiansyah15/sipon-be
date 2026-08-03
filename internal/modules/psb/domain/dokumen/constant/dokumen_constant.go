package constant

import "sipon-be/internal/shared/kernel"

type DokumenKind string

const (
	DokumenKindSuratPernyataan DokumenKind = "surat_pernyataan"
	DokumenKindKTP             DokumenKind = "ktp"
	DokumenKindKK              DokumenKind = "kk"
	DokumenKindMutasi          DokumenKind = "mutasi"
	DokumenKindPembayaran      DokumenKind = "pembayaran"
)

var ValidDokumenKinds = map[DokumenKind]bool{
	DokumenKindSuratPernyataan: true,
	DokumenKindKTP:             true,
	DokumenKindKK:              true,
	DokumenKindMutasi:          true,
	DokumenKindPembayaran:      true,
}

type DokumenStatus string

const (
	DokumenStatusPending  DokumenStatus = "pending"
	DokumenStatusVerified DokumenStatus = "verified"
	DokumenStatusRejected DokumenStatus = "rejected"
)

type DokumenStage string

const (
	StagePendaftaran DokumenStage = "pendaftaran"
	StageDaftarUlang DokumenStage = "daftar_ulang"
)

var AllowedContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"application/pdf": true,
}

const (
	CodeDokumenNotFound          kernel.Code = "DOKUMEN_NOT_FOUND"
	CodeDokumenPersistenceFailed kernel.Code = "DOKUMEN_PERSISTENCE_FAILED"
	CodeDokumenQueryFailed       kernel.Code = "DOKUMEN_QUERY_FAILED"
	CodeDokumenInvalidKind       kernel.Code = "DOKUMEN_INVALID_KIND"
	CodeDokumenInvalidStatus     kernel.Code = "DOKUMEN_INVALID_STATUS"
)
