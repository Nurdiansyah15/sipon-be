# Plan: Penyempurnaan Sistem Absensi

## Context

Sistem absensi saat ini bekerja secara **batch** dari admin: admin membuka sesi, lalu
merekam absensi santri satu per satu / massal dari halaman admin.

Kebutuhan baru:
1. **Halaman presensi** — admin buka sesi dari halaman admin, lalu admin membuka
   halaman presensi khusus untuk sesi itu. Santri cukup input NIS → tekan Enter →
   otomatis tercatat sebagai **hadir**.
2. **Validasi** tetap sama seperti sekarang: santri terdaftar pada periode akademik
   sesi tersebut, herregistrasi completed, status santri aktif, belum tercatat
   duplikat.
3. **Portal santri** — santri bisa melihat **history absensi** mereka, difilter per
   periode akademik dan/atau per sesi kegiatan.

---

## Alur Sistem

```
Admin buka sesi (status: scheduled → open) dari halaman admin.
    │
    ├── Admin klik "Buka Presensi" → membuka /presensi/:sessionId
    │
    ▼
Halaman presensi menampilkan:
    • Info sesi (nama kegiatan, jadwal, waktu)
    • Counter kehadiran
    • Input NIS + tombol Enter
    • List santri yang sudah hadir
    │
    ▼
Santri input NIS → Enter:
    • Backend lookup santri by NIS
    • Validasi: terdaftar, herreg completed, belum absen, sesi masih open
    • Auto-record attendance sebagai "present"
    • Frontend refresh list & counter
    │
    ▼
Admin bisa:
    • Menutup sesi (status: open → completed) dari halaman admin
    • Melihat ringkasan absensi
    • Manual override (ubah status absensi santri tertentu)

Portal santri (/akademik/absensi):
    • Pilih periode akademik
    • Lihat list sesi + status absensi per sesi
    • Filter per kegiatan/sesi (opsional)
```

---

## Endpoint HTTP Baru

### Halpresensi (tanpa permission, tanpa JWT — akses via URL)

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/presensi/:sessionId` | Info sesi + ringkasan absensi (untuk tampilan halaman presensi) |
| POST | `/presensi/:sessionId/checkin` | Check-in santri via NIS → auto-record sebagai hadir |
| GET | `/presensi/:sessionId/attendance` | List santri yang sudah hadir (real-time refresh) |

**Catatan**: endpoint ini **tidak butuh JWT** — presensi diakses via URL dari admin
yang sudah share link. Validasi ada di level session status (harus `open`).

### Portal Santri (JWT, tanpa permission)

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/my/absensi` | History absensi santri (filter: period_id, activity_schedule_id) |

### Admin (existing + penambahan)

Endpoint admin existing (`/sessions/:id/open`, `/sessions/:id/attendance`, dll) tidak
berubah. Cukup tambah **link "Buka Presensi"** di halaman sesi detail yang mengarah ke
halaman `/presensi/:sessionId`.

---

## Domain & Repository Changes

### Kesantrian: Lookup by NIS

Saat ini `FindByNIS` ada di repo kesantrian tapi **belum diekspos** lewat kontrak.
Perlu ditambahkan:

**`kesantrian.Contract`** (tambah method):
```go
GetSantriByNIS(ctx context.Context, nis string) (*SantriBasicInfo, error)
```

**`akademik/application/ports.KesantrianReader`** (tambah):
```go
GetSantriByNIS(ctx context.Context, nis string) (*SantriBasicInfo, error)
```

**Gateway adapter**: implementasi delegasi ke `kesantrian.Contract.GetSantriByNIS`.

### Repository: Attendance history query

Tambah method di `AttendanceRepository`:
```go
// ListBySantriAndPeriod returns attendance records for a santri within an
// academic period (via period resolution through session → schedule → activity_period).
ListBySantriAndPeriod(ctx context.Context, santriID, academicPeriodID string) ([]*entity.AttendanceWithSession, error)
```

Dimana `AttendanceWithSession` adalah enriched struct:
```go
type AttendanceWithSession struct {
    Attendance       entity.Attendance
    SessionID        string
    SessionStartsAt  time.Time
    SessionEndsAt    time.Time
    ActivityName     string
    ActivityCode     string
    ScheduleType     string
}
```

Implementasi SQL:
```sql
SELECT a.*, s.id as session_id, s.starts_at, s.ends_at,
       act.name as activity_name, act.code as activity_code,
       sch.type as schedule_type
FROM attendances a
JOIN activity_sessions s ON s.id = a.activity_session_id
JOIN activity_schedules sch ON sch.id = s.activity_schedule_id
JOIN activity_periods ap ON ap.id = sch.activity_period_id
JOIN activities act ON act.id = ap.activity_id
WHERE a.santri_id = $1
  AND ap.academic_period_id = $2
  AND a.deleted_at IS NULL
  AND s.deleted_at IS NULL
ORDER BY s.starts_at DESC
```

---

## DTO Baru

```go
// --- Presensi (halaman check-in) ---

type PresensiSessionInfo struct {
    ID           string `json:"id"`
    ActivityName string `json:"activity_name"`
    ActivityCode string `json:"activity_code"`
    ScheduleType string `json:"schedule_type"`
    StartsAt     string `json:"starts_at"`
    EndsAt       string `json:"ends_at"`
    Status       string `json:"status"`
    PeriodName   string `json:"period_name"`

    // Ringkasan kehadiran.
    TotalEligible int `json:"total_eligible"`
    TotalPresent  int `json:"total_present"`
}

type CheckinRequest struct {
    NIS string `json:"nis" binding:"required"`
}

type CheckinResponse struct {
    Attendance AttendanceResponse `json:"attendance"`
    Message    string             `json:"message"`
}

type PresensiAttendanceItem struct {
    SantriID   string  `json:"santri_id"`
    NIS        *string `json:"nis,omitempty"`
    Fullname   *string `json:"fullname,omitempty"`
    Status     string  `json:"status"`
    RecordedAt string  `json:"recorded_at"`
}

// --- History absensi santri ---

type MyAttendanceSessionItem struct {
    SessionID    string `json:"session_id"`
    ActivityName string `json:"activity_name"`
    ActivityCode string `json:"activity_code"`
    ScheduleType string `json:"schedule_type"`
    StartsAt     string `json:"starts_at"`
    EndsAt       string `json:"ends_at"`
    Status       string `json:"status"` // present / absent / excused / unrecorded
    RecordedAt   *string `json:"recorded_at,omitempty"`
}

type MyAttendanceSummary struct {
    TotalSessions int `json:"total_sessions"`
    Present       int `json:"present"`
    Absent        int `json:"absent"`
    Excused       int `json:"excused"`
    Unrecorded    int `json:"unrecorded"`
}

type MyAttendanceResponse struct {
    AcademicPeriod *AcademicPeriodResponse    `json:"academic_period"`
    Summary        MyAttendanceSummary        `json:"summary"`
    Sessions       []MyAttendanceSessionItem  `json:"sessions"`
}

type MyAttendanceListQuery struct {
    AcademicPeriodID    string `form:"academic_period_id"`
    ActivityScheduleID  string `form:"activity_schedule_id"`
}
```

---

## Queries & Commands Baru

### Commands
| Use Case | Deskripsi |
|----------|-----------|
| `CheckinByNISUseCase` | Lookup santri by NIS → validasi eligibility → record sebagai `present`. |

### Queries
| Use Case | Deskripsi |
|----------|-----------|
| `GetPresensiSessionInfoUseCase` | Info sesi + ringkasan absensi untuk halaman presensi. |
| `ListPresensiAttendanceUseCase` | List santri yang sudah hadir di sesi tertentu. |
| `GetMyAttendanceUseCase` | History absensi santri untuk portal santri. |

### Logic `CheckinByNISUseCase`

```go
func (uc *CheckinByNISUseCase) Execute(ctx, sessionID, nis string) (*dto.CheckinResponse, error) {
    // 1. Sesi harus ada & berstatus open.
    session, err := uc.sessionRepo.FindByID(ctx, sessionID)
    if err != nil → 404
    if session.Status != "open" → 422 "sesi tidak terbuka"

    // 2. Lookup santri by NIS.
    info, err := uc.kesantrianReader.GetSantriByNIS(ctx, nis)
    if err != nil → 404 "NIS tidak ditemukan"
    if info.Status != "SANTRI" → 422 "bukan santri aktif"

    // 3. Resolve periode akademik dari sesi.
    periodID, err := uc.periodResolver.Resolve(ctx, sessionID)

    // 4. Validasi herregistrasi completed.
    reg, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, info.SantriID, periodID)
    if err != nil → 422 "belum herregistrasi di periode ini"
    if reg.Status != "completed" → 422 "herregistrasi belum selesai"

    // 5. Cek duplikat absensi.
    existing, _ := uc.attendanceRepo.FindBySessionAndSantri(ctx, sessionID, info.SantriID)
    if existing != nil → 409 "sudah tercatat"

    // 6. Record sebagai present.
    att, err := entity.NewAttendance(uuid, sessionID, info.SantriID, "present")
    uc.attendanceRepo.Save(ctx, att)

    return CheckinResponse{
        Attendance: mapToAttendanceResponse(att, info),
        Message:    fmt.Sprintf("Selamat, %s! Kehadiran tercatat.", info.Fullname),
    }, nil
}
```

### Logic `GetMyAttendanceUseCase`

```go
func (uc *GetMyAttendanceUseCase) Execute(ctx, userID, periodID string) (*dto.MyAttendanceResponse, error) {
    // 1. Resolve santri dari JWT.
    info := resolveSantriByUserID(ctx, userID)

    // 2. Ambil semua sesi di periode ini untuk kegiatan yang applicable.
    //    (Filter by program santri — sama seperti kegiatan/jadwal.)
    periods := activityPeriodRepo.ListByPeriodAndProgram(ctx, periodID, programID)

    // 3. Untuk setiap activity_period, ambil sessions.
    sessions := collectSessions(periods)

    // 4. Ambil absensi santri untuk sesi-sesi tsb.
    attendanceMap := map[sessionID]*Attendance(attendances by session)

    // 5. Build response: untuk setiap sesi, tentukan status:
    //    - present / absent / excused (jika ada record)
    //    - "unrecorded" (jika belum ada record)

    // 6. Hitung summary.
    return response, nil
}
```

---

## Struktur Module

```
internal/modules/akademik/
  application/
    command/
      checkin_by_nis.go                  ← BARU
    query/
      get_presensi_session_info.go       ← BARU
      list_presensi_attendance.go        ← BARU
      get_my_attendance.go              ← BARU
    dto/
      presensi_dto.go                    ← BARU
      my_attendance_dto.go              ← BARU
    ports/
      kesantrian_reader.go              ← UPDATE: tambah GetSantriByNIS
  domain/
    attendance/
      repository/interfaces.go           ← UPDATE: tambah ListBySantriAndPeriod
  infrastructure/
    kesantriangateway/gateway.go         ← UPDATE: implementasi GetSantriByNIS
    persistence/
      postgres_attendance_repo.go        ← UPDATE: implementasi ListBySantriAndPeriod
  interfaces/http/
    handler.go                           ← UPDATE: tambah handler baru
    router.go                            ← UPDATE: tambah routes baru

internal/modules/kesantrian/
  contract.go                            ← UPDATE: tambah GetSantriByNIS
  domain/santri/repository/interfaces.go ← sudah ada FindByNIS
  interfaces/http/handler.go             ← UPDATE: expose GetSantriByNIS
```

---

## Frontend

### 1. Halaman Presensi (`/presensi/:sessionId`)

**Layout**: `default` (bukan `akademik` — ini halaman publik/standalone).

```
┌──────────────────────────────────────────────────┐
│  PRESENSI                                        │
├──────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ Shalat Subuh Berjamaah                   │    │
│  │ Senin, 12 Agustus 2026 · 04:30 – 05:00  │    │
│  │                                          │    │
│  │ Kehadiran: 12 / 45                       │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │  Masukkan NIS Anda                       │    │
│  │                                          │    │
│  │  [____________________]  ✓ Hadirkan      │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │  Sudah Hadir (12)                        │    │
│  │                                          │    │
│  │  1. Ahmad F. (1000126001)   04:32       │    │
│  │  2. Budi S. (1000126002)    04:33       │    │
│  │  3. ...                                  │    │
│  └──────────────────────────────────────────┘    │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Interaksi**:
- Input NIS → tekan Enter (atau klik tombol) → POST `/presensi/:sessionId/checkin`.
- Sukses → tampilkan pesan "Selamat, {nama}!" + auto-clear input, refresh list.
- Gagal (NIS tidak ditemukan, sudah tercatat, dll) → tampilkan error.
- Auto-refresh list kehadiran setiap beberapa detik (polling atau SSE).

**Komponen baru**: `app/pages/presensi/[sessionId].vue`

### 2. Link "Buka Presensi" di Halaman Admin Sesi Detail

Di `app/pages/admin/akademik/sesi/[id].vue`, saat status = `open`:
- Tampilkan tombol **"Buka Presensi"** yang membuka `/presensi/{sessionId}` di tab baru.

### 3. Halaman Riwayat Absensi Santri (`/akademik/absensi`)

```
┌──────────────────────────────────────────────────┐
│  ← Kembali ke Akademik                           │
├──────────────────────────────────────────────────┤
│  Riwayat Absensi                                 │
│                                                  │
│  Periode: [Periode 1 2026/2027 ▼]              │
│                                                  │
│  ┌───────────┐ ┌───────────┐ ┌───────────┐     │
│  │ Hadir     │ │ Alpa      │ │ Izin      │     │
│  │ 28        │ │ 3         │ │ 2         │     │
│  └───────────┘ └───────────┘ └───────────┘     │
│                                                  │
│  ┌──────────────────────────────────────────┐    │
│  │ Sesi                   │ Waktu     │ Status│   │
│  ├────────────────────────┼───────────┼───────┤   │
│  │ Shalat Subuh           │ 12 Agu    │ ✓ Hadir│   │
│  │ Kajian Kitab           │ 12 Agu    │ ✗ Alpa │   │
│  │ Shalat Subuh           │ 13 Agu    │ ✓ Hadir│   │
│  │ Mutaba'ah              │ 13 Agu    │ —      │   │
│  │ ...                    │           │       │   │
│  └────────────────────────┴───────────┴───────┘   │
│                                                  │
└──────────────────────────────────────────────────┘
```

**Komponen baru**: `app/pages/akademik/absensi.vue`

---

## Fase Pengerjaan

### Fase 1 — Kesantrian: expose GetSantriByNIS
- [ ] Tambah `GetSantriByNIS` di `kesantrian.Contract`
- [ ] Implementasi handler/query di kesantrian (gunakan existing `FindByNIS`)
- [ ] Tambah `GetSantriByNIS` di `KesantrianReader` port akademik
- [ ] Implementasi di gateway adapter
- [ ] `go build ./...`

### Fase 2 — Repository & Persistence
- [ ] Tambah method `ListBySantriAndPeriod` di `AttendanceRepository` interface
- [ ] Implementasi di `postgres_attendance_repo.go`
- [ ] `go build ./...`

### Fase 3 — Commands & Queries
- [ ] `CheckinByNISUseCase` (presensi check-in)
- [ ] `GetPresensiSessionInfoUseCase` (info sesi + ringkasan)
- [ ] `ListPresensiAttendanceUseCase` (list kehadiran real-time)
- [ ] `GetMyAttendanceUseCase` (history absensi santri)
- [ ] DTO: `presensi_dto.go`, `my_attendance_dto.go`
- [ ] `go build ./...`

### Fase 4 — Handler & Router
- [ ] Handler presensi: `PresensiSessionInfo`, `Checkin`, `ListPresensiAttendance`
- [ ] Handler santri: `MyAttendance`
- [ ] Route presensi (tanpa JWT): `/api/v1/web/akademik/presensi/:sessionId/*`
- [ ] Route santri (JWT): `/api/v1/web/akademik/my/absensi`
- [ ] Wire di `module.go`
- [ ] `go build ./...`

### Fase 5 — Frontend: Halaman Presensi
- [ ] `app/pages/presensi/[sessionId].vue`
- [ ] Input NIS + Enter → check-in → auto-refresh list
- [ ] Error handling (NIS tidak ditemukan, sudah tercatat, sesi tutup)
- [ ] Polling untuk refresh real-time
- [ ] Link "Buka Presensi" di halaman admin sesi detail

### Fase 6 — Frontend: Halaman Riwayat Absensi Santri
- [ ] `app/pages/akademik/absensi.vue`
- [ ] Pilih periode akademik
- [ ] Tampilkan summary cards (hadir/alpa/izin/belum)
- [ ] Tampilkan list sesi + status absensi
- [ ] Link di dashboard/sidenav santri

---

## Verifikasi

1. `go build ./...` dan `go test ./...` lolos di tiap fase.
2. Smoke test presensi:
   - Admin buka sesi → klik "Buka Presensi" → halaman presensi muncul.
   - Santri input NIS valid → Enter → tercatat hadir, counter bertambah.
   - Santri input NIS yang sudah tercatat → error "sudah tercatat".
   - Santri input NIS tidak terdaftar → error "NIS tidak ditemukan".
   - Santri input NIS yang belum herreg → error "belum herregistrasi".
   - Admin tutup sesi → halaman presensi menampilkan "Sesi sudah ditutup".
3. Smoke test riwayat santri:
   - Santri login → `/akademik/absensi` → tampil summary + list sesi.
   - Filter periode → list update.
   - Status absensi (hadir/alpa/izin/belum) sesuai dengan yang tercatat.
4. Test NIS lookup dengan NIS yang tidak unik (jika ada) → error handling.
5. Test concurrent check-in (2 santri input NIS bersamaan) → tidak ada race condition.
