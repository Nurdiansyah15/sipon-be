package command

// Routing key untuk event yang dipublish ke outbox, dikonsumsi oleh modul
// notification untuk mengirim notifikasi in-app/push ke pendaftar.
const (
	RoutingPendaftaranSubmitted         = "psb.pendaftaran.submitted"
	RoutingDaftarUlangSubmitted         = "psb.daftar_ulang.submitted"
	RoutingDokumenVerified              = "psb.dokumen.verified"
	RoutingDokumenRejected              = "psb.dokumen.rejected"
	RoutingRevisionRequested            = "psb.pendaftaran.revision_requested"
	RoutingRevisionRequestedDaftarUlang = "psb.daftar_ulang.revision_requested"
	RoutingPendaftaranAccepted          = "psb.pendaftaran.accepted"
	RoutingPendaftaranRejected          = "psb.pendaftaran.rejected"
	RoutingNISGenerated                 = "psb.pendaftaran.nis_generated"
)
