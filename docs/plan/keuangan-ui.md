# Plan UI: Implementasi Module Keuangan di `sipon-ui`

## Context

Backend module `keuangan` sudah lengkap (Fase 1 steps 1-18) dengan dua bounded context:
1. **Billing** — komponen biaya, skema tagihan, invoice, pembayaran manual, potongan
2. **Accounting** — chart of accounts (COA), jurnal double-entry, periode pembukuan, laporan keuangan

Module keuangan memiliki dua konteks pengguna:
- **Admin (Bendahara)** — CRUD billing, accounting, reports
- **Santri/Wali** — hanya bisa lihat tagihan & riwayat pembayaran sendiri

Module ini masuk ke **portal admin** dengan layout sendiri (seperti module lainnya), dan juga menyediakan **self-service portal** untuk santri.

---

## Temuan penting dari kode aktual

1. **Dua konteks akses**: Admin (`/admin/keuangan/`) dengan permission gates, dan self-service (`/keuangan/`) untuk santri.
2. **Permission keys**: 6 permission baru (`manage_keuangan`, `verify_payment`, `view_keuangan_reports`, `manage_accounts`, `manage_journal`, `close_period`) — sudah terdaftar di backend dan otomatis muncul di `GET /api/v1/web/role-permission/permission-keys`.
3. **Status invoice**: 6 state (`draft`, `issued`, `partial`, `paid`, `expired`, `cancelled`) — UI harus menampilkan badge warna sesuai status.
4. **Status payment**: 3 state (`pending`, `verified`, `rejected`) — pending perlu verifikasi, rejected tidak bisa diubah.
5. **Period status**: 4 state (`open`, `closing`, `closed`, `locked`) — hanya `open` yang bisa posting jurnal.
6. **Account hierarchy**: COA adalah tree structure (parent-child) — UI perlu recursive tree view.
7. **Auto-posting**: Jurnal dari billing otomatis ter-posting saat invoice issued/payment verified — UI hanya tampilkan, tidak bisa edit/cancel jurnal auto.
8. **PDF kwitansi**: Endpoint `GET /admin/payments/:id/receipt` return PDF binary — UI trigger download.
9. **Batch invoice**: `POST /admin/invoices/batch` generate massal berdasarkan skema santri — butuh konfirmasi modal karena bisa generate ratusan invoice.
10. **Reports**: 6 jenis laporan dengan query params berbeda — butuh form filter + tabel/chart.

---

## Pola yang sudah ada di `sipon-ui` dan harus di-reuse

- **API layer**: `useApi()`, `parseApiError()`, envelope types.
- **Store pattern**: Pinia store per domain (`app/stores/*.ts`).
- **List/table pattern**: `UTable` + `UPagination` + `usePermission()` inline `v-if="can('...')"`.
- **Modal pattern**: `ConfirmActionModal.vue` untuk destructive actions.
- **Tree view**: Belum ada komponen tree — perlu buat untuk COA (atau pakai `UAccordion` per level).
- **Date picker**: Nuxt UI v4 sudah include — pakai untuk period date range, invoice due date.
- **Currency formatter**: Helper `formatRupiah(amount)` di `app/utils/format.ts`.
- **File download**: Helper untuk trigger blob download.

---

## Struktur baru

### Types — `shared/types/Keuangan.ts`

```typescript
// Enums
export type FeeComponentType = 'ukt' | 'spp' | 'daftar_ulang' | 'insidental'
export type PeriodType = 'monthly' | 'semesterly' | 'yearly' | 'once'
export type InvoiceStatus = 'draft' | 'issued' | 'partial' | 'paid' | 'expired' | 'cancelled'
export type PaymentStatus = 'pending' | 'verified' | 'rejected'
export type PaymentMethod = 'transfer' | 'cash' | 'check'
export type AdjustmentType = 'beasiswa' | 'diskon' | 'penyesuaian'
export type AccountType = 'asset' | 'liability' | 'equity' | 'revenue' | 'expense'
export type NormalBalance = 'debit' | 'credit'
export type JournalStatus = 'draft' | 'posted' | 'cancelled'
export type SourceType = 'invoice_issued' | 'payment_verified' | 'invoice_cancelled' | 'adjustment' | 'closing' | 'manual'
export type PeriodStatus = 'open' | 'closing' | 'closed' | 'locked'

// Entities
export interface FeeComponent {
  id: string
  code: string
  name: string
  type: FeeComponentType
  amount: number
  is_periodic: boolean
  period_type: PeriodType | null
  description: string | null
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface BillingSchemeItem {
  id: string
  fee_component_id: string
  fee_component?: FeeComponent
  amount_override: number | null
  is_required: boolean
  sort_order: number
}

export interface BillingScheme {
  id: string
  name: string
  description: string | null
  is_active: boolean
  items: BillingSchemeItem[]
  created_at: string
  updated_at: string
}

export interface SantriBillingAssignment {
  id: string
  santri_id: string
  billing_scheme_id: string
  effective_from: string
  effective_until: string | null
  assigned_by: string
  created_at: string
}

export interface Invoice {
  id: string
  invoice_number: string
  santri_id: string
  user_id: string
  billing_scheme_id: string | null
  fee_component_id: string
  fee_component?: FeeComponent
  periode: string
  tahun_ajaran: string
  amount: number
  discount_amount: number
  paid_amount: number
  status: InvoiceStatus
  due_date: string
  issued_at: string | null
  notes: string | null
  created_at: string
  updated_at: string
}

export interface Payment {
  id: string
  payment_number: string
  invoice_id: string
  invoice?: Invoice
  debit_account_id: string | null
  debit_account?: Account
  amount: number
  method: PaymentMethod
  reference_number: string | null
  payment_date: string
  status: PaymentStatus
  verified_by: string | null
  verified_at: string | null
  notes: string | null
  proof_key: string | null
  created_at: string
  updated_at: string
}

export interface InvoiceAdjustment {
  id: string
  invoice_id: string
  type: AdjustmentType
  amount: number
  percentage: number | null
  description: string | null
  applied_by: string
  applied_at: string
}

export interface Account {
  id: string
  code: string
  name: string
  type: AccountType
  parent_id: string | null
  level: number
  is_postable: boolean
  normal_balance: NormalBalance
  description: string | null
  is_active: boolean
  is_system: boolean
  children?: Account[]
  created_at: string
  updated_at: string
}

export interface JournalEntryLine {
  id: string
  account_id: string
  account?: Account
  account_code: string
  description: string | null
  debit: number
  credit: number
}

export interface JournalEntry {
  id: string
  journal_number: string
  entry_date: string
  description: string
  source_type: SourceType | null
  source_id: string | null
  period_id: string
  period?: AccountingPeriod
  total_debit: number
  total_credit: number
  posted_by: string
  posted_at: string | null
  status: JournalStatus
  lines: JournalEntryLine[]
  created_at: string
  updated_at: string
}

export interface AccountingPeriod {
  id: string
  name: string
  start_date: string
  end_date: string
  status: PeriodStatus
  closed_by: string | null
  closed_at: string | null
  created_at: string
  updated_at: string
}

// Request DTOs
export interface CreateFeeComponentRequest {
  code: string
  name: string
  type: FeeComponentType
  amount: number
  is_periodic: boolean
  period_type?: PeriodType
  description?: string
}

export interface CreateBillingSchemeRequest {
  name: string
  description?: string
}

export interface AddSchemeItemRequest {
  fee_component_id: string
  amount_override?: number
  is_required: boolean
  sort_order: number
}

export interface CreateInvoiceRequest {
  santri_id: string
  fee_component_id: string
  periode: string
  tahun_ajaran: string
  amount: number
  due_date: string
  notes?: string
}

export interface CreateInvoiceBatchRequest {
  billing_scheme_id: string
  periode: string
  tahun_ajaran: string
  due_date: string
}

export interface ApplyAdjustmentRequest {
  type: AdjustmentType
  amount?: number
  percentage?: number
  description?: string
}

export interface CreatePaymentRequest {
  invoice_id: string
  debit_account_id?: string
  amount: number
  method: PaymentMethod
  reference_number?: string
  payment_date: string
  notes?: string
  proof_key?: string
}

export interface CreateAccountRequest {
  code: string
  name: string
  type: AccountType
  parent_id?: string
  normal_balance: NormalBalance
  description?: string
  is_postable: boolean
}

export interface CreateJournalEntryRequest {
  entry_date: string
  description: string
  period_id: string
  lines: Array<{
    account_id: string
    description?: string
    debit: number
    credit: number
  }>
}

export interface CreatePeriodRequest {
  name: string
  start_date: string
  end_date: string
}

// Query params
export interface InvoiceListQuery {
  santri_id?: string
  status?: InvoiceStatus
  periode?: string
  tahun_ajaran?: string
  page?: number
  limit?: number
}

export interface PaymentListQuery {
  invoice_id?: string
  status?: PaymentStatus
  page?: number
  limit?: number
}

// Reports
export interface InvoiceSummary {
  tahun_ajaran: string
  periode: string
  total_amount: number
  total_paid: number
  total_outstanding: number
  invoice_count: number
  paid_count: number
}

export interface OutstandingBySantri {
  santri_id: string
  total_outstanding: number
  invoice_count: number
}

export interface LedgerLine {
  date: string
  journal_number: string
  description: string
  debit: number
  credit: number
  balance: number
}

export interface TrialBalanceLine {
  account_id: string
  account_code: string
  account_name: string
  account_type: AccountType
  debit: number
  credit: number
}

export interface BalanceSheetLine {
  account_id: string
  account_code: string
  account_name: string
  amount: number
}

export interface IncomeStatementLine {
  account_id: string
  account_code: string
  account_name: string
  amount: number
}

export interface BalanceSheetResponse {
  as_of_date: string
  assets: BalanceSheetLine[]
  total_assets: number
  liabilities: BalanceSheetLine[]
  total_liabilities: number
  equities: BalanceSheetLine[]
  total_equities: number
}

export interface IncomeStatementResponse {
  period_id: string
  period_name: string
  revenues: IncomeStatementLine[]
  total_revenue: number
  expenses: IncomeStatementLine[]
  total_expense: number
  net_income: number
}
```

### Stores

#### `app/stores/keuangan.ts` (billing operations)
- **State**: `feeComponents[]`, `billingSchemes[]`, `invoices[]`, `payments[]`, `meta`, `isLoading`, `error`
- **Actions**:
  - `fetchFeeComponents(query)`
  - `createFeeComponent(payload)`
  - `updateFeeComponent(id, payload)`
  - `deleteFeeComponent(id)`
  - `fetchBillingSchemes(query)`
  - `createBillingScheme(payload)`
  - `addSchemeItem(schemeId, payload)`
  - `removeSchemeItem(schemeId, itemId)`
  - `fetchInvoices(query)`
  - `createInvoice(payload)`
  - `createInvoiceBatch(payload)`
  - `getInvoice(id)`
  - `cancelInvoice(id)`
  - `applyAdjustment(invoiceId, payload)`
  - `fetchPayments(query)`
  - `createPayment(payload)`
  - `verifyPayment(id)`
  - `rejectPayment(id)`
  - `downloadReceipt(paymentId)`

#### `app/stores/keuanganAccounting.ts` (accounting operations)
- **State**: `accounts[]`, `journalEntries[]`, `periods[]`, `activePeriod`, `meta`, `isLoading`, `error`
- **Actions**:
  - `fetchAccounts(query)`
  - `createAccount(payload)`
  - `updateAccount(id, payload)`
  - `deleteAccount(id)`
  - `fetchJournalEntries(query)`
  - `createJournalEntry(payload)`
  - `getJournalEntry(id)`
  - `cancelJournalEntry(id)`
  - `fetchPeriods(query)`
  - `createPeriod(payload)`
  - `closePeriod(id)`
  - `reopenPeriod(id)`
  - `lockPeriod(id)`
  - `fetchActivePeriod()`

#### `app/stores/keuanganReports.ts` (reports)
- **State**: `invoiceSummary[]`, `outstandingBySantri[]`, `ledgerLines[]`, `trialBalanceLines[]`, `balanceSheet`, `incomeStatement`, `isLoading`, `error`
- **Actions**:
  - `fetchInvoiceSummary(tahun_ajaran?, periode?)`
  - `fetchOutstandingBySantri(tahun_ajaran?)`
  - `fetchLedger(account_id, period_id)`
  - `fetchTrialBalance(period_id)`
  - `fetchBalanceSheet(period_id?, as_of_date?)`
  - `fetchIncomeStatement(period_id)`

### Routing

**Admin layout** (`app/layouts/keuangan.vue`):
- Sidebar navigation dengan menu:
  - Dashboard
  - Master Data
    - Komponen Biaya
    - Skema Tagihan
    - Skema Santri
  - Transaksi
    - Tagihan
    - Pembayaran
  - Akuntansi
    - Chart of Accounts
    - Jurnal
    - Periode
  - Laporan
    - Rekap Tagihan
    - Tunggakan
    - Buku Besar
    - Neraca Saldo
    - Neraca
    - Laba Rugi

**Admin pages** (`app/pages/admin/keuangan/`):

```
admin/keuangan/
├── index.vue                          # Dashboard dengan summary cards
├── komponen/
│   └── index.vue                      # List + CRUD fee components
├── skema/
│   ├── index.vue                      # List billing schemes
│   └── [id].vue                       # Detail scheme + manage items
├── santri/
│   └── index.vue                      # Assign schemes to santri
├── tagihan/
│   ├── index.vue                      # List invoices + filter
│   ├── [id].vue                       # Invoice detail + adjustments
│   └── batch.vue                      # Batch generate form
├── pembayaran/
│   ├── index.vue                      # List payments + filter
│   ├── [id].vue                       # Payment detail + verify/reject
│   └── manual.vue                     # Create manual payment form
├── akun/
│   └── index.vue                      # Tree view COA + CRUD
├── jurnal/
│   ├── index.vue                      # List journal entries
│   ├── [id].vue                       # Journal detail + cancel (manual only)
│   └── manual.vue                     # Create manual journal
├── periode/
│   └── index.vue                      # List periods + close/reopen/lock
└── laporan/
    ├── rekap.vue                      # Invoice summary report
    ├── tunggakan.vue                  # Outstanding by santri
    ├── buku-besar.vue                 # Ledger
    ├── neraca-saldo.vue               # Trial balance
    ├── neraca.vue                     # Balance sheet
    └── laba-rugi.vue                  # Income statement
```

**Self-service pages** (`app/pages/keuangan/`):

```
keuangan/
├── index.vue                          # Redirect to tagihan
├── tagihan/
│   ├── index.vue                      # My invoices
│   └── [id].vue                       # Invoice detail
└── riwayat.vue                        # My payments
```

### Komponen

#### Shared (`app/components/keuangan/*.vue`):
- **`KeuanganStatusBadge.vue`** — badge untuk invoice/payment/period/journal status
- **`KeuanganAmountDisplay.vue`** — formatted rupiah amount dengan warna (debit=red, credit=green)
- **`KeuanganAccountPicker.vue`** — dropdown/tree picker untuk account (COA)
- **`KeuanganJournalLineEditor.vue`** — editor untuk journal lines (dynamic rows)
- **`KeuanganPeriodSelector.vue`** — dropdown period selector

#### Admin (`app/components/admin/keuangan/*.vue`):
- **`AdminFeeComponentFormModal.vue`** — form create/edit fee component
- **`AdminBillingSchemeFormModal.vue`** — form create/edit billing scheme
- **`AdminSchemeItemsManager.vue`** — manage items in a scheme (add/remove components)
- **`AdminInvoiceFormModal.vue`** — form create single invoice
- **`AdminBatchInvoiceForm.vue`** — form batch generate invoices
- **`AdminAdjustmentFormModal.vue`** — form apply discount/scholarship
- **`AdminPaymentFormModal.vue`** — form create manual payment
- **`AdminPaymentVerificationModal.vue`** — modal verify/reject payment
- **`AdminAccountFormModal.vue`** — form create/edit account
- **`AdminAccountTreeView.vue`** — recursive tree view for COA
- **`AdminJournalFormModal.vue`** — form create manual journal
- **`AdminPeriodFormModal.vue`** — form create period
- **`AdminPeriodActionsModal.vue`** — modal close/reopen/lock period
- **`AdminReportFilter.vue`** — generic filter form for reports
- **`AdminReportTable.vue`** — generic table for report display

#### Self-service (`app/components/keuangan/*.vue`):
- **`KeuanganInvoiceCard.vue`** — card display for invoice (santri view)
- **`KeuanganPaymentHistory.vue`** — timeline display for payments

---

## Matriks visibility per status

### Invoice Status

| Status      | Actions tersedia                                    |
| ----------- | --------------------------------------------------- |
| `draft`     | Issue (not in UI yet), Cancel, Edit                 |
| `issued`    | Apply adjustment, Record payment, Cancel            |
| `partial`   | Apply adjustment, Record payment                    |
| `paid`      | -                                                   |
| `expired`   | -                                                   |
| `cancelled` | -                                                   |

### Payment Status

| Status      | Actions tersedia                                    |
| ----------- | --------------------------------------------------- |
| `pending`   | Verify, Reject                                      |
| `verified`  | Download receipt                                    |
| `rejected`  | -                                                   |

### Period Status

| Status      | Actions tersedia                                    |
| ----------- | --------------------------------------------------- |
| `open`      | Post journal, Close                                 |
| `closing`   | -                                                   |
| `closed`    | Reopen, Lock                                        |
| `locked`    | -                                                   |

### Journal Status

| Status      | Actions tersedia                                    |
| ----------- | --------------------------------------------------- |
| `draft`     | -                                                   |
| `posted`    | Cancel (only if manual)                             |
| `cancelled` | -                                                   |

---

## Fase Pengerjaan

**Fase 1 — Types.** `shared/types/Keuangan.ts`. Checkpoint: type-check lolos.

**Fase 2 — Stores.** `app/stores/keuangan.ts`, `keuanganAccounting.ts`, `keuanganReports.ts`. Checkpoint: compile lolos.

**Fase 3 — Layout + Navigation.** `app/layouts/keuangan.vue` dengan sidebar navigation. Checkpoint: layout render, menu navigasi bekerja.

**Fase 4 — Shared components.** `KeuanganStatusBadge.vue`, `KeuanganAmountDisplay.vue`, `KeuanganAccountPicker.vue`, dll. Checkpoint: render bersih.

**Fase 5 — Admin pages: Master Data.** `komponen/index.vue`, `skema/index.vue`, `skema/[id].vue`, `santri/index.vue` + modal components. Checkpoint: CRUD fee components, billing schemes, assign schemes.

**Fase 6 — Admin pages: Transaksi.** `tagihan/index.vue`, `tagihan/[id].vue`, `tagihan/batch.vue`, `pembayaran/index.vue`, `pembayaran/[id].vue`, `pembayaran/manual.vue` + modal components. Checkpoint: CRUD invoices, batch generate, apply adjustments, create payments, verify/reject, download receipt.

**Fase 7 — Admin pages: Akuntansi.** `akun/index.vue` (tree view COA), `jurnal/index.vue`, `jurnal/[id].vue`, `jurnal/manual.vue`, `periode/index.vue` + modal components. Checkpoint: CRUD accounts (hierarchical), manual journals, period lifecycle.

**Fase 8 — Admin pages: Laporan.** Semua 6 halaman laporan + filter/table components. Checkpoint: fetch & display reports dengan filter.

**Fase 9 — Self-service pages.** `keuangan/tagihan/index.vue`, `keuangan/tagihan/[id].vue`, `keuangan/riwayat.vue` + components. Checkpoint: santri bisa lihat tagihan & riwayat pembayaran sendiri.

**Fase 10 — Polish & Integration.** Dashboard summary, permission gating, edge cases, empty states. Checkpoint: end-to-end flow dari setup komponen → generate invoice → terima pembayaran → lihat laporan.

---

## Verifikasi

1. `npm run dev` jalan tanpa error di tiap checkpoint fase.
2. Uji manual browser: seluruh alur admin (Fase 5-8 checkpoint) end-to-end lawan `sipon-be` yang jalan lokal.
3. Uji self-service (Fase 9 checkpoint): santri login → lihat tagihan & pembayaran.
4. Uji permission gating: user tanpa `manage_keuangan`/`verify_payment`/etc tidak melihat menu/action terkait.
5. Uji status transitions: invoice draft → issued → partial → paid, payment pending → verified/rejected, period open → closed → locked.
6. Uji laporan: fetch summary, outstanding, ledger, trial balance, balance sheet, income statement dengan berbagai filter.
7. Uji tree view COA: expand/collapse, CRUD accounts di berbagai level.
8. Uji batch invoice: generate massal untuk skema tertentu, cek hasilnya.

---

## Catatan tambahan

- **Layout keuangan** terpisah dari layout admin umum karena navigasi sangat spesifik (banyak sub-menu).
- **Tree view COA** perlu komponen custom (belum ada di Nuxt UI v4) — pakai recursive component atau `UAccordion` per level.
- **Currency input** perlu custom component untuk format rupiah saat input (bisa pakai `@vueuse/core` `useCurrencyInput` atau buat sendiri).
- **Date picker** untuk period date range, invoice due date, payment date — pakai native `UInput type="date"` dari Nuxt UI.
- **PDF download** untuk kwitansi — pakai `useFetch` dengan `responseType: 'blob'` lalu trigger download via `URL.createObjectURL`.
- **Batch invoice** butuh progress indicator karena bisa lama (max 500 invoice per request).
- **Report charts** — opsional untuk Fase 2, Fokus Fase 1 pada tabel dulu.
- **Auto-posting indicator** — di halaman invoice/payment, tampilkan badge "Auto-jurnal" untuk menunjukkan jurnal sudah ter-posting otomatis.
