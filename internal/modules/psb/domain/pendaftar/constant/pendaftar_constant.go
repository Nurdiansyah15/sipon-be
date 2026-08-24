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

// Gender pendaftar — nilai kolom "gender" (VARCHAR(2) CHECK ('1','2')).
const (
	GenderMale   = "1"
	GenderFemale = "2"
)

// ProgramCode adalah kode program dari master `programs` (module akademik)
// yang saat ini dikenal PSB. Nama jurusan & gender volatile, sehingga untuk
// sementara kombinasi yang berlaku dijaga sebagai konstanta di sini — domain
// quota (kuota per program × gender) memakai kosakata yang sama agar dua
// domain ini konsisten saat memvalidasi pendaftar.
const (
	ProgramCodeTahfidz = "TAHFIDZ"
	ProgramCodeKitab   = "KITAB"
)

// QuotaCountedStatuses adalah status pendaftar yang menghabiskan kuota
// (rekalkulasi sisa kuota). Hanya pendaftar yang sudah "diterima" — ditolak
// dan mengundurkan_diri melepas slot, sedangkan draft/diajukan/perlu_revisi
// belum mengklaim slot.
var QuotaCountedStatuses = []PendaftarStatus{
	StatusDiterima,
	StatusDaftarUlang,
	StatusPerluRevisiDaftarUlang,
	StatusSelesai,
}

const (
	CodePendaftarNotFound          kernel.Code = "PENDAFTAR_NOT_FOUND"
	CodePendaftarPersistenceFailed kernel.Code = "PENDAFTAR_PERSISTENCE_FAILED"
	CodePendaftarQueryFailed       kernel.Code = "PENDAFTAR_QUERY_FAILED"
	CodePendaftarDuplicate         kernel.Code = "PENDAFTAR_DUPLICATE"
	CodePendaftarInvalidStatus     kernel.Code = "PENDAFTAR_INVALID_STATUS"
)
