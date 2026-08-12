# Plan: UX Sesi dari Jadwal (Frontend)

## Status: ✅ Diimplementasikan (sipon-ui)

Backend `generate-sessions` sudah tersedia; UI sudah diterapkan di repo `sipon-ui`:

- `app/components/admin/akademik/GenerateSessionsModal.vue` (baru) — dialog generate bulk
  dengan preview tanggal + toast hasil (X dibuat / Y dilewati).
- `app/pages/admin/akademik/jadwal/[id].vue` — section "Sesi Kegiatan" (tabel sesi per
  jadwal, tombol "Buat Sesi" & "Generate Sesi").
- `app/components/admin/akademik/SessionFormModal.vue` — prop `schedule` untuk pre-fill
  waktu (start_time/end_time) + kirim offset platform `+07:00`.
- `app/pages/admin/akademik/sesi/index.vue` — dropdown filter "Semua jadwal".
- `app/stores/akademik.ts` — action `generateSessions`.
- `app/utils/akademikSchedule.ts` — `PLATFORM_TZ_OFFSET`, `toRFC3339WithOffset`,
  `scheduleDatesInRange` (preview).
- `shared/types/Akademik.ts` — `GenerateSessionsRequest`/`GenerateSessionsResponse`.

---

## Context

Backend sudah mendukung relasi 1:N antara jadwal (`activity_schedule`) dan sesi
(`activity_session`). Satu jadwal bisa menghasilkan banyak sesi. Saat ini di UI,
sesi dibuat satu-per-satu lewat halaman sessions admin, terpisah dari jadwal.

**Tujuan UX:**
1. Dari halaman jadwal, admin bisa **langsung generate sesi** (banyak sekaligus)
   berdasarkan recurrence pattern + rentang tanggal.
2. Dari halaman jadwal, admin bisa **melihat semua sesi** untuk jadwal tersebut.
3. Form buat sesi tunggal **pre-fill waktu** dari jadwal (start_time / end_time),
   tapi tetap bisa diedit.

**Keputusan:**
- Generate bulk dari jadwal → **Ya**
- Sesi duplikat (tanggal+waktu sama) → **Skip** (backend idempotent, tidak error)
- Pre-fill waktu → Frontend baca `start_time`/`end_time` jadwal
- Relasi jadwal→sesi → 1:N

---

## API yang Digunakan (sudah tersedia di backend)

### 1. Generate Sessions Bulk

```
POST /api/v1/web/akademik/schedules/:id/generate-sessions
Content-Type: application/json
Authorization: Bearer <token>   (permission: manage_akademik)

{
  "from_date": "2026-08-01",   // required
  "to_date":   "2026-08-31"    // optional, default = from_date
}
```

Response:
```json
{
  "code": "OK",
  "message": "sesi kegiatan berhasil digenerate dari jadwal",
  "data": {
    "total_dates_expanded": 9,
    "total_created": 9,
    "total_skipped": 0,
    "sessions": [
      {
        "id": "...",
        "activity_schedule_id": "...",
        "starts_at": "2026-08-03T08:00:00+07:00",
        "ends_at": "2026-08-03T09:00:00+07:00",
        "status": "scheduled",
        "created_at": "...",
        "updated_at": "..."
      }
    ]
  }
}
```

- Waktu `starts_at`/`ends_at` diambil dari `start_time`/`end_time` jadwal (jam
  wall-clock dalam platform timezone).
- `total_dates_expanded` = jumlah tanggal yang memenuhi recurrence pattern.
- `total_skipped` = jumlah tanggal yang sudah punya sesi (tidak dibuat ulang).
- Perilaku idempotent: memanggil endpoint dua kali dengan range yang sama tidak
  akan membuat sesi ganda.

### 2. List Sesi (filter by jadwal) — sudah ada

```
GET /api/v1/web/akademik/sessions?activity_schedule_id=<scheduleId>&page=1&limit=20
Authorization: Bearer <token>   (permission: manage_akademik)
```

### 3. Detail Jadwal — sudah ada

```
GET /api/v1/web/akademik/schedules/:id
Authorization: Bearer <token>   (permission: manage_akademik)
```

Response memuat `start_time`, `end_time`, `type`, `start_date`, `end_date`,
`weekly_days` / `monthly_days` / `yearly_dates` — dipakai untuk preview tanggal
dan pre-fill waktu di form.

---

## Perubahan UI yang Disarankan

### A. Halaman Detail Jadwal — Section "Sesi"

Tambahkan section baru "Sesi Kegiatan" di bawah detail jadwal:

- **Tombol "Generate Sesi"** di header section → membuka dialog generate.
- **Tombol "Buat Sesi"** (single) → membuka form create session dengan pre-fill
  waktu dari jadwal.
- **Tabel/List sesi** — data dari `GET /sessions?activity_schedule_id=:id`:
  - Kolom: tanggal (`starts_at` di-format `DD/MM/YYYY`), waktu mulai, waktu
    selesai, status (scheduled/open/completed/cancelled), aksi
    (buka / buka-presensi / selesai / batal).
  - Pagination bila jumlah banyak.
  - Setelah generate, refresh list sesi.

### B. Dialog "Generate Sesi"

- **Input "Dari tanggal"** (required) dan **"Sampai tanggal"** (optional) —
  date picker.
- **Preview** (opsional tapi disarankan): tampilkan daftar tanggal yang akan
  di-generate. Frontend bisa menghitung sendiri dari `type` + detail recurrence
  yang ada di response detail jadwal, ATAU langsung tampilkan
  `total_dates_expanded` setelah backend memproses. Untuk pengalaman lebih baik:
  hitung preview di frontend (pola sama dengan halaman kalender).
- **Tombol "Generate"** → `POST /schedules/:id/generate-sessions`.
- Setelah sukses: tampilkan toast/info **"X sesi dibuat, Y dilewati (sudah ada)"**
  dari `total_created` dan `total_skipped`. Refresh list sesi.

### C. Form "Buat Sesi" (single) — Pre-fill Waktu

Saat admin klik "Buat Sesi" dari halaman jadwal (atau memilih jadwal di form
sessions global):

- `starts_at` → date picker kosong + **time picker diisi dari `schedule.start_time`**
- `ends_at` → date picker kosong + **time picker diisi dari `schedule.end_time`**
- User memilih tanggal; waktu sudah terisi (auto-fill) tapi tetap bisa diubah.
- Kirim `POST /sessions` dengan body:
  ```json
  { "activity_schedule_id": "<id>", "starts_at": "2026-08-12T08:00:00+07:00", "ends_at": "2026-08-12T09:00:00+07:00" }
  ```
- Pastikan format RFC3339 dengan offset timezone (backend memakai platform
  timezone; kirim offset `+07:00`).

### D. Halaman Sessions Admin (tetap ada)

Tidak diubah. Tetap bisa filter sesi global (semua jadwal). Tambahan kecil yang
disarankan: dropdown/select filter by jadwal agar admin bisa menemukan sesi dari
jadwal tertentu tanpa lewat halaman jadwal.

---

## Catatan Implementasi Frontend

1. **Format waktu**: backend mengirim `starts_at`/`ends_at` dalam platform
   timezone dengan offset (mis. `+07:00`). Untuk menampilkan waktu lokal di UI,
   gunakan library date (dayjs/moment/luxon) dengan `timeZone` atau langsung
   tampilkan jam dari string (karena offset sudah WIB). Hindari konversi ganda.
2. **Preview tanggal**: pola hitung recurrence (daily/weekly/monthly/yearly/once)
   sama seperti logika kalender di backend — bisa ditiru di frontend, atau
   cukup tampilkan `total_dates_expanded` dari response.
3. **Idempotent**: jika admin klik generate dua kali dengan range sama, backend
   mengembalikan `total_created=0`, `total_skipped=N` — UI harus menampilkan
   pesan yang jelas, bukan error.
4. **Loading & error**: tampilkan loading selama generate (transaksi backend
   all-or-nothing). Jika error validasi (mis. `to_date` sebelum `from_date`),
   tampilkan pesan dari backend.
