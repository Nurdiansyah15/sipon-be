# Plan: Platform Timezone — Penerapan untuk Module Lain (Non-Akademik)

## Context

Modul akademik sudah menerapkan timezone platform (lihat `docs/plan/platform-timezone.md`).
Dokumen ini menjabarkan perubahan yang harus dilakukan pada module lain agar konsisten
dengan prinsip yang sama.

**Prinsip timezone platform:**

| Kategori | Aturan |
|----------|--------|
| **TIMESTAMPTZ** (kolom `*_at`, `starts_at`, `ends_at`, `recorded_at`, dll) | Disimpan **UTC**. Output ke UI dikonversi ke platform timezone. |
| **DATE** (kolom `DATE`) | Bersifat timezone-naive: parse & format dalam platform timezone. |
| **TIME** (kolom `TIME`) | Bersifat timezone-naive: representasi wall-clock dalam platform timezone. |
| **Logic perbandingan waktu** | Semua `time.Now()` memakai instant UTC (aman selama input konsisten). |

**Timezone dikonfigurasi via env `APP_TIMEZONE` (default `Asia/Jakarta`), sama untuk
semua user.**

Foundation yang sudah ada (tidak perlu dibuat ulang):

- `internal/shared/timeutil/timeutil.go`
  - `Init(timezone string) error`
  - `Now() time.Time` — `time.Now()` dalam platform timezone
  - `ToPlatform(t time.Time) time.Time`
  - `ToPlatformPtr(t *time.Time) *time.Time`
  - `ParseDate(s string) (time.Time, error)` — parse `"2006-01-02"` dalam platform timezone
  - `ParseDatePtr(s *string) (*time.Time, error)`
  - `FormatDate(t time.Time) string`
  - `FormatDateTime(t time.Time) string` — RFC3339 dalam platform timezone
  - `DateOnly(t time.Time) time.Time`
- `config.go` — field `App.Timezone`
- Entrypoint `cmd/api`, `cmd/migrate`, `cmd/seeder` — sudah memanggil `timeutil.Init`

---

## Daftar Perubahan per Module

### 1. Module `keuangan` (paling banyak)

**Input parsing DATE — ganti `time.Parse("2006-01-02", ...)` → `timeutil.ParseDate`:**

| File | Baris |
|------|-------|
| `application/command/create_period.go` | 27, 31 |
| `application/command/create_invoice.go` | 82, 129 |
| `application/command/create_invoice_batch.go` | 114, 124 |
| `application/command/create_manual_payment.go` | 45 |
| `application/command/submit_payment.go` | 55 |
| `application/command/assign_scheme_to_santri.go` | 48, 55 |
| `application/command/update_assignment.go` | 49, 56 |
| `application/command/create_billing_period.go` | 28, 32 |
| `application/command/create_manual_journal.go` | 63 |
| `infrastructure/persistence/postgres_report_reader.go` | 152 |

**Output format DATE — ganti `.Format("2006-01-02")` → `timeutil.FormatDate`:**

- `query/get_active_period.go` 38-39
- `query/get_journal_entry.go` 48
- `query/report_ledger.go` 55
- `query/list_periods.go` 43-44
- `query/report_trial_balance.go` 30
- `query/billing_period_response.go` 13-14
- `query/invoice_response.go` 28
- `query/report_balance_sheet.go` 44
- `query/list_assignments.go` 43, 45, 48
- `query/list_journal_entries.go` 45
- `query/payment_response.go` 22
- `command/create_period.go` 71-72
- `command/create_invoice.go` 106, 223, 229
- `command/create_manual_payment.go` 114
- `command/create_invoice_batch.go` 121
- `command/create_manual_journal.go` 166
- `command/create_billing_period.go` 61-62

**Output format TIMESTAMPTZ — ganti `.Format("2006-01-02T15:04:05Z07:00")` → `timeutil.FormatDateTime`:**

53 kemunculan di query & command response builder. Daftar lengkap:
`query/list_billing_schemes.go`, `query/payment_response.go`, `query/list_accounts.go`,
`query/get_account.go`, `query/get_active_period.go`, `query/billing_batch_response.go`,
`query/get_journal_entry.go`, `query/list_periods.go`, `query/get_billing_scheme.go`,
`query/billing_period_response.go`, `query/invoice_response.go`,
`query/list_fee_components.go`, `query/list_journal_entries.go`, `command/create_period.go`,
`command/create_billing_scheme.go`, `command/create_invoice.go`, `command/create_account.go`,
`command/create_manual_payment.go`, `command/create_fee_component.go`,
`command/create_manual_journal.go`, `command/create_billing_period.go`.

> **Catatan:** DTO keuangan memakai tipe `string` (sudah manual format), jadi cukup
> ganti format call — tidak ada `time.Time` yang di-marshal JSON secara implisit.

**Logic `time.Now()` — ganti dengan `timeutil.Now()` agar konsisten platform TZ:**

| File | Baris | Keterangan |
|------|-------|-----------|
| `infrastructure/persistence/postgres_billing_scheme_repo.go` | 278 | `FindActiveBySantriIDAt(..., time.Now())` |
| `infrastructure/external/pdf_generator.go` | 64 | `time.Now().Format("02/01/2006 15:04")` — timestamp cetak PDF |
| `infrastructure/external/report_pdf.go` | 70 | Sama — timestamp cetak PDF |
| `domain/invoice/valueobject/invoice_number.go` | 19 | `now.Year()`, `now.Month()` untuk nomor invoice |
| `domain/journal/valueobject/journal_number.go` | 19 | Nomor jurnal |
| `domain/payment/valueobject/payment_number.go` | 19 | Nomor pembayaran |

> **Catatan penting:** Penomoran invoice/jurnal/payment memakai `Year()`/`Month()`
> dari server local time. Jika tidak dikonversi, pada tanggal 1 Januari dini hari WIB
> nomor bisa terlanjur memakai tahun sebelumnya (karena server UTC). Wajib ganti ke
> `timeutil.Now()`.

---

### 2. Module `kesantrian`

**Input parsing DATE → `timeutil.ParseDate`:**

| File | Baris |
|------|-------|
| `interfaces/http/import_excel.go` | 94 (tanggal lahir dari Excel) |
| `application/command/create_surat.go` | 45 (tanggal surat) |

**Output format DATE → `timeutil.FormatDate`:**

| File | Baris |
|------|-------|
| `application/query/get_surat.go` | 46 |
| `application/query/list_surat.go` | 53 |

**DTO `time.Time` yang di-marshal JSON → `timeutil.ToPlatform` / `ToPlatformPtr`:**

- `application/dto/santri_dto.go` — `DOB` (40), `CreatedAt`/`UpdatedAt` (88-89)
- `application/dto/admin_dto.go` — `CreatedAt` (34)
- `application/dto/dokumen_dto.go` — `CreatedAt` (29), `VerifiedAt`/`CreatedAt` (42-43)
- `application/dto/request_dto.go` — `CreatedAt` (25)
- `application/dto/surat_dto.go` — `CreatedAt` (19, 31)
- `application/dto/persuratan_dto.go` — `CreatedAt`/`UpdatedAt` (24-25)

Konversi dilakukan saat menyusun response (di query/command layer), bukan di DTO.

---

### 3. Module `psb`

**Input parsing DATE → `timeutil.ParseDate`:**

| File | Baris |
|------|-------|
| `application/command/manage_setting.go` | 26, 30 |

**Output format DATE → `timeutil.FormatDate`:**

| File | Baris |
|------|-------|
| `application/query/list_pendaftaran.go` | 56 |
| `application/command/manage_setting.go` | 92-93 |
| `application/query/setting_query.go` | 55-56 |

**Logic `time.Now().Format("06")` → `timeutil.Now().Format("06")` (penting!):**

| File | Baris | Keterangan |
|------|-------|-----------|
| `application/command/generate_nis.go` | 59 | Tahun masuk NIS |
| `application/command/upsert_formulir.go` | 223 | Tahun NoRegis |

> NIS/NoRegis memakai tahun server local. Di akhir tahun (31 Des malam WIB / 1 Jan
> dini hari WIB) tahun bisa salah jika server UTC. Ganti ke `timeutil.Now()`.

**DTO `time.Time` → `timeutil.ToPlatform`:**

- `application/dto/pendaftaran_dto.go` — `DOB` (13, 83), `AcceptedAt` (133), `CreatedAt`/`UpdatedAt` (138-139)
- `application/dto/dokumen_dto.go` — `VerifiedAt`/`CreatedAt` (38-39)
- `application/dto/review_dto.go` — `CreatedAt` (12)
- `application/dto/setting_dto.go` — `DataPurgedAt`/`CreatedAt`/`UpdatedAt` (36-38)

> **Inkonsistensi yang perlu diperbaiki:** `list_pendaftaran.go:56` memformat
> `CreatedAt` (full timestamp) sebagai date-only `"2006-01-02"` — memotong komponen
> waktu. Ini harus diseragamkan dengan DTO `PendaftarResponse` yang memakai
> `time.Time`.

---

### 4. Module `identity`

**Logic `time.Now()` — instant comparison, TIDAK perlu diubah (sudah benar):**

- `domain/user/entity/user.go` 133, 146 (lockout)
- `domain/verification/entity/verification_code.go` 41-42, 62 (OTP expiry)
- `domain/role/entity/user_role.go` 63 (role expiry)
- `infrastructure/external/jwt_token.go` 39, 55 (JWT)
- `infrastructure/cache/redis_rate_limiter.go` 48 (rate limit)

> Semua ini membandingkan instant (`time.Now()`) dengan nilai yang disimpan sebagai
> TIMESTAMPTZ UTC. Tidak ada masalah timezone selama input konsisten.

**DTO `time.Time` → `timeutil.ToPlatform` (untuk output ke UI):**

- `application/dto/auth_dto.go` — `CreatedAt` (39, 200), `CreatedAt`/`UpdatedAt`/`LastLoginAt` (263-265), `CreatedAt`/`UpdatedAt` (302-303), `AssignedAt`/`ExpiredAt`/`DeactivatedAt` (352-356)

> Field INPUT `ExpiredAt` (380, 385) berupa `*time.Time` yang di-parse oleh Gin dari
> JSON. Pastikan frontend mengirim RFC3339 dengan offset (default `+07:00`); jika
> dikirim tanpa offset, Go menganggap UTC. Normalisasi bisa dilakukan di command layer.

---

### 5. Module `feedback`

**DTO `time.Time` → `timeutil.ToPlatform`:**

- `application/dto/feedback_dto.go` — `CreatedAt`/`UpdatedAt` (43-44)
- `application/dto/comment_dto.go` — `CreatedAt`/`UpdatedAt` (26-27)
- `application/dto/attachment_dto.go` — `CreatedAt` (30)

---

### 6. Module `article`

**Logic `time.Now()`:**

| File | Baris | Perubahan |
|------|-------|-----------|
| `interfaces/http/helpers.go` | 11 | Thumbnail path `now.Format("2006/01/02")` → `timeutil.Now()` agar segmen tanggal path sesuai platform TZ |
| `infrastructure/persistence/postgres_article_repo.go` | 250 | Fallback pub date `time.Now()` — boleh tetap (instant), atau `timeutil.Now()` untuk konsistensi |
| `application/command/trigger_scrape.go` | 90 | `UpdateLastScrapedCategory(..., time.Now())` — boleh tetap |

**DTO `time.Time` → `timeutil.ToPlatform`:**

- `application/dto/article_dto.go` — `PublishedAt`/`CreatedAt` (67-68), `PublishedAt`/`ArchivedAt`/`CreatedAt`/`UpdatedAt` (86-89)
- `application/dto/source_dto.go` — `LastScrapedAt` (26, 59), `CreatedAt` (62)

---

### 7. Module `dokumen_aset`

**DTO `time.Time` → `timeutil.ToPlatform`:**

- `application/dto/dokumen_aset_dto.go` — `CreatedAt` (35), `CreatedAt`/`UpdatedAt` (55-56)

---

## Ringkasan Volume Perubahan

| Module | Parse DATE | Format DATE | Format TIMESTAMPTZ | Logic time.Now() | DTO time.Time |
|--------|:---:|:---:|:---:|:---:|:---:|
| keuangan | 16 | 28 | 53 | 6 | 0 |
| kesantrian | 2 | 2 | 0 | 0 | 11 |
| psb | 2 | 5 | 0 | 2 | 10 |
| identity | 0 | 0 | 0 | 0 (sudah benar) | 15 |
| feedback | 0 | 0 | 0 | 0 | 6 |
| article | 0 | 1 | 0 | 2 | 7 |
| dokumen_aset | 0 | 0 | 0 | 0 | 3 |

## Urutan Eksekusi yang Disarankan

1. **keuangan** — volume terbesar & ada logic penomoran (paling rentan bug pergantian tahun).
2. **psb** — logic NIS/NoRegis tahun + inkonsistensi format.
3. **kesantrian** — parsing tanggal lahir Excel.
4. **identity / feedback / article / dokumen_aset** — hanya DTO output, cepat.
5. Tambahkan unit test `timeutil` (parse/format di platform TZ, perilaku fallback UTC).
