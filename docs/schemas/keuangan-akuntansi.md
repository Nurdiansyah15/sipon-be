# Schema: Keuangan & Akuntansi

Dokumen ini merangkum skema DB modul `keuangan` **sebagaimana sudah berjalan** (hasil `migrations/*_create_keuangan_tables.up.sql` + migrasi lanjutannya), lalu bagian **perubahan yang diusulkan** untuk penyempurnaan tagihan/pembayaran/COA/jurnal/periode. Lihat aturan bisnis di baliknya di [`docs/rules/akuntansi.md`](../rules/akuntansi.md).

Catatan penomoran migrasi: nama file migrasi di repo ini memakai prefix timestamp (`20260808213506_create_keuangan_tables.up.sql`, dst), bukan nomor urut manual — golang-migrate membaca berdasarkan urutan nama file, jadi ini tidak memengaruhi apa pun selain penamaan.

---

## 1. Billing (sudah ada)

### `fee_components` — master jenis biaya
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| code | VARCHAR(20) UNIQUE | `SPP`, `UKT`, dst |
| name | VARCHAR(200) | |
| type | VARCHAR(30) CHECK | `ukt`, `spp`, `daftar_ulang`, `insidental` |
| amount | NUMERIC(14,2) | nominal default |
| is_periodic | BOOLEAN | |
| period_type | VARCHAR(20) CHECK nullable | `monthly`, `semesterly`, `yearly`, `once` |
| is_active | BOOLEAN | |
| created_by, created_at, updated_at, deleted_at | | soft delete |

### `billing_schemes`, `billing_scheme_items`, `santri_billing_assignments`
Skema paket komponen biaya per santri. Tidak berubah untuk penyempurnaan ini — lihat `docs/plan/keuangan-module.md` untuk detail kolom.

### `billing_periods` — periode **tagihan** (bukan periode akuntansi!)
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| name | VARCHAR(100) | mis. "Juli 2026" |
| period_type | VARCHAR(20) CHECK | `monthly`, `semesterly`, `yearly`, `once` |
| start_date, end_date | DATE | |
| status | VARCHAR(20) CHECK | `draft`, `open`, `closed` |
| created_by, created_at, updated_at | | |

⚠️ **Sengaja terpisah total dari `accounting_periods`** (lihat §2) — satu untuk siklus tagih-menagih ke santri, satu untuk tutup buku akuntansi. Jangan digabung.

### `invoices`

Tabel ini dibuat di migrasi `*_create_keuangan_tables` dengan kolom `periode`/`tahun_ajaran` (string bebas), lalu migrasi `*_create_billing_periods` menambah `billing_period_id` dan **men-drop** `periode`/`tahun_ajaran` (Fase 1, `docs/plan/perubahan-skema-tagihan.md`). Kolom aktif setelah kedua migrasi berjalan urut:

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| invoice_number | VARCHAR(30) UNIQUE | `INV/{tahun}/{bulan}/{urutan}`, generate lewat `finance_number_sequences` |
| santri_id, user_id | UUID | no FK (cross-module) |
| billing_scheme_id | UUID nullable | audit trail |
| fee_component_id | UUID FK → fee_components | |
| billing_period_id | UUID NOT NULL FK → billing_periods | periode tagihan (bukan periode akuntansi) |
| amount, discount_amount, paid_amount | NUMERIC(14,2) | |
| status | VARCHAR(20) CHECK | `draft`, `issued`, `partial`, `paid`, `expired`, `cancelled` |
| due_date | DATE | |
| issued_at | DATE nullable | |
| created_by, created_at, updated_at, deleted_at | | |

`periode`/`tahun_ajaran` tidak lagi ada di tabel final — jangan dipakai lagi di query/kode baru.

### `payments`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| payment_number | VARCHAR(30) UNIQUE | `PAY/{tahun}/{bulan}/{urutan}` |
| invoice_id | UUID FK → invoices | |
| debit_account_id | UUID nullable, FK → accounts | akun kas/bank tujuan |
| amount | NUMERIC(14,2) | |
| method | VARCHAR(20) CHECK | `transfer`, `cash`, `check` |
| reference_number | VARCHAR(100) nullable | |
| payment_date | DATE | |
| status | VARCHAR(20) CHECK | `pending`, `verified`, `rejected` |
| verified_by, verified_at | nullable | |
| proof_key | VARCHAR(512) nullable | key MinIO bukti transfer |
| created_by, created_at, updated_at | | |

**🔧 Perubahan diusulkan**: `debit_account_id` → **`NOT NULL`**. Alasan di `docs/rules/akuntansi.md` §2.2 — tanpa akun kas/bank tujuan, payment yang diverifikasi tidak bisa diposting ke jurnal.

```sql
-- Backfill dulu kalau ada baris lama tanpa debit_account_id (pilih akun kas default,
-- mis. kode '1101' dari seed COA) — aman dilewati kalau belum ada data produksi:
UPDATE payments SET debit_account_id = (SELECT id FROM accounts WHERE code = '1101')
WHERE debit_account_id IS NULL;

ALTER TABLE payments ALTER COLUMN debit_account_id SET NOT NULL;
```

Sejalan dengan itu, `dto.CreateManualPaymentRequest.DebitAccountID` (`application/dto/payment_dto.go:5`) berubah dari `*string` opsional jadi `string` dengan `binding:"required"`, dan `create_manual_payment.go` meneruskannya sebagai nilai non-pointer ke `payEntity.NewPayment(...)` (perhatikan signature `NewPayment` saat ini menerima `debitAccountID *string` — kalau field DTO jadi wajib, tetap boleh dikonversi ke pointer sebelum diteruskan, yang penting validasi "tidak boleh kosong" terjadi di DTO/handler, bukan cuma di constructor entity).

### `invoice_adjustments` — tidak berubah.

---

## 2. Akuntansi (sudah ada, sebagian dead code)

### `accounts` — Chart of Accounts
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| code | VARCHAR(20) UNIQUE | `1101`, dst |
| name | VARCHAR(200) | |
| type | VARCHAR(20) CHECK | `asset`, `liability`, `equity`, `revenue`, `expense` |
| parent_id | UUID nullable, self-FK | null = root |
| level | INTEGER | 0 = root |
| is_postable | BOOLEAN | hanya leaf yang boleh terima jurnal |
| normal_balance | VARCHAR(10) CHECK | `debit`/`credit` |
| is_active, is_system | BOOLEAN | |
| created_by, created_at, updated_at, deleted_at | | |

Tidak berubah. Lihat daftar akun seed lengkap di `docs/plan/keuangan-module.md` §"Default Seed Accounts" — pastikan `3200 Saldo Laba` dan `3201 Laba Tahun Berjalan` benar-benar ter-seed karena proses closing (§3 rules doc) bergantung padanya.

### `accounting_periods` — periode **akuntansi** (tutup buku)
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| name | VARCHAR(100) | |
| start_date, end_date | DATE | dijaga tidak overlap oleh trigger `check_period_overlap` |
| status | VARCHAR(20) CHECK | `open`, `closing`, `closed`, `locked` |
| closed_by, closed_at | nullable | |
| created_by, created_at, updated_at | | |

**🔧 Perubahan diusulkan**: hapus nilai `closing` dari CHECK constraint status, sisakan `open`/`closed`/`locked`. Alasan: status `closing` & method domain `StartClosing()` tidak pernah dipakai — closing dilakukan sebagai satu operasi atomik (§3.2 rules doc), tidak butuh status transisi terpisah. Ini juga menyederhanakan state machine yang perlu dijaga konsistensinya.

```sql
ALTER TABLE accounting_periods DROP CONSTRAINT accounting_periods_status_check;
ALTER TABLE accounting_periods ADD CONSTRAINT accounting_periods_status_check
    CHECK (status IN ('open', 'closed', 'locked'));
```

### `journal_entries`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| journal_number | VARCHAR(30) UNIQUE | `{PREFIX}/{tahun}/{bulan}/{urutan}` |
| entry_date | DATE | |
| description | TEXT | |
| source_type | VARCHAR(30) nullable CHECK | `invoice_issued`, `payment_verified`, `invoice_cancelled`, `adjustment`, `closing`, `manual` |
| source_id | UUID nullable | ID entity pemicu |
| period_id | UUID NOT NULL FK → accounting_periods | |
| total_debit, total_credit | NUMERIC(16,2) | `CHECK (total_debit = total_credit)` |
| posted_by | UUID | |
| posted_at | TIMESTAMPTZ nullable | |
| status | VARCHAR(20) CHECK | `draft`, `posted`, `cancelled` |
| created_at, updated_at | | |

**Tidak ada kolom `deleted_at`** — catat ini karena empat query laporan (`report_ledger.go`, `report_trial_balance.go`, `report_income_statement.go`, `report_balance_sheet.go`) saat ini mem-filter `WHERE je.deleted_at IS NULL`, yang berarti semuanya **akan gagal dieksekusi** (kolom tidak ada). Lihat [bug: 4 laporan query kolom deleted_at yang tidak ada](../bugs/akuntansi-laporan-jurnal-kolom-deleted-at-tidak-ada.md).

**🔧 Perubahan diusulkan**: tambah **unique partial index** untuk mencegah double-posting dari sumber otomatis:

```sql
CREATE UNIQUE INDEX idx_journal_source_unique
    ON journal_entries(source_type, source_id)
    WHERE source_type IS NOT NULL AND source_type != 'manual' AND status != 'cancelled';
```

Ini menjadikan aturan "satu invoice/payment/pembatalan → maksimal satu jurnal" (rules doc §2) ditegakkan juga di level DB, bukan cuma disiplin aplikasi — konsisten dengan pola yang sudah dipakai untuk `finance_number_sequences` dkk (jaga uniqueness di DB, bukan cuma cek-lalu-insert di aplikasi yang rawan race condition).

### `journal_entry_lines`
| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID PK | |
| journal_entry_id | UUID FK → journal_entries ON DELETE CASCADE | |
| account_id | UUID FK → accounts | |
| account_code | VARCHAR(20) | denormalized (read performance laporan) |
| description | TEXT nullable | |
| debit, credit | NUMERIC(16,2) | `CHECK (debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0)` |

Sama seperti header-nya, **tidak ada kolom `deleted_at`** di tabel ini — lihat bug terkait di atas.

### `finance_number_sequences` (sudah ada, hasil sesi sebelumnya)
| Kolom | Tipe | Keterangan |
|---|---|---|
| doc_type | VARCHAR(20) | `invoice`, `payment`, ... |
| year | INTEGER | |
| seq | INTEGER | |
| PK | `(doc_type, year)` | |

**🔧 Perubahan diusulkan**: pakai tabel yang sama untuk nomor jurnal — tambah pemanggilan dengan `doc_type='journal'` di `AutoPostingService`/`CreateManualJournalUseCase`, gantikan `generateJournalNumber(seq, ...)` yang saat ini menerima `seq` mentah dari pemanggil (rawan pola bug yang sama seperti nomor invoice/payment sebelum diperbaiki — lihat [bug: nomor jurnal manual selalu 1](../bugs/akuntansi-manual-journal-nomor-selalu-1.md)). Prefix kosmetik (`INV/PAY/CNL/ADJ/CLS/JRN`) tetap bisa dipertahankan di depan nomor, tidak perlu `doc_type` terpisah per prefix — cukup satu deret angka global untuk semua jurnal per tahun, sesuai prinsip "satu tabel ringkas" yang sudah dipakai.

### `finance_audit_logs` (sudah ada, dead table)
Tidak berubah strukturnya. Lihat rules doc §4.8 untuk keputusan yang perlu diambil (pakai atau dokumentasikan sebagai belum dipakai).

---

## 3. Ringkasan Perubahan Skema yang Diusulkan

| # | Tabel | Perubahan | Alasan |
|---|---|---|---|
| 1 | `payments` | `debit_account_id` jadi `NOT NULL` | Wajib ada akun kas/bank sebelum payment bisa diposting ke jurnal saat verifikasi |
| 2 | `accounting_periods` | Hapus status `closing` dari CHECK, sisakan `open`/`closed`/`locked` | `closing` tidak pernah dipakai; closing disederhanakan jadi satu operasi atomik |
| 3 | `journal_entries` | Tambah `UNIQUE(source_type, source_id) WHERE source_type != 'manual' AND status != 'cancelled'` | Cegah double-posting jurnal otomatis di level DB |
| 4 | `journal_entries`, `journal_entry_lines` | **Tidak** menambah `deleted_at` — sebaliknya, perbaiki `report_ledger.go`/`report_trial_balance.go` supaya tidak mereferensikan kolom itu | Kolom memang tidak seharusnya ada — jurnal yang sudah posted tidak pernah di-soft-delete, hanya `status='cancelled'` |
| 5 | `finance_number_sequences` | Tidak perlu perubahan struktur — tambah pemakaian `doc_type='journal'` | Reuse tabel yang sudah ada, hindari tabel counter baru |

Migrasi baru untuk perubahan #1–#3 bisa digabung jadi satu file (`..._refine_akuntansi.up/down.sql`) karena semuanya bagian dari paket penyempurnaan yang sama.
