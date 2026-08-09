# Rencana: Tambah `sub_type` pada Chart of Accounts (COA)

## Context

`accounts.type` saat ini hanya punya 5 nilai besar (`asset|liability|equity|revenue|expense` — `domain/account/constant/account_constant.go:15-22`). Ini terlalu kasar untuk beberapa kebutuhan validasi yang muncul dari diskusi modul komponen biaya (lihat `docs/plan/fee-component-account-id.md`):

- **Tidak bisa membedakan akun piutang dari akun asset lain.** `1103 Piutang Santri`, `1101 Kas`, `1102 Bank`, `1201 Tanah`, dst. semuanya `type = asset` — tidak ada cara memvalidasi "akun ini khusus piutang" tanpa menambah field baru.
- **Tidak ada validasi jenis akun untuk `payments.debit_account_id`** (akun kas/bank yang dipilih bendahara saat mencatat pembayaran manual). Riset menemukan gap nyata: `CreateManualPaymentUseCase` (`application/command/create_manual_payment.go:52-58`) menerima `req.DebitAccountID` **tanpa validasi tipe/sub-tipe apapun** — hanya `binding:"required"` di DTO (`application/dto/payment_dto.go:3-12`). Satu-satunya gate adalah `Account.EnsurePostable()` (`domain/account/entity/account.go:85-90`), yang cuma cek `is_postable && is_active`, **tidak cek `type`**. Akibatnya bendahara secara teknis bisa memilih `1103 Piutang Santri`, `2101 Utang Usaha`, `3100 Modal`, `4100 Pendapatan SPP`, atau bahkan `5100 Beban Gaji` sebagai "akun kas" pembayaran — semua accounts itu `is_postable = true` di seed data (`internal/seeders/coa_seeder.go:20-56`), tidak ada satupun yang ditolak sistem hari ini.

**Solusi**: tambah kolom `sub_type` di `accounts` — klasifikasi yang lebih halus di dalam tiap `type`, mengikuti pola umum teori akuntansi (kas & bank, piutang, aset tetap, dst.). `sub_type` ini lalu dipakai untuk:
1. Validasi `fee_components.receivable_account_id` harus akun `sub_type = receivable` (bukan sekadar `type = asset`) — lihat `docs/plan/fee-component-account-id.md`, yang **bergantung pada rencana ini**.
2. Validasi `payments.debit_account_id` harus akun `sub_type = cash_bank` — menutup gap yang ditemukan di atas.

Rencana ini **fondasi** (skema + validasi generik di COA), dipisah dari `fee-component-account-id.md` (yang jadi konsumen) karena scope masing-masing sudah cukup besar sendiri-sendiri.

---

## Taksonomi `sub_type` (mengikuti teori akuntansi umum)

Bukan cuma yang dipakai fee-component/payment — daftar ini dibuat lengkap per `type` supaya COA punya kerangka standar yang bisa dipakai untuk kebutuhan lain di masa depan (laporan yang lebih rinci, dst.), bukan sekadar 2 sub-tipe yang kepakai sekarang.

| `type` | `sub_type` | Nama Indonesia | Akun seed existing yang cocok |
|---|---|---|---|
| asset | `cash_bank` | Kas & Bank | 1101, 1102 |
| asset | `receivable` | Piutang | 1103 |
| asset | `prepaid_expense` | Biaya Dibayar Dimuka | *(baru)* |
| asset | `inventory` | Persediaan | *(baru)* |
| asset | `fixed_asset` | Aset Tetap | 1201, 1202, 1203, 1204 |
| asset | `accumulated_depreciation` | Akumulasi Penyusutan (kontra-aset) | *(baru)* |
| asset | `intangible_asset` | Aset Tidak Berwujud | *(baru)* |
| asset | `investment` | Investasi | *(baru)* |
| asset | `other_asset` | Aset Lainnya | *(baru)* |
| liability | `payable` | Utang Usaha | 2101 |
| liability | `customer_advance` | Uang Muka Pelanggan/Santri | 2102 |
| liability | `unearned_revenue` | Pendapatan/Biaya Diterima Dimuka | 2103 |
| liability | `tax_payable` | Utang Pajak | *(baru)* |
| liability | `accrued_liability` | Beban Masih Harus Dibayar | *(baru)* |
| liability | `long_term_liability` | Liabilitas Jangka Panjang | *(baru)* |
| liability | `other_liability` | Liabilitas Lainnya | *(baru)* |
| equity | `capital` | Modal | 3100 |
| equity | `retained_earnings` | Laba Ditahan/Saldo Laba | 3200 |
| equity | `current_year_earnings` | Laba Tahun Berjalan | 3201 |
| equity | `withdrawal` | Prive/Distribusi | *(baru)* |
| revenue | `operating_revenue` | Pendapatan Operasional | 4100, 4200, 4300, 4400 |
| revenue | `non_operating_revenue` | Pendapatan Non-Operasional | 4500, 4600 |
| expense | `cost_of_goods_sold` | Beban Pokok/HPP | *(baru)* |
| expense | `operating_expense` | Beban Operasional | 5100, 5200, 5300 |
| expense | `depreciation_expense` | Beban Penyusutan | 5400 |
| expense | `non_operating_expense` | Beban Non-Operasional | 5500 |
| expense | `tax_expense` | Beban Pajak | *(baru)* |

27 sub-tipe total. Akun grup/root (`1000`, `1100`, `1200`, `2000`, `2100`, `3000`, `4000`, `5000` — semua `is_postable = false`) **tidak diberi `sub_type`** (tetap `NULL`) — sub-tipe hanya relevan untuk akun yang benar-benar dipakai posting.

**Catatan khusus `accumulated_depreciation`**: ini akun kontra-aset — meski `type = asset`, saldo normalnya `credit` (bukan `debit` seperti akun asset lain). Ini pengecualian yang sah secara teori akuntansi, bukan bug — pastikan tidak "diperbaiki" jadi `debit` saat implementasi.

---

## Keputusan desain yang perlu dikonfirmasi user

1. **`sub_type` nullable di DB, wajib diisi di level aplikasi untuk akun postable.** Akun grup (`is_postable = false`) tidak butuh sub_type. Akun postable **wajib** punya sub_type — divalidasi di `CreateAccountUseCase`/`UpdateAccountUseCase`, bukan lewat DB `NOT NULL` (karena baris grup tetap butuh `NULL`).
2. **Cross-check `sub_type` harus valid untuk `type` yang dipilih** (mis. `cash_bank` tidak valid untuk `type = revenue`) — divalidasi di application layer via mapping tabel di atas, bukan lewat SQL `CHECK` lintas kolom (mengikuti pola yang sudah ada: `normal_balance` juga tidak divalidasi lintas kolom terhadap `type` oleh DB, konsisten dengan desain existing).
3. **`sub_type` boleh diubah setelah akun dibuat (lewat `UpdateAccount`)?** Diusulkan **boleh**, mengikuti pola `is_postable` yang juga bisa diubah di `Update()` (`account.go:50-59`) — beda dari `type` yang immutable. Perlu konfirmasi user.
4. **Skema baru (13 akun) yang ditambahkan ke seeder untuk melengkapi semua 27 sub-tipe** (lihat tabel di atas) bersifat **contoh/starter**, bukan akun wajib dipakai — semua `is_system = false` sehingga admin bebas menonaktifkan/mengedit/menghapus lewat COA management yang sudah ada kalau tidak relevan untuk institusi tertentu (mis. pesantren mungkin tidak butuh `inventory`/`cost_of_goods_sold` sama sekali).

---

## Langkah implementasi

### 1. Migrasi DB

Migrasi baru (menyusul migrasi terakhir keuangan saat ini):

```sql
ALTER TABLE accounts ADD COLUMN sub_type VARCHAR(40) CHECK (sub_type IN (
    'cash_bank','receivable','prepaid_expense','inventory','fixed_asset',
    'accumulated_depreciation','intangible_asset','investment','other_asset',
    'payable','customer_advance','unearned_revenue','tax_payable',
    'accrued_liability','long_term_liability','other_liability',
    'capital','retained_earnings','current_year_earnings','withdrawal',
    'operating_revenue','non_operating_revenue',
    'cost_of_goods_sold','operating_expense','depreciation_expense',
    'non_operating_expense','tax_expense'
));
CREATE INDEX idx_accounts_sub_type ON accounts(sub_type) WHERE deleted_at IS NULL;

-- backfill akun existing yang postable
UPDATE accounts SET sub_type = 'cash_bank' WHERE code IN ('1101','1102');
UPDATE accounts SET sub_type = 'receivable' WHERE code = '1103';
UPDATE accounts SET sub_type = 'fixed_asset' WHERE code IN ('1201','1202','1203','1204');
UPDATE accounts SET sub_type = 'payable' WHERE code = '2101';
UPDATE accounts SET sub_type = 'customer_advance' WHERE code = '2102';
UPDATE accounts SET sub_type = 'unearned_revenue' WHERE code = '2103';
UPDATE accounts SET sub_type = 'capital' WHERE code = '3100';
UPDATE accounts SET sub_type = 'retained_earnings' WHERE code = '3200';
UPDATE accounts SET sub_type = 'current_year_earnings' WHERE code = '3201';
UPDATE accounts SET sub_type = 'operating_revenue' WHERE code IN ('4100','4200','4300','4400');
UPDATE accounts SET sub_type = 'non_operating_revenue' WHERE code IN ('4500','4600');
UPDATE accounts SET sub_type = 'operating_expense' WHERE code IN ('5100','5200','5300');
UPDATE accounts SET sub_type = 'depreciation_expense' WHERE code = '5400';
UPDATE accounts SET sub_type = 'non_operating_expense' WHERE code = '5500';
```

`down.sql`: `DROP INDEX idx_accounts_sub_type; ALTER TABLE accounts DROP COLUMN sub_type;`.

### 2. Domain: `Account` entity & constant

`domain/account/constant/account_constant.go`:
- Tambah `type AccountSubType string` + 27 konstanta (lihat tabel taksonomi).
- Tambah `var SubTypesByType = map[AccountType][]AccountSubType{...}` (persis mengikuti pengelompokan di tabel) + helper `IsValidSubTypeForType(t AccountType, st AccountSubType) bool`.

`domain/account/entity/account.go`:
- Tambah field `SubType *constant.AccountSubType` (pointer — nullable, mengikuti pola `ParentID *string`/`Description *string` yang sudah ada di struct ini).
- `NewAccount(...)` tambah parameter `subType *constant.AccountSubType`; validasi (kalau `subType != nil`) via `constant.IsValidSubTypeForType(accType, *subType)` — kalau `IsPostable` true (yaitu `level > 0`, sesuai logika `NewAccount` baris 40) dan `subType == nil`, tolak (lihat keputusan #1).
- `Update(...)`: kalau keputusan #3 = boleh diubah, tambah parameter `subType *constant.AccountSubType` + validasi ulang terhadap `a.Type` (yang immutable). Kalau tidak, `Update()` tetap seperti sekarang.

### 3. Repository & DTO

- `infrastructure/persistence/postgres_account_repo.go`: tambah kolom `sub_type` ke semua SQL (`Save`/`Update`/`scan()`/`List`).
- `domain/account/repository/interfaces.go`: `AccountListQuery` tambah `SubType *string` (mengikuti pola `Type *string` yang sudah ada).
- `application/dto/account_dto.go`: `CreateAccountRequest` tambah `SubType *string`; `UpdateAccountRequest` tambah `SubType *string` (kalau keputusan #3 = boleh diubah); `AccountListQuery` (DTO) tambah `SubType *string`; `AccountResponse`/`AccountBriefResponse` tambah `SubType *string`.

### 4. Application layer

- `application/command/create_account.go`: setelah `accType := accConst.AccountType(req.Type)`, resolve `subType` dari `req.SubType` (kalau ada) dan validasi lewat `constant.IsValidSubTypeForType`; kalau akun postable (`req.IsPostable` true / `level > 0`) dan `req.SubType` kosong → tolak dengan pesan jelas ("akun postable wajib punya sub-tipe").
- `application/command/update_account.go`: sama, kalau keputusan #3 = boleh diubah.
- Tidak ada use case list/get yang perlu logika baru selain passthrough field (mapping ke response sudah cukup).

### 5. Seeder — lengkapi semua 27 sub-tipe

`internal/seeders/coa_seeder.go`:
- Tambah field `SubType string` ke struct `acct` (baris 16-19), isi untuk semua 31 baris existing (kosong/`""` untuk 8 baris grup/root) sesuai tabel taksonomi.
- Tambah `sub_type` ke kolom `INSERT`/`ON CONFLICT DO UPDATE` (baris 72-86).
- **Tambah 13 akun baru** untuk melengkapi sub-tipe yang belum terwakili (semua `IsSystem: 0`, level & postable sesuai pola existing):

  | Code | Name | Type | NormalBalance | SubType |
  |---|---|---|---|---|
  | 1104 | Biaya Dibayar Dimuka | asset | debit | prepaid_expense |
  | 1105 | Persediaan | asset | debit | inventory |
  | 1205 | Akumulasi Penyusutan Aset Tetap | asset | **credit** (kontra-aset) | accumulated_depreciation |
  | 1300 | Aset Tidak Berwujud | asset | debit | intangible_asset |
  | 1400 | Investasi Jangka Panjang | asset | debit | investment |
  | 1500 | Aset Lainnya | asset | debit | other_asset |
  | 2104 | Utang Pajak | liability | credit | tax_payable |
  | 2105 | Beban Masih Harus Dibayar | liability | credit | accrued_liability |
  | 2200 | Liabilitas Jangka Panjang | liability | credit | long_term_liability |
  | 2300 | Liabilitas Lainnya | liability | credit | other_liability |
  | 3300 | Prive/Distribusi | equity | debit | withdrawal |
  | 5600 | Beban Pokok Penjualan (HPP) | expense | debit | cost_of_goods_sold |
  | 5700 | Beban Pajak | expense | debit | tax_expense |

  Catatan: `1104`/`1105` sudah otomatis kebagian parent `1100` di `parentCodeFor()` (baris 122: range `1101 <= code <= 1105` sudah ada dari awal, belum pernah dipakai). Sisanya butuh **tambahan case baru** di `parentCodeFor()` (`internal/seeders/coa_seeder.go:118-141`): perluas `1201-1204` → `1201-1205`, `2101-2103` → `2101-2105`, dan tambah case eksplisit untuk `1300→1000`, `1400→1000`, `1500→1000`, `2200→2000`, `2300→2000`, `3300→3000`, `5600→5000`, `5700→5000`.

### 6. Konsumen sub_type #1 — `payments.debit_account_id` (menutup gap validasi)

- `application/command/create_manual_payment.go`: sebelum `payEntity.NewPayment(...)`, tambah `debitAcc, err := uc.accountRepo.FindByID(ctx, req.DebitAccountID)` → error kalau tidak ketemu; `debitAcc.EnsurePostable()`; **tambahan baru**: `if debitAcc.SubType == nil || *debitAcc.SubType != accConst.SubTypeCashBank { tolak }`. Ini butuh inject `AccountRepository` ke use case ini (cek constructor `NewCreateManualPaymentUseCase` — kemungkinan belum ada dependency ini).
- Sebagai defense-in-depth (opsional, tapi disarankan karena `AutoPostingService.PostPaymentVerified` adalah titik terakhir sebelum jurnal tercipta): tambah validasi yang sama di `domain/journal/service/auto_posting.go:151-157` (setelah `debitAcc.EnsurePostable()`) — cek `sub_type = cash_bank` juga di sana, supaya kalau ada jalur lain di masa depan yang memanggil `PostPaymentVerified` tanpa lewat `CreateManualPaymentUseCase`, validasi tetap tegak.
- **Catatan tambahan (perbaikan kecil yang bisa ikut, opsional)**: kolom `payments.debit_account_id` di migrasi (`migrations/20260808213506_create_keuangan_tables.up.sql:96`) ternyata **tidak punya FK constraint** ke `accounts(id)` (`UUID` polos). Kalau mau dibenahi sekalian saat menyentuh area ini: tambah `ALTER TABLE payments ADD CONSTRAINT fk_payments_debit_account FOREIGN KEY (debit_account_id) REFERENCES accounts(id)` di migrasi yang sama. Ini independen dari `sub_type` — boleh dilewati kalau ingin scope minimal.

### 7. Konsumen sub_type #2 — `fee_components` (lihat rencana terpisah)

Detail lengkap ada di `docs/plan/fee-component-account-id.md` (sudah direvisi agar bergantung pada rencana ini). Ringkasnya: `fee_components.receivable_account_id` harus divalidasi `sub_type = receivable`, `fee_components.revenue_account_id` harus divalidasi `type = revenue` (kedua sub-tipe `operating_revenue`/`non_operating_revenue` diterima — supaya `4500`/`4600` tetap bisa dipakai, sesuai motivasi awal rencana itu).

### 8. Test

- Tambah test untuk `Account.NewAccount`/`Update` (kalau ada `account_test.go` — cek dulu) mencakup: sub_type valid untuk type → ok; sub_type tidak valid untuk type (mis. `cash_bank` + `type=revenue`) → error; akun postable tanpa sub_type → error; akun grup (`is_postable=false`) tanpa sub_type → ok.
- Test baru untuk validasi `sub_type = cash_bank` di `CreateManualPaymentUseCase` (mock `AccountRepository`): akun kas/bank asli → lolos; akun piutang/utang/modal/revenue/expense yang postable → ditolak.

---

## Verifikasi

- `go build ./...`, `go vet ./...`, `go test ./internal/modules/keuangan/...`.
- Migrasi `up`/`down` bersih; jalankan seeder ulang (`ON CONFLICT DO UPDATE` sudah idempotent) → cek 13 akun baru muncul dengan `sub_type` benar, dan 31 akun lama ter-backfill `sub_type`-nya tanpa duplikat.
- Manual (API COA + pembayaran manual):
  1. Buat akun baru dengan `type=asset`, `sub_type=cash_bank` → berhasil.
  2. Buat akun baru dengan `type=asset` tapi `sub_type=operating_revenue` (mismatch) → ditolak.
  3. Buat akun postable tanpa `sub_type` sama sekali → ditolak.
  4. Catat pembayaran manual dengan `debit_account_id` = `1101 Kas` → berhasil.
  5. Catat pembayaran manual dengan `debit_account_id` = `1103 Piutang Santri` (bukan kas) → ditolak dengan pesan jelas — ini skenario yang **hari ini lolos begitu saja**, jadi pastikan regresi ini benar-benar tertutup.

---

## Di luar scope rencana ini

- Perubahan `fee_components` itu sendiri — di `docs/plan/fee-component-account-id.md` (konsumen dari rencana ini, dikerjakan setelah/bersama rencana ini).
- Laporan keuangan (income statement, balance sheet, dst.) belum diminta untuk mengelompokkan ulang berdasarkan `sub_type` (mis. pisahkan "Pendapatan Operasional" vs "Non-Operasional" di laporan) — sub_type baru dipakai untuk validasi input di rencana ini, bukan untuk restrukturisasi laporan. Bisa jadi rencana lanjutan kalau user mau.
- Perubahan UI COA di `sipon-ui` untuk memilih `sub_type` (dropdown bergantung pada `type` yang dipilih) — repo ini backend-only, kontrak API baru perlu dikomunikasikan ke tim frontend.
