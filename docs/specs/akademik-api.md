# API Spec: Akademik

Base path: `/api/v1/web/akademik`. Semua endpoint butuh JWT + permission `manage_akademik`.

---

## Program

### List Program
```
GET /programs
```
Query params: `status`, `search`, `page`, `limit`
Response: paginasi `ProgramResponse` + meta

### Detail Program
```
GET /programs/:id
```
Response: `ProgramResponse` atau 404

### Buat Program
```
POST /programs
{
  "code": "TAHFIDZ",
  "name": "Tahfidz"
}
```
Response: 201 + `ProgramResponse`
Error: 409 jika `code` sudah ada

### Update Program
```
PUT /programs/:id
{
  "code": "TAHFIDZ",
  "name": "Tahfidz Al-Quran",
  "status": "active"
}
```
Response: 200 + `ProgramResponse`
Error: 409 jika `code` sudah dipakai activity_period_programs, 404 jika tidak ada

---

## Academic Period

### List Academic Period
```
GET /periods
```
Query params: `status`, `search`, `page`, `limit`
Response: paginasi `AcademicPeriodResponse` + meta

### Detail Academic Period
```
GET /periods/:id
```
Response: `AcademicPeriodResponse` atau 404

### Buat Academic Period
```
POST /periods
{
  "code": "2026/2027-P1",
  "name": "Periode 1 2026/2027",
  "start_date": "2026-07-01",
  "end_date": "2026-12-31"
}
```
Response: 201 + `AcademicPeriodResponse`
Error: 409 jika `code` sudah ada, 422 jika `end_date < start_date`

### Update Academic Period
```
PUT /periods/:id
{
  "code": "2026/2027-P1",
  "name": "Periode 1 2026/2027",
  "start_date": "2026-07-01",
  "end_date": "2026-12-31"
}
```
Response: 200 + `AcademicPeriodResponse`
Error: 409 jika `code` sudah ada, 404 jika tidak ada, 422 jika date invalid

### Open Period
```
POST /periods/:id/open
```
Response: 200 + `AcademicPeriodResponse`
Error: 422 jika status bukan `draft`

### Close Period
```
POST /periods/:id/close
```
Response: 200 + `AcademicPeriodResponse`
Error: 422 jika status bukan `open`

---

## Santri Registration

### List Registration
```
GET /registrations
```
Query params: `academic_period_id`, `santri_id`, `status`, `page`, `limit`
Response: paginasi `SantriRegistrationResponse` + meta

### Detail Registration
```
GET /registrations/:id
```
Response: `SantriRegistrationResponse` atau 404

### Register Santri
```
POST /registrations
{
  "santri_id": "uuid",
  "academic_period_id": "uuid"
}
```
Response: 201 + `SantriRegistrationResponse`
Error: 409 jika sudah ada registration untuk santri+period yang sama,
       404 jika santri/period tidak ada,
       422 jika period bukan `open` atau santri bukan `active`

### Complete Registration
```
POST /registrations/:id/complete
```
Response: 200 + `SantriRegistrationResponse` (dengan `registered_at`)
Error: 422 jika status bukan `pending`

### Cancel Registration
```
POST /registrations/:id/cancel
```
Response: 200 + `SantriRegistrationResponse`
Error: 422 jika status bukan `pending`

---

## Activity

### List Activity
```
GET /activities
```
Query params: `status`, `search`, `page`, `limit`
Response: paginasi `ActivityResponse` + meta

### Detail Activity
```
GET /activities/:id
```
Response: `ActivityResponse` atau 404

### Buat Activity
```
POST /activities
{
  "code": "SHALAT_SUBUH",
  "name": "Shalat Subuh Berjamaah"
}
```
Response: 201 + `ActivityResponse`
Error: 409 jika `code` sudah ada

### Update Activity
```
PUT /activities/:id
{
  "code": "SHALAT_SUBUH",
  "name": "Shalat Subuh Berjamaah",
  "status": "active"
}
```
Response: 200 + `ActivityResponse`
Error: 404 jika tidak ada

---

## Activity Period

### List Activity Period
```
GET /activity-periods
```
Query params: `academic_period_id`, `activity_id`, `status`, `page`, `limit`
Response: paginasi `ActivityPeriodResponse` + meta

### Activate Activity Period
```
POST /activity-periods
{
  "activity_id": "uuid",
  "academic_period_id": "uuid"
}
```
Response: 201 + `ActivityPeriodResponse`
Error: 409 jika sudah ada activity_period untuk activity+period yang sama,
       404 jika activity/period tidak ada

### Activate
```
POST /activity-periods/:id/activate
```
Response: 200 + `ActivityPeriodResponse`
Error: 422 jika sudah `active`

### Deactivate
```
POST /activity-periods/:id/deactivate
```
Response: 200 + `ActivityPeriodResponse`
Error: 422 jika sudah `inactive`

---

## Activity Period Program

### List Programs
```
GET /activity-periods/:id/programs
```
Response: array of `ProgramResponse`
Error: 404 jika activity_period tidak ada

**Catatan:** Jika response array kosong, activity berlaku untuk **semua program**.

### Assign Program
```
POST /activity-periods/:id/programs
{
  "program_id": "uuid"
}
```
Response: 200 + `ActivityPeriodProgramResponse`
Error: 409 jika sudah ada, 404 jika activity_period/program tidak ada

### Remove Program
```
DELETE /activity-periods/:id/programs/:programId
```
Response: 200
Error: 404 jika tidak ada

---

## Activity Schedule

### List Schedules per Activity Period
```
GET /activity-periods/:id/schedules
```
Response: array of `ActivityScheduleResponse`
Error: 404 jika activity_period tidak ada

### Detail Schedule
```
GET /schedules/:id
```
Response: `ActivityScheduleDetailResponse` (include recurrence rules)
Error: 404 jika tidak ada

### Buat Schedule
```
POST /schedules
{
  "activity_period_id": "uuid",
  "type": "weekly",
  "start_date": "2026-07-01",
  "end_date": "2026-12-31",
  "start_time": "19:30:00",
  "end_time": "21:00:00",
  "weekly_days": ["monday", "thursday"]
}
```
Response: 201 + `ActivityScheduleDetailResponse`
Error: 404 jika activity_period tidak ada,
       422 jika type tidak sesuai atau data recurrence tidak valid

**Catatan:** Field recurrence tergantung type:
- `once`: tidak butuh field recurrence
- `daily`: tidak butuh field recurrence
- `weekly`: `weekly_days: ["monday", ...]`
- `monthly`: `monthly_days: [5, 20]`
- `yearly`: `yearly_dates: [{"month": 8, "day": 17}]`

### Update Schedule
```
PUT /schedules/:id
{
  "start_date": "2026-07-01",
  "end_date": "2026-12-31",
  "start_time": "19:30:00",
  "end_time": "21:00:00",
  "weekly_days": ["monday", "wednesday", "friday"]
}
```
Response: 200 + `ActivityScheduleDetailResponse`
Error: 404 jika tidak ada, 422 jika data invalid

### Delete Schedule
```
DELETE /schedules/:id
```
Response: 200
Error: 404 jika tidak ada

---

## Activity Session

### List Sessions
```
GET /sessions
```
Query params: `activity_schedule_id`, `academic_period_id`, `status`,
              `start_date`, `end_date`, `page`, `limit`
Response: paginasi `ActivitySessionResponse` + meta

### Detail Session
```
GET /sessions/:id
```
Response: `ActivitySessionDetailResponse`
Error: 404 jika tidak ada

### Buat Session
```
POST /sessions
{
  "activity_schedule_id": "uuid",
  "starts_at": "2026-08-10T19:30:00Z",
  "ends_at": "2026-08-10T21:00:00Z"
}
```
Response: 201 + `ActivitySessionResponse`
Error: 404 jika schedule tidak ada, 422 jika `ends_at <= starts_at`

### Cancel Session
```
POST /sessions/:id/cancel
```
Response: 200 + `ActivitySessionResponse`
Error: 422 jika status `completed` atau `cancelled`

### Complete Session
```
POST /sessions/:id/complete
```
Response: 200 + `ActivitySessionResponse`
Error: 422 jika status bukan `scheduled` atau `open`

---

## Attendance

### List Attendance per Session
```
GET /sessions/:id/attendance
```
Response: array of `AttendanceResponse`
Error: 404 jika session tidak ada

### Record Attendance (Batch)
```
POST /sessions/:id/attendance
{
  "records": [
    {"santri_id": "uuid1", "status": "present"},
    {"santri_id": "uuid2", "status": "absent"},
    {"santri_id": "uuid3", "status": "excused"}
  ]
}
```
Response: 201 + array of `AttendanceResponse`
Error: 409 jika sudah ada attendance untuk santri+session yang sama,
       404 jika session tidak ada atau status `cancelled`,
       422 jika santri tidak valid

### Update Attendance
```
PUT /sessions/:id/attendance/:santriId
{
  "status": "present"
}
```
Response: 200 + `AttendanceResponse`
Error: 404 jika tidak ada, 422 jika status invalid

---

## Response Shapes

### ProgramResponse
```json
{
  "id": "uuid",
  "code": "TAHFIDZ",
  "name": "Tahfidz",
  "status": "active",
  "created_at": "2026-07-01T00:00:00Z",
  "updated_at": "2026-07-01T00:00:00Z"
}
```

### AcademicPeriodResponse
```json
{
  "id": "uuid",
  "code": "2026/2027-P1",
  "name": "Periode 1 2026/2027",
  "start_date": "2026-07-01",
  "end_date": "2026-12-31",
  "status": "open",
  "created_at": "2026-07-01T00:00:00Z",
  "updated_at": "2026-07-01T00:00:00Z"
}
```

### SantriRegistrationResponse
```json
{
  "id": "uuid",
  "santri_id": "uuid",
  "santri_name": "Ahmad",
  "santri_nis": "1000126001",
  "academic_period_id": "uuid",
  "period_name": "Periode 1 2026/2027",
  "status": "completed",
  "registered_at": "2026-07-15T10:00:00Z",
  "created_at": "2026-07-15T09:00:00Z",
  "updated_at": "2026-07-15T10:00:00Z"
}
```

### ActivityResponse
```json
{
  "id": "uuid",
  "code": "SHALAT_SUBUH",
  "name": "Shalat Subuh Berjamaah",
  "status": "active",
  "created_at": "2026-07-01T00:00:00Z",
  "updated_at": "2026-07-01T00:00:00Z"
}
```

### ActivityPeriodResponse
```json
{
  "id": "uuid",
  "activity_id": "uuid",
  "activity_name": "Shalat Subuh Berjamaah",
  "activity_code": "SHALAT_SUBUH",
  "academic_period_id": "uuid",
  "period_name": "Periode 1 2026/2027",
  "status": "active",
  "created_at": "2026-07-01T00:00:00Z",
  "updated_at": "2026-07-01T00:00:00Z"
}
```

### ActivityPeriodProgramResponse
```json
{
  "id": "uuid",
  "activity_period_id": "uuid",
  "program_id": "uuid",
  "program_code": "TAHFIDZ",
  "program_name": "Tahfidz"
}
```

### ActivityScheduleResponse
```json
{
  "id": "uuid",
  "activity_period_id": "uuid",
  "type": "weekly",
  "start_date": "2026-07-01",
  "end_date": "2026-12-31",
  "start_time": "19:30:00",
  "end_time": "21:00:00",
  "created_at": "2026-07-01T00:00:00Z",
  "updated_at": "2026-07-01T00:00:00Z"
}
```

### ActivityScheduleDetailResponse
```json
{
  "id": "uuid",
  "activity_period_id": "uuid",
  "activity_name": "Kajian",
  "type": "weekly",
  "start_date": "2026-07-01",
  "end_date": "2026-12-31",
  "start_time": "19:30:00",
  "end_time": "21:00:00",
  "weekly_days": ["monday", "thursday"],
  "created_at": "2026-07-01T00:00:00Z",
  "updated_at": "2026-07-01T00:00:00Z"
}
```
**Catatan:** Field recurrence hanya muncul sesuai type:
- `weekly`: `weekly_days: ["monday", ...]`
- `monthly`: `monthly_days: [5, 20]`
- `yearly`: `yearly_dates: [{"month": 8, "day": 17}]`

### ActivitySessionResponse
```json
{
  "id": "uuid",
  "activity_schedule_id": "uuid",
  "activity_name": "Kajian",
  "starts_at": "2026-08-10T19:30:00Z",
  "ends_at": "2026-08-10T21:00:00Z",
  "status": "scheduled",
  "created_at": "2026-08-01T00:00:00Z",
  "updated_at": "2026-08-01T00:00:00Z"
}
```

### ActivitySessionDetailResponse
```json
{
  "id": "uuid",
  "activity_schedule_id": "uuid",
  "activity_name": "Kajian",
  "activity_code": "KAJIAN",
  "schedule_type": "weekly",
  "starts_at": "2026-08-10T19:30:00Z",
  "ends_at": "2026-08-10T21:00:00Z",
  "status": "completed",
  "attendance_summary": {
    "total": 30,
    "present": 25,
    "absent": 3,
    "excused": 2
  },
  "created_at": "2026-08-01T00:00:00Z",
  "updated_at": "2026-08-10T21:00:00Z"
}
```

### AttendanceResponse
```json
{
  "id": "uuid",
  "activity_session_id": "uuid",
  "santri_id": "uuid",
  "santri_name": "Ahmad",
  "santri_nis": "1000126001",
  "status": "present",
  "recorded_at": "2026-08-10T19:35:00Z",
  "created_at": "2026-08-10T19:35:00Z",
  "updated_at": "2026-08-10T19:35:00Z"
}
```
