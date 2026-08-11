# Rules: Akademik

Dokumen ini merangkum aturan bisnis modul akademik (sistem akademik pondok pesantren).

---

## 1. Program

### Create
- `code` wajib unique, case-sensitive.
- `code` dan `name` wajib diisi.
- Default `status = active`.

### Update
- `code` boleh diubah jika belum dipakai oleh `activity_period_programs`.
- `name` boleh diubah kapan saja.
- Status boleh diubah antara `active` / `inactive`.

### Delete
- Soft delete (`deleted_at`).
- Program yang sudah dipakai di `activity_period_programs` tetap bisa di-soft-delete
  (data historis tetap ada, cukup set `deleted_at`).

---

## 2. Academic Period

### Create
- `code` wajib unique.
- `start_date` dan `end_date` wajib, `end_date >= start_date`.
- Default `status = draft`.

### Lifecycle
```
draft → open → closed → archived
```
- `draft`: periode sedang disiapkan, belum bisa digunakan untuk registrasi/activity.
- `open`: periode aktif, santri bisa registrasi, activity bisa berjalan.
- `closed`: periode selesai, tidak ada activity baru.
- `archived`: periode diarsipkan, readonly.

### Transisi Status
- `draft → open`: hanya dari `draft`.
- `open → closed`: hanya dari `open`.
- `closed → archived`: hanya dari `closed`.
- Tidak ada mundur (closed → open, archived → closed, dst).

### Delete
- Period yang sudah punya data (`santri_registrations`, `activity_periods`)
  **tidak boleh di-hard-delete**.
- Gunakan archive untuk menonaktifkan tanpa menghapus histori.
- Soft delete hanya boleh jika belum ada data transaksional.

---

## 3. Santri Registration (Herregistrasi)

### Create
- `santri_id` wajib, harus ada di kesantrian (validasi via Contract).
- `academic_period_id` wajib, harus ada dan `status = open`.
- Santri harus `status = active` di kesantrian (tidak boleh GRADUATED/DROP_OUT).
- Default `status = pending`.

### Unique Constraint
- Satu santri hanya boleh punya **satu registration aktif** per academic period.
- `UNIQUE (santri_id, academic_period_id) WHERE deleted_at IS NULL`.
- Registration yang sudah `cancelled` tetap terhitung (tidak boleh create baru
  untuk kombinasi yang sama selama yang cancelled belum di-hard-delete).

### Transisi Status
```
pending → completed
pending → cancelled
```
- `completed`: proses herregistrasi selesai, `registered_at` diisi.
- `cancelled`: registrasi dibatalkan.
- Tidak ada transisi dari `completed` atau `cancelled` ke status lain.

### Complete
- Hanya dari `pending`.
- Set `registered_at = NOW()`.

### Cancel
- Hanya dari `pending`.
- Tidak mengubah `registered_at`.

---

## 4. Activity

### Create
- `code` wajib unique.
- `code` dan `name` wajib diisi.
- Default `status = active`.

### Status
- `status = active`: activity tersedia untuk digunakan.
- `status = inactive`: activity tidak tersedia untuk period baru,
  tapi period lama yang sudah pakai tetap bisa jalan.

### Delete
- Activity yang sudah dipakai di `activity_periods` **tidak boleh di-hard-delete**.
- Gunakan soft delete.
- Ubah `status = inactive` jika tidak ingin dipakai lagi.

---

## 5. Activity Period

### Create
- `activity_id` wajib, harus ada.
- `academic_period_id` wajib, harus ada.
- Default `status = active`.

### Unique Constraint
- Satu activity hanya boleh punya **satu assignment** per academic period.
- `UNIQUE (activity_id, academic_period_id) WHERE deleted_at IS NULL`.

### Status
- `active`: activity berjalan di period ini.
- `inactive`: activity tidak berjalan di period ini (tapi record tetap ada).
- Activity bisa `active` di period 1, `inactive` di period 2, `active` di period 3.

### Transisi
- `active ↔ inactive`: boleh bolak-balik.
- Tidak ada transisi ke status lain.

### Delete
- Soft delete. Activity period yang sudah punya `activity_schedules`
  tetap bisa di-soft-delete (schedules menjadi tidak aktif).

---

## 6. Activity Period Program

### Create
- `activity_period_id` wajib.
- `program_id` wajib.
- Unique: `(activity_period_id, program_id)`.

### Default Behavior
- Jika activity_period **tidak punya** record di tabel ini,
  activity berlaku untuk **semua program**.
- Jika ada record, activity **hanya berlaku** untuk program yang disebut.

### Delete
- Hard delete (ON DELETE CASCADE dari `activity_periods`).
- Menghapus program scope tidak mempengaruhi schedule/session.

---

## 7. Activity Schedule

### Create
- `activity_period_id` wajib, harus ada.
- `type` wajib: `once`, `daily`, `weekly`, `monthly`, `yearly`.
- `start_time` dan `end_time` wajib, `end_time > start_time`.
- `start_date` dan `end_date` opsional (batas berlakunya schedule).
- Jika `start_date` dan `end_date` diisi: `end_date >= start_date`.

### Type-Specific Rules
- `once`: tidak butuh detail table. `start_date` = tanggal kejadian.
- `daily`: berlaku setiap hari dalam rentang `start_date` s/d `end_date`.
- `weekly`: butuh `activity_schedule_weeklies` (minimal 1 hari).
- `monthly`: butuh `activity_schedule_monthlies` (minimal 1 tanggal).
- `yearly`: butuh `activity_schedule_yearlies` (minimal 1 bulan+tanggal).

### Recurrence Detail
- Detail table **hanya boleh** digunakan sesuai type.
- `weekly` → `activity_schedule_weeklies`.
- `monthly` → `activity_schedule_monthlies`.
- `yearly` → `activity_schedule_yearlies`.
- `once` dan `daily` tidak butuh detail table.

### Update
- Boleh diubah kapan saja.
- Perubahan schedule tidak mempengaruhi session yang sudah ada.

### Delete
- Soft delete.
- Schedule yang sudah punya session tetap bisa di-soft-delete
  (session yang sudah tercatat tidak terhapus).

---

## 8. Activity Schedule Weekly / Monthly / Yearly

### Weekly
- `day_of_week` wajib: `monday`, `tuesday`, ..., `sunday`.
- Unique: `(schedule_id, day_of_week)`.
- Satu schedule weekly bisa punya multiple days (multiple rows).

### Monthly
- `day_of_month` wajib: 1-31.
- Unique: `(schedule_id, day_of_month)`.
- Satu schedule monthly bisa punya multiple days (multiple rows).

### Yearly
- `month` wajib: 1-12.
- `day` wajib: 1-31.
- Unique: `(schedule_id, month, day)`.
- Satu schedule yearly bisa punya multiple occurrences (multiple rows).

### Cascade Delete
- Semua detail table `ON DELETE CASCADE` dari `activity_schedules`.
- Jika schedule dihapus, detail ikut terhapus.

---

## 9. Activity Session

### Create
- `activity_schedule_id` wajib, harus ada.
- `starts_at` dan `ends_at` wajib, `ends_at > starts_at`.
- Default `status = scheduled`.

### Status
- `scheduled`: session sudah dijadwalkan, belum dimulai.
- `open`: session sedang berlangsung.
- `completed`: session selesai.
- `cancelled`: session dibatalkan.

### Transisi Status
```
scheduled → open → completed
scheduled → cancelled
open → completed
open → cancelled
```
- Tidak ada mundur (completed → open, cancelled → scheduled, dst).

### Cancel
- Session bisa dibatalkan kapan saja (selama belum `completed`).
- Cancel session tidak mempengaruhi schedule.
- Histori session yang cancelled tetap ada (soft delete, tidak hard delete).

### Delete
- Soft delete.
- Session yang sudah punya attendance tetap bisa di-soft-delete
  (attendance tetap tercatat sebagai historis).

---

## 10. Attendance

### Create
- `activity_session_id` wajib, harus ada dan `status != cancelled`.
- `santri_id` wajib, harus ada di kesantrian (validasi via Contract).
- `status` wajib: `present`, `absent`, `excused`.
- `recorded_at` wajib, default `NOW()`.

### Unique Constraint
- Satu santri hanya boleh punya **satu attendance** per session.
- `UNIQUE (activity_session_id, santri_id) WHERE deleted_at IS NULL`.

### Status
- `present`: santri hadir.
- `absent`: santri tidak hadir.
- `excused`: santri tidak hadir dengan alasan (sakit, izin, dst).

### Update
- Status boleh diubah (mis. dari `absent` ke `present` jika ada koreksi).
- `recorded_at` tidak berubah saat update status.

### Delete
- Soft delete.
- Attendance yang sudah tercatat tidak boleh di-hard-delete (historis).

### Violation
- Attendance **tidak secara otomatis** membuat violation/pelanggaran.
- Domain pelanggaran/takziran **di luar scope** fase ini.
- Jangan membuat coupling ke domain pelanggaran berdasarkan asumsi.

---

## 11. Cross-Module Validation

### Santri Existence
- Semua operasi yang melibatkan `santri_id` harus validasi ke kesantrian
  lewat `kesantrian.Contract`.
- Santri harus `status = active` (tidak boleh `alumni` / `drop_out`).

### Batch Operations
- Untuk operasi batch (mis. record attendance untuk banyak santri),
  gunakan `ListSantriByIDs` untuk validasi semua santri sekaligus.

---

## 12. Histori & Audit

### Soft Delete
- Semua entity menggunakan soft delete (`deleted_at`).
- Entity yang sudah digunakan dalam transaksi akademik tidak boleh di-hard-delete.

### Immutable Records
- Session yang sudah `completed` tidak boleh diubah statusnya.
- Attendance yang sudah tercatat tidak boleh di-hard-delete.

### No Hard Delete
- Program, Activity, Activity Period, Schedule, Session, Attendance
  tidak boleh di-hard-delete jika sudah punya data transaksional.
- Gunakan soft delete atau status change.

---

## 13. Access Control

### Permission
- Semua endpoint butuh JWT valid + permission `manage_akademik`.
- Operational tasks (CRUD program, period, registration, activity, schedule,
  session, attendance) menggunakan `manage_akademik`.
- Configuration tasks (open/close/archive period) bisa dipisah ke
  `manage_akademik_settings` jika diperlukan.

### Self-Service
- Fase ini **tidak ada** endpoint self-service untuk santri.
- Semua endpoint admin-only.
