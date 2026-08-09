# Rencana: Ganti Pemetaan Komponen Biaya dari `type` (enum) ke Referensi Akun COA Langsung (Revenue + Piutang)

> **Bergantung pada**: `docs/plan/coa-sub-type.md` (menambah kolom `accounts.sub_type`). Rencana itu **harus dikerjakan lebih dulu** — validasi `receivable_account_id` di bawah ini (keputusan #3) memakai `sub_type = receivable` yang baru ada setelah rencana itu selesai.

## Context

Saat ini setiap `fee_component` (komponen biaya: SPP, UKT, Daftar Ulang, Insidental) punya kolom `type` (enum string: `ukt|spp|daftar_ulang|insidental`). `AutoPostingService` (`internal/modules/keuangan/domain/journal/service/auto_posting.go`) memakai `type` itu untuk menentukan akun jurnal lewat **dua mekanisme hardcode berbeda**:

1. **Akun pendapatan (revenue)** — map hardcoded di Go (baris 40-45):
   ```go
   var feeTypeRevenueAccount = map[feeConst.FeeComponentType]string{
       feeConst.FeeTypeSPP:         "4100",
       feeConst.FeeTypeUKT:         "4200",
       feeConst.FeeTypeDaftarUlang: "4300",
       feeConst.FeeTypeInsidental:  "4400",
   }
   ```
   Dipakai di `PostInvoiceIssued` (baris 89), `PostInvoiceCancelled` (baris 210), `PostAdjustment` (baris 270).

2. **Akun piutang (receivable)** — kode akun `"1103"` di-hardcode **langsung sebagai string literal**, dipakai identik di **SEMUA EMPAT** fungsi posting termasuk `PostPaymentVerified` (baris 94, 159, 215, 275):
   ```go
   piutang, err := s.accountRepo.FindByCode(ctx, "1103")
   ```
   Ini berarti *semua* komponen biaya, apapun jenisnya, selalu dicatat ke satu akun piutang yang sama — tidak mungkin memisahkan piutang SPP dari piutang UKT dari piutang Daftar Ulang di buku besar, walau secara bisnis itu mungkin diinginkan.

Kedua hardcode ini sudah ditandai sebagai keterbatasan sementara di `docs/rules/akuntansi.md:39` (khusus untuk mapping revenue; piutang belum disinggung sama sekali di dokumen itu — ikut jadi bagian dari revisi ini).

**Revisi dari diskusi**: bukan cuma akun pendapatan yang harus lepas dari hardcode — akun piutang **juga** harus bisa dipilih per komponen biaya (bisa saja nanti "Piutang SPP" dialihkan ke akun piutang yang berbeda dari "Piutang UKT"), bukan satu akun `1103` berlaku untuk semua. Jadi `fee_components` butuh **dua** referensi akun, bukan satu:

- `revenue_account_id` — akun yang **dikredit** saat invoice terbit (dan dibalik saat cancel/adjustment).
- `receivable_account_id` — akun **piutang** yang didebit saat invoice terbit, dan dikredit saat pembayaran diverifikasi / invoice dibatalkan / ada adjustment pengurangan.

**Tujuan perubahan**: admin/bendahara memilih kedua akun ini langsung dari COA saat membuat komponen biaya, alih-alih sistem menebak dari `type` (revenue) atau mengasumsikan satu akun piutang universal.

### Yang TIDAK berubah

- Struktur baris jurnal (debit piutang/kredit revenue saat issued, dst.) tetap sama — yang berubah hanya **dari mana** kedua akun itu di-resolve.
- `feeConst.PeriodType` (monthly/semesterly/yearly/once) — enum berbeda di package yang sama, tidak disentuh.

### Dampak tambahan yang baru ketahuan: `PostPaymentVerified` ikut terdampak

`PostPaymentVerified` (`auto_posting.go:142-199`, dipanggil dari `application/command/verify_payment.go:87-91`) **tidak menerima `feeType` sama sekali** — ia hanya menerima `debitAccountID` (akun kas/bank pilihan bendahara) dari input, lalu meng-kredit piutang hardcoded `"1103"`. Supaya akun piutang bisa berbeda per komponen biaya, `verify_payment.go` (yang saat ini hanya fetch `payment` dan `inv`/Invoice — **belum** fetch `FeeComponent`) harus ditambah dependency `FeeComponentRepository`, resolve `inv.FeeComponentID → fee.ReceivableAccountID`, lalu diteruskan ke `PostPaymentVerified`. Ini titik yang tidak muncul di draft rencana sebelumnya.

---

## Keputusan desain yang perlu dikonfirmasi user

1. **Kolom `type` dihapus langsung atau dipertahankan sementara (untuk backfill)?**
   Rencana sebelumnya (`docs/plan/perubahan-skema-tagihan.md`, 2026-08-08) mencatat *"belum ada data production nyata — migrasi skema boleh dilakukan bebas"*. Kalau masih berlaku, `type` bisa langsung di-drop dalam satu migrasi tanpa backfill. Rencana di bawah tetap menulis langkah backfill (aman di kedua kondisi) — kalau dikonfirmasi belum ada data nyata, langkah backfill itu boleh dilewati.

2. **`revenue_account_id`/`receivable_account_id` boleh diubah setelah komponen biaya dibuat (lewat `UpdateFeeComponent`), atau immutable seperti `type` sekarang?**
   `FeeComponent.Update()` saat ini sengaja tidak menerima `type` baru. Diusulkan **boleh diubah** untuk kedua akun — jurnal yang sudah terbit menyimpan `account_id` hasil resolusi saat itu di `journal_entry_lines` (tidak berubah retroaktif), jadi mengubah mapping ke depan aman. Perlu konfirmasi user — kalau tidak, keduanya tetap immutable pasca-create.

3. **Validasi tipe akun yang boleh dipilih** — sekarang bisa lebih presisi karena `accounts.sub_type` sudah ada (`docs/plan/coa-sub-type.md`):
   - `revenue_account_id` → akun harus `type = 'revenue'` (kedua sub-tipe `operating_revenue` **dan** `non_operating_revenue` diterima — sengaja tidak dipersempit ke `operating_revenue` saja, supaya `4500 Pendapatan Donasi`/`4600 Pendapatan Lainnya` tetap bisa dipakai komponen biaya, sesuai motivasi awal rencana ini).
   - `receivable_account_id` → akun harus `sub_type = 'receivable'` **secara spesifik** (bukan sekadar `type = asset` seperti draft sebelumnya) — inilah gunanya `sub_type`: sekarang sistem bisa menolak kalau admin salah pilih akun kas/aset tetap sebagai "akun piutang" komponen biaya, bukan cuma menerima akun apa saja bertipe `asset`.
   - Keduanya harus `is_postable = true` dan `is_active = true` (mirror `Account.EnsurePostable()`).
   - Validasi ini letaknya di `application/command/create_fee_component.go`/`update_fee_component.go` (langkah 6 di bawah), memanggil `accountRepo.FindByID` lalu cek `Type`/`SubType` sesuai aturan di atas.

---

## Langkah implementasi

### 1. Migrasi DB

File baru menyusul `20260809140000_invoice_billing_period_nullable.up.sql` (migrasi terakhir saat ini):

```sql
ALTER TABLE fee_components
    ADD COLUMN revenue_account_id UUID REFERENCES accounts(id),
    ADD COLUMN receivable_account_id UUID REFERENCES accounts(id);

-- backfill dari mapping revenue lama
UPDATE fee_components fc SET revenue_account_id = a.id
FROM accounts a
WHERE (fc.type = 'spp' AND a.code = '4100')
   OR (fc.type = 'ukt' AND a.code = '4200')
   OR (fc.type = 'daftar_ulang' AND a.code = '4300')
   OR (fc.type = 'insidental' AND a.code = '4400');

-- backfill piutang: semua baris lama memakai akun 1103 yang sama
UPDATE fee_components fc SET receivable_account_id = a.id
FROM accounts a WHERE a.code = '1103';

ALTER TABLE fee_components
    ALTER COLUMN revenue_account_id SET NOT NULL,
    ALTER COLUMN receivable_account_id SET NOT NULL,
    DROP COLUMN type;
DROP INDEX IF EXISTS idx_fee_components_type;
CREATE INDEX idx_fee_components_revenue_account ON fee_components(revenue_account_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_fee_components_receivable_account ON fee_components(receivable_account_id) WHERE deleted_at IS NULL;
```

Kalau keputusan #1 = "belum ada data production": kedua langkah `UPDATE ... backfill` boleh dihapus (kolom langsung `NOT NULL` di database kosong). Siapkan `down.sql` yang membalikkan (drop kedua kolom akun, tambah balik `type` + CHECK constraint + index lama) — tanpa merekonstruksi data lama.

### 2. Domain: `FeeComponent` entity

`internal/modules/keuangan/domain/feecomponent/entity/fee_component.go`:
- Ganti field `Type constant.FeeComponentType` → dua field `RevenueAccountID string` dan `ReceivableAccountID string`.
- `NewFeeComponent(...)` ganti parameter `feeType constant.FeeComponentType` → `revenueAccountID, receivableAccountID string`; hapus validasi `constant.IsValidFeeType`, ganti validasi non-empty untuk kedua id (validasi "akun ini benar-benar tipe & status yang tepat" dilakukan di application layer yang punya akses `AccountRepository` — lihat langkah 6).
- Kalau keputusan #2 → "boleh diubah": tambah kedua parameter ke `Update()`. Kalau immutable: `Update()` tetap seperti sekarang.
- Hapus (setelah cek tidak ada pemakaian lain) `constant.FeeComponentType`, `FeeTypeUKT/SPP/DaftarUlang/Insidental`, `ValidFeeTypes`, `IsValidFeeType` di `domain/feecomponent/constant/fee_component_constant.go` — `PeriodType` di file yang sama **tetap**.

### 3. `AutoPostingService` — inti perubahan (4 fungsi, bukan 3)

`internal/modules/keuangan/domain/journal/service/auto_posting.go`:
- Hapus map `feeTypeRevenueAccount` (baris 40-45) dan **hapus semua `s.accountRepo.FindByCode(ctx, "1103")`** (4 tempat: baris 94, 159, 215, 275).
- `PostInvoiceIssued`, `PostInvoiceCancelled`, `PostAdjustment`: ganti parameter `feeType feeConst.FeeComponentType` → **dua** parameter `revenueAccountID, receivableAccountID string`. Resolusi: `accountRepo.FindByID(ctx, revenueAccountID)` + `accountRepo.FindByID(ctx, receivableAccountID)`, masing-masing dicek `EnsurePostable()` dan tipe yang sesuai (`revenue`/`asset`).
- `PostPaymentVerified`: tambah parameter baru `receivableAccountID string` (di samping `debitAccountID` yang sudah ada); ganti `FindByCode(ctx, "1103")` → `accountRepo.FindByID(ctx, receivableAccountID)` + `EnsurePostable()`.
- `journalConst.CodeJournalAccountMappingNotFound` jadi sebagian besar obsolete (dulu untuk "type belum dipetakan") — ganti pesan ke skenario baru ("akun pendapatan/piutang tidak ditemukan, tidak aktif, atau bukan tipe yang sesuai"), atau pertahankan kode error yang sama dengan pesan baru supaya konsumen API yang sudah cek kode ini tidak break.

### 4. Repository & interface komponen biaya

- `infrastructure/persistence/postgres_fee_component_repo.go`: kolom `type` → `revenue_account_id, receivable_account_id` di `feeComponentColumns`, `Save`/`Update` SQL, `scan()` (buang parsing `constant.FeeComponentType(feeType)`), filter `List()` (`WHERE type=$n` → filter baru, misal `WHERE revenue_account_id=$n` kalau memang butuh filter list by akun — cek dulu apakah filter ini masih relevan secara UX, bisa juga filter ini di-drop kalau tidak dipakai).
- `domain/feecomponent/repository/interfaces.go`: `FeeComponentListQuery.Type *string` → sesuaikan (kemungkinan dihapus kalau filter by type tidak lagi masuk akal, karena tidak ada lagi "jenis" biaya yang mengelompokkan).

### 5. DTO (`application/dto/fee_component_dto.go`)

- `CreateFeeComponentRequest.Type string` → `RevenueAccountID string` + `ReceivableAccountID string` (keduanya `binding:"required,uuid"`).
- `UpdateFeeComponentRequest`: tambah kedua field kalau keputusan #2 = boleh diubah (kalau tidak, tetap seperti sekarang).
- `FeeComponentListQuery.Type *string` → hapus atau ganti sesuai keputusan filter di langkah 4.
- `FeeComponentResponse.Type string` → `RevenueAccount *AccountBriefResponse{ID, Code, Name}` dan `ReceivableAccount *AccountBriefResponse{ID, Code, Name}` (nested, supaya konsumen API tahu nama/kode akun tanpa request terpisah).

### 6. Application layer

- `application/command/create_fee_component.go`: ganti validasi `constant.IsValidFeeType(req.Type)` dengan dua validasi terpisah:
  - `accountRepo.FindByID(ctx, req.RevenueAccountID)` → cek `Type == revenue` + `EnsurePostable()`.
  - `accountRepo.FindByID(ctx, req.ReceivableAccountID)` → cek `Type == asset` + `EnsurePostable()`.
  Butuh inject `AccountRepository` ke use case ini (belum ada — cek constructor `NewCreateFeeComponentUseCase`).
- `application/command/update_fee_component.go`: sama seperti di atas, hanya kalau keputusan #2 = boleh diubah.
- `application/query/list_fee_components.go`: response mapping `Type: string(fc.Type)` → resolve & sertakan info kedua akun (batch-load semua akun relevan sekali per request untuk hindari N+1, bukan query per baris).

### 7. Titik pemanggilan `AutoPostingService` (5 tempat — bukan 4)

- `application/command/create_invoice.go:183` (`PostInvoiceIssued`) — kirim `fee.RevenueAccountID, fee.ReceivableAccountID`.
- `application/command/create_invoice_batch.go:284` (`processTarget`, loop batch) — sama.
- `application/command/cancel_invoice.go:55` (`PostInvoiceCancelled`) — sama.
- `application/command/apply_adjustment.go:75` (`PostAdjustment`) — sama.
- **`application/command/verify_payment.go:87-91`** (`PostPaymentVerified`) — **titik baru**: use case ini saat ini hanya punya `paymentRepo` dan `invoiceRepo` (lihat `NewVerifyPaymentUseCase`, `verify_payment.go:21-30`), belum fetch `FeeComponent`. Perlu:
  1. Tambah dependency `feeRepo feecomponentRepo.FeeComponentRepository` ke struct & constructor.
  2. Di `Execute`, setelah `inv, err := uc.invoiceRepo.FindByID(...)` (baris 45), tambah `fee, err := uc.feeRepo.FindByID(ctx, inv.FeeComponentID)`.
  3. Teruskan `fee.ReceivableAccountID` ke `PostPaymentVerified` (baris 87-91).

### 8. Test

- `domain/feecomponent/entity/fee_component_test.go`: update semua pemanggilan `NewFeeComponent(..., feeConst.FeeTypeSPP, ...)` → pakai dua UUID dummy (`revenueAccountID`, `receivableAccountID`).
- **Tambah test baru untuk `AutoPostingService`** (`domain/journal/service/auto_posting.go`) — saat ini **tidak ada test sama sekali** untuk file ini, dan perubahan ini menyentuh langsung logika inti resolusi akun di 4 fungsi. Minimal butuh unit test (mock `AccountRepository`) untuk `PostInvoiceIssued`/`PostPaymentVerified`/`PostInvoiceCancelled`/`PostAdjustment` yang menutupi: kedua akun ditemukan & valid → jurnal benar dengan akun sesuai per fee component (termasuk kasus 2 fee component berbeda memakai 2 akun piutang berbeda); salah satu akun tidak ditemukan/tidak postable/tidak aktif → error; akun dengan tipe salah (revenue dipakai sebagai piutang atau sebaliknya) → error.
- Test baru untuk `VerifyPaymentUseCase` (kalau belum ada) mencakup resolusi `fee.ReceivableAccountID` dari `inv.FeeComponentID`.

### 9. Dokumentasi yang harus diupdate

- `docs/rules/akuntansi.md:39` — ganti catatan "mapping type→code hardcode revenue" jadi menjelaskan desain baru (kedua akun — revenue & piutang — dipilih dari COA, tervalidasi tipe+postable+active).
- `docs/schemas/keuangan-akuntansi.md:18` — update definisi kolom `fee_components.type VARCHAR(30) CHECK (...)` jadi `revenue_account_id UUID REFERENCES accounts(id)` + `receivable_account_id UUID REFERENCES accounts(id)`.
- `docs/plan/keuangan-module.md:338-345` — tandai "Fase 2" (mapping bisa diedit dari COA) sebagai selesai/digantikan rencana ini.

---

## Verifikasi

- `go build ./...`, `go vet ./...`, `go test ./internal/modules/keuangan/...`.
- Migrasi: `up` lalu `down` jalan bersih tanpa error di database kosong maupun (kalau relevan) database dengan data existing hasil backfill.
- Manual (lewat API, `admin/keuangan/fee-components` + `admin/keuangan/invoices` + `admin/keuangan/payments`):
  1. Buat 2 fee component dengan `receivable_account_id` **berbeda** (misal "Piutang SPP" vs "Piutang UKT", kalau akun-akun itu sudah ada di COA — kalau belum, tambah dulu via COA sebelum test) dan `revenue_account_id` yang juga berbeda (termasuk memakai `4600 Pendapatan Lainnya`, yang sebelumnya tidak bisa dipakai sama sekali).
  2. Terbitkan invoice dari masing-masing fee component → cek jurnal yang tercipta mendebit **akun piutang yang sesuai per komponen**, bukan `1103` untuk semua.
  3. Verifikasi pembayaran salah satu invoice tersebut → cek jurnal payment verified mengkredit akun piutang yang **sama** dengan yang didebit saat invoice terbit (bukan `1103` hardcoded).
  4. Batalkan salah satu invoice yang belum dibayar → cek jurnal pembalik memakai akun piutang & revenue yang benar.
  5. Ajukan adjustment atas invoice serupa → cek jurnal adjustment memakai akun yang benar.
  6. Coba buat fee component dengan `receivable_account_id` menunjuk akun `type=asset` tapi `sub_type` **bukan** `receivable` (mis. `1101 Kas`) → ditolak dengan pesan jelas — ini skenario yang sebelumnya (draft tanpa `sub_type`) lolos begitu saja. Sebaliknya untuk `revenue_account_id` menunjuk akun `type=asset`/`expense`/dll → ditolak.
  7. Coba menunjuk akun yang `is_active = false` / `is_postable = false` untuk salah satu dari kedua field → ditolak.
  8. (Kalau keputusan #2 = boleh diubah) Ubah `receivable_account_id` sebuah fee component yang sudah punya invoice & pembayaran lama → jurnal **lama** tidak berubah, hanya transaksi **baru** setelah perubahan yang memakai akun baru.

---

## Di luar scope rencana ini

- Halaman/komponen frontend untuk memilih kedua akun saat membuat komponen biaya (repo ini backend-only; perubahan kontrak API di atas perlu dikomunikasikan ke tim frontend `sipon-ui` secara terpisah).
- Perubahan struktur `invoice_items`/alokasi pembayaran lintas komponen (dicatat sebagai isu terpisah di `docs/plan/perubahan-skema-tagihan.md`).
- Validasi `sub_type = cash_bank` untuk `debitAccountID` di `PostPaymentVerified`/`CreateManualPaymentUseCase` (akun kas/bank yang dipilih bendahara saat mencatat pembayaran) — ini "lawan" dari validasi `receivable_account_id` di atas, tapi dikerjakan di `docs/plan/coa-sub-type.md` (bagian "Konsumen sub_type #1"), bukan di rencana ini, karena tidak menyentuh `fee_components` sama sekali.
