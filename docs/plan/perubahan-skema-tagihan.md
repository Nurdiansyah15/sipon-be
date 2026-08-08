# Rencana: Perbaikan Bug + Evolusi Konsep Sub-Modul Tagihan (Billing)

## Context

Selama audit menu Tagihan & Pembayaran (admin + santri) di modul keuangan, ditemukan beberapa bug fungsional yang cukup serius pada fitur "Buat Tagihan" (single) dan "Generate Massal" (batch) — termasuk satu bug kritis yang membuat santri kemungkinan **tidak pernah bisa melihat tagihan mereka sendiri** di halaman self-service. Di sisi lain, user membawa skema referensi (students, fee_components, billing_periods, billing_batches, billing_batch_targets, invoices, invoice_items) untuk mengevolusi struktur data tagihan supaya lebih matang.

Setelah riset menyeluruh (lihat ringkasan temuan di bawah) dan diskusi dengan user, disepakati:

1. **Bug diperbaiki dulu**, sebelum ada perubahan konsep apapun — supaya pondasi yang ada benar dulu.
2. Dari skema referensi, yang diadopsi sekarang: **`billing_periods`** (formalisasi periode tagihan) dan **`billing_batches` + `billing_batch_targets` dengan snapshot target** (audit trail generate massal). **`invoice_items`** (konsolidasi satu-invoice-banyak-komponen) **sengaja TIDAK dimasukkan** — terlalu berisiko (perlu desain alokasi pembayaran yang belum ada sama sekali) dan pola "satu invoice per komponen" yang ada sekarang justru punya keuntungan UX (wali santri bisa bayar SPP duluan, UKT nanti, independen).
3. Belum ada data production nyata — migrasi skema boleh dilakukan bebas tanpa script backfill data lama.
4. `students` dan `fee_components` dari skema referensi **tidak dibuat baru** — sudah ada (santri di modul kesantrian, fee_components di modul keuangan); konsepnya diadaptasi, bukan didirikan ulang.

### Ringkasan temuan bug (dasar Fase 0)

1. **`invoice.user_id` salah di KEDUA jalur pembuatan invoice** (paling kritis):
   - `CreateInvoice` (single) — `internal/modules/keuangan/interfaces/http/handler.go:392-410` — mengisi `UserID` dengan ID **admin yang login** (`middleware.GetUserID(c)`), bukan user account milik santri.
   - `CreateInvoiceBatchUseCase` (batch) — `internal/modules/keuangan/application/command/create_invoice_batch.go:87` — mengisi `userID := assignment.SantriID` (ID santri dipakai ulang sebagai user id).
   - Akibatnya `MyInvoicesUseCase`/`MyPaymentsUseCase` (`application/query/my_invoices.go:24`, `my_payments.go:35`) yang memfilter `user_id` asli dari JWT santri **tidak akan pernah cocok** — tagihan yang dibuat lewat kedua jalur ini invisible di halaman "Tagihan Saya"/"Riwayat Pembayaran" milik santri.
   - Modul kesantrian **sudah punya** `user_id` di entity `Santri` (`internal/modules/kesantrian/domain/santri/entity/santri.go:14-17`) dan repo `FindByID` sudah bisa resolve itu — tapi belum diekspos lewat `Contract` (`internal/modules/kesantrian/contract.go:92-96` cuma punya `GetSantriByUserID`, arah kebalikannya).

2. **Nomor invoice gampang bentrok** — `invVO.NewInvoiceNumberNow(1)` di single-create (`create_invoice.go:71`, SELALU seq=1) dan `NewInvoiceNumberNow(created + 1)` di batch (`create_invoice_batch.go:103`, counter lokal reset tiap run). Karena `invoice_number` `UNIQUE` di DB, tagihan kedua yang dibuat di bulan yang sama (lewat jalur manapun) akan gagal insert.

3. **Unique constraint & pengecekan duplikat tagihan tidak menyertakan `tahun_ajaran`** — `idx_invoices_unique_periode` (`migrations/20260808213506_create_keuangan_tables.up.sql:88-90`) dan `FindBySantriComponentPeriode` (`domain/invoice/repository/interfaces.go:30`, `infrastructure/persistence/postgres_invoice_repo.go:163-170`) cuma key ke `(santri_id, fee_component_id, periode)`. Kalau string periode dipakai ulang tiap tahun ajaran (memang itu tujuannya), tagihan tahun berikutnya akan dikira duplikat dan di-skip selamanya.

4. **Error "duplikat" nyasar jadi 500 generik** — `create_invoice.go`/`create_invoice_batch.go` memanggil `application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)` padahal kode error yang keluar dari repo adalah `CodeInvoiceDuplicate` (tidak match) — jatuh ke `ErrCodeInternal` (500), pola yang sama seperti bug skema-tagihan/skema-santri yang sudah diperbaiki sebelumnya di sesi ini.

5. **Tidak ada jejak audit untuk Generate Massal** — `CreateInvoiceBatchUseCase.Execute` (`create_invoice_batch.go:57-132`) murni operasi in-memory, cuma balikin pesan ringkasan string. Tidak ada batch id, tidak ada daftar santri yang jadi target tersimpan di manapun. Tabel `finance_audit_logs` sudah ada di migrasi tapi **tidak pernah dipakai** di kode manapun (dead table).

---

## FASE 0 — Perbaikan Bug (harus selesai duluan)

### 0.1 Perbaiki `invoice.user_id` (paling prioritas)

**Kesantrian module** (tambahan kapabilitas, ikuti pola `ListActiveSantriIDs`/`GetSantriByUserID` yang sudah ada):

- Tambah use-case baru di `internal/modules/kesantrian/application/query/` — bulk resolver, misal `ListActiveSantriWithUserIDUseCase` yang mengembalikan `[]{SantriID, UserID}` untuk semua santri aktif sekaligus (hindari N+1 saat dipanggil dari batch). Bisa dibangun dari repo method yang sudah ada (`santriRepo.FindByID`/list aktif) — cek `internal/modules/kesantrian/domain/santri/repository/interfaces.go` untuk method list aktif yang sudah ada dan tambahkan variant yang menyertakan `user_id`.
- Tambah method baru ke `Contract` (`internal/modules/kesantrian/contract.go:92-96`), implementasikan di `*Module` (ikuti pola persis `ListActiveSantriIDs`/`GetSantriByUserID` di `internal/modules/kesantrian/module.go:208-215`).

**Keuangan module**:

- Tambah method baru ke `ports.KesantrianReader` (`internal/modules/keuangan/application/ports/kesantrian_reader.go`) + pass-through di `infrastructure/kesantriangateway/gateway.go`.
- `create_invoice_batch.go`: ganti `userID := assignment.SantriID` dengan hasil resolve dari kapabilitas baru di atas (ambil sekali di awal `Execute`, bentuk jadi map `santriID -> userID`, dipakai di dalam loop — bukan panggil satu-satu per santri).
- `create_invoice.go` / `CreateInvoiceCmd`: tambah field `UserID` yang diisi dari resolve santri_id→user_id (bukan dari `middleware.GetUserID(c)`, yang tetap dipakai HANYA untuk `CreatedBy`). Perbaiki `handler.go`'s `CreateInvoice` supaya tidak lagi menyamakan `UserID` dengan admin yang login — resolve dulu via kapabilitas baru (lewat use-case, bukan langsung di handler).

### 0.2 Perbaiki penomoran invoice

- Ganti `NewInvoiceNumberNow(1)`/`NewInvoiceNumberNow(created+1)` dengan sequence yang benar-benar atomic. Pendekatan paling murah tanpa infra baru: pakai Postgres sequence per tahun (`CREATE SEQUENCE invoice_number_seq`) atau tabel counter kecil dengan `UPDATE ... RETURNING` di dalam transaksi row-lock (`SELECT ... FOR UPDATE`), dieksekusi di `PostgresInvoiceRepository.Save` (atau method baru `NextInvoiceNumber(ctx) (string, error)`) — bukan dihitung di application layer dari counter lokal.
- Migrasi baru `013_add_invoice_number_sequence` (naik dari `012` yang terakhir).

### 0.3 Perbaiki unique constraint duplikat tagihan (tactical — akan digantikan Fase 1)

- Migrasi: drop `idx_invoices_unique_periode`, buat ulang menyertakan `tahun_ajaran`: `UNIQUE(santri_id, fee_component_id, periode, tahun_ajaran)`.
- Update `FindBySantriComponentPeriode` (interface + postgres impl) untuk menerima & memfilter `tahunAjaran` juga; update kedua caller (`create_invoice.go`, `create_invoice_batch.go`).
- **Catatan**: ini perbaikan cepat/aman sekarang. Fase 1 di bawah akan menggantikan mekanisme ini sepenuhnya dengan `billing_period_id` (FK, bukan pasangan string) — jadi perubahan di sini kecil & sementara secara sengaja.

### 0.4 Perbaiki pemetaan kode error "duplikat tagihan"

- `create_invoice.go`/`create_invoice_batch.go`: ganti `application.WrapRepoErr(err, invConst.CodeInvoiceNotFound)` (salah kode) jadi `application.WrapConflictErr(err, invConst.CodeInvoiceDuplicate)` di titik yang relevan — pola yang sama persis dengan fix `AssignSchemeToSantri`/`AddSchemeItem` sebelumnya di sesi ini.

**Verifikasi Fase 0**: `go build ./...`, `go vet ./...`, `go test ./internal/modules/keuangan/... ./internal/modules/kesantrian/...`. Manual: (a) buat 2 tagihan individual di bulan yang sama → nomor tidak boleh bentrok; (b) generate massal 2 skema berbeda di bulan yang sama → nomor tidak boleh bentrok; (c) buat tagihan dengan periode yang sama di tahun_ajaran berbeda → tidak boleh dianggap duplikat; (d) login sebagai santri yang baru saja dibuatkan tagihan (individual maupun batch) → harus muncul di "Tagihan Saya".

---

## FASE 1 — Formalisasi Periode Tagihan (`billing_periods`)

Adaptasi dari skema referensi: `periode`/`tahun_ajaran` (string bebas) diganti jadi entity master `billing_periods` dengan lifecycle draft→open→closed. Karena belum ada data production, kolom string lama **dihapus langsung** (bukan dipertahankan berdampingan) — normalisasi sekarang jauh lebih murah daripada nanti.

**Domain baru** — ikuti pola persis `internal/modules/keuangan/domain/billingscheme/` (constant/entity/repository) dan `domain/period/` (untuk lifecycle status):

- `domain/billingperiod/entity/billing_period.go`: `BillingPeriod{ID, Name, PeriodType (reuse enum monthly/semesterly/yearly/once dari `feecomponent/constant`), StartDate, EndDate, Status(draft/open/closed), CreatedBy, CreatedAt, UpdatedAt}`.
- `domain/billingperiod/constant/`, `domain/billingperiod/repository/interfaces.go`.
- `infrastructure/persistence/postgres_billing_period_repo.go`.
- `application/command/create_billing_period.go`, `open_billing_period.go`, `close_billing_period.go`.
- `application/query/list_billing_periods.go`, `get_billing_period.go`.
- `application/dto/billing_period_dto.go` (request/response, ikuti pola `billing_scheme_dto.go`).

**Migrasi** (`014_create_billing_periods.up/down.sql`):

- `CREATE TABLE billing_periods (...)`.
- `ALTER TABLE invoices ADD COLUMN billing_period_id UUID NOT NULL REFERENCES billing_periods(id)`, **DROP COLUMN periode, DROP COLUMN tahun_ajaran**.
- Ganti `idx_invoices_unique_periode` jadi `UNIQUE(santri_id, fee_component_id, billing_period_id)` — menggantikan tactical fix Fase 0.3.
- Sesuaikan index `idx_invoices_periode`/`idx_invoices_tahun_ajaran` jadi index ke `billing_period_id`.

**Perubahan di keuangan module**:

- `dto.CreateInvoiceRequest`/`CreateInvoiceBatchRequest`: ganti `Periode`/`TahunAjaran` (2 string) jadi `BillingPeriodID string`.
- `dto.InvoiceResponse`: ganti `Periode`/`TahunAjaran` jadi `BillingPeriod *BillingPeriodBriefResponse` (nested, pola sama seperti `FeeComponentBriefResponse`), enrichment lewat `buildInvoiceResponse` (`application/query/invoice_response.go`).
- `FindBySantriComponentPeriode` → `FindBySantriComponentPeriod(ctx, santriID, feeComponentID, billingPeriodID)`.
- `create_invoice.go`, `create_invoice_batch.go`: pakai `BillingPeriodID` alih-alih string.
- `dto.InvoiceListQuery`: `Periode`/`TahunAjaran` filter jadi `BillingPeriodID` filter.
- `report_summary.go`, `report_outstanding.go`: `JOIN billing_periods bp ON bp.id = i.billing_period_id`, `GROUP BY bp.id` (tampilkan `bp.name` alih-alih string periode+tahun_ajaran mentah).
- Handler baru di `KeuanganHandler` + route di `router.go`: `GET/POST /admin/billing-periods`, `POST /admin/billing-periods/:id/open`, `POST /admin/billing-periods/:id/close`.

**Frontend (sipon-ui)**:

- Halaman admin baru: `pages/admin/keuangan/periode-tagihan/index.vue` (list + form create), mirip pola `pages/admin/keuangan/skema/index.vue`.
- `AdminInvoiceFormModal.vue` & `tagihan/batch.vue`: ganti 2 input teks bebas (Periode, Tahun Ajaran) jadi 1 dropdown pilih Billing Period (hanya yang status `open`).
- `tagihan/index.vue`, `tagihan/[id].vue` (admin & santri-facing): tampilkan nama billing period alih-alih `periode + tahun_ajaran` mentah.
- `shared/types/Keuangan.ts`: update tipe `Invoice`, request/query types.

**Verifikasi Fase 1**: migrasi up/down jalan bersih, `go build/vet/test`, manual: buat billing period baru → buka statusnya → buat tagihan terhadap period itu → cek muncul benar di list/detail/laporan.

---

## FASE 2 — Formalisasi Batch (`billing_batches` + `billing_batch_targets`, snapshot target)

Adaptasi dari skema referensi: setiap kali "Generate Massal" dijalankan, sistem mencatat **batch header** + **snapshot daftar santri yang jadi target beserta status per-santri** (bukan cuma pesan ringkasan sekali pakai). Ini juga jadi tempat wajar untuk akhirnya memakai `transactor` yang sudah di-inject tapi tidak pernah dipanggil (`create_invoice_batch.go`).

**Domain baru** — ikuti pola yang sama:

- `domain/billingbatch/entity/billing_batch.go`: `BillingBatch{ID, Name, BillingSchemeID, BillingPeriodID, Status(processing/completed/failed), CreatedBy, CreatedAt, CompletedAt, TotalCreated, TotalSkipped, TotalError}`.
- `domain/billingbatch/entity/billing_batch_target.go`: `BillingBatchTarget{ID, BatchID, SantriID, Status(pending/created/skipped_no_assignment/skipped_wrong_scheme/skipped_already_invoiced/skipped_component_missing/error), InvoiceID *string, Reason *string, ProcessedAt *time.Time}`.
- `domain/billingbatch/repository/interfaces.go`, `infrastructure/persistence/postgres_billing_batch_repo.go`.

**Migrasi** (`015_create_billing_batches.up/down.sql`):

- `CREATE TABLE billing_batches (...)`.
- `CREATE TABLE billing_batch_targets (..., UNIQUE(batch_id, santri_id))`.

**Rombak `CreateInvoiceBatchUseCase.Execute`** (`create_invoice_batch.go`) jadi 3 langkah eksplisit:

1. **Buat `billing_batches` row** (status `processing`) di awal.
2. **Snapshot target**: loop semua santri aktif, tentukan status awal tiap target (pakai kapabilitas resolve santri_id→user_id dari Fase 0.1 sekalian untuk validasi), tulis semua `billing_batch_targets` sekaligus (termasuk yang langsung berstatus `skipped_no_assignment`/`skipped_wrong_scheme` — tidak diproses lebih lanjut).
3. **Proses tiap target yang eligible** dalam transaksi per-target (pakai `uc.transactor.WithTx` yang sudah ada — akhirnya dipakai): cek duplikat → buat invoice → update baris target itu jadi `created` (simpan `invoice_id`) atau `skipped_already_invoiced`/`error` (simpan `reason`). Kegagalan satu target tidak menggagalkan target lain (masing-masing transaksi kecil sendiri), tapi sekarang **semuanya tercatat**, bukan cuma dihitung.
4. Update `billing_batches.status` jadi `completed` + totals di akhir.

**Use-case & endpoint baru**:

- `ListBillingBatchesUseCase`, `GetBillingBatchUseCase` (detail termasuk breakdown target).
- `GET /admin/billing-batches`, `GET /admin/billing-batches/:id`.
- `CreateInvoiceBatch` response berubah dari pesan string jadi `{batch_id, status}` — frontend redirect ke halaman detail batch untuk lihat hasil lengkap.

**Frontend (sipon-ui)**:

- `tagihan/batch.vue`: setelah generate, redirect/tampilkan link ke halaman detail batch (bukan lagi dump pesan/JSON).
- Halaman baru `pages/admin/keuangan/tagihan/batch/[id].vue`: tampilkan ringkasan (created/skipped/error) + tabel per-santri dengan status & alasan jelas — ini yang menyelesaikan keluhan "pesan hasil generate tidak merinci alasan skip" dari diskusi sebelumnya.
- Halaman baru `pages/admin/keuangan/tagihan/batch/index.vue` (opsional tapi disarankan): riwayat semua batch yang pernah dijalankan.

**Verifikasi Fase 2**: migrasi up/down jalan bersih, `go build/vet/test`, manual: jalankan generate massal dengan campuran santri (ada yang punya assignment cocok, ada yang tidak, ada yang sudah pernah ditagih) → cek halaman detail batch menunjukkan status yang benar & granular per santri; jalankan generate massal 2x berturut-turut dengan skema sama → target kedua semuanya `skipped_already_invoiced`, bukan diproses ulang.

---

## Catatan untuk fase selanjutnya (di luar rencana ini)

- **`invoice_items`** (konsolidasi satu-invoice-banyak-komponen): sengaja tidak masuk rencana ini. Kalau nanti mau dikerjakan, titik paling berisiko adalah desain alokasi pembayaran lintas komponen (`Invoice.AddPayment`/`RemainingAmount` di `domain/invoice/entity/invoice.go` saat ini murni skalar) — perlu keputusan desain terpisah sebelum implementasi.
- **`auto_posting.go`** (posting jurnal otomatis ke akuntansi) saat ini kode mati, tidak pernah dipanggil dari manapun — aman diubah/diwiring kapan saja tanpa risiko regresi, relevan untuk pembahasan modul akunting berikutnya.
- **`accounting_periods`** (tutup buku GL) tetap terpisah total dari `billing_periods` (periode tagihan) — keduanya punya tujuan berbeda, sengaja tidak digabung.
