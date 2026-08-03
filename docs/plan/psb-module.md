# Plan: Module PSB (Penerimaan Santri Baru)

## Context

Saat ini `kesantrian` hanya punya alur admisi minimal (`santri_requests`): user yang sudah punya akun mengajukan diri jadi santri, admin approve dengan NIS yang diinput manual. Tidak ada konsep periode/gelombang, kuota, biaya pendaftaran, verifikasi berkas bertahap, atau daftar ulang.

Tujuan plan ini: menambah module baru `psb` yang meng-cover alur pendaftaran santri baru secara penuh — dari registrasi akun (email belum verified), pengisian formulir + upload dokumen, verifikasi berkas oleh admin, keputusan diterima, daftar ulang (isi ulang formulir + bukti bayar administrasi awal), sampai NIS di-generate otomatis dan data pendaftar "naik kelas" jadi `santri` sungguhan di module `kesantrian`. Selain itu, `kesantrian` perlu tambahan kecil: field status siklus-hidup santri (`SANTRI` / `ALUMNI` / `DROP_OUT`) yang belum ada sama sekali hari ini.

`santri_requests` yang lama **tidak diubah/dihapus** — tetap ada sebagai jalur manual terpisah (mis. untuk mutasi/kasus khusus). `psb` adalah jalur baru yang lebih lengkap, tabelnya sendiri.

**Langkah pertama setelah plan ini disetujui:** salin dokumen ini apa adanya ke `docs/plan/psb-module.md` di repo `sipon-be` (dokumentasi hidup, sejajar dengan `docs/architecture/module-boundaries.md`) sebelum mulai implementasi kode.

## Keputusan arsitektur kunci

1. **Reuse identity's existing self-service register + OTP flow as-is.** Identity module sudah punya `POST /register` (bikin akun, email login-identity dibuat *unverified*) dan `request-otp` / `verify-otp` (OTP-based, bukan token-link) untuk verifikasi email. `psb` tidak menambah endpoint auth baru — pendaftar cukup register & login lewat identity seperti user biasa, lalu memakai JWT itu untuk hit endpoint `psb`. `Pendaftar` (entity baru) dibuat lazily di `psb` begitu user pertama kali mengisi formulir, terikat ke `user_id` (pola sama seperti `Santri` 1:1 `user_id`, lihat `internal/modules/kesantrian/domain/santri/entity/santri.go`).
2. **Tabel psb terpisah dari `santri`/`santri_dokumen`**, field formulir & jenis dokumen di-mirror dari `Santri`/`SantriDokumen` (bukan diduplikasi sebagai relasi). Baru saat NIS di-generate, datanya disalin jadi baris `santri` + `santri_dokumen` sungguhan.
3. **`kesantrian` dapat `Contract` baru** (module ini belum punya `contract.go` sama sekali hari ini — jadi ini file baru) dengan satu method `CreateSantriFromPendaftaran(...)` yang secara atomic: generate NIS, insert `santri`, copy dokumen jadi `santri_dokumen`. `psb` memanggil ini lewat port+gateway miliknya sendiri (pola persis `kesantrian`'s `ports.AccountProvisioner` + `infrastructure/identitygateway`, lihat `docs/architecture/module-boundaries.md`).
4. **Generasi NIS dimiliki `kesantrian`**, bukan `psb` — karena `kesantrian` yang punya `valueobject.NIS` dan yang menegakkan uniqueness di tabel `santri`. `psb` hanya mengirim gender + tahun angkatan + profil + dokumen; `kesantrian` yang menghitung nomor urut berikutnya per (tahun, gender) dan reset otomatis karena prefix tahun beda per angkatan.
5. **Cleanup data psb bersifat manual, per-periode, hanya lewat aksi admin eksplisit** (bukan otomatis saat proses psb selesai) — sesuai revisi user. Admin memilih satu `psb_setting` (periode) yang statusnya sudah `CLOSED`, lalu memicu aksi "hapus data periode" yang hard-delete semua `pendaftar` + `pendaftar_dokumen` (+ object storage-nya) untuk periode itu. `psb_setting` sendiri **tidak ikut terhapus** (tetap ada untuk histori kuota/laporan), hanya ditandai `data_purged_at`.
6. **Status dropout santri** memakai istilah `DROP_OUT` (dikonfirmasi user).

## Alur status pendaftaran (dengan negative flow + revisi berkali-kali)

```
                     ┌───────────────request revision───────────────┐
                     ▼                                              │
DRAFT ──submit──▶ DIAJUKAN ───────────────────────────────▶ PERLU_REVISI
                     │  │                                      │ submit lagi (balik ke DIAJUKAN)
                     │  └──reject (final)──▶ DITOLAK (terminal, daftar lagi tahun depan)
                     │
                   accept (cek kuota)
                     ▼
                  DITERIMA ──(tidak daftar ulang sebelum deadline, admin tandai)──▶ MENGUNDURKAN_DIRI (terminal)
                     │
              submit daftar ulang
                     ▼
                     ┌───────────────request revision───────────────┐
                     ▼                                              │
                DAFTAR_ULANG ───────────────────────▶ PERLU_REVISI_DAFTAR_ULANG
                     │                                    │ submit lagi (balik ke DAFTAR_ULANG)
                     └─verifikasi lolos, admin "Generate NIS"──▶ SELESAI (terminal, sukses)
```

- `DRAFT`: pendaftar masih isi formulir + upload dokumen tahap pendaftaran (termasuk bukti bayar biaya pendaftaran). Bisa edit bebas.
- `DIAJUKAN`: submitted, menunggu verifikasi berkas admin. Dokumen diverifikasi satu-satu (status per dokumen: pending/verified/rejected, pola sama `SantriDokumen.Verify/Reject`). Dari sini admin punya **3 aksi**: minta revisi, tolak final, atau terima.
- `PERLU_REVISI`: admin minta revisi (mis. dokumen buram/kurang lengkap, formulir belum sesuai) dengan catatan spesifik. Pendaftar edit formulir/upload ulang dokumen yang diminta, lalu submit lagi → balik ke `DIAJUKAN` untuk direview ulang. **Bisa terjadi berkali-kali** sebelum akhirnya diterima/ditolak — setiap putaran tercatat di tabel riwayat (lihat di bawah), beda dengan `DITOLAK` yang final.
- `DITOLAK`: keputusan final tidak lolos — terminal untuk periode ini, pendaftar baru bisa daftar lagi di periode/tahun berikutnya (bukan resubmit di periode yang sama).
- `DITERIMA`: lolos verifikasi berkas, menunggu daftar ulang. Pengecekan kuota (per kategori program, lihat quota di `psb_setting`) dilakukan di sini, bukan saat submit — supaya orang tetap bisa apply meski kuota kelihatan penuh (ada yang diterima tapi tidak daftar ulang).
- `MENGUNDURKAN_DIRI`: admin menandai manual kalau pendaftar tidak melakukan daftar ulang sampai batas waktu — terminal negative flow kedua yang diminta user.
- `DAFTAR_ULANG`: pendaftar sudah upload ulang formulir + dokumen tahap daftar ulang (full administrasi awal + bukti bayar), menunggu verifikasi final. Simetris dengan tahap pertama: admin bisa minta revisi (`PERLU_REVISI_DAFTAR_ULANG`, bisa berkali-kali, pendaftar submit ulang → balik ke `DAFTAR_ULANG`) sebelum akhirnya verifikasi lolos dan NIS di-generate.
- `SELESAI`: admin generate NIS → santri tercipta di `kesantrian`. Pendaftar menyimpan `santri_id`/`nis` hasil untuk jejak (sebelum datanya nanti dihapus manual oleh admin).

Dokumen di `pendaftar_dokumen` punya kolom `stage` (`pendaftaran` | `daftar_ulang`) selain `kind` (reuse vocabulary `DokumenKind` yang sudah ada: `surat_pernyataan, ktp, kk, mutasi, pembayaran`) — jadi dokumen yang sama jenisnya bisa diupload ulang di tahap daftar ulang tanpa bentrok dengan punya tahap pendaftaran, dan juga tetap dipakai saat versi direvisi (upload baru menggantikan yang lama, kind+stage yang sama). Saat generate NIS, dokumen yang disalin ke `santri_dokumen` adalah dokumen stage `daftar_ulang` (set administrasi paling final/lengkap).

### Riwayat revisi (`pendaftar_reviews`)

Setiap keputusan admin di kedua tahap (`request-revision`, `reject`, `accept` / `request-revision-daftar-ulang`, `generate-nis`) menambah satu baris baru ke tabel append-only ini — bukan overwrite — supaya histori berkali-kali revisi tetap lengkap:
```sql
id            UUID PK
pendaftar_id  UUID NOT NULL REFERENCES pendaftar(id) ON DELETE CASCADE
stage         VARCHAR(20) NOT NULL CHECK (stage IN ('pendaftaran','daftar_ulang'))
action        VARCHAR(20) NOT NULL CHECK (action IN ('perlu_revisi','ditolak','diterima'))
notes         TEXT                          -- catatan detail apa yang perlu direvisi / alasan tolak
reviewed_by   UUID NOT NULL
created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
```
Tidak ada kolom counter revisi terpisah di `pendaftar` — jumlah revisi cukup dihitung dari `COUNT(*) WHERE action='perlu_revisi'` di tabel ini. Pendaftar bisa lihat riwayatnya sendiri (`GET /psb/pendaftaran/riwayat`), admin bisa lihat punya siapapun (`GET /psb/admin/pendaftaran/:id/riwayat`).

## Skema data baru

### `psb_settings` (periode PSB)
```sql
id            UUID PK
name          VARCHAR(200) NOT NULL          -- "PSB 2026/2027" dst, bebas
start_period  DATE NOT NULL
end_period    DATE NOT NULL
status        VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','closed'))
quota         JSONB NOT NULL DEFAULT '{}'    -- map[string]int, mis. {"tahfidh_pa":30,"tahfidh_pi":30,"kitab_pa":20,"kitab_pi":20}
reg_fee       NUMERIC(12,2) NOT NULL DEFAULT 0
bank_accounts JSONB NOT NULL DEFAULT '[]'    -- [{"name":"BSI","no":"1234567890"}, ...]
data_purged_at TIMESTAMPTZ                  -- diisi saat admin melakukan purge manual
created_at, updated_at, deleted_at
```
`quota` dipakai `map[string]int` (bukan struct Go tetap) supaya "bisa tambah jika dibutuhkan" (kategori baru) tidak perlu migration/redeploy — cukup key JSON baru. `bank_accounts` demikian juga array of `{name, no}`.

### `pendaftar` (mirror field `Santri`, lihat `santri.go` untuk daftar lengkap field profil: nickname, hobby, purpose, motivation_entry, pob, dob, blood, address/sub_district/district/province/postal_code, previous_pondok_*, nik/no_kk/nisn/no_kip/no_kks/no_pkh, workplace/department, home_status, father_*/mother_*/guardian_* — semua kolom itu diduplikasi persis nama & tipenya di tabel ini)
```sql
id              UUID PK
user_id         UUID NOT NULL              -- no FK, cross-module (ke identity)
psb_setting_id  UUID NOT NULL REFERENCES psb_settings(id)
gender          VARCHAR(2) NOT NULL CHECK (gender IN ('1','2'))   -- diisi di formulir, dipakai utk cek kuota & generate NIS nanti
program         VARCHAR(50)                -- key kategori kuota, mis. "tahfidh_pa"
-- ...seluruh kolom profil sama seperti santri (lihat migrations/005_create_kesantrian_tables.up.sql)...

status          VARCHAR(30) NOT NULL DEFAULT 'draft'
                CHECK (status IN ('draft','diajukan','perlu_revisi','ditolak','diterima',
                                   'mengundurkan_diri','daftar_ulang','perlu_revisi_daftar_ulang','selesai'))
accepted_by     UUID, accepted_at TIMESTAMPTZ         -- keputusan diterima
santri_id       UUID, nis VARCHAR(10)       -- hasil generate NIS, utk jejak sebelum data dipurge

created_at, updated_at, deleted_at
```
Unique partial index: `(user_id, psb_setting_id) WHERE deleted_at IS NULL` — satu pendaftaran per user per periode. Catatan reject/revisi & siapa/kapan me-review **tidak** disimpan sebagai kolom di `pendaftar` (tidak representatif untuk kejadian berkali-kali) — semuanya tercatat sebagai baris baru di `pendaftar_reviews` (lihat bagian "Riwayat revisi" di atas), termasuk untuk tahap daftar ulang.

### `pendaftar_reviews`
Lihat DDL lengkap di bagian "Riwayat revisi" di atas — tabel append-only, satu baris per keputusan admin (revisi/tolak/terima di kedua tahap).

### `pendaftar_dokumen`
```sql
id               UUID PK
pendaftar_id     UUID NOT NULL REFERENCES pendaftar(id) ON DELETE CASCADE
stage            VARCHAR(20) NOT NULL CHECK (stage IN ('pendaftaran','daftar_ulang'))
kind             VARCHAR(30) NOT NULL CHECK (kind IN ('surat_pernyataan','ktp','kk','mutasi','pembayaran'))
key              TEXT NOT NULL
status           VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','verified','rejected'))
original_filename VARCHAR(500), mime_type VARCHAR(200), size BIGINT, notes TEXT
verified_by UUID, verified_at TIMESTAMPTZ
created_at, updated_at, deleted_at
```
Unique partial index: `(pendaftar_id, stage, kind) WHERE deleted_at IS NULL`.

Migration file baru: `migrations/006_create_psb_tables.up.sql` / `.down.sql`, mengikuti gaya `005_create_kesantrian_tables.up.sql` (UUID PK `gen_random_uuid()`, no FK lintas-module, CHECK constraint untuk enum, index eksplisit).

## Tambahan kecil di `kesantrian`: status siklus-hidup santri

`Santri` hari ini tidak punya kolom status sama sekali (dicek langsung ke kode/migration — tidak ada). Tambahkan:
- Migration baru `migrations/007_add_status_to_santri.up.sql`: `ALTER TABLE santri ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'SANTRI' CHECK (status IN ('SANTRI','ALUMNI','DROP_OUT'))`, plus `status_changed_by UUID`, `status_changed_at TIMESTAMPTZ`, `status_notes TEXT`.
- `domain/santri/constant/santri_constant.go`: tambah `SantriStatus` type + `SantriStatusSantri/Alumni/DropOut`.
- `domain/santri/entity/santri.go`: field `Status`, method `MarkAlumni(changedBy string)`, `MarkDropOut(changedBy string, notes *string)` (keduanya hanya valid dari status `SANTRI`, pola sama seperti `SantriDokumen.Verify/Reject`).
- Use case baru `application/command/change_santri_status.go`, endpoint admin baru `POST /santri/admin/:id/status` (gated `manage_santri`, permission sudah ada — tidak perlu key baru untuk ini).
- Santri yang baru dibuat lewat `CreateSantriFromPendaftaran` (dari psb) otomatis `status='SANTRI'`.

## Cross-module: `kesantrian` Contract baru

`internal/modules/kesantrian/contract.go` (file baru — module ini belum pernah expose Contract):
```go
package kesantrian

type SantriDokumenInput struct {
    Kind             string
    Key              string
    OriginalFilename *string
    MimeType         *string
    Size             *int64
    VerifiedBy       *string
    VerifiedAt       *time.Time
}

type CreateSantriFromPendaftaranInput struct {
    UserID    string
    Gender    string // "1" atau "2"
    EntryYear string // 2 digit, mis. "26"
    // ...seluruh field profil, sama seperti dto.CreateSantriRequest tapi lebih lengkap...
    Dokumen []SantriDokumenInput
}

type CreateSantriFromPendaftaranResult struct {
    SantriID string
    NIS      string
}

type Contract interface {
    CreateSantriFromPendaftaran(ctx context.Context, in CreateSantriFromPendaftaranInput) (*CreateSantriFromPendaftaranResult, error)
}

var _ Contract = (*Module)(nil)
```

Use case baru `internal/modules/kesantrian/application/command/create_santri_from_pendaftaran.go`:
- Repo baru: `SantriRepository.FindMaxSequence(ctx, prefix string) (int, error)` — cari nomor urut 3-digit terbesar untuk NIS yang match `prefix + '%'` (prefix = `1000<gender><year>`).
- Di dalam `transactor.WithTx`: pakai `pg_advisory_xact_lock(hashtext(prefix))` sebelum hitung `FindMaxSequence`, supaya dua admin generate NIS bersamaan untuk kombinasi tahun+gender yang sama tidak collide — baru insert `santri` dengan NIS = prefix + urut+1 (zero-padded 3 digit), lalu insert satu baris `santri_dokumen` per item `Dokumen` (langsung status `verified`, bawa `verified_by/verified_at` dari input — karena sudah diverifikasi di tahap psb).
- Setelah commit: best-effort `provisioner.AddNISLoginIdentity(ctx, userID, nis)` (pola sama `ApproveSantriRequestUseCase`).
- `kesantrian/module.go`: implement method `Contract` di `*Module`, delegasikan ke use case baru.

`psb` module sisi pemanggil, ikuti pola `docs/architecture/module-boundaries.md` persis:
- `internal/modules/psb/application/ports/kesantrian_provisioner.go` — port versi psb (`CreateSantriFromPendaftaran(ctx, kesantrian.CreateSantriFromPendaftaranInput) (*kesantrian.CreateSantriFromPendaftaranResult, error)`).
- `internal/modules/psb/infrastructure/kesantriangateway/gateway.go` — adapter delegasi ke `kesantrian.Contract`.
- `psb.NewModule(..., identityContract identity.Contract, kesantrianContract kesantrian.Contract, ...)`.
- Wiring baru di `cmd/api/main.go`: bikin urutan `identity → kesantrian → psb` (psb butuh kesantrian.Contract), lalu `psb.RegisterRoutes(engine)`.

## Struktur module `psb` (mengikuti skeleton `kesantrian` persis)

```
internal/modules/psb/
  module.go
  domain/
    setting/{entity,constant,repository}       -- PsbSetting
    pendaftar/{entity,constant,repository,valueobject}  -- Pendaftar (status state machine)
    dokumen/{entity,constant,repository}        -- PendaftarDokumen
    review/{entity,constant,repository}         -- PendaftarReview (riwayat revisi/tolak/terima, append-only)
  application/
    command/  -- upsert_formulir.go, submit_pendaftaran.go, request_revision.go, reject_pendaftaran.go,
              -- accept_pendaftaran.go, mark_not_reregistered.go, submit_daftar_ulang.go,
              -- request_revision_daftar_ulang.go, generate_nis.go,
              -- dokumen_upload.go (presign/confirm/delete), dokumen_review.go (verify/reject),
              -- manage_setting.go (create/update/close), purge_period.go
    query/    -- get_pendaftaran.go, list_pendaftaran.go, list_dokumen.go, list_reviews.go, get_setting.go, list_settings.go
    dto/      -- setting_dto.go, pendaftaran_dto.go, dokumen_dto.go, review_dto.go
    ports/    -- kesantrian_provisioner.go, storage.go (copy shape dari kesantrian), transactor.go
    errors.go
  infrastructure/
    persistence/  -- postgres_setting_repo.go, postgres_pendaftar_repo.go, postgres_dokumen_repo.go, postgres_review_repo.go, postgres_transactor.go, helpers.go
    external/     -- minio_uploader.go (copy pola kesantrian)
    kesantriangateway/gateway.go
  interfaces/http/
    handler.go, router.go
```

Reuse langsung (bukan tulis ulang): pola `helpers.go` (nullStr/nullTime/isUniqueViolation), pola `postgres_transactor.go` (ctx-based tx via `execerFromContext`), pola presign/confirm dokumen dari `kesantrian/application/command/dokumen_*.go`, pola handler/router dari `kesantrian/interfaces/http/*`.

## Permission keys baru

Ditambahkan di `internal/modules/identity/domain/role/constant/permission_constant.go` (satu-satunya tempat semua permission key didefinisikan, termasuk yang module-specific — lihat `manage_santri`):
- `PermissionManagePSB = "manage_psb"` — verifikasi berkas, accept/reject, verifikasi daftar ulang, generate NIS (operasional harian).
- `PermissionManagePSBSettings = "manage_psb_settings"` — kelola periode/kuota/biaya/rekening, dan aksi purge data periode (destruktif, dipisah dari operasional harian — sama pola `manage_system_settings` vs `manage_users`).

User yang menentukan role custom mana yang dapat key ini — tidak ditambahkan ke `RolePermissions` default manapun kecuali diminta.

## Ringkasan endpoint HTTP (`/api/v1/web/psb/...`)

Self-service (butuh JWT, tanpa permission khusus — siapapun yang login boleh daftar):
- `GET /psb/setting/active` — lihat periode yang sedang buka (kuota, biaya, rekening)
- `GET /psb/pendaftaran`, `PUT /psb/pendaftaran` (upsert formulir, hanya saat `draft`/`perlu_revisi`)
- `POST /psb/pendaftaran/submit` (`draft|perlu_revisi → diajukan`)
- `GET /psb/pendaftaran/riwayat` (lihat riwayat revisi/keputusan milik sendiri)
- `POST /psb/dokumen/presign`, `POST /psb/dokumen/confirm`, `GET /psb/dokumen`, `DELETE /psb/dokumen/:id`
- `POST /psb/daftar-ulang/submit` (`diterima|perlu_revisi_daftar_ulang → daftar_ulang`)

Admin (`RequirePermission("manage_psb")`):
- `GET /psb/admin/pendaftaran`, `GET /psb/admin/pendaftaran/:id`, `GET /psb/admin/pendaftaran/:id/riwayat`
- `POST /psb/admin/pendaftaran/:id/dokumen/:dokumenId/verify|reject`
- `POST /psb/admin/pendaftaran/:id/request-revision` (`diajukan → perlu_revisi`, dengan catatan; bisa berkali-kali)
- `POST /psb/admin/pendaftaran/:id/reject` (`diajukan → ditolak`, final)
- `POST /psb/admin/pendaftaran/:id/accept` (`diajukan → diterima`, cek kuota di sini)
- `POST /psb/admin/pendaftaran/:id/mark-not-reregistered`
- `POST /psb/admin/pendaftaran/:id/request-revision-daftar-ulang` (`daftar_ulang → perlu_revisi_daftar_ulang`, bisa berkali-kali)
- `POST /psb/admin/pendaftaran/:id/generate-nis`

Admin settings (`RequirePermission("manage_psb_settings")`):
- `GET/POST /psb/admin/settings`, `PUT /psb/admin/settings/:id` (termasuk set `status=closed`)
- `POST /psb/admin/settings/:id/purge` (hard-delete `pendaftar`+`pendaftar_dokumen`+file storage utk periode itu; ditolak kalau `status != closed`)

## File yang dibuat/diubah (ringkas)

**Baru (module `psb`, ~28 file):** seluruh struktur di atas (termasuk domain `review`) + `migrations/006_create_psb_tables.up.sql`/`.down.sql` (mencakup `psb_settings`, `pendaftar`, `pendaftar_dokumen`, `pendaftar_reviews`).

**Diubah (`kesantrian`):**
- `domain/santri/constant/santri_constant.go`, `domain/santri/entity/santri.go` — tambah status
- `domain/santri/repository/interfaces.go` — tambah `FindMaxSequence`
- `infrastructure/persistence/postgres_santri_repo.go` — implement `FindMaxSequence`, kolom `status`
- **Baru:** `contract.go`, `application/command/create_santri_from_pendaftaran.go`, `application/command/change_santri_status.go`
- `module.go` — wire use case baru, implement `Contract`
- `interfaces/http/router.go`, `handler.go` — endpoint status baru
- `migrations/007_add_status_to_santri.up.sql`/`.down.sql`

**Diubah (identity):**
- `domain/role/constant/permission_constant.go` — 2 key baru

**Diubah (`cmd/api/main.go`):** wiring `psb` module setelah `kesantrian`.

## Verifikasi

1. `go build ./...` harus lolos setelah semua file baru & perubahan.
2. Jalankan migration (`go run cmd/migrate/main.go up`) di DB lokal/dev, pastikan `006`/`007` apply bersih dan `down` bisa rollback.
3. Manual smoke test end-to-end pakai curl/Postman: register akun (identity) → verify OTP email → isi formulir psb → upload dokumen (presign+confirm) → submit → (sebagai admin) verify dokumen satu-satu → accept → cek kuota berkurang → (sebagai pendaftar) submit daftar ulang + dokumen → (admin) verify + generate-nis → cek baris baru muncul di `santri`/`santri_dokumen` dengan NIS sesuai format, dan `pendaftar.status='selesai'`.
4. Test alur revisi berkali-kali: dari `diajukan`, admin `request-revision` 2-3 kali berturut-turut (pendaftar edit & submit ulang tiap kali) sebelum akhirnya `accept` — pastikan tiap putaran menambah baris baru di `pendaftar_reviews` (bukan overwrite) dan `GET .../riwayat` menampilkan urutan lengkapnya. Ulangi juga untuk tahap daftar ulang (`perlu_revisi_daftar_ulang`).
5. Test race NIS: generate NIS dua pendaftar gender+tahun sama nyaris bersamaan, pastikan tidak ada duplicate NIS (advisory lock bekerja).
6. Test permission gating: user tanpa `manage_psb`/`manage_psb_settings` dapat 403 di endpoint admin.
7. Test purge: coba purge periode yang masih `active` → ditolak; set `closed` → purge sukses, baris `pendaftar`/`pendaftar_dokumen`/`pendaftar_reviews` hilang, `psb_settings.data_purged_at` terisi, row `psb_settings` tetap ada.
