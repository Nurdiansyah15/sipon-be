# Fix: Konsistensi DB↔MinIO + Orphan Upload Cleanup (PSB, Kesantrian, Identity)

## Context

Analisis alur presign/confirm/delete storage di 3 modul (avatar profile, santri
dokumen, PSB dokumen) menemukan beberapa celah:

1. **Bug purge PSB**: endpoint admin `POST /admin/settings/:id/purge` melakukan
   hard-delete `pendaftar` yang cascade menghapus `pendaftar_dokumen` di DB —
   **tanpa pernah menghapus file fisiknya di MinIO** — sehingga jadi orphan
   permanen tak terlacak (key sudah hilang dari DB karena hard-delete).
   → **Keputusan user: fitur ini ditunda dulu (dinonaktifkan), bukan diperbaiki
   sekarang.**

2. **Urutan delete yang tidak aman** di avatar & santri dokumen: DB ditulis
   (soft-delete / hapus referensi) **dulu**, baru file MinIO dihapus, dan error
   dari langkah MinIO **diabaikan** (`_ = ...`). Kalau penghapusan MinIO gagal,
   DB sudah terlanjur bilang "sudah tidak ada" padahal file masih nyangkut —
   orphan lagi, dengan cara berbeda. PSB dokumen delete justru sudah melakukan
   ini dengan urutan yang benar (hapus MinIO dulu, cek error, baru update DB).

3. **File yang di-presign tapi tidak pernah di-confirm** (client batal upload,
   presigned URL expired sebelum sempat confirm) — tidak ada catatan apa pun
   di DB untuk file ini, dan tidak ada mekanisme pembersihan otomatis sama
   sekali. File menumpuk selamanya di MinIO.

Tujuan perbaikan (sesuai arahan user): **fokus dulu ke poin 2 dan 3** — pastikan
begitu suatu key dihapus/diganti di DB, file MinIO-nya juga benar-benar
terhapus; dan bangun mekanisme pelacakan + pembersihan berkala untuk upload yang
tidak pernah di-confirm. Poin 1 (purge) cukup dinonaktifkan sementara, tidak
diperbaiki di iterasi ini.

**Riset teknis yang mengubah pendekatan awal**: sempat dipertimbangkan pakai
MinIO Lifecycle Policy + tagging (expire object yang tidak ditag "confirmed"),
tapi dicek langsung ke source `minio-go/v7` — **S3/MinIO lifecycle filter tidak
punya "kecuali tag ini"**, hanya bisa filter inclusive (assign rule ke object
yang PUNYA tag/prefix tertentu). Sempat dipertimbangkan juga tabel tracking di
DB + scheduled job, tapi **keputusan final user: pakai solusi native S3 yang
benar — staging-prefix + promote-on-confirm**. Ini lebih invasif (ubah key yang
di-generate presign, tambah `CopyObject`, sentuh presign+confirm di 3 modul
sekaligus) tapi tidak butuh tabel/scheduler baru sama sekali — MinIO yang
menangani expiry secara native lewat lifecycle rule berbasis prefix.

## Bagian 1 — Nonaktifkan Fitur Purge Period

File: `internal/modules/psb/interfaces/http/router.go`

Comment-out baris registrasi route (bukan hapus use case/handler-nya, supaya
mudah diaktifkan lagi nanti setelah bug MinIO-nya beneran diperbaiki):

```go
settings := admin.Group("/settings")
settings.Use(middleware.RequirePermission("manage_psb_settings"))
{
	settings.GET("", h.ListSettings)
	settings.POST("", h.CreateSetting)
	settings.PUT("/:id", h.UpdateSetting)
	// TODO: dinonaktifkan sementara — hard-delete pendaftar/pendaftar_dokumen
	// belum membersihkan file MinIO terkait (orphan permanen). Aktifkan lagi
	// setelah PurgePeriodUseCase diperbaiki untuk menghapus objek MinIO dulu
	// sebelum hard-delete DB.
	// settings.POST("/:id/purge", h.PurgePeriod)
}
```

`PurgePeriodUseCase`, handler `PurgePeriod`, dan wiring di `module.go` **tidak**
diubah/dihapus — hanya endpoint HTTP-nya yang dicabut dari routing (jadi 404
kalau dipanggil).

## Bagian 2 — Perbaiki Urutan Delete (guarantee: DB bilang hilang ⇒ MinIO benar hilang)

Pola yang benar (sudah dipakai `DokumenDeleteUseCase` PSB,
`internal/modules/psb/application/command/dokumen_upload.go:161-168`): **hapus
MinIO dulu, cek error (gagal → return error, DB tidak disentuh), baru
soft-delete/update DB.** Terapkan pola yang sama ke 2 tempat yang masih pakai
urutan terbalik + error diabaikan:

### 2a. `AvatarDeleteUseCase.Execute` — `internal/modules/identity/application/command/avatar.go:141-171`

Ubah dari (DB dulu, MinIO diabaikan):
```go
user.AvatarKey = nil
if err := uc.userRepo.Update(ctx, user); err != nil { ... }
if oldKey != nil && *oldKey != "" {
	_ = uc.fileUploader.MarkDeleted(ctx, *oldKey)
}
```
menjadi (MinIO dulu + cek error, baru DB):
```go
if oldKey != nil && *oldKey != "" {
	if err := uc.fileUploader.MarkDeleted(ctx, *oldKey); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
}
user.AvatarKey = nil
if err := uc.userRepo.Update(ctx, user); err != nil {
	return nil, kernel.Wrap(application.ErrCodeInternal, err)
}
```

### 2b. `DokumenDeleteUseCase.Execute` (kesantrian) — `internal/modules/kesantrian/application/command/dokumen_manage.go:37-63`

Ubah dari (soft-delete dulu, MinIO diabaikan):
```go
dokumen.SoftDelete()
if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
	return uc.dokumenRepo.Update(txCtx, dokumen)
}); err != nil { ... }
_ = uc.fileUploader.DeleteObject(ctx, dokumen.Key, ports.PrivacyPrivate)
```
menjadi (MinIO dulu + cek error, baru soft-delete):
```go
if err := uc.fileUploader.DeleteObject(ctx, dokumen.Key, ports.PrivacyPrivate); err != nil {
	return nil, kernel.Wrap(application.ErrCodeInternal, err)
}
dokumen.SoftDelete()
if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
	return uc.dokumenRepo.Update(txCtx, dokumen)
}); err != nil {
	return nil, kernel.Wrap(application.ErrCodeInternal, err)
}
```

### 2c. Best-effort secondary cleanup — ganti silent-ignore jadi logged

Tempat yang menghapus file LAMA sebagai efek samping (bukan aksi delete
eksplisit user) tetap best-effort (tidak boleh menggagalkan aksi utama), tapi
error-nya jangan didiamkan total — pakai `slog.Warn` seperti idiom yang sudah
ada di modul kesantrian (`internal/modules/kesantrian/application/command/update_santri.go:39`,
dll: `slog.Warn("kesantrian: best-effort ... failed", "user_id", ..., "error", err)`):

- `AvatarConfirmUseCase.Execute` (avatar.go:117-119) — cleanup `oldKey` saat ganti avatar:
  ```go
  if oldKey != nil && *oldKey != normalizedKey {
  	if err := uc.fileUploader.MarkDeleted(ctx, *oldKey); err != nil {
  		slog.Warn("identity: best-effort hapus avatar lama gagal", "key", *oldKey, "error", err)
  	}
  }
  ```
- `DokumenConfirmUseCase.Execute` (PSB, `dokumen_upload.go:90-97`) — cleanup dokumen lama saat replace stage+kind sama:
  ```go
  for _, d := range existing {
  	if d.Kind == kind && d.DeletedAt == nil {
  		if err := uc.fileUploader.DeleteObject(ctx, d.Key, ports.PrivacyPrivate); err != nil {
  			slog.Warn("psb: best-effort hapus dokumen lama gagal", "key", d.Key, "error", err)
  		}
  		d.SoftDelete()
  		if err := uc.dokumenRepo.Update(ctx, d); err != nil {
  			slog.Warn("psb: gagal update soft-delete dokumen lama", "id", d.ID, "error", err)
  		}
  	}
  }
  ```

## Bagian 3 — Staging-Prefix + Promote-on-Confirm untuk Presign yang Tidak Di-confirm

Solusi native S3/MinIO: object yang di-presign selalu masuk dulu ke bawah
prefix `pending/` di bucket yang sama. Sebuah lifecycle rule di bucket
(`Filter.Prefix = "pending/"`, `Expiration.Days = N`) otomatis menghapus
object yang tertinggal di prefix itu — artinya upload yang tidak pernah
di-confirm otomatis lenyap dari MinIO tanpa kode aplikasi mana pun perlu tahu.
Saat confirm, object "dipromosikan" keluar dari zona expiring itu: di-copy ke
key final (di luar prefix `pending/`), staging object lama dihapus, dan key
FINAL itulah yang disimpan ke DB — bukan key hasil presign.

Sudah dicek ke source `minio-go/v7`: `Client.CopyObject(ctx, dst CopyDestOptions,
src CopySrcOptions) (UploadInfo, error)` (`api-copy-object.go:27`) mendukung
server-side copy dalam satu bucket yang sama, dan `Client.SetBucketLifecycle`
(`api-bucket-lifecycle.go:34`) menerima `*lifecycle.Configuration` dengan
`Rule.RuleFilter.Prefix` + `Rule.Expiration.Days` — persis yang dibutuhkan.

Pola ini **identik di ketiga modul** — dijelaskan sekali pakai identity/avatar
sebagai contoh, lalu tabel perbedaan nama per modul di bagian akhir. Tidak ada
tabel/scheduler baru yang perlu ditambahkan — hanya perubahan pada
key-generation, interface `FileUploader`, dan setup lifecycle sekali di startup.

### 3a. Tambah method `PromoteUpload` ke `ports.FileUploader` (3 modul)

`internal/modules/identity/application/ports/storage.go` (dan file setara di
kesantrian, psb):
```go
type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	ConfirmUpload(ctx context.Context, key string) error
	MarkDeleted(ctx context.Context, key string) error   // (avatar) / DeleteObject (kesantrian & psb) — tetap seperti sekarang
	PromoteUpload(ctx context.Context, stagingKey, finalKey string, privacy PrivacyRule) error // BARU
	EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error                    // BARU
	PublicURL(key string) string
	KeyFromURL(url string) string
}
```

Implementasi di `MinioFileUploader` (contoh `internal/modules/identity/infrastructure/external/minio_uploader.go`,
identik di kesantrian/psb kecuali pemilihan bucket lewat parameter `privacy`
yang sudah ada di semua adapter ini):
```go
func (u *MinioFileUploader) PromoteUpload(ctx context.Context, stagingKey, finalKey string, privacy ports.PrivacyRule) error {
	if u == nil {
		return nil
	}
	bucket := u.bucket
	if privacy == ports.PrivacyPrivate {
		bucket = u.privateBucket
	}
	dst := minio.CopyDestOptions{Bucket: bucket, Object: finalKey}
	src := minio.CopySrcOptions{Bucket: bucket, Object: stagingKey}
	if _, err := u.client.CopyObject(ctx, dst, src); err != nil {
		return fmt.Errorf("promote upload copy: %w", err)
	}
	if err := u.client.RemoveObject(ctx, bucket, stagingKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("promote upload cleanup staging: %w", err)
	}
	return nil
}

func (u *MinioFileUploader) EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error {
	if u == nil {
		return nil
	}
	cfg := lifecycle.NewConfiguration()
	cfg.Rules = []lifecycle.Rule{{
		ID:         "expire-pending-uploads",
		Status:     "Enabled",
		RuleFilter: lifecycle.Filter{Prefix: "pending/"},
		Expiration: lifecycle.Expiration{Days: lifecycle.ExpirationDays(expireDays)},
	}}
	// avatar (identity) hanya pakai bucket public; kesantrian/psb hanya privateBucket —
	// tiap adapter cukup panggil SetBucketLifecycle ke bucket yang benar-benar dipakainya.
	return u.client.SetBucketLifecycle(ctx, u.bucket, cfg)
}
```
Catatan: idempotent — dipanggil ulang di setiap startup aman, cukup
menimpa rule yang sama.

### 3b. Presign — generate key di bawah prefix `pending/`

- Avatar (`avatar.go`, `AvatarPresignUseCase.Execute`):
  `objectName := path.Join("pending", "avatars", uuid.NewString()+ext)`
- Kesantrian dokumen (`dokumen_upload.go`, `DokumenPresignUseCase.Execute`):
  `objectName := path.Join("pending", "santri", "dokumen", string(kind), uuid.NewString()+ext)`
- PSB dokumen (`dokumen_upload.go`, `DokumenPresignUseCase.Execute`):
  `objectName := fmt.Sprintf("pending/psb/dokumen/%s/%s/%s%s", req.Stage, req.Kind, uuid.NewString(), ext)`

Response presign (`Key`) tetap dikirim apa adanya ke client — client upload
lalu mengirim key yang sama itu ke endpoint confirm, **tidak ada perubahan
kontrak API ke frontend**.

### 3c. Confirm — validasi prefix, promote, baru simpan key FINAL ke DB

Pola yang sama di ketiga `*ConfirmUseCase.Execute`, urutan:
1. Normalisasi key seperti sekarang (mis. `KeyFromURL` di avatar).
2. **Validasi** `strings.HasPrefix(stagingKey, "pending/")` — kalau tidak,
   reject `ErrCodeBadRequest`/`ErrCodeUnprocessableEntity` (mencegah client
   mengirim key sembarangan yang bukan hasil presign miliknya sendiri).
3. `ConfirmUpload(ctx, stagingKey)` seperti sekarang — verifikasi upload
   benar-benar selesai (stat object).
4. `finalKey := strings.TrimPrefix(stagingKey, "pending/")`
5. `fileUploader.PromoteUpload(ctx, stagingKey, finalKey, privacy)` — kalau
   gagal, **return error, jangan lanjut simpan apa pun ke DB** (staging object
   tetap ada, nanti otomatis expire lewat lifecycle rule kalau memang tidak
   pernah berhasil dipromosikan).
6. Simpan **`finalKey`** (bukan `stagingKey`) ke DB — `user.AvatarKey`,
   `entity.NewSantriDokumen(..., finalKey)`, `entity.NewPendaftarDokumen(..., finalKey)`.

Contoh perubahan `AvatarConfirmUseCase.Execute` (avatar.go:83-124):
```go
normalizedKey := uc.fileUploader.KeyFromURL(key)
if normalizedKey == "" || !strings.HasPrefix(normalizedKey, "pending/") {
	return nil, kernel.New(application.ErrCodeUnprocessableEntity)
}

user, err := uc.userRepo.FindByID(ctx, userID)
...

if err := uc.fileUploader.ConfirmUpload(ctx, normalizedKey); err != nil {
	return nil, kernel.Wrap(application.ErrCodeInternal, err)
}

finalKey := strings.TrimPrefix(normalizedKey, "pending/")
if err := uc.fileUploader.PromoteUpload(ctx, normalizedKey, finalKey, ports.PrivacyPublic); err != nil {
	return nil, kernel.Wrap(application.ErrCodeInternal, err)
}

oldKey := user.AvatarKey
user.AvatarKey = &finalKey

if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
	return uc.userRepo.Update(txCtx, user)
}); err != nil {
	return nil, kernel.Wrap(application.ErrCodeInternal, err)
}

if oldKey != nil && *oldKey != finalKey {
	if err := uc.fileUploader.MarkDeleted(ctx, *oldKey); err != nil {
		slog.Warn("identity: best-effort hapus avatar lama gagal", "key", *oldKey, "error", err)
	}
}

return &dto.AvatarConfirmResponse{AvatarURL: uc.fileUploader.PublicURL(finalKey)}, nil
```
(`ConfirmUpload` dipindah ke atas sebelum promote — urutan lama memanggilnya
setelah `transactor.WithTx`, di desain baru harus sebelum promote karena
promote butuh object staging benar-benar sudah ada.)

PSB dan kesantrian confirm mengikuti urutan yang sama: promote dulu (dengan
`ports.PrivacyPrivate`), baru bangun entity pakai `finalKey`, baru `Save`.

### 3d. Setup lifecycle rule sekali di startup — `cmd/api/main.go`

Tidak perlu goroutine/ticker/scheduler apa pun. Cukup panggil sekali per modul
setelah modul dikonstruksi (idempotent, aman dipanggil di setiap boot):
```go
const pendingUploadExpireDays = 1 // cukup: promote terjadi detik/menit setelah upload; ini murni jaring pengaman

if err := identity.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
	lg.Warn("gagal set lifecycle pending upload (identity)", slog.Any("error", err))
}
if err := kesantrian.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
	lg.Warn("gagal set lifecycle pending upload (kesantrian)", slog.Any("error", err))
}
if err := psb.EnsurePendingUploadLifecycle(context.Background(), pendingUploadExpireDays); err != nil {
	lg.Warn("gagal set lifecycle pending upload (psb)", slog.Any("error", err))
}
```
Tiap `Module` expose method tipis yang meneruskan ke `fileUploader.EnsurePendingUploadLifecycle(ctx, expireDays)`
miliknya sendiri. Kegagalan di sini hanya di-log (warning), tidak menghentikan
startup — server tetap boleh jalan meski lifecycle rule gagal terpasang
(mis. MinIO belum siap), staging object saat itu cukup tidak ter-expire
otomatis sampai rule berhasil dipasang di boot berikutnya.

### Perbedaan per modul (pola sama persis, cuma nama beda)

| | Identity (avatar) | Kesantrian (santri dokumen) | PSB (pendaftar dokumen) |
|---|---|---|---|
| Bucket yang dipakai | `sipon-public` (selalu `PrivacyPublic`) | `sipon-private` | `sipon-private` |
| Prefix staging | `pending/avatars/...` | `pending/santri/dokumen/<kind>/...` | `pending/psb/dokumen/<stage>/<kind>/...` |
| Method delete lama (tetap ada) | `MarkDeleted(ctx, key)` | `DeleteObject(ctx, key, privacy)` | `DeleteObject(ctx, key, privacy)` |
| File presign use case | `identity/application/command/avatar.go` (`AvatarPresignUseCase`) | `kesantrian/application/command/dokumen_upload.go` (`DokumenPresignUseCase`) | `psb/application/command/dokumen_upload.go` (`DokumenPresignUseCase`) |
| File confirm use case | sama file, `AvatarConfirmUseCase` | sama file, `DokumenConfirmUseCase` | sama file, `DokumenConfirmUseCase` |
| Adapter yang diubah | `identity/infrastructure/external/minio_uploader.go` | `kesantrian/infrastructure/external/minio_uploader.go` | `psb/infrastructure/external/minio_uploader.go` |

## File yang Diubah/Ditambah — Ringkasan

**Bagian 1**: `internal/modules/psb/interfaces/http/router.go` (comment 1 baris)

**Bagian 2**: `internal/modules/identity/application/command/avatar.go`,
`internal/modules/kesantrian/application/command/dokumen_manage.go` (reorder +
error check), `internal/modules/psb/application/command/dokumen_upload.go`
(logging saja, sudah benar urutannya)

**Bagian 3** (×3 modul, pola sama, tanpa migrasi/tabel baru sama sekali):
`application/ports/storage.go` (tambah `PromoteUpload` + `EnsurePendingUploadLifecycle`
ke interface `FileUploader`), `infrastructure/external/minio_uploader.go`
(implementasi kedua method itu pakai `CopyObject`/`SetBucketLifecycle`),
presign use case (key digenerate di bawah prefix `pending/`), confirm use case
(validasi prefix → `ConfirmUpload` → `PromoteUpload` → simpan `finalKey`),
`module.go` (expose method `EnsurePendingUploadLifecycle` di `Module`), dan
`cmd/api/main.go` (panggil sekali di startup, tanpa goroutine/ticker).

## Verifikasi

1. `go build ./...` dan `go vet ./...` — pastikan semua interface/constructor
   yang berubah konsisten di seluruh pemanggil (khususnya `FileUploader` yang
   diimplementasikan ulang di 3 adapter).
2. Manual sanity check (kalau environment dev MinIO/Postgres tersedia):
   - Purge: pastikan `POST .../settings/:id/purge` sekarang 404.
   - Delete flow: hapus 1 avatar dan 1 santri dokumen, verifikasi lewat
     MinIO console/client bahwa objeknya benar-benar hilang.
   - Lifecycle rule: setelah startup, cek via `mc ilm ls` (atau MinIO console)
     bahwa rule `expire-pending-uploads` pada prefix `pending/` benar-benar
     terpasang di bucket publik & privat.
   - Promote-on-confirm: lakukan presign → upload → confirm seperti biasa,
     lalu cek di MinIO bahwa objek akhirnya ada di key FINAL (tanpa prefix
     `pending/`) dan staging object sudah tidak ada; cek juga kolom
     `avatar_key`/`key` di DB menyimpan `finalKey`, bukan key hasil presign.
   - Orphan tanpa confirm: minta presign tapi JANGAN confirm — objek staging
     harus tetap ada segera setelah upload (belum expire), dan (kalau mau
     diuji cepat) coba set `Expiration.Days` sangat kecil di lingkungan test
     untuk memverifikasi MinIO benar-benar membersihkannya otomatis setelah
     rentang waktu tsb berlalu.
