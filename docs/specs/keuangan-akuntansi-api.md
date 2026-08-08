# API Spec: Keuangan & Akuntansi

Base path: `/api/v1/web/keuangan`. Semua endpoint admin butuh JWT + permission (lihat kolom Permission). Endpoint santri butuh JWT biasa (role member).

Dokumen ini memuat **(A)** endpoint yang sudah berjalan (referensi cepat, sinkron dengan `internal/modules/keuangan/interfaces/http/router.go`) dan **(B)** perubahan/endpoint baru untuk penyempurnaan akuntansi. Aturan bisnis di balik tiap endpoint ada di [`docs/rules/akuntansi.md`](../rules/akuntansi.md).

---

## A. Endpoint yang Sudah Berjalan

### Billing — Santri (self-service)
| Method | Path | Deskripsi |
|---|---|---|
| GET | `/invoices` | Tagihan saya |
| GET | `/invoices/:id` | Detail tagihan saya |
| GET | `/payments` | Riwayat pembayaran saya |

### Billing — Admin (`manage_keuangan`)
| Method | Path | Deskripsi |
|---|---|---|
| GET/POST | `/admin/components` | List/buat komponen biaya |
| PUT/DELETE | `/admin/components/:id` | Update/nonaktifkan komponen |
| GET/POST | `/admin/schemes` | List/buat skema tagihan |
| GET/PUT/DELETE | `/admin/schemes/:id` | Detail/update/nonaktifkan skema |
| POST/DELETE | `/admin/schemes/:id/items[/:itemId]` | Tambah/hapus komponen dari skema |
| POST/GET | `/admin/assignments` | Assign skema ke santri / list assignment |
| GET/POST | `/admin/invoices` | List / buat invoice individual |
| POST | `/admin/invoices/batch` | Generate tagihan massal |
| GET | `/admin/invoices/:id` | Detail invoice |
| POST | `/admin/invoices/:id/cancel` | Batalkan invoice |
| POST | `/admin/invoices/:id/adjustment` | Tambah diskon/beasiswa |
| GET | `/admin/payments` / `/admin/payments/:id` | List/detail payment |
| GET | `/admin/payments/:id/receipt` | Download kwitansi PDF |
| POST | `/admin/payments/manual` | Catat pembayaran manual |
| POST | `/admin/payments/:id/verify` (`verify_payment`) | Verifikasi payment |
| POST | `/admin/payments/:id/reject` (`verify_payment`) | Tolak payment |
| GET/POST | `/admin/billing-periods` | List/buat periode tagihan |
| POST | `/admin/billing-periods/:id/open` `/close` | Buka/tutup periode tagihan |
| GET | `/admin/billing-batches[/:id]` | List/detail batch generate massal |

### Akuntansi — Admin
| Method | Path | Permission | Deskripsi |
|---|---|---|---|
| GET/POST | `/admin/accounts` | `manage_accounts` | List/buat akun COA |
| GET/PUT/DELETE | `/admin/accounts/:id` | `manage_accounts` | Detail/update/nonaktifkan akun |
| GET | `/admin/journal-entries` | `manage_journal` | List jurnal |
| GET | `/admin/journal-entries/:id` | `manage_journal` | Detail jurnal + lines |
| POST | `/admin/journal-entries` | `manage_journal` | Buat jurnal manual |
| POST | `/admin/journal-entries/:id/cancel` | `manage_journal` | Batalkan jurnal (manual saja) |
| GET | `/admin/periods` | `close_period` | List periode akuntansi |
| GET | `/admin/periods/active` | `close_period` | Periode akuntansi aktif |
| POST | `/admin/periods` | `close_period` | Buat periode akuntansi |
| POST | `/admin/periods/:id/close` | `close_period` | Tutup periode |
| POST | `/admin/periods/:id/reopen` | `close_period` | Buka kembali periode closed |
| POST | `/admin/periods/:id/lock` | `close_period` | Kunci permanen periode closed |

### Laporan — Admin (`view_keuangan_reports`)
| Method | Path | Deskripsi |
|---|---|---|
| GET | `/admin/reports/summary` | Rekap tagihan & pembayaran |
| GET | `/admin/reports/outstanding` | Tunggakan per santri |
| GET | `/admin/reports/ledger` | Buku besar per akun (`?account_id=&period_id=`) |
| GET | `/admin/reports/trial-balance` | Neraca saldo (`?period_id=`) |
| GET | `/admin/reports/balance-sheet` | Neraca |
| GET | `/admin/reports/income-statement` | Laba rugi |

---

## B. Perubahan & Endpoint Baru (Penyempurnaan Akuntansi)

Prinsip: **tidak menambah endpoint baru untuk hal yang bisa jadi efek samping endpoint yang sudah ada** — mengeluarkan tagihan, memverifikasi pembayaran, dan membatalkan tagihan sudah masing-masing punya endpoint; yang berubah adalah *apa yang terjadi di baliknya* (sekarang juga memposting jurnal), bukan bentuk request/response-nya secara drastis.

### B.1 `POST /admin/invoices`, `POST /admin/invoices/batch` — tidak berubah request

Response `InvoiceResponse` **tidak perlu** ditambah field jurnal (jurnal bisa ditelusuri lewat B.4). Perilaku baru: use-case ini sekarang memanggil `AutoPostingService.PostInvoiceIssued` di transaksi yang sama dengan pembuatan invoice. Error baru yang mungkin muncul — **semua wajib dibungkus `application.WrapConflictErr(err, kode)` sebelum dikembalikan dari use-case**, bukan `kernel.New(kode)` telanjang, supaya statusnya benar 409 dan bukan 500 (lihat `docs/bugs/keuangan-kernel-error-tanpa-wrap-selalu-500.md`):

| Error | Kapan | Kode domain | HTTP (setelah di-`WrapConflictErr`) |
|---|---|---|---|
| Periode akuntansi untuk tanggal invoice tidak ditemukan/tidak `open` | Tidak ada `accounting_periods` yang mencakup tanggal invoice (`FindByDate` return not-found), atau ada tapi sudah `closed`/`locked` | `JOURNAL_PERIOD_CLOSED` (existing, `journal_constant.go`) | 409 |
| Mapping akun pendapatan untuk tipe komponen tidak ditemukan | `fee_component.type` baru yang belum ada di `feeTypeRevenueAccount` (`auto_posting.go:37-42`) | `JOURNAL_ACCOUNT_MAPPING_NOT_FOUND` (kode baru, tambahkan ke `journal_constant.go` — lihat `docs/bugs/akuntansi-auto-posting-tidak-terpasang.md` langkah 1) | 409 |

Generate massal (`/invoices/batch`) memposting satu jurnal `invoice_issued` **per invoice yang berhasil dibuat** (bukan satu jurnal gabungan) — konsisten dengan billing_batch_targets yang juga mencatat granular per santri.

### B.2 `POST /admin/payments/:id/verify` — tidak berubah request

Perilaku baru: memanggil `AutoPostingService.PostPaymentVerified` di transaksi yang sama dengan `payment.Verify()` + `invoice.AddPayment()` (sudah dalam satu `transactor.WithTx`, tinggal disisipkan). Validasi baru sebelum verifikasi diizinkan:

- `payment.debit_account_id` harus terisi (lihat perubahan skema — jadi wajib diisi saat `POST /admin/payments/manual`, bukan baru dicek saat verify).
- Akun tersebut harus `is_postable` & `is_active`.

`POST /admin/payments/manual` — `debit_account_id` di `CreateManualPaymentRequest` berubah dari opsional jadi **wajib** (`binding:"required"`).

### B.3 `POST /admin/invoices/:id/cancel` — tambah validasi

Tolak pembatalan kalau `invoice.paid_amount > 0` (lihat rules doc §2.3) — bukan cuma status `paid`. Response error baru pakai kode existing `CodeInvoiceInvalidStatus` dengan pesan yang membedakan ("tidak bisa membatalkan invoice yang sudah ada pembayaran terverifikasi").

Perilaku baru: kalau invoice sebelumnya sudah `issued` (sudah pernah menghasilkan jurnal), pembatalan memposting jurnal `invoice_cancelled` di transaksi yang sama.

### B.4 Endpoint baru: telusur jurnal per dokumen sumber

```
GET /admin/journal-entries/by-source?source_type=invoice_issued&source_id=<invoice_id>
```

Response: `JournalEntryResponse` tunggal (atau 404 kalau belum pernah diposting — mis. invoice masih `draft` atau baru dibuat sebelum fitur ini aktif). Berguna untuk tombol "Lihat Jurnal" di halaman detail invoice/payment admin, tanpa perlu invoice/payment tahu ID jurnalnya sendiri.

Backend: tinggal expose `JournalRepository.FindBySource` (sudah ada implementasinya, belum ada use-case/handler) lewat use-case query baru + route baru.

### B.5 `POST /admin/periods/:id/close` — perilaku berubah signifikan

Sekarang menjalankan proses closing penuh (rules doc §3.2) dalam satu transaksi, bukan cuma flip status:

**Request**: tidak berubah (tidak perlu body).

**Response** (`PeriodResponse` diperluas):
```json
{
  "id": "...",
  "status": "closed",
  "closing_journal_entry_id": "...",
  "total_revenue": 15000000,
  "total_expense": 8000000,
  "net_income": 7000000,
  ...
}
```

**Error baru**:
| Kondisi | Kode domain | HTTP |
|---|---|---|
| Akun `3200`/`3201` tidak ditemukan/tidak aktif di COA | `PERIOD_CLOSING_ACCOUNT_MISSING` (kode baru di `period_constant.go`, dibungkus `application.WrapConflictErr`) | 409 |
| Periode bukan `open` | `PERIOD_INVALID_STATUS` (existing) — **saat ini use-case lain sudah membungkusnya lewat `WrapRepoErr` yang salah memetakan ke 404, bukan cuma soal wrap/tidak-wrap** (lihat `docs/bugs/keuangan-kernel-error-tanpa-wrap-selalu-500.md` untuk pola yang benar); untuk endpoint ini pakai `application.WrapConflictErr(err, periodConst.CodePeriodInvalidStatus)` → 409, lebih sesuai semantik "state tidak valid" dibanding 404 | 409 |

**`POST /admin/periods/:id/reopen`**: kalau periode punya jurnal `closing`, jurnal itu ikut di-`cancel` (bukan dihapus) sebagai bagian dari reopen — satu transaksi.

Status `closing` dihapus dari domain (`AccountingPeriod.StartClosing()` dihapus, tidak ada endpoint terpisah untuk memulai closing — closing selalu atomik dari `open` langsung ke `closed`).

### B.6 `POST /admin/invoices/:id/adjustment` — perilaku berubah

Setelah menyimpan `invoice_adjustments` dan mengubah `discount_amount`, posting jurnal koreksi (`source_type='adjustment'`) sebesar nilai adjustment — lihat rules doc §2.1. Hanya berlaku kalau invoice sudah `issued` (sudah punya jurnal awal untuk dikoreksi); kalau masih `draft`, cukup ubah `discount_amount` seperti sekarang tanpa jurnal (belum ada apa pun untuk dikoreksi).

---

## C. Ringkasan Dampak ke Permission

Tidak ada permission baru yang diperlukan — semua perubahan di atas menempel ke endpoint dan permission yang sudah ada (`manage_keuangan`, `verify_payment`, `close_period`, `manage_journal`). Endpoint baru B.4 memakai `manage_journal` (sama seperti endpoint jurnal lain).
