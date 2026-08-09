# Rules: Akuntansi Modul Keuangan

Dokumen ini adalah rujukan aturan bisnis akuntansi untuk modul `keuangan` — lingkup organisasi pengurus pondok pesantren/sekolah (bukan perusahaan komersial): sederhana, cukup untuk transparansi ke wali santri & yayasan, tidak perlu selengkap PSAK entitas komersial.

Status tiap aturan ditandai:
- ✅ **Sudah berjalan** — sudah diimplementasi dan dipakai di alur nyata.
- ⚠️ **Ada kodenya tapi belum terpasang** — logic sudah ditulis tapi tidak pernah dipanggil (dead code).
- ❌ **Belum ada** — perlu dibangun.

Lihat juga: [`docs/schemas/keuangan-akuntansi.md`](../schemas/keuangan-akuntansi.md) (struktur tabel), [`docs/specs/keuangan-akuntansi-api.md`](../specs/keuangan-akuntansi-api.md) (endpoint), [`docs/bugs/`](../bugs/) (bug konkret yang ditemukan saat menyusun dokumen ini).

---

## 1. Prinsip Dasar

1. **Double-entry** — setiap transaksi keuangan dicatat sebagai jurnal dengan minimal 2 baris (line), total debit harus sama dengan total kredit. ✅ Ditegakkan di domain (`JournalEntry.Validate()`) maupun DB (`CHECK (total_debit = total_credit)`, `CHECK (debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0)` per line).
2. **Accrual basis** — pendapatan diakui saat tagihan (invoice) diterbitkan (`issued`), bukan saat uang diterima. Piutang santri dicatat di titik itu. Saat kas benar-benar diterima (payment verified), yang terjadi hanya pelunasan piutang, bukan pengakuan pendapatan baru.
3. **Semua jurnal wajib menempel ke satu Periode Akuntansi (`accounting_periods`)** — tidak ada jurnal "lepas" tanpa periode. Periode inilah yang nanti ditutup (closing) secara berkala (biasanya bulanan/semesteran, ditentukan pengurus).
4. **Hanya akun *postable* yang boleh menerima jurnal** — akun header/golongan (mis. `1000 ASET`, `1100 Aset Lancar`) hanya untuk pengelompokan tampilan laporan, tidak pernah menerima baris jurnal langsung.
5. **Tiga sumber jurnal, satu tabel** — jurnal manual dan jurnal otomatis (dari billing) hidup di tabel yang sama (`journal_entries`), dibedakan lewat `source_type`/`source_id`. Laporan (buku besar, neraca saldo, dst) tidak peduli asalnya — semua diperlakukan sama selama `status = 'posted'`.
6. **Non-tujuan (sengaja tidak dibuat, sesuai skala pesantren/sekolah)**: tidak ada multi-currency, tidak ada pajak otomatis, tidak ada depresiasi otomatis aset tetap, tidak ada refund, tidak ada denda keterlambatan otomatis. Kalau nanti dibutuhkan, itu keputusan terpisah — jangan diam-diam ditambahkan saat mengerjakan poin di dokumen ini.

---

## 2. Tiga Sumber Pencatatan Jurnal

### 2.1 Invoice diterbitkan (issued) → jurnal otomatis pengakuan piutang & pendapatan

⚠️ **Ada kodenya (`AutoPostingService.PostInvoiceIssued`), tidak pernah dipanggil.** Lihat [bug: auto-posting tidak terpasang](../bugs/akuntansi-auto-posting-tidak-terpasang.md).

```
Dr. Piutang Santri (1103)                 = amount - discount_amount
    Cr. Pendapatan [sesuai tipe komponen] = amount - discount_amount
```

Aturan:
- Dipicu saat invoice pindah status `draft → issued` (bukan saat dibuat sebagai draft — kalau ke depan ada alur draft, jangan posting sebelum `Issue()` dipanggil). Saat ini `create_invoice.go`/`create_invoice_batch.go` langsung membuat invoice berstatus `issued` (tidak lewat `draft` dulu secara terpisah), jadi titik posting = titik pembuatan invoice.
- **Tanggal terbit (`issued_date`) kini input eksplisit** — bendahara mengisi tanggal terbit invoice (`POST /admin/invoices` & `/batch`), dan `Invoice.Issue(issuedDate)` memakainya sebagai `IssuedAt` (bukan `time.Now()` implisit). Titik penentuan periode akuntansi = `issued_date` ini; auto-posting menolak kalau periode akuntansi untuk tanggal itu sudah ditutup. `issued_date` harus dalam rentang periode tagihan yang dipilih (kalau ada). Periode tagihan **opsional untuk komponen non-periodik** (`is_periodic = false`, mis. insidental) — invoice non-periodik bisa dibuat tanpa `billing_period_id`; untuk komponen periodik, `billing_period_id` wajib.
- Mapping `fee_component.type → account.code` (SPP→4100, UKT→4200, Daftar Ulang→4300, Insidental→4400) sudah cukup untuk kebutuhan sekarang; ini seharusnya bisa disunting dari UI COA nanti (Fase 2 pada `docs/plan/keuangan-module.md`), tidak hardcode selamanya.
- **Wajib idempoten**: satu invoice hanya boleh menghasilkan satu jurnal `invoice_issued`. Sebelum posting, cek dulu apakah `journal_entries` sudah punya baris dengan `(source_type='invoice_issued', source_id=<invoice_id>)` — kalau sudah ada, jangan posting ulang (mis. request di-retry, atau batch generate dijalankan dua kali). Idealnya ditegakkan juga di level DB lewat unique index (lihat schema doc).
- Diskon/adjustment yang diterapkan **setelah** invoice issued (`ApplyAdjustment`) idealnya juga menghasilkan jurnal penyesuaian sendiri (`source_type='adjustment'`) agar piutang & pendapatan tetap sinkron dengan jurnal — saat ini `apply_adjustment.go` hanya mengubah `discount_amount` di tabel `invoices`, tidak menyentuh jurnal sama sekali. Untuk skala pesantren, cukup: jurnal balik sebagian dengan pola yang sama seperti pembatalan (lihat 2.3) tapi sebesar nilai adjustment saja.
- ❌ **Belum ada**: `issued_date` (tanggal terbit invoice, yang jadi `entry_date` jurnal) harus jadi input eksplisit dari bendahara di semua cara membuat invoice — bukan `time.Now()` implisit seperti sekarang. Kalau komponen biayanya periodik, `issued_date` wajib berada dalam rentang `billing_period` yang dipilih; kalau non-periodik (`fee_component.is_periodic = false`, mis. insidental), `billing_period_id` boleh dikosongkan. Rencana lengkap & alasan di [`docs/plan/invoice-issued-date-dan-periode-opsional.md`](../plan/invoice-issued-date-dan-periode-opsional.md).

### 2.2 Payment diverifikasi (verified) → jurnal otomatis penerimaan kas

⚠️ **Ada kodenya (`AutoPostingService.PostPaymentVerified`), tidak pernah dipanggil.**

```
Dr. {akun kas/bank pilihan bendahara}   = amount
    Cr. Piutang Santri (1103)          = amount
```

Aturan:
- **Titik posting = saat status payment berubah jadi `verified`, bukan saat payment dibuat (`pending`).** Alasan: payment yang masih `pending` belum tentu benar (bisa `rejected`), dan menunda posting sampai verifikasi menghindari kebutuhan jurnal pembalik untuk payment yang ditolak — lebih sederhana sesuai prinsip di dokumen ini. Ini konsisten dengan field `verified_by`/`verified_at` yang sudah ada di tabel `payments`.
- `debit_account_id` (akun kas/bank tujuan) **wajib diisi** sebelum payment bisa diverifikasi — tanpa ini `AutoPostingService.PostPaymentVerified` tidak tahu akun mana yang di-debit. Saat ini kolom itu nullable (`payments.debit_account_id UUID` tanpa `NOT NULL`, DTO `CreateManualPaymentRequest.DebitAccountID *string` opsional). Begitu auto-posting dipasang, ini harus jadi wajib — baik di validasi aplikasi maupun (disarankan) constraint DB. Lihat schema doc.
- Akun yang dipilih harus `is_postable = true` dan `is_active = true` (pola sama seperti validasi di jurnal manual, `Account.EnsurePostable()`).
- Idempoten sama seperti invoice: satu payment → maksimal satu jurnal `payment_verified`.
- Payment `rejected` **tidak menghasilkan jurnal apa pun** (tidak pernah sempat diakui sebagai kas masuk).

### 2.3 Invoice dibatalkan (cancelled) → jurnal pembalik (reversal)

⚠️ **Ada kodenya (`AutoPostingService.PostInvoiceCancelled`), tidak pernah dipanggil.**

```
Dr. Pendapatan [sesuai tipe komponen] = amount yang pernah diakui saat issued
    Cr. Piutang Santri (1103)        = amount yang sama
```

Aturan:
- Hanya diperlukan **jika invoice tersebut sudah pernah menghasilkan jurnal `invoice_issued`** (yaitu sudah pernah issued). Kalau invoice batal sebelum sempat issued (mis. dibuat lalu langsung dibatalkan tanpa pernah ada jurnal), tidak perlu jurnal pembalik apa pun.
- **Invoice yang sudah punya pembayaran terverifikasi (`paid_amount > 0`) tidak boleh dibatalkan begitu saja** — uang sudah masuk kas dan sudah ada jurnal `payment_verified` yang menunjuk ke invoice itu. `Invoice.Cancel()` saat ini sudah menolak status `paid`, **tapi tidak menolak status `partial`** (yang paid_amount-nya > 0 juga). Ini harus diperbaiki: `Cancel()` idealnya menolak setiap invoice dengan `paid_amount > 0`, dan mengarahkan bendahara ke alur lain (mis. adjustment/koreksi manual) kalau memang perlu dibatalkan. Ini gap desain, dicatat di [bug: cancel invoice tidak menolak partial](../bugs/akuntansi-cancel-invoice-partial-payment.md).
- Pembatalan **tidak menghapus** jurnal `invoice_issued` yang lama — cukup tambah jurnal pembalik baru, demi jejak audit (tidak pernah delete/update jurnal yang sudah posted).

### 2.4 Jurnal manual

✅ Sudah ada (`CreateManualJournalUseCase`), tapi punya bug penomoran — lihat [bug: nomor jurnal manual selalu 1](../bugs/akuntansi-manual-journal-nomor-selalu-1.md) — dan bug baris jurnal tidak tersimpan — lihat [bug: journal lines tidak tersimpan](../bugs/akuntansi-journal-lines-tidak-tersimpan.md).

Aturan:
- Dipakai bendahara untuk transaksi yang tidak berasal dari billing: setor modal awal, beli inventaris, bayar gaji, biaya operasional, dsb.
- Minimal 2 baris, total debit = total kredit (sudah ditegakkan).
- Hanya ke akun `is_postable` & `is_active` (sudah ditegakkan lewat `EnsurePostable()`).
- Hanya bisa posting ke periode berstatus `open` (sudah ditegakkan lewat `period.CanPost()`).
- `source_type = 'manual'` — `source_id` sebaiknya diisi `NULL`, bukan menunjuk ke ID jurnal itu sendiri seperti sekarang (`entry.SetSource(constant.SourceManual, entry.ID)` di `create_manual_journal.go:57`). Self-reference tidak salah secara teknis, tapi membingungkan kalau nanti ada laporan/anti-duplikasi yang mengandalkan `(source_type, source_id)` sebagai penunjuk ke dokumen sumber asli.
- Jurnal manual **boleh dibatalkan langsung** oleh bendahara (`CancelJournalUseCase`) — beda dengan jurnal otomatis yang harus dibatalkan lewat pembatalan dokumen sumbernya (invoice/payment), bukan lewat endpoint jurnal. Aturan ini sudah ditegakkan di domain (`JournalEntry.Cancel()` menolak kalau `SourceType != manual`).
- **Pembatalan jurnal (manual maupun via reversal) tidak boleh menembus periode yang sudah `closed`/`locked`** — saat ini `CancelJournalUseCase` tidak mengecek status periode sama sekali. Lihat [bug: cancel journal tidak cek status periode](../bugs/akuntansi-cancel-journal-tidak-cek-periode.md).

---

## 3. Periode Akuntansi & Proses Closing

### 3.1 Status & siapa yang boleh apa

| Status | Boleh posting jurnal baru? | Boleh dibuka lagi (reopen)? | Catatan |
|---|---|---|---|
| `open` | Ya | — | Status normal berjalan |
| `closed` | Tidak | Ya (kembali ke `open`) | Sudah ditutup, tapi masih bisa dikoreksi kalau ternyata ada salah catat |
| `locked` | Tidak | **Tidak** | Permanen — dipakai setelah laporan periode itu sudah difinalisasi/diaudit/dipakai untuk laporan ke yayasan |

**Simplifikasi dari desain awal**: rencana lama (`docs/plan/keuangan-module.md`) punya status antara `closing` (sedang proses tutup buku) di antara `open` dan `closed`, dengan method `AccountingPeriod.StartClosing()`. Status ini ❌ **tidak pernah dipakai** — tidak ada use-case yang memanggilnya, `ClosePeriodUseCase` langsung lompat `open → closed`. Karena closing di skala pesantren cukup dilakukan sebagai satu operasi atomik (generate jurnal penutup + ubah status, dalam satu transaksi DB), status `closing` tidak menambah nilai dan sebaiknya **dihapus** (sederhanakan jadi 3 status: `open`, `closed`, `locked`). Lihat [bug: status closing dead code](../bugs/akuntansi-closing-period-tidak-generate-jurnal.md) untuk detail dan opsi lain.

### 3.2 Proses Closing (❌ belum ada — baru rencana di `auto_posting.go`/dokumen lama, belum diimplementasikan sama sekali)

Tujuan closing: memindahkan saldo akun nominal (Pendapatan & Beban) ke ekuitas, supaya periode berikutnya mulai dari saldo nol untuk akun-akun itu (akun riil — Aset/Kewajiban/Ekuitas — saldonya jalan terus/carry-forward).

Langkah (satu transaksi DB, dipicu satu aksi "Tutup Periode" — `ClosePeriodUseCase.Execute`):

1. **Validasi pra-syarat**:
   - Periode berstatus `open` (`period.CanPost()`/`IsOpen()`).
   - Akun `3200 Saldo Laba` dan `3201 Laba Tahun Berjalan` harus ada & aktif (`accountRepo.FindByCode(ctx, "3200")`, `FindByCode(ctx, "3201")`) — kalau tidak ada, tolak dengan kode error baru (mis. `CodePeriodClosingAccountMissing`), **wajib** dibungkus `application.WrapConflictErr` (lihat `docs/bugs/keuangan-kernel-error-tanpa-wrap-selalu-500.md` — jangan sampai closing yang gagal karena COA belum lengkap malah tampil sebagai 500).
2. **Hitung saldo tiap akun Pendapatan (4xxx) & Beban (5xxx)** untuk periode ini — reuse query yang sama persis dengan `ReportIncomeStatementUseCase` (`application/query/report_income_statement.go:36-45`, setelah kolom `deleted_at` dihapus dari query itu — lihat bug terkait):
   ```sql
   SELECT jel.account_id, COALESCE(SUM(jel.debit),0) AS total_debit, COALESCE(SUM(jel.credit),0) AS total_credit
   FROM journal_entry_lines jel
   JOIN journal_entries je ON je.id = jel.journal_entry_id
   WHERE je.period_id = $1 AND je.status = 'posted'
   GROUP BY jel.account_id
   ```
   Untuk tiap akun `type='revenue'`: `balance = total_credit - total_debit`. Untuk tiap akun `type='expense'`: `balance = total_debit - total_credit`. **Lewati akun dengan `balance <= 0`** (tidak ada aktivitas — baris jurnal dengan nominal 0 melanggar `CHECK (debit > 0 AND credit = 0) OR (debit = 0 AND credit > 0)`).
3. **Susun baris jurnal penutup** (`source_type='closing'`, `source_id=<period_id>`), satu jurnal untuk semuanya:
   - Untuk tiap akun revenue dengan `balance > 0`: baris `Debit = balance` ke akun itu (menutup saldo kreditnya ke nol). Jumlahkan semua ini sebagai `totalRevenue`.
   - Satu baris `Credit = totalRevenue` ke `3201 Laba Tahun Berjalan` (kalau `totalRevenue > 0`).
   - Untuk tiap akun expense dengan `balance > 0`: baris `Credit = balance` ke akun itu (menutup saldo debitnya ke nol). Jumlahkan semua ini sebagai `totalExpense`.
   - Satu baris `Debit = totalExpense` ke `3201 Laba Tahun Berjalan` (kalau `totalExpense > 0`).
   - Hitung `diff = totalRevenue - totalExpense`. Kalau `diff > 0` (laba): baris `Debit = diff` ke `3201`, baris `Credit = diff` ke `3200 Saldo Laba`. Kalau `diff < 0` (rugi): baris `Debit = -diff` ke `3200`, baris `Credit = -diff` ke `3201`. Kalau `diff == 0`: **jangan** tambah baris ini (nominal nol tidak valid).
   - Total debit di jurnal ini akan selalu sama dengan total kredit secara matematis (lihat pembuktian di `docs/bugs/akuntansi-closing-period-tidak-generate-jurnal.md`) — tetap panggil `JournalEntry.Validate()`/`Post()` seperti biasa sebagai jaring pengaman, jangan di-skip.
   - **Kalau `totalRevenue == 0 && totalExpense == 0`** (tidak ada aktivitas revenue/expense sama sekali di periode ini — kurang dari 2 baris): **lewati pembuatan jurnal closing sama sekali** (langsung ke langkah 4), jangan paksa `Save` jurnal dengan < 2 baris (akan gagal di `Validate()`).
4. **Ubah status periode jadi `closed`**, isi `closed_by`/`closed_at`, dalam transaksi yang sama dengan langkah 3 (jangan sampai jurnal penutup kebuat tapi status gagal ke-update, atau sebaliknya).
5. Kalau closing ternyata salah (mis. ada koreksi yang kelewat), **reopen** (`ReopenPeriodUseCase`): kalau periode itu punya jurnal `closing` (`journalRepo.FindBySource(ctx, "closing", periodID)`), batalkan (`entry.Cancel()` → `status='cancelled'`, bukan dihapus) sebagai bagian dari transaksi reopen yang sama, lalu kembalikan status periode ke `open`. Setelah dikoreksi, closing bisa dijalankan ulang dan akan membuat jurnal `closing` baru untuk `source_id` (period_id) yang sama — ini **aman** terhadap unique partial index anti-dobel-posting yang diusulkan di `docs/schemas/keuangan-akuntansi.md` (`... WHERE ... AND status != 'cancelled'`) karena jurnal `closing` lama sudah `status='cancelled'` (dikecualikan dari constraint) saat jurnal baru dibuat — tidak perlu pengecualian khusus untuk `source_type='closing'`.
6. **`locked` tidak bisa reopen** — kalau sudah dikunci, tutup buku periode itu final selamanya. Perubahan salah harus dikoreksi lewat jurnal di periode berjalan (periode baru), bukan mengubah masa lalu — ini prinsip akuntansi standar (tidak mengubah catatan historis yang sudah difinalisasi).

### 3.3 Setiap dokumen (invoice/payment/jurnal manual) harus jatuh di periode yang benar

- Titik penentuan periode: `entry_date` jurnal harus berada di antara `start_date` dan `end_date` periode yang dipilih. Saat ini `AutoPostingService` malah mengambil periode lewat `FindActive()` (periode manapun yang sedang `status='open'`) tanpa mencocokkan ke tanggal transaksi — kalau suatu saat ada lebih dari satu periode `open` bersamaan (mis. periode bulan lalu belum ditutup, sudah dibuat periode bulan ini), atau transaksi di-backdate, jurnal bisa "salah kamar" (entry_date di bulan A tapi period_id nunjuk ke bulan B). Perbaikannya: cari periode lewat `start_date <= entry_date <= end_date`, bukan `FindActive()`. Kalau tidak ketemu periode yang cocok (belum dibuat, atau sudah closed), tolak transaksi dengan pesan jelas ("Periode akuntansi untuk tanggal ini belum dibuat/sudah ditutup"), jangan biarkan asal masuk ke periode aktif yang salah.

---

## 4. Aturan Lain yang Perlu Ada (COA & integritas umum)

1. **Kode akun tidak boleh berubah setelah dipakai jurnal** — `code` dipakai sebagai referensi denormalized di `journal_entry_lines.account_code`. Kalau `code` akun diubah setelah ada histori jurnal, laporan lama jadi tidak konsisten dengan COA sekarang. Aturan: `UpdateAccount` tidak boleh mengubah `code`, hanya `name`/`description`/`is_postable` (perlu dicek apakah `update_account.go` sekarang sudah membatasi ini).
2. **Akun yang sudah pernah dipakai jurnal tidak boleh dihapus/dinonaktifkan begitu saja** — sudah ada kode error `CodeAccountHasJournal`, pastikan use-case delete/deactivate akun benar-benar mengecek ke `journal_entry_lines` sebelum mengizinkan.
3. **Root/header account (`is_system=true`) tidak bisa dihapus** — sudah ditegakkan (`Account.SoftDelete()`/`Deactivate()` menolak kalau `IsSystem`).
4. **Penomoran dokumen (invoice/payment/jurnal) harus atomic, tidak boleh tabrakan** — sudah dibereskan untuk invoice & payment lewat tabel bersama `finance_number_sequences` (lihat `docs/plan/perubahan-skema-tagihan.md`). Jurnal (`JRN/PAY/INV/CNL/ADJ/CLS`, prefix beda-beda per `source_type` di `generateJournalNumber`) harus pakai mekanisme yang sama — tambah `doc_type` baru di tabel yang sama per prefix, atau satu `doc_type='journal'` kalau prefix-nya tidak perlu ikut disegmentasi ke sequence yang beda (lebih sederhana: satu `doc_type='journal'`, prefix cuma kosmetik string, tidak memengaruhi urutan angka).
5. **Tidak ada UPDATE/DELETE pada jurnal yang sudah `posted`** — koreksi selalu lewat jurnal baru (manual atau reversal), tidak pernah mengubah baris/nominal jurnal lama. `PostgresJournalRepository.Update` saat ini hanya dipakai untuk transisi status (`posted → cancelled`), bukan mengubah nominal — pertahankan pola ini, jangan tambah endpoint "edit jurnal".
6. **Semua endpoint yang menghasilkan efek akuntansi wajib satu transaksi DB** — pembuatan invoice/pelunasan payment + jurnal otomatisnya harus sukses/gagal bersama (`transactor.WithTx`), tidak boleh ada state "invoice issued tapi jurnalnya gagal kebuat" atau sebaliknya.
7. **Laporan (ledger, neraca saldo, neraca, laba rugi) hanya menghitung jurnal `status='posted'`** — jurnal `draft`/`cancelled` tidak pernah masuk hitungan.
8. **Audit trail** — `finance_audit_logs` sudah ada di skema tapi ❌ tidak pernah diisi kode manapun (dead table, sama seperti disebut di `docs/plan/perubahan-skema-tagihan.md`). Kalau mau dipakai: minimal catat setiap posting jurnal (siapa, kapan, sumber apa) dan setiap perubahan status periode (close/reopen/lock) — dua hal paling sensitif secara akuntansi. Kalau tidak akan dipakai dalam waktu dekat, sebaiknya didokumentasikan sebagai keputusan sadar "belum dipakai", bukan dibiarkan sebagai tabel misterius.
