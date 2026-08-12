package entity

import (
	"time"

	"sipon-be/internal/modules/fingerprint/domain/scanlog/constant"
	"sipon-be/internal/shared/kernel"
)

// ScanLog merepresentasikan satu baris scan mentah yang dikirim mesin
// fingerprint ke database (skema identik dengan yang dikirim hardware).
type ScanLog struct {
	ID         string
	SN         string
	ScanDate   time.Time
	PIN        string
	VerifyMode int
	InOutMode  int
	DeviceIP   string
	CreatedAt  time.Time
}

func NewScanLog(id, sn, pin, deviceIP string, scanDate time.Time, verifyMode, inOutMode int) (*ScanLog, error) {
	if id == "" || sn == "" || pin == "" {
		return nil, kernel.New(constant.CodeScanLogInvalid)
	}
	return &ScanLog{
		ID:         id,
		SN:         sn,
		ScanDate:   scanDate,
		PIN:        pin,
		VerifyMode: verifyMode,
		InOutMode:  inOutMode,
		DeviceIP:   deviceIP,
		CreatedAt:  time.Now(),
	}, nil
}
