# Plan: Auto-Absen Alpa & Validasi Program pada Sesi

## Context

Santri aktif dikelompokkan ke dalam **program** (melalui `santri_programs`). Setiap program
memiliki **kegiatan wajib** masing-masing yang ditentukan melalui relasi
`ActivityPeriodProgram` (kegiatan mana yang berlaku untuk program mana).

Setiap kegiatan memiliki **sesi** (`activity_session`) yang dijadwalkan, dibuka, dan
diselesaikan. Sesi yang diliburkan dicatat sebagai **dibatalkan** (`cancelled`).

**Kebutuhan baru:**
1. **Auto-absen alpa** — saat sesi diselesaikan (`completed`), sistem secara otomatis
   mencatat absensi sebagai `absent` (alpa) untuk **semua santri pada program yang
   terkait** yang belum melakukan absensi.
2. **Sesi diliburkan** → status `cancelled` (sudah ada di domain). Tidak ada auto-absen
   untuk sesi yang dibatalkan.
3. **Validasi program** — hanya santri yang berada pada program yang berkaitan dengan
   kegiatan tersebut yang boleh melakukan absensi (check-in).

**Domain chain (sudah ada):**
```
Program ← ActivityPeriodProgram → ActivityPeriod → ActivitySchedule → ActivitySession → Attendance
SantriProgram: SantriID → ProgramID (aktif)
```

---

## Alur Sistem

```
Sesi dijadwalkan (status: scheduled)
    │
    ├── Admin / sistem membuka sesi (status: scheduled → open)
    │
    ▼
Santri check-in (via presensi NIS):
    • Validasi sesi status = open
    • Validasi santri belongs to program yang terkait dengan kegiatan ini
    • Record attendance sebagai "present"
    │
    ▼
Admin / sistem menyelesaikan sesi (status: open → completed):
    │
    ├── Sistem query semua santri aktif di program terkait
    ├── Sistem query santri yang sudah absen di sesi ini
    ├── Selisih = santri yang tidak absen
    ├── Untuk setiap santri yang tidak absen → record attendance sebagai "absent"
    │
    ▼
Sesi dibatalkan (status: * → cancelled):
    • TIDAK ada auto-absen alpa
    • Sesi dianggap diliburkan / tidak dilaksanakan
```

---

## Keputusan Desain

| No | Pertanyaan | Keputusan |
|----|-----------|-----------|
| 1 | Kapan auto-absen alpa dijalankan? | Saat sesi di-complete (transisi ke `completed`). Bisa dipicu dari admin action atau scheduled job. |
| 2 | Siapa yang terkena auto-absen? | Semua santri yang punya `santri_program` aktif di program yang terkait dengan kegiatan sesi tersebut, dan belum punya record absensi di sesi itu. Termasuk santri yang belum herregistrasi. |
| 3 | Bagaimana resolve program dari sesi? | Session → Schedule → ActivityPeriod → ActivityPeriodProgram → Program. Jika ActivityPeriodProgram kosong (tidak ada scope), kegiatan berlaku untuk **semua** program → auto-absen untuk semua santri aktif. |
| 4 | Sesi dibatalkan (cancelled) apakah auto-absen? | **Tidak.** Sesi cancelled = diliburkan, tidak ada kewajiban hadir. |
| 5 | Bagaimana validasi check-in? | Saat check-in, resolve program dari sesi, lalu cek apakah santri memiliki `santri_program` aktif di program tersebut. Jika tidak → tolak. |
| 6 | Bagaimana jika santri belum herregistrasi? | **Auto-absen**: tetap dicatat sebagai `absent` (tidak ada filter herregistrasi). **Check-in**: validasi herregistrasi tetap berlaku (existing check). |
| 7 | Apakah auto-absen bisa dilakukan ulang? | Tidak — auto-absen hanya mencatat santri yang **belum** punya record absensi. Santri yang sudah tercatat (present/excused) tidak ditimpa. |

---

## Domain & Repository Changes

### 1. Resolve Programs dari Session

Helper/service untuk resolve daftar program yang terkait dengan sebuah sesi:

```go
// akademik/application/ports/program_resolver.go

type ProgramResolver interface {
    // ResolveProgramIDsFromSession returns the list of program IDs associated
    // with the given session. If ActivityPeriodProgram has no rows (no scope),
    // returns all active programs.
    ResolveProgramIDsFromSession(ctx context.Context, sessionID string) ([]string, error)
}
```

Implementasi:
1. `session → ActivityScheduleID`
2. `schedule → ActivityPeriodID`
3. `activity_period → List ActivityPeriodProgram`
4. Jika kosong → ambil semua program aktif
5. Jika ada → return daftar `ProgramID` dari junction

### 2. Resolve Santri di Program (untuk Auto-Absen)

Butuh method untuk mengambil semua santri aktif di suatu program:

```go
// akademik/domain/santri_program/repository/interfaces.go (UPDATE)

type SantriProgramRepository interface {
    // ... existing methods

    // ListActiveSantriIDsByProgramID returns all santri IDs with active program
    // membership for the given program.
    ListActiveSantriIDsByProgramID(ctx context.Context, programID string) ([]string, error)
}
```

### 3. Check Existing Attendance IDs

```go
// akademik/domain/attendance/repository/interfaces.go (UPDATE)

type AttendanceRepository interface {
    // ... existing methods

    // ListBySession returns all attendance records for the given session.
    ListBySession(ctx context.Context, sessionID string) ([]*entity.Attendance, error)

    // ListSantriIDsBySession returns santri IDs that already have attendance
    // records in the given session.
    ListSantriIDsBySession(ctx context.Context, sessionID string) ([]string, error)
}
```

### 4. Resolve Academic Period dari Session

Helper untuk resolve academic_period_id dari session (untuk filter herregistrasi):

```go
// akademik/application/ports/period_resolver.go (UPDATE jika belum ada)

type PeriodResolver interface {
    // ResolveAcademicPeriodIDFromSession returns the academic period ID
    // for the given session (session → schedule → activity_period → academic_period).
    ResolveAcademicPeriodIDFromSession(ctx context.Context, sessionID string) (string, error)
}
```

---

## Use Case: Auto-Absen Alpa (Complete Session)

### `CompleteSessionWithAutoAbsenUseCase`

Logic:

```go
func (uc *CompleteSessionWithAutoAbsenUseCase) Execute(ctx context.Context, sessionID string) error {
    // 1. Load session, validasi bisa di-complete.
    session, err := uc.sessionRepo.FindByID(ctx, sessionID)
    if err != nil → 404
    err = session.Complete() // domain state transition
    if err != nil → 422 "tidak bisa diselesaikan"

    // 2. Resolve daftar program terkait sesi ini.
    programIDs, err := uc.programResolver.ResolveProgramIDsFromSession(ctx, sessionID)
    if err != nil → 500

    // 3. Kumpulkan semua santri aktif di program terkait (tanpa filter herregistrasi).
    var allSantriIDs []string
    for _, pid := range programIDs {
        santriIDs, _ := uc.santriProgramRepo.ListActiveSantriIDsByProgramID(ctx, pid)
        allSantriIDs = append(allSantriIDs, santriIDs...)
    }
    allSantriIDs = unique(allSantriIDs)

    // 4. Ambil santri yang sudah punya record absensi di sesi ini.
    recordedIDs, _ := uc.attendanceRepo.ListSantriIDsBySession(ctx, sessionID)
    recordedSet := toSet(recordedIDs)

    // 5. Selisih = santri yang belum absen → auto-absen sebagai absent.
    for _, sid := range allSantriIDs {
        if _, exists := recordedSet[sid]; exists {
            continue // sudah tercatat, skip
        }
        att, _ := entity.NewAttendance(uuid.NewString(), sessionID, sid, "absent")
        uc.attendanceRepo.Save(ctx, att)
    }

    // 6. Update session status ke completed.
    uc.sessionRepo.Update(ctx, session)

    return nil
}
```

**Catatan**: Jika sesi di-complete tapi tidak ada program terkait (mis. kegiatan tanpa
scope program), maka auto-absen berlaku untuk **semua** santri aktif — sesuai
keputusan #3 di tabel. Santri yang belum herregistrasi tetap dicatat sebagai absent.

---

## Use Case: Cancel Session (Libur)

### `CancelSessionUseCase`

Logic sudah ada di domain (`session.Cancel()`), tapi perlu dipastikan:

```go
func (uc *CancelSessionUseCase) Execute(ctx context.Context, sessionID string) error {
    session, err := uc.sessionRepo.FindByID(ctx, sessionID)
    if err != nil → 404

    err = session.Cancel()
    if err != nil → 422

    // TIDAK ada auto-absen alpa untuk sesi yang dibatalkan.

    uc.sessionRepo.Update(ctx, session)
    return nil
}
```

Jika sudah ada record absensi sebelumnya (mis. santri sudah check-in sebelum sesi
dibatalkan), record tersebut **tetap ada** — tidak dihapus. Admin bisa manual
override jika diperlukan.

---

## Use Case: Validasi Check-in by Program

### Update `CheckinByNISUseCase`

Tambah validasi bahwa santri belongs to program terkait:

```go
func (uc *CheckinByNISUseCase) Execute(ctx, sessionID, nis string) (*dto.CheckinResponse, error) {
    // 1-2. Existing: sesi open, lookup santri by NIS, validasi status santri.
    // 3. Resolve periode akademik (existing).
    // 4. Validasi herregistrasi (existing).

    // --- BARU ---
    // 5. Resolve program terkait sesi ini.
    programIDs, err := uc.programResolver.ResolveProgramIDsFromSession(ctx, sessionID)
    if err != nil → 500

    // 6. Cek apakah santri memiliki santriprogram aktif di salah satu program tsb.
    santriProgram, _ := uc.santriProgramRepo.FindBySantriID(ctx, info.SantriID)
    if santriProgram == nil → 422 "santri tidak terdaftar di program manapun"

    found := false
    for _, pid := range programIDs {
        if santriProgram.ProgramID == pid {
            found = true
            break
        }
    }
    if !found → 422 "santri tidak terdaftar di program kegiatan ini"

    // 7. Existing: cek duplikat, record present.
}
```

---

## Endpoint HTTP

### Update: Complete Session

Endpoint existing `POST /sessions/:id/complete` akan menjalankan
`CompleteSessionWithAutoAbsenUseCase` (auto-absen alpa dijalankan otomatis).

### Update: Cancel Session

Endpoint existing `POST /sessions/:id/cancel` akan menjalankan
`CancelSessionUseCase` (tanpa auto-absen).

### Tidak ada endpoint baru

Tidak ada endpoint baru — perubahan behavior ada di existing complete/cancel session.

---

## Struktur Module

```
internal/modules/akademik/
  application/
    command/
      complete_session.go                    ← UPDATE: tambah auto-absen alpa
      cancel_session.go                      ← UPDATE: pastikan tidak ada auto-absen
      checkin_by_nis.go                      ← UPDATE: tambah validasi program
    ports/
      program_resolver.go                    ← BARU: resolve programs dari session
      period_resolver.go                     ← BARU/UPDATE: resolve academic period dari session
  domain/
    santri_program/
      repository/interfaces.go               ← UPDATE: tambah ListActiveSantriIDsByProgramID
    attendance/
      repository/interfaces.go               ← UPDATE: tambah ListSantriIDsBySession
    activity_session/
      repository/interfaces.go               ← UPDATE: tambah ListByStatus jika perlu batch
  infrastructure/
    persistence/
      postgres_santri_program_repo.go        ← UPDATE: implementasi ListActiveSantriIDsByProgramID
      postgres_attendance_repo.go            ← UPDATE: implementasi ListSantriIDsBySession
```

---

## Fase Pengerjaan

### Fase 1 — Repository & Resolver
- [ ] Tambah `ListActiveSantriIDsByProgramID` di `SantriProgramRepository`
- [ ] Implementasi di `postgres_santri_program_repo.go`
- [ ] Tambah `ListSantriIDsBySession` di `AttendanceRepository`
- [ ] Implementasi di `postgres_attendance_repo.go`
- [ ] Buat `ProgramResolver` port + implementasi
- [ ] Buat/update `PeriodResolver` port + implementasi
- [ ] `go build ./...`

### Fase 2 — Update Complete Session (Auto-Absen)
- [ ] Update `CompleteSessionUseCase` → `CompleteSessionWithAutoAbsenUseCase`
- [ ] Inject dependencies: `programResolver`, `periodResolver`, `santriProgramRepo`, `attendanceRepo`, `registrationRepo`
- [ ] Logic: resolve programs → collect eligible santri → diff with existing attendance → auto-record absent
- [ ] Wire di `module.go`
- [ ] `go build ./...`

### Fase 3 — Update Cancel Session
- [ ] Pastikan `CancelSessionUseCase` tidak menjalankan auto-absen
- [ ] `go build ./...`

### Fase 4 — Update Check-in Validasi
- [ ] Update `CheckinByNISUseCase`: tambah validasi program membership
- [ ] Inject `programResolver` + `santriProgramRepo`
- [ ] `go build ./...`

### Fase 5 — Testing
- [ ] Unit test: auto-absen alpa saat complete session
- [ ] Unit test: cancel session tidak auto-absen
- [ ] Unit test: check-in ditolak jika santri bukan dari program terkait
- [ ] Unit test: check-in diterima jika santri dari program terkait
- [ ] Unit test: auto-absen skip santri yang sudah tercatat
- [ ] Unit test: auto-absen tetap mencatat santri yang belum herreg completed sebagai absent
- [ ] `go test ./...`

---

## Verifikasi

1. `go build ./...` dan `go test ./...` lolos di tiap fase.
2. **Auto-absen alpa**:
   - Admin complete session → semua santri program yang belum absen tercatat sebagai `absent`.
   - Santri yang sudah `present`/`excused` tidak ditimpa.
   - Santri dari program lain tidak terpengaruh.
   - Santri yang belum herreg completed **tetap** di-absen sebagai absent.
3. **Cancel session**:
   - Admin cancel session → status menjadi `cancelled`, tidak ada auto-absen.
   - Jika ada santri yang sudah check-in sebelum cancel → record tetap ada.
4. **Validasi check-in**:
   - Santri dari program yang sesuai → check-in berhasil.
   - Santri dari program lain → ditolak dengan pesan "santri tidak terdaftar di program kegiatan ini".
   - Santri tanpa program aktif → ditolak dengan pesan "santri tidak terdaftar di program manapun".
5. **Kegiatan tanpa scope program** (ActivityPeriodProgram kosong):
   - Auto-absen berlaku untuk semua santri aktif.
   - Check-in diterima untuk santri dari program manapun.

---

## Catatan

- **Performance**: jika jumlah santri besar, pertimbangkan batch insert untuk auto-absen
  (bulk `INSERT INTO attendances` langsung via SQL, bukan satu-per-satu `Save`).
- **Idempotency**: auto-absen aman dipanggil ulang — hanya insert untuk santri yang
  belum punya record.
- **Existing attendance records**: auto-absen tidak mengubah record yang sudah ada
  (present/excused). Admin tetap bisa manual override.
- **Session from scheduled → completed**: jika sesi di-complete tanpa pernah di-open,
  auto-absen tetap berjalan (domain allow transisi `scheduled → completed`).
