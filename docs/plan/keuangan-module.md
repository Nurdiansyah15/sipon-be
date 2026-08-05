# Plan: Module Keuangan (Billing + Akuntansi)

## Context

Sistem SIPON saat ini sudah memiliki modul `identity` (autentikasi & RBAC), `kesantrian` (data santri), `psb` (penerimaan santri baru), `dokumen_aset`, dan `article`. Namun belum ada modul yang menangani **tagihan/pembayaran keuangan santri** maupun **pencatatan akuntansi**.

Tujuan plan ini: menambah modul baru `keuangan` yang meng-cover dua bounded context:

1. **Billing** — komponen biaya, skema tagihan, invoice per santri, pembayaran manual, potongan/beasiswa, dan pelaporan tagihan.
2. **Akuntansi** — chart of accounts (COA), jurnal double-entry (auto-posting dari billing + manual entry), periode pembukuan (closing/reopen/lock), dan laporan keuangan (buku besar, neraca saldo, neraca, laba rugi).

Modul ini mengikuti arsitektur modular monolith DDD yang sama dengan modul existing (lihat `docs/architecture/module-boundaries.md`).

---

## Keputusan Arsitektur Kunci

1. **Satu modul `keuangan`**, dua bounded context di dalamnya (billing & accounting). Dipisah menjadi satu modul karena keduanya terikat secara transaksional — setiap aksi billing yang terverifikasi langsung memicu auto-posting jurnal dalam transaksi yang sama.
2. **Cross-module via Contract** — `keuangan` memanggil `identity.Contract` (UserSummary) dan `kesantrian.Contract` (data santri). `kesantrian` perlu menambah 2 method baru di Contract-nya. `keuangan` sendiri mengekspos `Contract` untuk module lain (mis. PSB butuh cek tagihan daftar ulang sebelum generate NIS).
3. **Tidak ada FK ke tabel module lain** — `santri_id` dan `user_id` di tabel keuangan adalah plain UUID tanpa FK, konsisten dengan pola existing (lihat migration 005 dan 006).
4. **Accrual basis** — pendapatan diakui saat invoice issued (bukan saat pembayaran diterima). Piutang santri dicatat saat tagihan diterbitkan.
5. **Manual payment di Fase 1** — bendahara input pembayaran secara manual. Payment gateway (VA, auto-verify) ditunda ke Fase 2.
6. **Skema tagihan (billing scheme)** — grouping komponen pembayaran yang menentukan komponen apa saja yang menjadi tagihan seorang santri. Tidak ada penggolongan berdasarkan pendapatan wali. Contoh: "Skema Reguler" → SPP + UKT, "Skema Spesial" → SPP + UKT + Kitab + Kursus.
7. **Tidak ada denda keterlambatan** — `due_date` adalah lifecycle document. Invoice yang melewati `due_date` berstatus `expired`.
8. **Tidak ada refund** — ditunda sampai ada kebutuhan bisnis yang jelas.
9. **Multi-tahun ajaran** — satu santri bisa punya tagihan dari tahun ajaran sebelumnya yang belum lunas.
10. **COA mengikuti teori akuntansi** — 5 golongan utama (Aset, Kewajiban, Ekuitas, Pendapatan, Beban) sebagai root. User bisa menambah akun child di bawahnya.
11. **Periode pembukuan** — ada opsi close, reopen, dan lock. Closing otomatis membuat jurnal pemindahan saldo revenue/expense ke saldo laba.
12. **PDF kwitansi** — diperlukan di Fase 1.

---

## Aktor

| Aktor | Peran |
|---|---|
| **Bendahara** (custom role, subset permission) | CRUD komponen biaya & skema, generate tagihan, catat pembayaran, kelola beasiswa/diskon, jurnal manual, closing periode |
| **Santri/Wali** (role: member) | Lihat tagihan, lihat riwayat pembayaran |
| **Admin PSB/Kesantrian** | Trigger verifikasi keuangan saat daftar ulang |
| **System** | Auto-expire invoice yang melewati jatuh tempo |

---

## Modul: Billing

### Entity: KomponenBiaya (fee_component)

Master data jenis tagihan.

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| Code | VARCHAR(20) | unique, mis. `SPP`, `UKT`, `DU` |
| Name | VARCHAR(200) | nama komponen |
| Type | ENUM | `ukt`, `spp`, `daftar_ulang`, `insidental` |
| Amount | NUMERIC(14,2) | nominal default |
| IsPeriodic | BOOLEAN | apakah berulang |
| PeriodType | ENUM nullable | `monthly`, `semesterly`, `yearly`, `once` |
| Description | TEXT | |
| IsActive | BOOLEAN | |
| CreatedBy | UUID | |
| CreatedAt, UpdatedAt | TIMESTAMPTZ | |
| DeletedAt | TIMESTAMPTZ nullable | soft delete |

### Entity: SkemaTagihan (billing_scheme)

Paket/grouping komponen pembayaran yang menentukan komponen apa saja yang jadi tagihan seorang santri.

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| Name | VARCHAR(100) | mis. "Reguler", "Spesial" |
| Description | TEXT | |
| IsActive | BOOLEAN | |
| CreatedBy | UUID | |
| CreatedAt, UpdatedAt | TIMESTAMPTZ | |

### Entity: SkemaItem (billing_scheme_item)

Join table antara skema dan komponen, dengan opsi override nominal.

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| BillingSchemeID | UUID | FK ke billing_schemes |
| FeeComponentID | UUID | FK ke fee_components |
| AmountOverride | NUMERIC(14,2) nullable | jika beda dari default fee_component.amount |
| IsRequired | BOOLEAN | wajib dibayar atau tidak |
| SortOrder | INTEGER | urutan |
| CreatedAt | TIMESTAMPTZ | |

Unique constraint: `(billing_scheme_id, fee_component_id)`.

### Entity: SantriBillingAssignment

Mapping santri ke skema. Satu santri aktif hanya punya 1 skema pada satu waktu.

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| SantriID | UUID | no FK (cross-module) |
| BillingSchemeID | UUID | FK ke billing_schemes |
| EffectiveFrom | DATE | mulai berlaku |
| EffectiveUntil | DATE nullable | sampai kapan (null = indefinite) |
| AssignedBy | UUID | |
| CreatedAt | TIMESTAMPTZ | |

### Entity: Tagihan (invoice)

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| InvoiceNumber | VARCHAR(30) | unique, auto-generated: `INV/{tahun}/{bulan}/{urutan}` |
| SantriID | UUID | no FK (cross-module) |
| UserID | UUID | no FK (cross-module) |
| BillingSchemeID | UUID nullable | audit trail |
| FeeComponentID | UUID | FK ke fee_components |
| Periode | VARCHAR(20) | mis. `2026-1` (semester), `2026-07` (bulanan) |
| TahunAjaran | VARCHAR(10) | mis. `2025/2026` |
| Amount | NUMERIC(14,2) | |
| DiscountAmount | NUMERIC(14,2) | default 0 |
| PaidAmount | NUMERIC(14,2) | default 0 |
| Status | VARCHAR(20) | `draft`, `issued`, `partial`, `paid`, `expired`, `cancelled` |
| DueDate | DATE | |
| IssuedAt | DATE nullable | |
| Notes | TEXT | |
| CreatedBy | UUID | |
| CreatedAt, UpdatedAt | TIMESTAMPTZ | |
| DeletedAt | TIMESTAMPTZ nullable | |

Status transitions:

```
draft → issued → partial → paid
                → expired (due_date lewat, auto atau manual)
       → cancelled (hanya dari draft/issued tanpa payment verified)
```

Unique constraint: `(santri_id, fee_component_id, periode)` kecuali status = `cancelled`.

### Entity: Pembayaran (payment)

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| PaymentNumber | VARCHAR(30) | unique, auto-generated: `PAY/{tahun}/{bulan}/{urutan}` |
| InvoiceID | UUID | FK ke invoices |
| DebitAccountID | UUID nullable | FK ke accounts (kas/bank mana yang menerima) |
| Amount | NUMERIC(14,2) | |
| Method | VARCHAR(20) | `transfer`, `cash`, `check` |
| ReferenceNumber | VARCHAR(100) | no. referensi bank |
| PaymentDate | DATE | |
| Status | VARCHAR(20) | `pending`, `verified`, `rejected` |
| VerifiedBy | UUID nullable | |
| VerifiedAt | TIMESTAMPTZ nullable | |
| Notes | TEXT | |
| ProofKey | VARCHAR(512) | MinIO key bukti transfer |
| CreatedBy | UUID | |
| CreatedAt, UpdatedAt | TIMESTAMPTZ | |

### Entity: Potongan (invoice_adjustment)

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| InvoiceID | UUID | FK ke invoices |
| Type | VARCHAR(20) | `beasiswa`, `diskon`, `penyesuaian` |
| Amount | NUMERIC(14,2) | nominal (salah satu dengan Percentage) |
| Percentage | NUMERIC(5,2) nullable | persentase (salah satu dengan Amount) |
| Description | TEXT | |
| AppliedBy | UUID | |
| AppliedAt | TIMESTAMPTZ | |

### Business Rules Billing

1. Satu santri, satu skema aktif pada satu waktu.
2. Invoice unik per (santri, komponen, periode) — tidak boleh duplikasi kecuali yang cancelled.
3. `paid_amount` bertambah hanya saat payment status = `verified`.
4. Invoice `paid` jika `paid_amount >= (amount - discount_amount)`.
5. Invoice `expired` jika `due_date < today AND status IN ('issued', 'partial')`.
6. Invoice `cancelled` hanya jika status = `draft` atau `issued` tanpa payment verified.
7. Generate batch: loop semua santri dalam skema × semua komponen di skema → buat invoice, skip yang sudah ada.
8. Due date adalah lifecycle — tidak ada denda otomatis.

---

## Modul: Akuntansi

### Entity: Akun (account) — Chart of Accounts

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| Code | VARCHAR(20) | unique, mis. `1101` |
| Name | VARCHAR(200) | mis. "Kas" |
| Type | VARCHAR(20) | `asset`, `liability`, `equity`, `revenue`, `expense` |
| ParentID | UUID nullable | FK ke accounts (self-referencing), null = root |
| Level | INTEGER | 0=root, 1+=child |
| IsPostable | BOOLEAN | hanya akun leaf yang bisa menerima jurnal |
| NormalBalance | VARCHAR(10) | `debit` atau `credit` |
| Description | TEXT | |
| IsActive | BOOLEAN | |
| IsSystem | BOOLEAN | default accounts dari seed, tidak bisa dihapus |
| CreatedBy | UUID | |
| CreatedAt, UpdatedAt | TIMESTAMPTZ | |
| DeletedAt | TIMESTAMPTZ nullable | |

**Default Seed Accounts:**

```
1000  ASET                         (root, header, is_system)
1100  Aset Lancar                  (group)
1101  Kas                          (postable, debit)
1102  Bank                         (postable, debit)
1103  Piutang Santri               (postable, debit)
1200  Aset Tetap                   (group)
1201  Tanah                        (postable, debit)
1202  Bangunan                     (postable, debit)
1203  Peralatan                    (postable, debit)
1204  Kendaraan                    (postable, debit)

2000  KEWAJIBAN                    (root, header, is_system)
2100  Kewajiban Lancar             (group)
2101  Utang Usaha                  (postable, credit)
2102  Uang Muka Santri             (postable, credit)
2103  Biaya Diterima Dimuka        (postable, credit)

3000  EKUITAS                      (root, header, is_system)
3100  Modal                        (postable, credit)
3200  Saldo Laba                   (postable, credit)
3201  Laba Tahun Berjalan          (postable, credit)

4000  PENDAPATAN                   (root, header, is_system)
4100  Pendapatan SPP               (postable, credit)
4200  Pendapatan UKT               (postable, credit)
4300  Pendapatan Daftar Ulang      (postable, credit)
4400  Pendapatan Insidental        (postable, credit)
4500  Pendapatan Donasi            (postable, credit)
4600  Pendapatan Lainnya          (postable, credit)

5000  BEBAN                        (root, header, is_system)
5100  Beban Gaji                   (postable, debit)
5200  Beban Operasional            (postable, debit)
5300  Beban Pemeliharaan           (postable, debit)
5400  Beban Penyusutan             (postable, debit)
5500  Beban Lainnya               (postable, debit)
```

Business rules:
- Root accounts (level 0, is_system=true) tidak bisa dihapus, hanya rename.
- Akun non-postable (header/group) tidak bisa menerima entry jurnal.
- User bisa menambah akun di bawah group manapun, kode harus unik.
- Akun yang sudah punya jurnal tidak bisa dihapus, hanya dinonaktifkan (IsActive=false).

### Entity: Jurnal (journal_entry)

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| JournalNumber | VARCHAR(30) | unique, `JRN/{tahun}/{bulan}/{urutan}` |
| EntryDate | DATE | |
| Description | TEXT | |
| SourceType | VARCHAR(30) nullable | `invoice_issued`, `payment_verified`, `invoice_cancelled`, `adjustment`, `closing`, `manual` |
| SourceID | UUID nullable | ID entity pemicu |
| PeriodID | UUID | FK ke accounting_periods |
| TotalDebit | NUMERIC(16,2) | harus = TotalCredit |
| TotalCredit | NUMERIC(16,2) | harus = TotalDebit |
| PostedBy | UUID | |
| PostedAt | TIMESTAMPTZ nullable | |
| Status | VARCHAR(20) | `draft`, `posted`, `cancelled` |
| CreatedAt, UpdatedAt | TIMESTAMPTZ | |

Business rules:
- TotalDebit = TotalCredit (CHECK constraint + validasi domain).
- Minimal 2 lines (satu debit, satu kredit).
- Hanya akun postable & active yang bisa menerima entry.
- Tidak bisa posting ke periode yang sudah closed/locked.
- Jurnal auto-generated tidak bisa di-cancel langsung — harus cancel dari sumbernya.
- Jurnal manual bisa di-cancel (mark as cancelled, exclude dari laporan).
- Cancel tidak membuat reversal journal — cukup mark cancelled untuk audit trail.

### Entity: Jurnal Item (journal_entry_line)

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| JournalEntryID | UUID | FK ke journal_entries |
| AccountID | UUID | FK ke accounts |
| AccountCode | VARCHAR(20) | denormalized untuk read performance |
| Description | TEXT nullable | |
| Debit | NUMERIC(16,2) | |
| Credit | NUMERIC(16,2) | |

CHECK constraint: setiap line harus debit > 0 XOR credit > 0 (tidak boleh keduanya).

### Entity: Periode Akuntansi (accounting_period)

| Atribut | Tipe | Keterangan |
|---|---|---|
| ID | UUID | PK |
| Name | VARCHAR(100) | mis. "Juli 2026", "Semester Ganjil 2025/2026" |
| StartDate | DATE | |
| EndDate | DATE | |
| Status | VARCHAR(20) | `open`, `closing`, `closed`, `locked` |
| ClosedBy | UUID nullable | |
| ClosedAt | TIMESTAMPTZ nullable | |
| CreatedBy | UUID | |
| CreatedAt, UpdatedAt | TIMESTAMPTZ | |

Status transitions:

```
open → closing → closed → locked
                  ↑
                  └── reopen (hapus closing entries, kembali ke open)
```

- `open`: boleh posting jurnal.
- `closing`: sedang proses closing.
- `closed`: closing entries sudah dibuat, tidak boleh posting. Bisa di-reopen.
- `locked`: permanen, tidak bisa reopen.

Trigger check: tidak boleh ada overlapping periods.

---

## Auto-Posting Rules

Semua auto-posting terjadi dalam **satu transaksi database** dengan operasi billing.

### Rule 1: Invoice ISSUED → Pengakuan Piutang & Pendapatan

```
Dr. Piutang Santri (1103)            = (amount - discount_amount)
  Cr. Pendapatan [type-mapped]       = (amount - discount_amount)
```

Mapping `fee_component.type` → `account.code`:

| Type | Account Code |
|---|---|
| `spp` | `4100` |
| `ukt` | `4200` |
| `daftar_ulang` | `4300` |
| `insidental` | `4400` |

### Rule 2: Payment VERIFIED → Penerimaan Kas

```
Dr. {debit_account_id} (pilihan bendahara)    = amount
  Cr. Piutang Santri (1103)                   = amount
```

Bendahara memilih akun debit saat create/verify payment dari daftar akun asset yang postable (Kas, Bank, dll).

### Rule 3: Invoice CANCELLED → Reversal

```
Dr. Pendapatan [type-mapped]       = original amount
  Cr. Piutang Santri (1103)        = original amount
```

Hanya jika invoice sudah pernah issued (sudah ada jurnal sebelumnya).

### Rule 4: Closing Periode

```
Dr. Semua akun Revenue (4xxx) dengan saldo > 0
  Cr. Laba Tahun Berjalan (3201)

Dr. Laba Tahun Berjalan (3201)
  Cr. Semua akun Expense (5xxx) dengan saldo > 0

Jika laba:
  Dr. Laba Tahun Berjalan (3201)
    Cr. Saldo Laba (3200)

Jika rugi:
  Dr. Saldo Laba (3200)
    Cr. Laba Tahun Berjalan (3201)
```

---

## Event dan Integrasi Cross-Module

### `keuangan.Contract` (untuk module lain)

```go
type Contract interface {
    GetOutstandingInvoices(ctx context.Context, santriID string) (*OutstandingSummary, error)
    HasPaidComponent(ctx context.Context, santriID, componentCode, periode string) (bool, error)
}

type OutstandingSummary struct {
    HasOutstanding   bool
    TotalOutstanding float64
    Count            int
}
```

PSB bisa panggil `HasPaidComponent` untuk verifikasi santri sudah bayar daftar ulang sebelum generate NIS.

### Perubahan ke `kesantrian.Contract`

Perlu tambah 2 method baru:

```go
type SantriBasicInfo struct {
    SantriID string
    UserID   string
    NIS      *string
    Status   string
}

// Tambah ke Contract interface:
ListActiveSantriIDs(ctx context.Context) ([]string, error)
GetSantriByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error)
```

### Perubahan ke `identity`

Tambah 6 permission keys baru di `permission_constant.go`.

---

## API Endpoints

### Billing — Admin (`manage_keuangan`)

| Method | Path | Deskripsi |
|---|---|---|
| GET | `/api/v1/web/keuangan/admin/components` | List komponen biaya |
| POST | `/api/v1/web/keuangan/admin/components` | Buat komponen biaya |
| PUT | `/api/v1/web/keuangan/admin/components/:id` | Update komponen |
| DELETE | `/api/v1/web/keuangan/admin/components/:id` | Nonaktifkan komponen |
| GET | `/api/v1/web/keuangan/admin/schemes` | List skema tagihan |
| POST | `/api/v1/web/keuangan/admin/schemes` | Buat skema |
| PUT | `/api/v1/web/keuangan/admin/schemes/:id` | Update skema |
| DELETE | `/api/v1/web/keuangan/admin/schemes/:id` | Nonaktifkan skema |
| POST | `/api/v1/web/keuangan/admin/schemes/:id/items` | Tambah komponen ke skema |
| DELETE | `/api/v1/web/keuangan/admin/schemes/:id/items/:itemId` | Hapus komponen dari skema |
| POST | `/api/v1/web/keuangan/admin/assignments` | Assign skema ke santri |
| GET | `/api/v1/web/keuangan/admin/invoices` | List invoice (filter: status, periode, tahun ajaran, santri) |
| POST | `/api/v1/web/keuangan/admin/invoices` | Buat invoice individual |
| POST | `/api/v1/web/keuangan/admin/invoices/batch` | Generate tagihan massal |
| GET | `/api/v1/web/keuangan/admin/invoices/:id` | Detail invoice |
| POST | `/api/v1/web/keuangan/admin/invoices/:id/cancel` | Batal invoice |
| POST | `/api/v1/web/keuangan/admin/invoices/:id/adjustment` | Tambah diskon/beasiswa |
| GET | `/api/v1/web/keuangan/admin/payments` | List pembayaran |
| GET | `/api/v1/web/keuangan/admin/payments/:id` | Detail pembayaran |
| POST | `/api/v1/web/keuangan/admin/payments/manual` | Catat pembayaran manual |
| POST | `/api/v1/web/keuangan/admin/payments/:id/verify` | Verifikasi pembayaran |
| POST | `/api/v1/web/keuangan/admin/payments/:id/reject` | Tolak pembayaran |
| GET | `/api/v1/web/keuangan/admin/payments/:id/receipt` | Download kwitansi PDF |

### Billing — Santri (authenticated)

| Method | Path | Deskripsi |
|---|---|---|
| GET | `/api/v1/web/keuangan/invoices` | Tagihan saya |
| GET | `/api/v1/web/keuangan/invoices/:id` | Detail tagihan |
| GET | `/api/v1/web/keuangan/payments` | Riwayat pembayaran saya |

### Accounting — Admin

| Method | Path | Permission |
|---|---|---|
| GET | `/api/v1/web/keuangan/admin/accounts` | `manage_accounts` |
| GET | `/api/v1/web/keuangan/admin/accounts/:id` | `manage_accounts` |
| POST | `/api/v1/web/keuangan/admin/accounts` | `manage_accounts` |
| PUT | `/api/v1/web/keuangan/admin/accounts/:id` | `manage_accounts` |
| DELETE | `/api/v1/web/keuangan/admin/accounts/:id` | `manage_accounts` |
| GET | `/api/v1/web/keuangan/admin/journal-entries` | `manage_journal` |
| GET | `/api/v1/web/keuangan/admin/journal-entries/:id` | `manage_journal` |
| POST | `/api/v1/web/keuangan/admin/journal-entries` | `manage_journal` |
| POST | `/api/v1/web/keuangan/admin/journal-entries/:id/cancel` | `manage_journal` |
| GET | `/api/v1/web/keuangan/admin/periods` | `close_period` |
| GET | `/api/v1/web/keuangan/admin/periods/active` | `close_period` |
| POST | `/api/v1/web/keuangan/admin/periods` | `close_period` |
| POST | `/api/v1/web/keuangan/admin/periods/:id/close` | `close_period` |
| POST | `/api/v1/web/keuangan/admin/periods/:id/reopen` | `close_period` |
| POST | `/api/v1/web/keuangan/admin/periods/:id/lock` | `close_period` |

### Reports — Admin (`view_keuangan_reports`)

| Method | Path | Deskripsi |
|---|---|---|
| GET | `/api/v1/web/keuangan/admin/reports/summary` | Rekap tagihan & pembayaran per periode |
| GET | `/api/v1/web/keuangan/admin/reports/outstanding` | Tunggakan per santri |
| GET | `/api/v1/web/keuangan/admin/reports/ledger` | Buku besar per akun |
| GET | `/api/v1/web/keuangan/admin/reports/trial-balance` | Neraca saldo |
| GET | `/api/v1/web/keuangan/admin/reports/balance-sheet` | Neraca |
| GET | `/api/v1/web/keuangan/admin/reports/income-statement` | Laporan laba rugi |

---

## Hak Akses

### Permission baru (tambah ke `permission_constant.go`)

```go
PermissionManageKeuangan      PermissionKey = "manage_keuangan"
PermissionVerifyPayment       PermissionKey = "verify_payment"
PermissionViewKeuanganReports PermissionKey = "view_keuangan_reports"
PermissionManageAccounts      PermissionKey = "manage_accounts"
PermissionManageJournal       PermissionKey = "manage_journal"
PermissionClosePeriod         PermissionKey = "close_period"
```

### Mapping ke system roles

```go
// usergod, superadmin, admin — semua mendapat:
PermissionManageKeuangan,
PermissionVerifyPayment,
PermissionViewKeuanganReports,
PermissionManageAccounts,
PermissionManageJournal,
PermissionClosePeriod,

// member — tidak ada permission keuangan (cuma akses endpoint santri biasa)
```

Custom role "bendahara" bisa dibuat nanti dan di-assign subset permissions.

---

## Laporan

| Laporan | Deskripsi | Query Params |
|---|---|---|
| **Rekap Tagihan & Pembayaran** | Total tagihan, terbayar, tunggakan per periode | `tahun_ajaran`, `periode` |
| **Tunggakan per Santri** | Outstanding balance per santri | `tahun_ajaran`, `status` |
| **Buku Besar (Ledger)** | Saldo per akun untuk periode tertentu, transaksi per transaksi | `account_id`, `period_id` |
| **Neraca Saldo (Trial Balance)** | Semua akun + saldo debit/kredit, total debit = total kredit | `period_id` |
| **Neraca (Balance Sheet)** | Aset = Kewajiban + Ekuitas pada tanggal tertentu | `period_id` atau `as_of_date` |
| **Laporan Laba Rugi (Income Statement)** | Pendapatan - Beban = Laba/Rugi untuk periode | `period_id` |

---

## Database Design

Migration file: `012_create_keuangan_tables.up.sql`

```sql
-- ============================================================
-- BILLING TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS fee_components (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(20) NOT NULL UNIQUE,
    name        VARCHAR(200) NOT NULL,
    type        VARCHAR(30) NOT NULL CHECK (type IN ('ukt', 'spp', 'daftar_ulang', 'insidental')),
    amount      NUMERIC(14,2) NOT NULL DEFAULT 0,
    is_periodic BOOLEAN NOT NULL DEFAULT false,
    period_type VARCHAR(20) CHECK (period_type IN ('monthly', 'semesterly', 'yearly', 'once')),
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);
CREATE INDEX idx_fee_components_type ON fee_components(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_fee_components_active ON fee_components(is_active) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS billing_schemes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_by  UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_scheme_items (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    billing_scheme_id UUID NOT NULL REFERENCES billing_schemes(id) ON DELETE CASCADE,
    fee_component_id  UUID NOT NULL REFERENCES fee_components(id),
    amount_override   NUMERIC(14,2),
    is_required       BOOLEAN NOT NULL DEFAULT true,
    sort_order        INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(billing_scheme_id, fee_component_id)
);
CREATE INDEX idx_bsi_scheme ON billing_scheme_items(billing_scheme_id);

-- No FK to santri — cross-module reference.
CREATE TABLE IF NOT EXISTS santri_billing_assignments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id         UUID NOT NULL,
    billing_scheme_id UUID NOT NULL REFERENCES billing_schemes(id),
    effective_from    DATE NOT NULL,
    effective_until   DATE,
    assigned_by       UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sba_santri ON santri_billing_assignments(santri_id);
CREATE INDEX idx_sba_active ON santri_billing_assignments(santri_id, effective_from)
    WHERE effective_until IS NULL OR effective_until >= CURRENT_DATE;

-- No FK to santri/users — cross-module references.
CREATE TABLE IF NOT EXISTS invoices (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number    VARCHAR(30) NOT NULL UNIQUE,
    santri_id         UUID NOT NULL,
    user_id           UUID NOT NULL,
    billing_scheme_id UUID,
    fee_component_id  UUID NOT NULL REFERENCES fee_components(id),
    periode           VARCHAR(20) NOT NULL,
    tahun_ajaran      VARCHAR(10) NOT NULL,
    amount            NUMERIC(14,2) NOT NULL,
    discount_amount   NUMERIC(14,2) NOT NULL DEFAULT 0,
    paid_amount       NUMERIC(14,2) NOT NULL DEFAULT 0,
    status            VARCHAR(20) NOT NULL DEFAULT 'draft'
                      CHECK (status IN ('draft', 'issued', 'partial', 'paid', 'expired', 'cancelled')),
    due_date          DATE NOT NULL,
    issued_at         DATE,
    notes             TEXT,
    created_by        UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);
CREATE UNIQUE INDEX idx_invoices_number ON invoices(invoice_number) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_santri ON invoices(santri_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_user ON invoices(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_status ON invoices(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_tahun_ajaran ON invoices(tahun_ajaran) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_periode ON invoices(periode) WHERE deleted_at IS NULL;
CREATE INDEX idx_invoices_due_date ON invoices(due_date) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_invoices_unique_periode
    ON invoices(santri_id, fee_component_id, periode)
    WHERE deleted_at IS NULL AND status NOT IN ('cancelled');

CREATE TABLE IF NOT EXISTS payments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_number    VARCHAR(30) NOT NULL UNIQUE,
    invoice_id        UUID NOT NULL REFERENCES invoices(id),
    debit_account_id  UUID,
    amount            NUMERIC(14,2) NOT NULL,
    method            VARCHAR(20) NOT NULL CHECK (method IN ('transfer', 'cash', 'check')),
    reference_number  VARCHAR(100),
    payment_date      DATE NOT NULL,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending'
                      CHECK (status IN ('pending', 'verified', 'rejected')),
    verified_by       UUID,
    verified_at       TIMESTAMPTZ,
    notes             TEXT,
    proof_key         VARCHAR(512),
    created_by        UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_payments_number ON payments(payment_number);
CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_date ON payments(payment_date);

CREATE TABLE IF NOT EXISTS invoice_adjustments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id    UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    type          VARCHAR(20) NOT NULL CHECK (type IN ('beasiswa', 'diskon', 'penyesuaian')),
    amount        NUMERIC(14,2) NOT NULL DEFAULT 0,
    percentage    NUMERIC(5,2),
    description   TEXT,
    applied_by    UUID NOT NULL,
    applied_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_adjustments_invoice ON invoice_adjustments(invoice_id);

-- ============================================================
-- ACCOUNTING TABLES
-- ============================================================

CREATE TABLE IF NOT EXISTS accounts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code           VARCHAR(20) NOT NULL UNIQUE,
    name           VARCHAR(200) NOT NULL,
    type           VARCHAR(20) NOT NULL CHECK (type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    parent_id      UUID REFERENCES accounts(id),
    level          INTEGER NOT NULL DEFAULT 0,
    is_postable    BOOLEAN NOT NULL DEFAULT false,
    normal_balance VARCHAR(10) NOT NULL CHECK (normal_balance IN ('debit', 'credit')),
    description    TEXT,
    is_active      BOOLEAN NOT NULL DEFAULT true,
    is_system      BOOLEAN NOT NULL DEFAULT false,
    created_by     UUID NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);
CREATE INDEX idx_accounts_type ON accounts(type) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_parent ON accounts(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_accounts_postable ON accounts(is_postable) WHERE deleted_at IS NULL AND is_active = true;

CREATE TABLE IF NOT EXISTS accounting_periods (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(100) NOT NULL,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'open'
               CHECK (status IN ('open', 'closing', 'closed', 'locked')),
    closed_by  UUID,
    closed_at  TIMESTAMPTZ,
    created_by UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_periods_status ON accounting_periods(status);
CREATE INDEX idx_periods_date_range ON accounting_periods(start_date, end_date);

CREATE OR REPLACE FUNCTION check_period_overlap()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM accounting_periods
        WHERE id != NEW.id
          AND start_date <= NEW.end_date
          AND end_date >= NEW.start_date
    ) THEN
        RAISE EXCEPTION 'period overlaps with existing period';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_period_overlap
    BEFORE INSERT OR UPDATE ON accounting_periods
    FOR EACH ROW
    EXECUTE FUNCTION check_period_overlap();

CREATE TABLE IF NOT EXISTS journal_entries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_number  VARCHAR(30) NOT NULL UNIQUE,
    entry_date      DATE NOT NULL,
    description     TEXT NOT NULL,
    source_type     VARCHAR(30) CHECK (source_type IN (
                        'invoice_issued', 'payment_verified', 'invoice_cancelled',
                        'adjustment', 'closing', 'manual'
                    )),
    source_id       UUID,
    period_id       UUID NOT NULL REFERENCES accounting_periods(id),
    total_debit     NUMERIC(16,2) NOT NULL,
    total_credit    NUMERIC(16,2) NOT NULL,
    posted_by       UUID NOT NULL,
    posted_at       TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft', 'posted', 'cancelled')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_journal_number ON journal_entries(journal_number);
CREATE INDEX idx_journal_date ON journal_entries(entry_date);
CREATE INDEX idx_journal_period ON journal_entries(period_id);
CREATE INDEX idx_journal_source ON journal_entries(source_type, source_id) WHERE source_id IS NOT NULL;
CREATE INDEX idx_journal_status ON journal_entries(status);
ALTER TABLE journal_entries ADD CONSTRAINT chk_journal_balance
    CHECK (total_debit = total_credit);

CREATE TABLE IF NOT EXISTS journal_entry_lines (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id       UUID NOT NULL REFERENCES accounts(id),
    account_code     VARCHAR(20) NOT NULL,
    description      TEXT,
    debit            NUMERIC(16,2) NOT NULL DEFAULT 0,
    credit           NUMERIC(16,2) NOT NULL DEFAULT 0
);
CREATE INDEX idx_journal_lines_entry ON journal_entry_lines(journal_entry_id);
CREATE INDEX idx_journal_lines_account ON journal_entry_lines(account_id);
ALTER TABLE journal_entry_lines ADD CONSTRAINT chk_line_debit_xor_credit
    CHECK ((debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0));

-- ============================================================
-- AUDIT
-- ============================================================

CREATE TABLE IF NOT EXISTS finance_audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id   UUID NOT NULL,
    action      VARCHAR(50) NOT NULL,
    actor_id    UUID NOT NULL,
    changes     JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_fal_entity ON finance_audit_logs(entity_type, entity_id);
CREATE INDEX idx_fal_actor ON finance_audit_logs(actor_id);
```

---

## Struktur Direktori

```
internal/modules/keuangan/
├── module.go
├── contract.go
├── application/
│   ├── command/
│   │   ├── create_fee_component.go
│   │   ├── update_fee_component.go
│   │   ├── create_billing_scheme.go
│   │   ├── update_billing_scheme.go
│   │   ├── assign_scheme_to_santri.go
│   │   ├── create_invoice.go
│   │   ├── create_invoice_batch.go
│   │   ├── cancel_invoice.go
│   │   ├── apply_adjustment.go
│   │   ├── create_manual_payment.go
│   │   ├── verify_payment.go
│   │   ├── reject_payment.go
│   │   ├── create_account.go
│   │   ├── update_account.go
│   │   ├── create_manual_journal.go
│   │   ├── cancel_journal.go
│   │   ├── create_period.go
│   │   ├── close_period.go
│   │   ├── reopen_period.go
│   │   └── lock_period.go
│   ├── dto/
│   │   ├── fee_component_dto.go
│   │   ├── billing_scheme_dto.go
│   │   ├── invoice_dto.go
│   │   ├── payment_dto.go
│   │   ├── account_dto.go
│   │   ├── journal_dto.go
│   │   ├── period_dto.go
│   │   └── report_dto.go
│   ├── errors.go
│   ├── ports/
│   │   ├── transactor.go
│   │   ├── storage.go
│   │   ├── kesantrian_reader.go
│   │   └── identity_reader.go
│   └── query/
│       ├── list_fee_components.go
│       ├── list_billing_schemes.go
│       ├── list_invoices.go
│       ├── get_invoice.go
│       ├── my_invoices.go
│       ├── list_payments.go
│       ├── my_payments.go
│       ├── list_accounts.go
│       ├── get_account.go
│       ├── list_journal_entries.go
│       ├── get_journal_entry.go
│       ├── list_periods.go
│       ├── report_summary.go
│       ├── report_outstanding.go
│       ├── report_ledger.go
│       ├── report_trial_balance.go
│       ├── report_balance_sheet.go
│       └── report_income_statement.go
├── domain/
│   ├── feecomponent/
│   │   ├── constant/
│   │   ├── entity/
│   │   └── repository/
│   ├── billingscheme/
│   │   ├── constant/
│   │   ├── entity/
│   │   └── repository/
│   ├── invoice/
│   │   ├── constant/
│   │   ├── entity/
│   │   ├── repository/
│   │   └── valueobject/
│   │       └── invoice_number.go
│   ├── payment/
│   │   ├── constant/
│   │   ├── entity/
│   │   ├── repository/
│   │   └── valueobject/
│   │       └── payment_number.go
│   ├── adjustment/
│   │   ├── constant/
│   │   ├── entity/
│   │   └── repository/
│   ├── account/
│   │   ├── constant/
│   │   ├── entity/
│   │   └── repository/
│   ├── journal/
│   │   ├── constant/
│   │   ├── entity/
│   │   ├── repository/
│   │   └── service/
│   │       └── auto_posting.go
│   └── period/
│       ├── constant/
│       ├── entity/
│       └── repository/
├── infrastructure/
│   ├── external/
│   │   ├── minio_uploader.go
│   │   └── pdf_generator.go
│   ├── kesantriangateway/
│   │   └── gateway.go
│   ├── identitygateway/
│   │   └── gateway.go
│   └── persistence/
│       ├── postgres_fee_component_repo.go
│       ├── postgres_billing_scheme_repo.go
│       ├── postgres_invoice_repo.go
│       ├── postgres_payment_repo.go
│       ├── postgres_adjustment_repo.go
│       ├── postgres_account_repo.go
│       ├── postgres_journal_repo.go
│       ├── postgres_period_repo.go
│       ├── postgres_transactor.go
│       └── helpers.go
└── interfaces/
    └── http/
        ├── handler.go
        └── router.go
```

---

## Perubahan di Modul Existing

### `identity` — `permission_constant.go`

Tambah 6 permission keys:

```go
PermissionManageKeuangan      PermissionKey = "manage_keuangan"
PermissionVerifyPayment       PermissionKey = "verify_payment"
PermissionViewKeuanganReports PermissionKey = "view_keuangan_reports"
PermissionManageAccounts      PermissionKey = "manage_accounts"
PermissionManageJournal       PermissionKey = "manage_journal"
PermissionClosePeriod         PermissionKey = "close_period"
```

Tambahkan ke `RolePermissions` untuk usergod, superadmin, admin. Tambahkan ke `AllPermissionDefinitions` dan `DefaultPermissionsInit`.

### `kesantrian` — `contract.go`

Tambah 2 method baru:

```go
type SantriBasicInfo struct {
    SantriID string
    UserID   string
    NIS      *string
    Status   string
}

type Contract interface {
    // ... existing methods ...
    ListActiveSantriIDs(ctx context.Context) ([]string, error)
    GetSantriByUserID(ctx context.Context, userID string) (*SantriBasicInfo, error)
}
```

### `cmd/api/main.go`

Wire `keuangan.NewModule(...)`:

```go
keuangan := keuanganModule.NewModule(
    db, cfg,
    identity,        // identity.Contract
    kesantrian,      // kesantrian.Contract
    identity.AuthMiddleware(),
    identity.PrincipalMiddleware(),
)
// ...
keuangan.RegisterRoutes(engine)
```

### `seeders`

Seeder untuk default COA accounts (30+ akun sesuai daftar di atas).

---

## Non-Functional Requirements

### Keamanan
- Semua endpoint dilindungi JWT + permission check.
- Cross-module references tanpa FK (pattern existing).
- Audit log (`finance_audit_logs`) untuk setiap aksi finansial.
- Amount disimpan dengan `NUMERIC(14,2)`, tidak pernah float.
- Race condition pada update `paid_amount` ditangani dengan `SELECT ... FOR UPDATE` atau optimistic locking.
- CHECK constraint `total_debit = total_credit` di level database.

### Audit Log
- Setiap perubahan status invoice & payment dicatat.
- Setiap jurnal yang dibuat/di-cancel dicatat.
- Actor (user ID) dan timestamp disimpan.

### Performa
- Index pada kolom filter (santri_id, status, periode, due_date, account_id).
- Batch generate tagihan dibatasi max 500 per request.
- Pagination pada semua list endpoint.
- `account_code` di-denormalize di `journal_entry_lines` untuk menghindari JOIN saat laporan.

### Skalabilitas
- Modul terisolasi, bisa di-extract jadi microservice jika perlu.
- Payment gateway integration via port/adapter (Fase 2).

---

## Roadmap Implementasi

### Fase 1 — MVP: Billing + Akuntansi Dasar

| Step | Deliverable |
|---|---|
| 1 | Migration `012_create_keuangan_tables.up.sql` + `.down.sql` |
| 2 | Tambah 6 permission keys di `identity` + assign ke system roles |
| 3 | Tambah method ke `kesantrian.Contract` (`ListActiveSantriIDs`, `GetSantriByUserID`) |
| 4 | Seeder default COA accounts |
| 5 | Domain entities: feecomponent, billingscheme, invoice, payment, adjustment |
| 6 | Domain entities: account, journal, period |
| 7 | Repository interfaces + Postgres implementations |
| 8 | Auto-posting service |
| 9 | Use cases billing: CRUD komponen, CRUD skema, assign skema |
| 10 | Use cases billing: create invoice (individual + batch), cancel, adjustment |
| 11 | Use cases billing: manual payment, verify, reject |
| 12 | Use cases accounting: CRUD akun, jurnal manual, cancel jurnal |
| 13 | Use cases accounting: create period, close, reopen, lock |
| 14 | HTTP handlers + router |
| 15 | Wiring di `cmd/api/main.go` |
| 16 | Laporan: Buku Besar, Neraca Saldo, Neraca, Laba Rugi, Summary, Outstanding |
| 17 | PDF kwitansi |
| 18 | Unit tests domain entities |

### Fase 2 — Enhanced Features

| Deliverable | Deskripsi |
|---|---|
| Upload bukti bayar | Presign/confirm pattern (MinIO) untuk santri upload bukti transfer |
| Mapping configurable | fee_component.type → account.code bisa dikonfigurasi via UI |
| Notifikasi | Reminder jatuh tempo via email/WA (Fonnte) |
| Laporan extended | Arus kas, rekap per tahun ajaran, export Excel |

### Fase 3 — Future

| Deliverable | Deskripsi |
|---|---|
| Payment gateway | Virtual account (Midtrans/Xendit), auto-verify |
| Rekonsiliasi bank | Upload mutasi bank → matching dengan pembayaran |
| Biaya Diterima Dimuka | Auto-detect pembayaran sebelum invoice → jurnal unearned revenue |
| Export laporan | PDF, Excel untuk semua laporan |
| Refund | Jika ada kebutuhan bisnis |
