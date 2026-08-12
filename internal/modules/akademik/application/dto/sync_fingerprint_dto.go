package dto

// SyncFingerprintError adalah scan yang gagal dicatat, dengan alasan error.
type SyncFingerprintError struct {
	PIN    string `json:"pin"`
	Reason string `json:"reason"`
}

// SyncFingerprintResponse adalah ringkasan sinkronisasi absensi dari scan
// mesin fingerprint. Scan yang NIS-nya sudah tercatat hadir dihitung sebagai
// skipped (idempotent), bukan error.
type SyncFingerprintResponse struct {
	TotalScans int                    `json:"total_scans"`
	Recorded   int                    `json:"recorded"`
	Skipped    int                    `json:"skipped"`
	Errors     []SyncFingerprintError `json:"errors"`
}
