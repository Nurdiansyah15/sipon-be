# Rencana: Input `issued_date` Eksplisit pada Invoice + Periode Tagihan Opsional untuk Komponen Non-Periodik

## Context

Saat ini `Invoice.Issue()` (`domain/invoice/entity/invoice.go:52-60`) selalu memakai `time.Now()` untuk `IssuedAt` — bendahara tidak punya kendali sama sekali atas tanggal terbit invoice, dan tanggal itu **tidak pernah dicocokkan** dengan periode tagihan (`billing_period`) yang dipilih. Akibatnya:

1. Kalau bendahara membuat invoice "untuk Agustus" tapi baru menginputnya bulan berikutnya (atau setelah periode akuntansi 2026 ditutup), invoice tetap tercatat tanggal **hari ini** — bisa jatuh ke periode akuntansi yang berbeda dari yang dimaksud, dan kalau periode akuntansi untuk hari ini ternyata sudah ditutup, error "periode akuntansi ditutup" muncul tanpa bendahara sadar itu terkait tanggal apa (karena mereka tidak pernah menginput tanggal apa pun).
2. Semua invoice — termasuk yang sifatnya sekali-jalan/tidak berulang (fee_component dengan `is_periodic = false`, mis. tipe `insidental`: denda, uang gedung, dll.) — **wajib** diberi `billing_period_id`, padahal secara konsep komponen non-periodik tidak "milik" periode tagihan tertentu.

Rencana ini menyelesaikan keduanya sekaligus: `issued_date` jadi input eksplisit di semua cara membuat invoice, divalidasi terhadap periode tagihan (kalau ada), dan periode tagihan jadi opsional untuk komponen non-periodik. Dengan ini, error "periode akuntansi sudah ditutup" (lihat diskusi sebelumnya, akan diperbaiki di Fase 0 di bawah) jadi jelas sebab-akibatnya: karena bendahara memang menginput tanggal yang jatuh di periode yang sudah ditutup, bukan lagi karena tanggal implisit yang tidak terlihat.

**Keputusan desain** (dikonfirmasi):
- Kalau `fee_component.is_periodic == false`: `billing_period_id` **opsional bebas** — boleh diisi atau tidak. Kalau diisi (walau komponennya non-periodik), `issued_date` tetap divalidasi harus dalam rentang periode itu — **satu aturan validasi saja**, tidak ada percabangan "wajib kosong vs boleh diisi".
- Aturan opsional ini **hanya berlaku di endpoint invoice single/manual** (`POST /admin/invoices`). Generate massal (`POST /admin/invoices/batch`) **tidak berubah** — `billing_period_id` tetap selalu wajib di sana, karena batch memang khusus penagihan periodik massal.

Lihat juga: [`docs/rules/akuntansi.md`](../rules/akuntansi.md) §2.1, [`docs/schemas/keuangan-akuntansi.md`](../schemas/keuangan-akuntansi.md), [`docs/specs/keuangan-akuntansi-api.md`](../specs/keuangan-akuntansi-api.md).

---

## FASE 0 — Prasyarat: `AutoPostingService` harus menolak posting ke periode akuntansi tertutup

Ini **wajib dikerjakan sebelum/bersama** fase-fase di bawah — kalau tidak, `issued_date` yang divalidasi dengan benar tetap bisa diam-diam terposting ke periode akuntansi yang sudah `closed`/`locked` tanpa error apa pun (bug yang sudah ditemukan sebelumnya, `AutoPostingService` cuma pakai `FindByDate` tanpa cek status).

1. Tambah pengecekan di **keempat** method `AutoPostingService` (`PostInvoiceIssued`, `PostPaymentVerified`, `PostInvoiceCancelled`, `PostAdjustment`, `internal/modules/keuangan/domain/journal/service/auto_posting.go`), tepat setelah `period, err := s.periodRepo.FindByDate(ctx, entryDate)`:
   ```go
   if !period.CanPost() {
       return kernel.WrapMsg(journalConst.CodeJournalPeriodClosed, "Periode akuntansi untuk tanggal ini sudah ditutup", nil)
   }
   ```
   (`journalConst.CodeJournalPeriodClosed` sudah ada, cuma belum dipakai di file ini.)
2. Tambah `case journalConst.CodeJournalPeriodClosed: return ..., kernel.WrapMsg(application.ErrCodeConflict, ke.Message, ke)` di error-switch pemanggil: `create_invoice.go`, `create_invoice_batch.go` (di `processTarget`), `verify_payment.go`, `cancel_invoice.go`, `apply_adjustment.go` — semuanya sudah punya switch serupa untuk `CodeJournalAccountMappingNotFound`, tinggal tambah satu `case` lagi di tempat yang sama.
3. Tulis sebagai bug baru di `docs/bugs/akuntansi-auto-posting-tidak-cek-status-periode.md` (ikuti format bug lain), tandai "Diperbaiki" begitu langkah 1-2 selesai.

**Verifikasi Fase 0**: buat payment dengan `payment_date` sengaja diisi tanggal yang jatuh di periode akuntansi yang sudah `closed` → verifikasi harus ditolak 409, bukan sukses diam-diam.

---

## FASE 1 — Skema: `invoices.billing_period_id` jadi nullable

Migrasi baru:
```sql
ALTER TABLE invoices ALTER COLUMN billing_period_id DROP NOT NULL;
```
Down migration re-`SET NOT NULL` **tidak aman** kalau sudah ada baris dengan `billing_period_id IS NULL` (tidak ada "default periode" yang masuk akal untuk di-backfill, beda dari kasus `debit_account_id` yang bisa di-backfill ke akun kas default). Catat ini di komentar migration down — kalau dijalankan saat ada baris NULL, akan gagal, itu memang harus jadi keputusan manual.

Index & constraint yang **tidak perlu diubah**: `idx_invoices_unique_period` (`UNIQUE(santri_id, fee_component_id, billing_period_id) WHERE deleted_at IS NULL AND status NOT IN ('cancelled')`) — di Postgres, `NULL` di kolom unique index dianggap berbeda satu sama lain, jadi beberapa invoice non-periodik untuk santri+komponen yang sama (mis. beberapa "denda" terpisah) otomatis **tidak** dianggap duplikat. Ini pas dengan kebutuhan (komponen non-periodik memang boleh berulang tanpa dianggap dobel).

---

## FASE 2 — Domain: `Invoice` entity

`internal/modules/keuangan/domain/invoice/entity/invoice.go`:

1. `BillingPeriodID string` → `BillingPeriodID *string`.
2. `NewInvoice(id, invoiceNumber, santriID, userID, feeComponentID string, billingPeriodID *string, amount float64, dueDate time.Time, createdBy string) (*Invoice, error)` — hapus `billingPeriodID` dari validasi "tidak boleh kosong" (baris 32), karena sekarang legal untuk nil.
3. `Issue()` → **`Issue(issuedDate time.Time) error`** — hapus `now := time.Now()` untuk `IssuedAt`, ganti jadi parameter:
   ```go
   func (i *Invoice) Issue(issuedDate time.Time) error {
       if i.Status != constant.StatusDraft {
           return kernel.WrapMsg(constant.CodeInvoiceInvalidStatus, "Hanya invoice berstatus draft yang dapat diterbitkan", nil)
       }
       i.Status = constant.StatusIssued
       i.IssuedAt = &issuedDate
       i.UpdatedAt = time.Now()
       return nil
   }
   ```
   (`UpdatedAt` tetap boleh `time.Now()` — itu metadata teknis "kapan baris ini diubah", bukan tanggal akuntansi.)

---

## FASE 3 — Use-case: validasi & wiring

### 3.1 `CreateInvoiceUseCase` (single, `application/command/create_invoice.go`)

`CreateInvoiceCmd` (baris 52-62): ubah `BillingPeriodID string` → `BillingPeriodID *string`, tambah `IssuedDate string`.

Urutan validasi baru di `Execute` (sisipkan setelah fetch `fee`, sebelum bagian yang sudah ada soal billing period):

```go
issuedDate, err := time.Parse("2006-01-02", cmd.IssuedDate)
if err != nil {
    return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal terbit tidak valid", err)
}

var billingPeriod *bpEntity.BillingPeriod
if fee.IsPeriodic && (cmd.BillingPeriodID == nil || *cmd.BillingPeriodID == "") {
    return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "Komponen biaya ini periodik, periode tagihan wajib diisi", nil)
}
if cmd.BillingPeriodID != nil && *cmd.BillingPeriodID != "" {
    billingPeriod, err = uc.billingPeriodRepo.FindByID(ctx, *cmd.BillingPeriodID)
    if err != nil {
        var ke *kernel.AppError
        if errors.As(err, &ke) && ke.Code == bpConst.CodeBillingPeriodNotFound {
            return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
        }
        return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
    }
    if !billingPeriod.IsOpen() {
        return nil, kernel.WrapMsg(application.ErrCodeConflict, "Status periode tagihan tidak valid", nil)
    }
    if issuedDate.Before(billingPeriod.StartDate) || issuedDate.After(billingPeriod.EndDate) {
        return nil, kernel.WrapMsg(application.ErrCodeBadRequest,
            fmt.Sprintf("Tanggal terbit harus dalam rentang periode tagihan %s (%s s.d. %s)",
                billingPeriod.Name, billingPeriod.StartDate.Format("2006-01-02"), billingPeriod.EndDate.Format("2006-01-02")), nil)
    }
}
```

Sisanya di `Execute`:
- Pengecekan duplikat (`FindBySantriComponentPeriod`) — **lewati kalau `cmd.BillingPeriodID` nil/kosong** (tidak ada konsep duplikat periodik untuk invoice tanpa periode tagihan):
  ```go
  if billingPeriod != nil {
      existing, _ := uc.invoiceRepo.FindBySantriComponentPeriod(ctx, cmd.SantriID, cmd.FeeComponentID, billingPeriod.ID)
      if existing != nil {
          return nil, kernel.WrapMsg(application.ErrCodeConflict, "Invoice duplikat", nil)
      }
  }
  ```
- `NewInvoice(..., cmd.BillingPeriodID, ...)` — langsung pakai `cmd.BillingPeriodID` (`*string`, sudah sesuai signature baru).
- `if cmd.Issue { inv.Issue(issuedDate) }` — ganti dari `inv.Issue()` tanpa argumen.
- `toInvoiceResponse(inv, billingPeriod)` di akhir — sudah menerima `*bpEntity.BillingPeriod`, tinggal pastikan dipanggil dengan `billingPeriod` yang bisa nil (fungsi ini sudah punya guard `if period != nil` di baris 213 versi sebelumnya — pastikan tetap ada).

**Opsional (nice-to-have, longgar dikerjakan sekalian atau tidak)**: validasi `dueDate` tidak boleh sebelum `issuedDate` — kalau mau ditambah, satu `if dueDate.Before(issuedDate) { return ... ErrCodeBadRequest ... }` setelah `dueDate` diparse.

### 3.2 `CreateInvoiceBatchUseCase` (batch, `application/command/create_invoice_batch.go`)

`billing_period_id` tetap wajib (tidak berubah). Tambahan: `CreateInvoiceBatchCmd` (baris 69-74) dapat field baru `IssuedDate string`.

Di `Execute`, setelah `period` (billing period) berhasil di-fetch & dicek `IsOpen()` (baris 98-111 saat ini), tambahkan:
```go
issuedDate, err := time.Parse("2006-01-02", cmd.IssuedDate)
if err != nil {
    return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "format tanggal terbit tidak valid", err)
}
if issuedDate.Before(period.StartDate) || issuedDate.After(period.EndDate) {
    return nil, kernel.WrapMsg(application.ErrCodeBadRequest,
        fmt.Sprintf("Tanggal terbit harus dalam rentang periode tagihan %s (%s s.d. %s)",
            period.Name, period.StartDate.Format("2006-01-02"), period.EndDate.Format("2006-01-02")), nil)
}
```
Teruskan `issuedDate` ke `processTarget` (tambah parameter di signature baris 214-220), lalu di dalam loop (baris 251-264):
- `NewInvoice(..., &cmd.BillingPeriodID, ...)` — bungkus jadi pointer (batch selalu punya nilai, jadi aman `&cmd.BillingPeriodID`).
- `inv.Issue(issuedDate)` — ganti dari `inv.Issue()`.

### 3.3 Use-case lain yang ikut kena dampak tipe `*string`

- `apply_adjustment.go`, `cancel_invoice.go`: tidak memanggil `Issue()`, cuma membaca `inv.IssuedAt` — tidak ada perubahan logic, aman karena `IssuedAt` tetap `*time.Time`.
- Tidak ada use-case lain yang memanggil `NewInvoice`/`Issue` di luar dua di atas.

---

## FASE 4 — DTO, Repository, Response

### 4.1 DTO (`application/dto/invoice_dto.go`)

```go
type CreateInvoiceRequest struct {
	SantriID        string  `json:"santri_id" binding:"required"`
	FeeComponentID  string  `json:"fee_component_id" binding:"required"`
	BillingPeriodID *string `json:"billing_period_id,omitempty"`
	IssuedDate      string  `json:"issued_date" binding:"required"`
	Amount          float64 `json:"amount" binding:"required"`
	DueDate         string  `json:"due_date" binding:"required"`
	Notes           *string `json:"notes,omitempty"`
}

type CreateInvoiceBatchRequest struct {
	BillingSchemeID string `json:"billing_scheme_id" binding:"required"`
	BillingPeriodID string `json:"billing_period_id" binding:"required"`
	IssuedDate      string `json:"issued_date" binding:"required"`
	DueDate         string `json:"due_date" binding:"required"`
}
```
`InvoiceResponse.BillingPeriodID string` → `*string` (baris 30) — `BillingPeriod *BillingPeriodBriefResponse` sudah nullable, tidak berubah.

### 4.2 Handler (`interfaces/http/handler.go`)

`CreateInvoice` (baris ~456-479): tambah `IssuedDate: req.IssuedDate` ke `cmd`. `BillingPeriodID: req.BillingPeriodID` (sudah cocok tipe `*string`).
`CreateInvoiceBatch` (baris ~481-500): tambah `IssuedDate: req.IssuedDate` ke `cmd`.

### 4.3 Repository (`infrastructure/persistence/postgres_invoice_repo.go`)

- `Save`/`Update` (baris ~56, ~81): ganti `inv.BillingPeriodID` jadi `nullStr(inv.BillingPeriodID)` di parameter query (sekarang tipenya sudah `*string`, cocok dengan helper `nullStr` yang sudah dipakai kolom nullable lain di file yang sama).
- `scan` (sekitar baris 253): kolom `billing_period_id` dibaca sebagai `sql.NullString` lalu `strFromNull(...)` — pola yang sama seperti `billing_scheme_id` di tabel yang sama (sudah nullable, sudah ada contohnya persis di baris atasnya).

### 4.4 Response enrichment

- `application/query/invoice_response.go` (`buildInvoiceResponse`, baris 15-58): baris 23 `BillingPeriodID: inv.BillingPeriodID` — otomatis cocok karena tipe sudah `*string` di kedua sisi. Baris 48-56 (`if billingPeriodRepo != nil { ... FindByID(ctx, inv.BillingPeriodID) ... }`) — tambah guard `inv.BillingPeriodID != nil &&`:
  ```go
  if billingPeriodRepo != nil && inv.BillingPeriodID != nil {
      if bp, err := billingPeriodRepo.FindByID(ctx, *inv.BillingPeriodID); err == nil {
          ...
      }
  }
  ```
- `application/command/create_invoice.go`'s lokal `toInvoiceResponse(inv, period)` — pastikan `resp.BillingPeriodID = inv.BillingPeriodID` (tipe sudah cocok, tidak perlu ubah selain assignment langsung).

---

## FASE 5 — API Spec & Rules (dokumentasi)

Update `docs/specs/keuangan-akuntansi-api.md` (tambah subbagian baru, mis. §B.7):
- `POST /admin/invoices` — request body baru (`issued_date` wajib, `billing_period_id` opsional), tabel error baru:

  | Kondisi | Kode domain | HTTP |
  |---|---|---|
  | Komponen periodik tapi `billing_period_id` kosong | (validasi inline, `ErrCodeBadRequest`) | 400 |
  | `issued_date` di luar rentang periode tagihan yang dipilih | (validasi inline, `ErrCodeBadRequest`) | 400 |
  | Periode tagihan tidak `open` | (existing) | 409 |
  | Periode akuntansi untuk `issued_date` sudah ditutup | `journalConst.CodeJournalPeriodClosed` (Fase 0) | 409 |

- `POST /admin/invoices/batch` — tambah `issued_date` wajib di body, error baru "issued_date di luar rentang periode tagihan" (400).

Update `docs/rules/akuntansi.md` §2.1 ("Invoice diterbitkan → jurnal otomatis"): tambah catatan bahwa titik posting sekarang berdasarkan `issued_date` yang diinput eksplisit (bukan `time.Now()` implisit), dan billing period opsional untuk komponen non-periodik.

Update `docs/schemas/keuangan-akuntansi.md` (tabel `invoices`): `billing_period_id` jadi nullable, catat aturannya.

---

## Verifikasi keseluruhan

1. Buat komponen biaya periodik (SPP) → buat invoice tanpa `billing_period_id` → ditolak 400.
2. Buat invoice periodik dengan `issued_date` di luar rentang billing period yang dipilih → ditolak 400.
3. Buat invoice periodik dengan `issued_date` valid dalam rentang → sukses, `IssuedAt` = `issued_date` yang diinput (bukan hari ini).
4. Buat komponen biaya non-periodik (insidental) → buat invoice **tanpa** `billing_period_id` → sukses, `billing_period_id` NULL di DB, tidak ada validasi rentang.
5. Buat invoice non-periodik **dengan** `billing_period_id` diisi tapi `issued_date` di luar rentangnya → ditolak 400 (aturan tunggal tetap berlaku).
6. Buat invoice (periodik atau tidak) dengan `issued_date` yang jatuh di periode akuntansi yang sudah `closed` → ditolak 409 dengan pesan jelas "periode akuntansi ... sudah ditutup" (Fase 0).
7. Generate batch dengan `issued_date` di luar rentang billing period yang dipilih → seluruh batch ditolak 400 sebelum ada satu invoice pun dibuat.
8. `go build/vet/test` lulus.
