package constant

import "sipon-be/internal/shared/kernel"

type PendaftarStatus string

const (
	StatusDraft                   PendaftarStatus = "draft"
	StatusDiajukan               PendaftarStatus = "diajukan"
	StatusPerluRevisi            PendaftarStatus = "perlu_revisi"
	StatusDitolak                PendaftarStatus = "ditolak"
	StatusDiterima               PendaftarStatus = "diterima"
	StatusMengundurkanDiri       PendaftarStatus = "mengundurkan_diri"
	StatusDaftarUlang            PendaftarStatus = "daftar_ulang"
	StatusPerluRevisiDaftarUlang PendaftarStatus = "perlu_revisi_daftar_ulang"
	StatusSelesai                PendaftarStatus = "selesai"
)

const (
	CodePendaftarNotFound          kernel.Code = "PENDAFTAR_NOT_FOUND"
	CodePendaftarPersistenceFailed kernel.Code = "PENDAFTAR_PERSISTENCE_FAILED"
	CodePendaftarQueryFailed       kernel.Code = "PENDAFTAR_QUERY_FAILED"
	CodePendaftarDuplicate         kernel.Code = "PENDAFTAR_DUPLICATE"
	CodePendaftarInvalidStatus     kernel.Code = "PENDAFTAR_INVALID_STATUS"
)
