package dto

import "time"

// SimulateScanRequest adalah payload endpoint sandbox (generator mesin
// fingerprint palsu). Semua field opsional kecuali pin.
type SimulateScanRequest struct {
	SN         *string    `json:"sn"`
	PIN        string     `json:"pin" binding:"required"`
	ScanDate   *time.Time `json:"scan_date"`
	VerifyMode *int       `json:"verifymode"`
	InOutMode  *int       `json:"inoutmode"`
	DeviceIP   *string    `json:"deviceip"`
}

type ScanLogListQuery struct {
	From string `form:"from"`
	To   string `form:"to"`
}

// ScanLogResponse adalah satu scan mentah (skema identik dengan mesin).
type ScanLogResponse struct {
	ID         string    `json:"id"`
	SN         string    `json:"sn"`
	ScanDate   time.Time `json:"scan_date"`
	PIN        string    `json:"pin"`
	VerifyMode int       `json:"verifymode"`
	InOutMode  int       `json:"inoutmode"`
	DeviceIP   string    `json:"deviceip"`
	CreatedAt  time.Time `json:"created_at"`
}

// ScanPin adalah scan pertama per pin dalam rentang waktu.
type ScanPin struct {
	PIN         string    `json:"pin"`
	SN          string    `json:"sn"`
	FirstScanAt time.Time `json:"first_scan_at"`
}
