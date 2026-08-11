# Plan UI: Implementasi Module Akademik di `sipon-ui`

## Context

Backend module `akademik` sudah lengkap dengan 12 tabel dan ~30 endpoint.
Domain utama: Program, Santri, Academic Period, Herregistrasi, Activity,
Activity Schedule, Activity Session, Attendance.

Module ini masuk ke **portal admin** dengan layout sendiri (seperti module keuangan),
karena domain sangat spesifik dan membutuhkan banyak navigasi. **Tidak ada**
self-service portal santri di fase ini.

---

## Temuan penting dari kode aktual

1. **Single permission**: `manage_akademik` — semua endpoint admin-only.
2. **Status lifecycle**: Period (`draft → open → closed → archived`), Registration
   (`pending → completed/cancelled`), Session (`scheduled → open → completed/cancelled`),
   ActivityPeriod (`active ↔ inactive`).
3. **Schedule recurrence**: 5 tipe (`once`, `daily`, `weekly`, `monthly`, `yearly`)
   — UI butuh dynamic form berdasarkan tipe.
4. **Cross-module**: `santri_id` adalah referensi ke kesantrian — UI perlu tampilkan
   NIS (bukan nama, karena contract kesantrian tidak expose nama).
5. **Batch attendance**: Record absensi untuk banyak santri sekaligus per session.
6. **Enriched responses**: Beberapa endpoint mengembalikan data join (period_name,
   activity_name, program_code) — UI tidak perlu fetch terpisah.
7. **Schedule detail**: GET `/schedules/:id` mengembalikan recurrence rules
   (`weekly_days`, `monthly_days`, `yearly_dates`) — UI perlu parse dinamis.
8. **Session filter**: List sessions bisa filter by `academic_period_id`,
   `activity_schedule_id`, `status`, `start_date`, `end_date`.
9. **Default behavior**: Activity tanpa `ActivityPeriodProgram` berarti berlaku
   untuk semua program — UI perlu tampilkan badge "Semua Program" jika list kosong.
10. **Time format**: Schedule `start_time`/`end_time` adalah TIME (HH:MM:SS),
    session `starts_at`/`ends_at` adalah TIMESTAMPTZ.

---

## Pola yang sudah ada di `sipon-ui` dan harus di-reuse

- **API layer**: `useApi()`, `parseApiError()`, envelope types.
- **Store pattern**: Pinia store per domain (`app/stores/*.ts`).
- **List/table pattern**: `UTable` + `UPagination` + `usePermission()` inline `v-if="can('...')"`.
- **Modal pattern**: `ConfirmActionModal.vue` untuk destructive actions.
- **Date picker**: Nuxt UI v4 `UInput type="date"` untuk period dates, session dates.
- **Time picker**: Nuxt UI v4 `UInput type="time"` untuk schedule times.
- **Status badge**: `UBadge` dengan color mapping per status.
- **Pagination**: `useApi()` helper + `UPagination` component.

---

## Struktur baru

### Types — `shared/types/Akademik.ts`

```typescript
// Enums
export type ProgramStatus = 'active' | 'inactive'
export type AcademicPeriodStatus = 'draft' | 'open' | 'closed' | 'archived'
export type SantriRegistrationStatus = 'pending' | 'completed' | 'cancelled'
export type ActivityStatus = 'active' | 'inactive'
export type ActivityPeriodStatus = 'active' | 'inactive'
export type ActivityScheduleType = 'once' | 'daily' | 'weekly' | 'monthly' | 'yearly'
export type DayOfWeek = 'monday' | 'tuesday' | 'wednesday' | 'thursday' | 'friday' | 'saturday' | 'sunday'
export type ActivitySessionStatus = 'scheduled' | 'open' | 'completed' | 'cancelled'
export type AttendanceStatus = 'present' | 'absent' | 'excused'

// Entities
export interface Program {
  id: string
  code: string
  name: string
  status: ProgramStatus
  created_at: string
  updated_at: string
}

export interface AcademicPeriod {
  id: string
  code: string
  name: string
  start_date: string
  end_date: string
  status: AcademicPeriodStatus
  created_at: string
  updated_at: string
}

export interface SantriRegistration {
  id: string
  santri_id: string
  santri_nis?: string
  academic_period_id: string
  period_name?: string
  status: SantriRegistrationStatus
  registered_at: string | null
  created_at: string
  updated_at: string
}

export interface Activity {
  id: string
  code: string
  name: string
  status: ActivityStatus
  created_at: string
  updated_at: string
}

export interface ActivityPeriod {
  id: string
  activity_id: string
  activity_code?: string
  activity_name?: string
  academic_period_id: string
  period_name?: string
  status: ActivityPeriodStatus
  created_at: string
  updated_at: string
}

export interface ActivityPeriodProgram {
  id: string
  activity_period_id: string
  program_id: string
  program_code?: string
  program_name?: string
}

export interface YearlyDate {
  month: number
  day: number
}

export interface ActivitySchedule {
  id: string
  activity_period_id: string
  activity_name?: string
  activity_code?: string
  type: ActivityScheduleType
  start_date?: string
  end_date?: string
  start_time: string
  end_time: string
  weekly_days?: DayOfWeek[]
  monthly_days?: number[]
  yearly_dates?: YearlyDate[]
  created_at: string
  updated_at: string
}

export interface AttendanceSummary {
  total: number
  present: number
  absent: number
  excused: number
}

export interface ActivitySession {
  id: string
  activity_schedule_id: string
  activity_name?: string
  activity_code?: string
  schedule_type?: ActivityScheduleType
  starts_at: string
  ends_at: string
  status: ActivitySessionStatus
  attendance_summary?: AttendanceSummary
  created_at: string
  updated_at: string
}

export interface Attendance {
  id: string
  activity_session_id: string
  santri_id: string
  santri_nis?: string
  status: AttendanceStatus
  recorded_at: string
  created_at: string
  updated_at: string
}

// Request DTOs
export interface CreateProgramRequest {
  code: string
  name: string
}

export interface UpdateProgramRequest {
  code?: string
  name?: string
  status?: ProgramStatus
}

export interface CreateAcademicPeriodRequest {
  code: string
  name: string
  start_date: string
  end_date: string
}

export interface UpdateAcademicPeriodRequest {
  code?: string
  name?: string
  start_date?: string
  end_date?: string
}

export interface CreateSantriRegistrationRequest {
  santri_id: string
  academic_period_id: string
}

export interface CreateActivityRequest {
  code: string
  name: string
}

export interface UpdateActivityRequest {
  code?: string
  name?: string
  status?: ActivityStatus
}

export interface CreateActivityPeriodRequest {
  activity_id: string
  academic_period_id: string
}

export interface AssignProgramRequest {
  program_id: string
}

export interface CreateScheduleRequest {
  activity_period_id: string
  type: ActivityScheduleType
  start_date?: string
  end_date?: string
  start_time: string
  end_time: string
  weekly_days?: DayOfWeek[]
  monthly_days?: number[]
  yearly_dates?: YearlyDate[]
}

export interface UpdateScheduleRequest {
  start_date?: string
  end_date?: string
  start_time?: string
  end_time?: string
  weekly_days?: DayOfWeek[]
  monthly_days?: number[]
  yearly_dates?: YearlyDate[]
}

export interface CreateSessionRequest {
  activity_schedule_id: string
  starts_at: string
  ends_at: string
}

export interface AttendanceRecordInput {
  santri_id: string
  status: AttendanceStatus
}

export interface RecordAttendanceRequest {
  records: AttendanceRecordInput[]
}

export interface UpdateAttendanceRequest {
  status: AttendanceStatus
}

// Query params
export interface ProgramListQuery {
  status?: ProgramStatus
  search?: string
  page?: number
  limit?: number
}

export interface AcademicPeriodListQuery {
  status?: AcademicPeriodStatus
  search?: string
  page?: number
  limit?: number
}

export interface SantriRegistrationListQuery {
  academic_period_id?: string
  santri_id?: string
  status?: SantriRegistrationStatus
  page?: number
  limit?: number
}

export interface ActivityListQuery {
  status?: ActivityStatus
  search?: string
  page?: number
  limit?: number
}

export interface ActivityPeriodListQuery {
  activity_id?: string
  academic_period_id?: string
  status?: ActivityPeriodStatus
  page?: number
  limit?: number
}

export interface ActivitySessionListQuery {
  activity_schedule_id?: string
  academic_period_id?: string
  status?: ActivitySessionStatus
  start_date?: string
  end_date?: string
  page?: number
  limit?: number
}
```

### Stores

#### `app/stores/akademik.ts`
- **State**: `programs[]`, `periods[]`, `registrations[]`, `activities[]`,
  `activityPeriods[]`, `schedules[]`, `sessions[]`, `attendances[]`,
  `meta`, `isLoading`, `isSubmitting`, `error`
- **Actions**:
  - `fetchPrograms(query)` / `createProgram` / `updateProgram`
  - `fetchPeriods(query)` / `createPeriod` / `updatePeriod` / `openPeriod` / `closePeriod`
  - `fetchRegistrations(query)` / `createRegistration` / `completeRegistration` / `cancelRegistration`
  - `fetchActivities(query)` / `createActivity` / `updateActivity`
  - `fetchActivityPeriods(query)` / `createActivityPeriod` / `activatePeriod` / `deactivatePeriod`
  - `fetchPeriodPrograms(periodId)` / `assignProgram` / `removeProgram`
  - `fetchSchedules(periodId)` / `getSchedule(id)` / `createSchedule` / `updateSchedule` / `deleteSchedule`
  - `fetchSessions(query)` / `getSession(id)` / `createSession` / `cancelSession` / `completeSession`
  - `fetchAttendance(sessionId)` / `recordAttendance(sessionId, records)` / `updateAttendance(sessionId, santriId, status)`

### Routing

**Admin layout** (`app/layouts/akademik.vue`):
- Sidebar navigation dengan menu:
  - Dashboard
  - Master
    - Program
    - Periode Akademik
    - Kegiatan
  - Operasional
    - Herregistrasi
    - Aktivasi Kegiatan
    - Jadwal
  - Pelaksanaan
    - Sesi

**Admin pages** (`app/pages/admin/akademik/`):

```
admin/akademik/
├── index.vue                          # Dashboard dengan summary cards
├── program/
│   └── index.vue                      # CRUD program (list + create/update modal)
├── periode/
│   └── index.vue                      # CRUD period + lifecycle (open/close)
├── kegiatan/
│   └── index.vue                      # CRUD activity (list + create/update modal)
├── herregistrasi/
│   └── index.vue                      # List registrations + create/complete/cancel
├── aktivasi/
│   ├── index.vue                      # List activity-periods
│   └── [id].vue                       # Detail: info + program scope + schedules
├── jadwal/
│   ├── index.vue                      # List all schedules (filter by period)
│   └── [id].vue                       # Detail: schedule + recurrence rules
└── sesi/
    ├── index.vue                      # List sessions (filter by period/schedule/status)
    └── [id].vue                       # Detail: session info + attendance list + record
```

### Komponen

#### Shared (`app/components/akademik/*.vue`):
- **`AkademikStatusBadge.vue`** — badge untuk period/registration/activity/session/attendance status
- **`AkademikTimeDisplay.vue`** — formatted time range (start_time - end_time)
- **`AkademikScheduleTypeBadge.vue`** — badge untuk schedule type (once/daily/weekly/monthly/yearly)
- **`AkademikDayOfWeekPicker.vue`** — multi-select untuk weekly days (checkbox grid)
- **`AkademikRecurrenceEditor.vue`** — dynamic form berdasarkan schedule type:
  - `weekly`: DayOfWeekPicker
  - `monthly`: array input day_of_month (1-31)
  - `yearly`: array input month+day

#### Admin (`app/components/admin/akademik/*.vue`):
- **`AdminProgramFormModal.vue`** — form create/edit program
- **`AdminPeriodFormModal.vue`** — form create/edit period + date range picker
- **`AdminActivityFormModal.vue`** — form create/edit activity
- **`AdminRegistrationFormModal.vue`** — form create registration (santri picker + period picker)
- **`AdminActivityPeriodFormModal.vue`** — form create activity-period (activity picker + period picker)
- **`AdminProgramScopeManager.vue`** — manage program scope (assign/remove programs)
- **`AdminScheduleFormModal.vue`** — form create/edit schedule + RecurrenceEditor
- **`AdminSessionFormModal.vue`** — form create session (schedule picker + datetime)
- **`AdminAttendanceRecorder.vue`** — form batch record attendance (santri list + status dropdown)
- **`AdminAttendanceList.vue`** — list attendance dengan inline update status

---

## Matriks visibility per status

### Academic Period Status

| Status      | Actions tersedia                                    |
|-------------|-----------------------------------------------------|
| `draft`     | Edit, Open                                          |
| `open`      | -                                                   |
| `closed`    | -                                                   |
| `archived`  | -                                                   |

### Santri Registration Status

| Status      | Actions tersedia                                    |
|-------------|-----------------------------------------------------|
| `pending`   | Complete, Cancel                                    |
| `completed` | -                                                   |
| `cancelled` | -                                                   |

### Activity Session Status

| Status       | Actions tersedia                                   |
|--------------|----------------------------------------------------|
| `scheduled`  | Open (implicit saat record attendance), Cancel, Complete |
| `open`       | Cancel, Complete                                   |
| `completed`  | -                                                  |
| `cancelled`  | -                                                  |

### Attendance Status

| Status      | Actions tersedia                                    |
|-------------|-----------------------------------------------------|
| `present`   | Update (ke absent/excused)                          |
| `absent`    | Update (ke present/excused)                         |
| `excused`   | Update (ke present/absent)                          |

---

## Halaman Detail

### Aktivasi Kegiatan (`/admin/akademik/aktivasi/[id]`)
Menampilkan:
- Info activity-period (activity name, period name, status)
- Tombol activate/deactivate
- **Program Scope**: list program yang berlaku (atau "Semua Program" jika kosong)
  - Tombol assign program (modal picker)
  - Tombol remove per program
- **Jadwal**: list schedules untuk activity-period ini
  - Tombol create schedule
  - Tombol edit/delete per schedule

### Sesi Kegiatan (`/admin/akademik/sesi/[id]`)
Menampilkan:
- Info session (activity name, schedule type, starts_at, ends_at, status)
- Tombol complete/cancel
- **Attendance Summary** (present/absent/excused counts)
- **Attendance List**: tabel santri dengan status, bisa inline update
- **Record Attendance** button (modal dengan santri list + status)

---

## Fase Pengerjaan

**Fase 1 — Types.** `shared/types/Akademik.ts`. Checkpoint: type-check lolos.

**Fase 2 — Store.** `app/stores/akademik.ts`. Checkpoint: compile lolos.

**Fase 3 — Layout + Navigation.** `app/layouts/akademik.vue` + `app/components/AppAkademikNavbar.vue` + `app/composables/useAkademikSidebar.ts`. Checkpoint: layout render, menu navigasi bekerja.

**Fase 4 — Shared components.** `AkademikStatusBadge.vue`, `AkademikTimeDisplay.vue`, `AkademikScheduleTypeBadge.vue`, `AkademikDayOfWeekPicker.vue`, `AkademikRecurrenceEditor.vue`. Checkpoint: render bersih.

**Fase 5 — Admin pages: Master.** `program/index.vue`, `periode/index.vue`, `kegiatan/index.vue` + modal components. Checkpoint: CRUD program, period (dengan lifecycle), activity.

**Fase 6 — Admin pages: Operasional.** `herregistrasi/index.vue`, `aktivasi/index.vue`, `aktivasi/[id].vue`, `jadwal/index.vue`, `jadwal/[id].vue` + modal components. Checkpoint: CRUD registration, activity-period, schedule (dengan recurrence).

**Fase 7 — Admin pages: Pelaksanaan.** `sesi/index.vue`, `sesi/[id].vue` + modal components. Checkpoint: CRUD session, record/update attendance.

**Fase 8 — Dashboard + Polish.** Dashboard index.vue dengan summary cards, permission gating, edge cases, empty states. Checkpoint: end-to-end flow dari setup program → create period → activate activity → create schedule → create session → record attendance.

---

## Verifikasi

1. `npm run dev` jalan tanpa error di tiap checkpoint fase.
2. Uji manual browser: seluruh alur admin (Fase 5-7 checkpoint) end-to-end lawan `sipon-be` yang jalan lokal.
3. Uji lifecycle: period draft → open, registration pending → completed, session scheduled → completed.
4. Uji schedule recurrence: create weekly/monthly/yearly schedule, verify detail menampilkan recurrence rules.
5. Uji batch attendance: record attendance untuk multiple santri, verify summary update.
6. Uji status badge: verify warna badge sesuai status (active=green, inactive=gray, pending=amber, completed=green, cancelled=red, dll).
7. Uji empty state: list kosong menampilkan pesan "Belum ada data" + tombol create (jika ada permission).
8. Uji responsive: sidebar collapse/expand, mobile drawer.

---

## Catatan tambahan

- **Layout akademik** terpisah dari layout admin umum (seperti keuangan) karena navigasi sangat spesifik (banyak sub-menu untuk schedule, session, attendance).
- **Recurrence editor** perlu komponen custom untuk weekly/monthly/yearly input — belum ada di Nuxt UI v4.
- **Santri picker**: untuk registration dan attendance, perlu dropdown/table picker santri dari kesantrian. Bisa pakai endpoint `/api/v1/web/santri/admin` (sudah ada di module kesantrian) atau create manual santri input.
- **Activity picker**: untuk activity-period, perlu dropdown activity dari `/admin/akademik/kegiatan`.
- **Period picker**: untuk registration dan activity-period, perlu dropdown period dari `/admin/akademik/periode`.
- **Time input**: gunakan `UInput type="time"` dari Nuxt UI untuk schedule start_time/end_time.
- **Date range**: untuk period start_date/end_date, gunakan `UInput type="date"`.
- **Batch attendance**: UI perlu tampilkan list santri aktif (dari kesantrian), lalu admin pilih status per santri (present/absent/excused) atau pilih semua present lalu override个别.
- **Attendance inline update**: di halaman session detail, attendance list bisa inline edit status (klik row → toggle status).
- **Default behavior**: jika activity-period tidak punya program scope, tampilkan badge "Semua Program" (tidak perlu tampilkan list kosong).
- **Session status**: saat record attendance pertama kali, session otomatis jadi "open" (implicit transition) — UI tidak perlu tampilkan tombol "Open", cukup tampilkan status "open" setelah ada attendance.
