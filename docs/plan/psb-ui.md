# Plan UI: Implementasi PSB (Penerimaan Santri Baru) di `sipon-ui`

## Context

`docs/plan/psb.md` di repo `sipon-be` mendeskripsikan modul `psb` — dan modul itu **sudah diimplementasikan penuh di backend** (commit `43f3cb8`, `internal/modules/psb/` sudah ada lengkap: domain, application, infrastructure, interfaces/http). Jadi tugas sekarang murni membangun UI-nya di `sipon-ui` (Nuxt 4), berdasarkan kontrak API **nyata** yang sudah saya baca langsung dari kode (router.go, semua file `dto/*.go`, use case command), bukan cuma dari draf `psb.md` — karena beberapa detail penting berbeda dari draf awal (lihat "Temuan penting" di bawah).

`sipon-ui` sendiri belum ada UI kesantrian/psb sama sekali — hanya ada folder kosong placeholder (`app/pages/pengelola-santri/*`, `app/pages/santri/*`) yang ditujukan untuk domain `kesantrian` (santri existing), bukan untuk `psb` (pendaftar baru). PSB akan jadi area baru sepenuhnya.

**Keputusan user**: admin PSB UI reuse layout `system-admin` yang sudah ada (bukan bikin layout baru), jadi route admin di bawah `/system-admin/psb/*`.

## Temuan penting dari kode aktual (bukan asumsi dari draf plan)

1. **Gender di-hardcode `"1"` di backend** (`upsert_formulir.go`: `p.Gender = "1"` selalu, `UpsertFormulirRequest` bahkan tidak punya field gender). UI **tidak boleh** menampilkan input gender yang bisa diedit di form pendaftar — akan menyesatkan karena backend mengabaikannya. Tampilkan gender di response sebagai read-only saja (kalaupun ditampilkan).
2. **Tidak ada enforcement kuota di backend sama sekali** — `AdminReviewUseCase.Accept()` cuma cek status `diajukan`, tidak pernah baca `psb_setting.quota`. Jadi UI "kuota" hanya boleh berupa indikator informatif (dihitung di client dari list pendaftar vs `setting.quota[program]`), **tidak pernah** memblokir tombol Accept.
3. **Daftar ulang tidak membuka ulang form profil.** `Pendaftar.UpsertFormulir()` cuma boleh dipanggil saat status `draft`/`perlu_revisi` (selain itu dapat 409 `PENDAFTAR_INVALID_STATUS`). `POST /psb/daftar-ulang/submit` **tidak punya request body** — murni transisi status. Jadi wizard profil cuma dipakai **sekali**, sebelum diterima; halaman daftar-ulang hanya berisi: ringkasan profil read-only + uploader dokumen `stage=daftar_ulang` + tombol submit.
4. **Upload/hapus dokumen tidak digating oleh status pendaftar di backend** — jadi kontrol visibility per-stage di UI adalah kenyamanan UX saja, bukan pengaman keamanan.
5. **404 harus dibedakan maknanya**: `GET /psb/pendaftaran` 404 kalau belum ada setting aktif (→ tampilkan empty state "pendaftaran belum dibuka") ATAU kalau user belum pernah upsert formulir (→ tampilkan "mulai isi formulir", bukan error).
6. **Content-type dokumen yang diizinkan**: hanya `image/jpeg`, `image/png`, `application/pdf` (constant `AllowedContentTypes` di `dokumen_constant.go`) — beda dari avatar upload yang juga terima webp/gif.
7. **Confirm dokumen auto-replace**: `DokumenConfirmUseCase` otomatis soft-delete dokumen lama dengan `(stage, kind)` sama dan insert baris baru (id baru, status balik ke `pending`). Tidak ada endpoint "replace" terpisah — re-upload = presign+confirm ulang untuk kind/stage yang sama, lalu refetch list.
8. **`GET/POST /system-admin/psb/settings` return array flat, tanpa pagination meta** — halaman settings tidak perlu `UPagination`.
9. `manage_psb`/`manage_psb_settings` sudah terdaftar sebagai permission key dinamis (`GET /api/v1/web/role-permission/permission-keys`), dan halaman assign-permission-ke-role yang sudah ada (`app/pages/system-admin/roles/[id]/permissions.vue`) otomatis mendukungnya tanpa perubahan.
10. `quota`/`bank_accounts` adalah `json.RawMessage` di backend (bebas bentuk) — form admin settings perlu editor baris key→angka (quota) dan editor baris `{name, no}` (bank_accounts), bukan field tetap.

## Pola yang sudah ada di `sipon-ui` dan harus di-reuse (bukan ditulis ulang)

- **API layer**: `useApi()` (`app/composables/useApi.ts`), envelope `ApiSuccess<T>`/`ApiError` (`shared/types/ApiResponse.ts`), error selalu lewat `parseApiError()` (`app/utils/errorParser.ts`).
- **Store pattern**: satu Pinia store per domain seperti `app/stores/userManagement.ts` (state `items/meta/isLoading/isSubmitting/error`, tiap action try/catch + `parseApiError` + rethrow).
- **List/table page pattern**: seperti `app/pages/system-admin/users/index.vue` — debounce search lokal, `UTable` + named `#<key>-cell` slot, `UPagination`, `usePermission()`'s `can()`/`canAny()` inline, row actions via `AppRowActions` + `DropdownMenuItem[]`, modal terpisah per aksi.
- **Presign/confirm upload pattern**: seperti `app/components/profile/AvatarUploadModal.vue` — validasi client dulu, `requestXPresign()` → raw `fetch(presign_url, {method:'PUT'})` ke storage → `confirmX(key)` ke backend.
- **OTP modal**: `app/components/profile/VerifyEmailModal.vue` sudah reusable persis untuk gate "verifikasi email dulu sebelum lanjut ke PSB" — **tidak perlu bikin OTP UI baru**.
- **Belum ada** di codebase: komponen wizard/stepper (akan pakai `UStepper` dari `@nuxt/ui` v4, terpasang tapi belum dipakai di mana pun), komponen status-badge reusable (baru ada inline per halaman), komponen `<PermissionGate>` atau middleware permission per-route (masih inline `v-if="can(...)"` saja — ikuti pola ini, jangan bikin abstraksi baru).

## Struktur baru

### Types — `shared/types/Psb.ts` (satu file, mirror `Auth.ts`)

Semua tipe status (`PendaftarStatus` 9 state, `DokumenStage`, `DokumenKind`, `DokumenStatus`, `ReviewAction`, `PsbSettingStatus`), `PendaftarProfileFields` (~40 field mirror `Santri`, semua nullable), `UpsertFormulirRequest`/`PendaftarResponse`/`ListPendaftarItem`/`ListPendaftarQuery`, `DokumenPresignRequest/Response`, `DokumenConfirmRequest/Response`, `DokumenItemResponse`, `ReviewResponse`, `PsbQuota` (`Record<string, number>`), `PsbBankAccount`, `CreateSettingRequest`/`UpdateSettingRequest`/`SettingResponse`, `MessageResponse`. Field & nama JSON persis sama seperti di `internal/modules/psb/application/dto/*.go` (sudah diverifikasi baris-per-baris).

### Stores

- `app/stores/psb.ts` (self-service): `setting`, `pendaftar`, `dokumen[]`, `reviews[]` + loading/error state. Actions: `fetchActiveSetting()`, `fetchPendaftaran()` (404 → `pendaftar=null`, bukan error), `upsertFormulir(payload)`, `submitPendaftaran()`, `submitDaftarUlang()`, `fetchRiwayat()`, `fetchDokumen()`, `requestDokumenPresign(...)`, `confirmDokumen(...)`, `deleteDokumen(id)`.
- `app/stores/psbAdmin.ts` (`manage_psb` tier): `items`, `meta`, `selected` (detail+dokumen+riwayat). Actions: `fetchPendaftaranList(query)`, `fetchPendaftaranDetail(id)`, `fetchDetailRiwayat(id)`, `fetchDetailDokumen(id)`, `verifyDokumen(...)`, `rejectDokumen(...)`, `requestRevision(id, notes)`, `reject(id, notes)`, `accept(id)`, `markNotReregistered(id)`, `requestRevisionDaftarUlang(id, notes)`, `generateNIS(id)`.
- `app/stores/psbSetting.ts` (`manage_psb_settings` tier): `items` (flat, tanpa meta). Actions: `fetchSettings()`, `createSetting(payload)`, `updateSetting(id, payload)` (juga dipakai untuk close period lewat `{status:'closed'}`), `purgePeriod(id)`.

### Routing (final, sesuai keputusan reuse system-admin)

Self-service (layout `default`, tanpa gating permission — siapapun login boleh):

- `app/pages/psb/index.vue` — entry: cek active setting → cek pendaftar → gate email-verified (buka `VerifyEmailModal` yang sudah ada) → tampilan status-driven dengan CTA sesuai status.
- `app/pages/psb/formulir.vue` — wizard, redirect kalau status bukan `draft`/`perlu_revisi`.
- `app/pages/psb/riwayat.vue` — timeline riwayat penuh.
- `app/pages/psb/daftar-ulang.vue` — ringkasan profil read-only + uploader `stage=daftar_ulang` + submit, aktif hanya saat `diterima`/`perlu_revisi_daftar_ulang`.

Admin (layout `system-admin`):

- `app/pages/system-admin/psb/index.vue` — shortcut cards, "Pendaftar" (`can('manage_psb')`) dan "Periode PSB" (`can('manage_psb_settings')`) masing-masing gated terpisah.
- `app/pages/system-admin/psb/pendaftaran/index.vue` — list (`GET /admin/pendaftaran`), filter status+periode, `UTable`+`UPagination`.
- `app/pages/system-admin/psb/pendaftaran/[id]/index.vue` — detail: profil read-only, panel verify/reject per dokumen, timeline riwayat, action bar status-driven.
- `app/pages/system-admin/psb/settings/index.vue` — list periode (flat, tanpa pagination), create/update modal, close, purge.

### Komponen

Shared (`app/components/psb/*.vue` → `Psb*`): `StatusBadge.vue`, `ReviewTimeline.vue`, `DocumentUploader.vue` (presign/confirm/list/delete generik per `stage`+`kind`, reuse pola `AvatarUploadModal`), `ProfileSummary.vue`, `PeriodInfoCard.vue`.

Wizard (`app/components/psb/*.vue`): `FormWizard.vue` (orkestrator `UStepper`, autosave tiap step via `upsertFormulir`), `StepDataPribadi.vue`, `StepAlamatSekolah.vue`, `StepDataKependudukan.vue`, `StepOrangTuaWali.vue`, `StepDokumen.vue` (embed `DocumentUploader stage=pendaftaran`), `StepReview.vue` (embed `ProfileSummary` + submit).

Admin (`app/components/system-admin/psb/*.vue` → mengikuti prefix domain admin yang sudah ada, mis. pola `SystemAdminCreateUserModal`): `ReviewActionModal.vue` (generik utk request-revision/reject/request-revision-daftar-ulang + textarea notes), `AcceptConfirmModal.vue` (banner kuota informatif non-blocking, lihat temuan #2), `GenerateNisConfirmModal.vue`, `DokumenReviewPanel.vue`, `SettingFormModal.vue` (editor baris quota + bank_accounts), `PurgeConfirmModal.vue` (destructive, ketik ulang nama periode utk konfirmasi, tombol disabled kecuali `status==='closed' && !data_purged_at` — mirror `PsbSetting.CanPurge()`).

## Matriks visibility per status (self-service & admin)

| Status                      | Form editable            | Tombol submit        | Daftar-ulang uploader/submit   | Aksi admin                                       |
| --------------------------- | ------------------------ | -------------------- | ------------------------------ | ------------------------------------------------ |
| `draft`                     | ya                       | "Ajukan Pendaftaran" | hidden                         | -                                                |
| `diajukan`                  | tidak (redirect)         | hidden               | hidden                         | verify dokumen, request-revision, reject, accept |
| `perlu_revisi`              | ya                       | "Ajukan Ulang"       | hidden                         | -                                                |
| `ditolak`                   | tidak, terminal          | hidden               | hidden                         | -                                                |
| `diterima`                  | tidak                    | hidden               | uploader tampil, submit aktif  | mark-not-reregistered                            |
| `mengundurkan_diri`         | tidak, terminal          | hidden               | hidden                         | -                                                |
| `daftar_ulang`              | tidak                    | hidden               | uploader tampil, submit hidden | request-revision-daftar-ulang, generate-nis      |
| `perlu_revisi_daftar_ulang` | tidak                    | hidden               | uploader tampil, submit aktif  | -                                                |
| `selesai`                   | tidak, terminal (sukses) | hidden               | hidden, tampilkan NIS          | -                                                |

## Fase Pengerjaan

**Fase 1 — Types.** `shared/types/Psb.ts`. Checkpoint: type-check lolos, belum dipakai di mana pun.

**Fase 2 — Store self-service.** `app/stores/psb.ts`. Checkpoint: compile lolos.

**Fase 3 — Komponen display bersama.** `StatusBadge.vue`, `ReviewTimeline.vue`, `DocumentUploader.vue`, `ProfileSummary.vue`, `PeriodInfoCard.vue`. Checkpoint: render bersih (uji lewat stub sementara di satu halaman).

**Fase 4 — Komponen wizard.** `FormWizard.vue` + 6 `Step*.vue`. Checkpoint: compile lolos, belum di-routing.

**Fase 5 — Halaman self-service + nav.** `psb/index.vue`, `formulir.vue`, `riwayat.vue`, `daftar-ulang.vue`; tambah tile "Pendaftaran Santri Baru" ke `FeatureModuleGrid.vue` (tanpa gating permission — terbuka untuk semua user login). **Checkpoint: alur self-service demo end-to-end lawan API asli** — register → verifikasi email → isi wizard → upload dokumen → submit → lihat riwayat.

**Fase 6 — Store admin.** `app/stores/psbAdmin.ts`, `app/stores/psbSetting.ts`. Checkpoint: compile lolos.

**Fase 7 — Komponen admin.** Semua modal/panel di `app/components/system-admin/psb/*.vue`. Checkpoint: render via stub page.

**Fase 8 — Halaman admin operasional (`manage_psb`).** `system-admin/psb/index.vue`, `pendaftaran/index.vue`, `pendaftaran/[id]/index.vue`. **Checkpoint: user dengan `manage_psb` bisa review pendaftar asli dari Fase 5** — verify dokumen, request-revision, accept, mark-not-reregistered, request-revision-daftar-ulang, generate-nis — dan hasilnya kelihatan balik di halaman self-service.

**Fase 9 — Halaman settings (`manage_psb_settings`).** `system-admin/psb/settings/index.vue` + `SettingFormModal`/`PurgeConfirmModal`. Checkpoint: buat periode, close, purge, `data_purged_at` terisi, row periode tetap ada.

**Fase 10 — Polish & edge case.** Verifikasi gate email-verified di `/psb`, empty state "belum ada periode aktif", 404-sebagai-draft-baru, banner kuota informatif di `AcceptConfirmModal`. Tidak ada file baru kecuali ada celah yang ketemu saat uji manual.

## Verifikasi

1. `npm run dev` jalan tanpa error di tiap checkpoint fase, tidak ada import mati.
2. Uji manual browser: seluruh alur self-service (Fase 5 checkpoint) dan alur admin (Fase 8-9 checkpoint) end-to-end lawan `sipon-be` yang jalan lokal.
3. Uji visibility matrix: untuk tiap 9 status, pastikan tombol/form yang tampil/aktif sesuai tabel di atas (terutama form terkunci di luar `draft`/`perlu_revisi`, dan daftar-ulang tidak membuka ulang wizard — sesuai temuan #3).
4. Uji re-upload dokumen (kind+stage sama) → pastikan status balik ke `pending` dan list ter-refresh dengan id baru (temuan #7).
5. Uji permission gating: user tanpa `manage_psb`/`manage_psb_settings` tidak melihat card/menu terkait (dan API tetap 403 kalau dipaksa lewat URL langsung — backend tetap sumber kebenaran).
6. Uji purge: tombol disabled kecuali periode `closed` & belum di-purge; setelah purge, data hilang tapi row `psb_settings` tetap ada.
