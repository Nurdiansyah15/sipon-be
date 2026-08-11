package constant

import "sipon-be/internal/shared/kernel"

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusAbsent  AttendanceStatus = "absent"
	AttendanceStatusExcused AttendanceStatus = "excused"
)

const (
	CodeAttendanceNotFound          kernel.Code = "ATTENDANCE_NOT_FOUND"
	CodeAttendanceDuplicate         kernel.Code = "ATTENDANCE_DUPLICATE"
	CodeAttendanceInvalidStatus     kernel.Code = "ATTENDANCE_INVALID_STATUS"
	CodeAttendancePersistenceFailed kernel.Code = "ATTENDANCE_PERSISTENCE_FAILED"
	CodeAttendanceQueryFailed       kernel.Code = "ATTENDANCE_QUERY_FAILED"
)
