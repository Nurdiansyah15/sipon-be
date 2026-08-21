package command

// Routing key untuk event akademik yang dipublish ke outbox, dikonsumsi oleh
// modul notification untuk mengirim notifikasi in-app/push.
const (
	RoutingAttendanceRecorded = "akademik.attendance.recorded"
)

// Sumber pencatatan absensi — dipakai untuk membedakan notifikasi dari
// check-in manual via NIS dan sinkronisasi fingerprint.
const (
	AttendanceSourceNIS         = "nis"
	AttendanceSourceFingerprint = "fingerprint"
)

// attendanceRecordedPayload adalah payload event akademik.attendance.recorded
// yang dikonsumsi modul notification.
type attendanceRecordedPayload struct {
	UserID       string `json:"user_id"`
	AttendanceID string `json:"attendance_id"`
	SantriID     string `json:"santri_id"`
	NIS          string `json:"nis"`
	Name         string `json:"name"`
	SessionID    string `json:"session_id"`
	Source       string `json:"source"`
}
