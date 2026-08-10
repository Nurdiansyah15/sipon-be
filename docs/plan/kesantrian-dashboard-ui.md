# Plan: Dashboard Utama = Halaman Santri (Summary Personal) di `sipon-ui` + endpoint summary di `sipon-be`

## Context

Halaman **dashboard utama (`/dashboard`)** menjadi halaman milik santri (user biasa) yang menampilkan ringkasan personal dari beberapa module. Grid card aplikasi yang sekarang ada di dashboard **dipindah ke menu "Aplikasi"** di layout utama (navbar), menggantikan link placeholder `#`. Urutan informasi di dashboard:

1. **Artikel terbaru** — module `article`
2. **User summary status kesantrian + informasi akun** — module `kesantrian`
3. **User summary tagihan keuangan** — module `keuangan`

Semua endpoint auth-only (tanpa permission key) karena diakses user biasa — sudah diverifikasi: `GET /articles`, `GET /santri/profile`, `GET /keuangan/invoices`, `GET /keuangan/summary` semuanya hanya `jwtAuth + principalLoad`, tanpa `RequirePermission`.

## Temuan penting dari kode aktual

1. **`GET /api/v1/web/articles` auth-only**, tapi `page` & `limit` wajib (binding `min=1` → 422 kalau kosong) dan **tanpa default filter status** → harus kirim `status=published&sort_by=published_at&sort_type=DESC`. Item list tidak punya `summary`/`slug`.
2. **`GET /api/v1/web/santri/profile` auth-only**, gabung profil santri + akun (`username`, `email`, `fullname`, `avatar_url`). **`status` kesantrian (`SANTRI`/`ALUMNI`/`DROP_OUT`) telah ditambahkan** ke `GetSantriResponse` (DTO `santri_dto.go` + mapper `get_santri.go`).
3. **`GET /api/v1/web/keuangan/summary` telah ditambahkan** (auth-only, tanpa permission). Response: `total_tagihan`, `total_terbayar`, `total_tunggakan`, `jumlah_invoice`, `jumlah_lunas`, `jumlah_belum`.
4. Invoice kini pakai `billing_period_id` + nested `billing_period {id,name,status}` (dok `keuangan-ui.md` usang soal `periode`/`tahun_ajaran`).
5. `AppNavbar.vue` & `AppMobileBottomNav.vue` link **"Aplikasi"** saat ini `to: '#'` → diubah ke `/aplikasi`.
6. Enums: invoice `draft|issued|partial|paid|expired|cancelled`; kesantrian `SANTRI|ALUMNI|DROP_OUT`; `RemainingAmount()` = `amount - discount - paid` (floor 0).

## Backend — `sipon-be` (selesai)

### A. Kesantrian: `status` di response profile
- `internal/modules/kesantrian/application/dto/santri_dto.go`: `Status *string \`json:"status,omitempty"\`` di `GetSantriResponse`.
- `internal/modules/kesantrian/application/query/get_santri.go` (`mapSantriToResponse`): isi dari `s.Status`.

### B. Keuangan: `GET /api/v1/web/keuangan/summary` (auth-only)
- **DTO**: `MyInvoiceSummaryResponse` di `application/dto/report_dto.go`.
- **Repo**: `FindSummaryByUserID` + struct `InvoiceSummary` di `domain/invoice/repository/interfaces.go`; implementasi SQL agregat di `postgres_invoice_repo.go`.
- **Use case**: `my_invoice_summary.go` di `application/query/`.
- **Handler**: `MySummary` di `interfaces/http/handler.go`.
- **Route**: `router.go` grup self-service → `santri.GET("/summary", h.MySummary)` tanpa `RequirePermission`.
- **Wiring**: `module.go`.

## Frontend — struktur baru di `sipon-ui`

### Types
- `shared/types/Kesantrian.ts`: tambah `status?: string | null` di `SantriProfile`.
- `shared/types/Keuangan.ts`: tambah `MyInvoiceSummary`:
  ```ts
  export interface MyInvoiceSummary {
    total_tagihan: number
    total_terbayar: number
    total_tunggakan: number
    jumlah_invoice: number
    jumlah_lunas: number
    jumlah_belum: number
  }
  ```

### Store
- **`useKeuanganStore`** (`app/stores/keuangan.ts`): tambah state `myInvoiceSummary: MyInvoiceSummary | null` + action `fetchMyInvoiceSummary()` → `GET /api/v1/web/keuangan/summary`.
- **`app/stores/dashboard.ts`** (baru): action `fetchAll()` → `Promise.all` dari `articleStore.fetchList(...)`, `kesantrianStore.fetchMyProfile()`, `keuanganStore.fetchMyInvoiceSummary()`. Expose `isLoading`/`error` agregat; data dibaca dari ketiga store domain.

### Navigasi — pindah grid aplikasi ke menu "Aplikasi" di layout utama
- **`app/pages/aplikasi/index.vue`** (baru, `layout: 'default'`): render `FeatureModuleGrid`.
- **`AppNavbar.vue`**: ubah `{ label: 'Aplikasi', to: '#' }` → `to: '/aplikasi'`.
- **`AppMobileBottomNav.vue`**: ubah item `Aplikasi` → `to: '/aplikasi'`.
- **`dashboard/index.vue`**: **hapus** render `FeatureModuleGrid`.

### Halaman dashboard — `app/pages/dashboard/index.vue` (tetap `layout: 'default'`)
Pertahankan `HeroBanner` + "Selamat datang, {nama}!", lalu 3 blok info:

1. **`DashboardArticleSection.vue`** (`app/components/dashboard/`) — grid artikel terbaru, reuse kartu dari `artikel/index.vue`, link `/artikel/:id`, empty state.
2. **`DashboardSantriCard.vue`** — status kesantrian + akun: `UAvatar`, fullname/username, NIS, program, email, badge status (`SANTRI`→success, `ALUMNI`→neutral, `DROP_OUT`→error), link `/profile`. 404 → CTA `/psb`.
3. **`DashboardBillingCard.vue`** — KPI: Total Tagihan, Total Dibayar, Total Tunggakan (`KeuanganAmountDisplay` + `formatRupiah`) + jumlah, link `/keuangan/tagihan`.

Urutan: Hero + sambutan → artikel → status kesantrian → tagihan.

## Fase Pengerjaan

**Fase 1 — Backend A**: `status` di `GetSantriResponse`. Checkpoint: `go build` + curl `/santri/profile` tampil `status`.

**Fase 2 — Backend B**: endpoint `GET /keuangan/summary`. Checkpoint: curl auth-only 200 + angka benar.

**Fase 3 — Types & Store**: update `Kesantrian.ts`/`Keuangan.ts`, `fetchMyInvoiceSummary`, `dashboard.ts`. Checkpoint: type-check.

**Fase 4 — Navigasi "Aplikasi"**: `aplikasi/index.vue` + update `AppNavbar`/`AppMobileBottomNav`, hapus grid dari dashboard. Checkpoint: menu Aplikasi jadi halaman nyata.

**Fase 5 — Dashboard santri**: 3 komponen + rakit `dashboard/index.vue`. Checkpoint: demo end-to-end lawan `sipon-be` lokal.

**Fase 6 — Polish & edge**: 404 not-a-santri → CTA `/psb`, artikel kosong, tagihan kosong, badge status variant.

## Verifikasi

1. `npm run dev` (sipon-ui) & `go build`/`go test` (sipon-be) lolos tiap fase.
2. Login user biasa → `/dashboard` = halaman santri (artikel → status+akun → tagihan), menu "Aplikasi" di navbar membuka halaman grid modul.
3. Ketiga endpoint di-hit tanpa permission key → 200 (bukan 403) untuk role `member`.
4. Uji empty/error state tiap blok.
5. Backend tetap sumber kebenaran: route admin/keuangan tetap require permission masing-masing.
