# Plan: Pembayaran Manual Santri (Upload Bukti Transfer + Verifikasi Admin) di `sipon-be` & `sipon-ui`

## Context

Santri (halaman `/dashboard`) belum bisa melakukan pembayaran sendiri. Saat ini hanya **admin** yang bisa membuat pembayaran manual (`POST /api/v1/web/keuangan/admin/payments/manual`) dan memverifikasi (`verify`/`reject`). Santri hanya bisa melihat tagihan & riwayat pembayaran.

Fitur ini menambahkan alur **self-service payment** untuk santri:
1. Santri mengisi form pembayaran pada sebuah invoice miliknya.
2. Santri **wajib** mengupload bukti transfer (disimpan di MinIO private bucket).
3. Pembayaran masuk status `pending`.
4. **Admin** melihat daftar pembayaran pending, memeriksa bukti transfer, lalu memverifikasi atau menolak (alur verify/reject sudah ada).

### Keputusan yang sudah dikonfirmasi

| No | Pertanyaan | Jawaban |
|----|-----------|---------|
| 1 | Santri bisa bayar sebagian (partial payment)? | **Ya** |
| 2 | Admin perlu notifikasi email? | **Tidak** (cek manual di halaman pembayaran) |
| 3 | Halaman riwayat pembayaran santri? | **Perlu** (enhance `/keuangan/riwayat` yang sudah ada) |
| 4 | Bukti transfer wajib? | **Ya** (`proof_key` required) |

## Temuan penting dari kode aktual

1. **Skema `payments` sudah siap** — tidak perlu migration baru. Kolom yang ada: `proof_key`, `status` (`pending|verified|rejected`), `verified_by`, `verified_at`, `reference_number`, `method` (`transfer|cash|check`), `amount`, `payment_date`, `notes`, `created_by`.
2. **Admin flow sudah lengkap**: `CreateManualPaymentUseCase` (`create_manual_payment.go`), `VerifyPaymentUseCase`, `RejectPaymentUseCase`, route admin di `interfaces/http/router.go` (baris 46-51).
3. **Belum ada** endpoint self-service: santri belum bisa presign upload bukti transfer maupun submit pembayaran.
4. **Presign pattern sudah ada di module lain** dan harus di-reuse:
   - `kesantrian`: `POST /santri/dokumen/presign` → `DokumenPresign` handler + `dokumen_upload.go` use case + `application/ports/storage.go` (`RequestUpload`).
   - `identity`: avatar presign.
   - `dokumen_aset`, `article`, `psb`: pola sama (MinIO private bucket untuk file sensitif).
5. **MinIO uploader BELUM ada** di keuangan module (tidak ada `minio_uploader.go`/`storage.go`). Perlu dibuat baru: `infrastructure/external/minio_uploader.go` + `application/ports/storage.go`, mengikuti pattern module `kesantrian` (private bucket `sipon-private`).
6. **Verifikasi debit account**: `verify_payment.go:104` memanggil `accountID(payment.DebitAccountID)`. **Keuangan Settings sudah diimplementasikan** (lihat `keuangan-settings.md`) — `default_payment_debit_account_id` akan dipakai sebagai `debit_account_id` pembayaran santri. Karena `payments.debit_account_id` **NOT NULL** (migration `20260809130000`), `SubmitPaymentUseCase` harus mengambil account default dari settings dan mengisinya saat membuat payment.
7. **UI santri sudah punya** halaman `/keuangan/tagihan/[id].vue` (detail invoice + riwayat pembayaran) dan `/keuangan/riwayat.vue`. Komponen upload pattern ada di `PsbDocumentUploader.vue` (presign → PUT file → simpan key).
8. **UI store `keuangan.ts`** sudah punya `fetchMyInvoices`, `fetchMyInvoice`, `fetchMyPayments`, `fetchMyInvoiceSummary` + action admin. Belum ada `submitPayment`/`requestPaymentProofPresign` untuk santri.
9. **Permission keys** yang relevan: `manage_keuangan` (list/detail), `verify_payment` (verify/reject). Endpoint self-service santri **tanpa** `RequirePermission` (cukup `jwtAuth + principalLoad`) — konsisten dengan `GET /invoices`, `GET /summary`.
10. **Otentikasi kepemilikan invoice**: invoice punya `santri_id` + `user_id`. Submit payment harus memvalidasi bahwa invoice milik santri yang sedang login (`middleware.GetUserID(c)` → lookup santri → validasi invoice).

## Backend — `sipon-be`

### A. Presign upload bukti transfer (self-service)

**Endpoint:** `POST /api/v1/web/keuangan/payments/proof/presign` (auth-only, tanpa permission)

**Request DTO** (`application/dto/payment_dto.go`):
```go
type PresignPaymentProofRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}
```

**Response DTO**:
```go
type PresignPaymentProofResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	ExpiresIn  int    `json:"expires_in"`
}
```

**Use case:** `application/command/create_payment_proof_presign.go`
- Validasi `ContentType` terhadap allowlist: `image/jpeg`, `image/png`, `image/webp`, `application/pdf` (pattern `kesantrian/domain/dokumen/constant/dokumen_constant.go`).
- Object name: `payment-proofs/{user_id}/{uuid}_{filename}`.
- Panggil `storage.RequestUpload(ctx, objectName, ct, 15*time.Minute, ports.PrivacyPrivate)`.
- TTL presign: **15 menit**.

**Storage port:** `application/ports/storage.go` (baru jika belum ada di keuangan) — ikuti pattern `kesantrian/application/ports/storage.go`:
```go
type PrivacyRule string

const (
	PrivacyPublic  PrivacyRule = "public"
	PrivacyPrivate PrivacyRule = "private"
)

type FileUploader interface {
	RequestUpload(ctx context.Context, objectName, contentType string, expiry time.Duration, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
	RequestDownload(ctx context.Context, key string, expiry time.Duration) (string, error)
}
```

**Infrastructure:** `infrastructure/external/minio_uploader.go` (baru, ikuti `kesantrian/infrastructure/external/minio_uploader.go`).

**Handler:** `PaymentProofPresign` di `interfaces/http/handler.go`.

**Wiring:** `module.go` — inject uploader + use case ke handler.

### B. Submit pembayaran santri (self-service)

**Endpoint:** `POST /api/v1/web/keuangan/payments` (auth-only, tanpa permission)

**Request DTO** (`application/dto/payment_dto.go`):
```go
type SubmitPaymentRequest struct {
	InvoiceID       string  `json:"invoice_id" binding:"required"`
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Method          string  `json:"method" binding:"required,oneof=transfer"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	PaymentDate     string  `json:"payment_date" binding:"required"`
	ProofKey        string  `json:"proof_key" binding:"required"` // WAJIB
	Notes           *string `json:"notes,omitempty"`
}
```

**Use case:** `application/command/submit_payment.go`
- Ambil `userID` dari context (bukan dari body).
- Validasi invoice milik user: invoice → `user_id`/`santri_id` harus cocok dengan user yang login.
- Validasi status invoice: hanya `issued` atau `partial` yang bisa dibayar.
- Validasi **partial payment**: `amount` ≤ `outstanding` (`amount - discount - paid`, floor 0).
- `DebitAccountID` **tidak diisi** (di-set admin saat verifikasi, atau di-null — konsisten dengan entity yang sudah ada; `debit_account_id` nullable).
- Method: `transfer`.
- Status awal: `pending`.
- `CreatedBy` = userID.
- Buat payment via repository `Create`.

**Catatan entity:** `NewPayment` di `domain/payment/entity/payment.go` sudah valid — `amount <= 0` ditolak, `debitAccountID` nullable.

### C. Download bukti transfer untuk admin

**Endpoint:** `GET /api/v1/web/keuangan/admin/payments/:id/proof` (permission `verify_payment`)

**Handler:** `GetPaymentProofURL` — ambil payment → ambil `proof_key` → `storage.RequestDownload(ctx, proofKey, 15*time.Minute)` → return `{ url, expires_in }`.

**Response DTO:**
```go
type PaymentProofResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
}
```

### D. Routing (`interfaces/http/router.go`)

Grup self-service `santri` (auth-only):
```go
santri.POST("/payments/proof/presign", h.PaymentProofPresign)
santri.POST("/payments", h.SubmitPayment)
```

Grup admin (permission `verify_payment`):
```go
admin.GET("/payments/:id/proof", middleware.RequirePermission("verify_payment"), h.GetPaymentProofURL)
```

### E. File yang dibuat/diubah di `sipon-be`

```
internal/modules/keuangan/
├── application/
│   ├── command/
│   │   ├── create_payment_proof_presign.go      (BARU)
│   │   └── submit_payment.go                    (BARU)
│   ├── dto/
│   │   └── payment_dto.go                       (UPDATE: +Presign/Submit/ProofResponse)
│   └── ports/
│       └── storage.go                           (BARU)
├── infrastructure/external/
│   └── minio_uploader.go                        (BARU)
├── interfaces/http/
│   ├── handler.go                               (UPDATE: +PaymentProofPresign, +SubmitPayment, +GetPaymentProofURL)
│   ├── router.go                                (UPDATE: +3 route)
│   └── module.go                                (UPDATE: wiring uploader + use cases)
```

## Frontend — `sipon-ui`

### A. Store (`app/stores/keuangan.ts`)

Tambah action:
```ts
async requestPaymentProofPresign(payload: PresignPaymentProofRequest): Promise<PresignPaymentProofResponse>
async submitPayment(payload: SubmitPaymentRequest): Promise<Payment>
async getPaymentProofURL(paymentId: string): Promise<PaymentProofResponse>  // admin
```

### B. Types (`shared/types/Keuangan.ts`)

```ts
export interface PresignPaymentProofRequest {
  filename: string
  content_type: string
}
export interface PresignPaymentProofResponse {
  presign_url: string
  key: string
  expires_in: number
}
export interface SubmitPaymentRequest {
  invoice_id: string
  amount: number
  method: PaymentMethod
  reference_number?: string
  payment_date: string
  proof_key: string       // wajib
  notes?: string
}
export interface PaymentProofResponse {
  url: string
  expires_in: number
}
```

### C. Komponen upload bukti transfer

**`app/components/keuangan/KeuanganPaymentProofUploader.vue`** (baru)
- Pattern: `PsbDocumentUploader.vue` (presign → PUT ke `presign_url` → simpan `key`).
- Validasi: max **5MB**, format **JPG/PNG/PDF**.
- Emit `update:modelValue` / `proof_key` setelah upload sukses.
- Menampilkan preview gambar/PDF + tombol ganti/hapus.

### D. Halaman pembayaran santri

**`app/pages/keuangan/tagihan/[id]/bayar.vue`** (baru, `layout: 'default'`)
- Fetch invoice (`fetchMyInvoice`), tampilkan ringkasan: total, diskon, dibayar, sisa.
- Form (Zod + `UForm`):
  ```ts
  const schema = z.object({
    amount: z.number().min(1, 'Nominal harus lebih dari 0'),
    reference_number: z.string().min(1, 'Nomor referensi wajib diisi'),
    payment_date: z.string().min(1, 'Tanggal pembayaran wajib diisi'),
    proof_key: z.string().min(1, 'Bukti transfer wajib diupload'),
    notes: z.string().optional(),
  })
  ```
- `amount` default = sisa tagihan; max = sisa tagihan (validasi client + server).
- Submit → `submitPayment()` → toast sukses → redirect ke `/keuangan/tagihan/[id]`.
- Tampilkan info "Pembayaran Anda sedang menunggu verifikasi admin."

### E. Tombol "Bayar" di detail tagihan

**`app/pages/keuangan/tagihan/[id].vue`** (update)
- Header: tambah `UButton` "Bayar Tagihan" jika `outstandingAmount > 0 && ['issued','partial'].includes(invoice.status)`.
- Link ke `/keuangan/tagihan/[id]/bayar`.

### F. Enhance riwayat pembayaran santri

**`app/pages/keuangan/riwayat.vue`** (update)
- Pastikan tiap baris menampilkan `KeuanganStatusBadge` status payment (`pending`/`verified`/`rejected`).
- Status `pending` → label "Menunggu Verifikasi" + warna warning.
- Status `rejected` → tampilkan alasan (notes) bila ada.
- Status `verified` → tampilkan tanggal verifikasi.

### G. Enhance admin payment (list & detail)

**`app/pages/admin/keuangan/pembayaran/index.vue`** (update)
- Tambah kolom "Bukti Transfer" dengan tombol "Lihat" (buka modal preview).
- Filter status `pending|verified|rejected` sudah ada — pastikan berfungsi.

**`app/pages/admin/keuangan/pembayaran/[id].vue`** (update)
- Tampilkan bukti transfer (preview gambar/PDF) via `getPaymentProofURL`.
- Info lengkap: pengaju (`created_by`), tanggal pengajuan (`created_at`), nomor referensi.
- Modal verify/reject yang ada sudah dipakai.

## Fase Pengerjaan (status: selesai)

**Fase 1 — Backend A & B** ✅: storage port + minio uploader + presign use case + submit payment use case + DTO. Checkpoint: `go build` + `go test ./internal/modules/keuangan/...` lolos.

**Fase 2 — Backend C & D** ✅: endpoint `GET /admin/payments/:id/proof` + routing + wiring `module.go`. Checkpoint: curl ketiga endpoint (presign, submit, download proof) berjalan.

**Fase 3 — UI Types & Store** ✅: `Keuangan.ts` + action store. Checkpoint: `npx nuxi typecheck` lolos.

**Fase 4 — UI Santri** ✅: `KeuanganPaymentProofUploader.vue`, halaman `/keuangan/tagihan/[id]/bayar.vue`, tombol "Bayar" di `[id].vue`. Checkpoint: santri bisa presign → upload → submit payment end-to-end.

**Fase 5 — UI Admin** ✅: aksi/preview bukti transfer di list & detail pembayaran admin + modal verifikasi mewajibkan akun debit. Checkpoint: admin bisa lihat bukti dan verify/reject.

**Fase 6 — Polish** ✅: enhance riwayat santri (pesan status pending/verified/rejected), validasi partial payment, edge cases.

## Verifikasi

1. `go build ./...` dan `go test ./internal/modules/keuangan/...` di `sipon-be` lolos.
2. `npm run dev` di `sipon-ui` tanpa error; type-check lolos.
3. **Alur santri**: login santri → `/dashboard` → tagihan → "Bayar Tagihan" → isi form + upload bukti (wajib, validasi 5MB/format) → submit → toast sukses → status `pending` muncul di detail & riwayat.
4. **Validasi kepemilikan**: santri tidak bisa submit payment untuk invoice milik santri lain (403/404).
5. **Validasi partial**: amount > sisa tagihan ditolak; amount sesuai bisa.
6. **Alur admin**: login admin → `/admin/keuangan/pembayaran` → filter `pending` → lihat bukti transfer (preview) → verify → invoice ter-update (`paid`/`partial`), jurnal otomatis ter-posting; atau reject → status `rejected`.
7. **Bukti transfer**: akses proof di MinIO private bucket hanya via presigned URL (expired 15 menit).

## Catatan tambahan (hasil implementasi)

- **Migration baru**: `20260811100000_payments_debit_account_nullable` — `payments.debit_account_id` dibuat **nullable**. Pembayaran santri dibuat status `pending` **tanpa** `debit_account_id`.
- **Debit account diisi saat verifikasi (bukan dari settings)**: sesuai keputusan, **keuangan settings TIDAK dipakai** untuk saat ini. `VerifyPaymentUseCase.Execute` sekarang menerima `debitAccountID`:
  - `debit_account_id` **wajib** (bad request jika kosong).
  - Akun divalidasi `postable` + aktif + `sub_type = cash_bank`.
  - `payment.DebitAccountID` di-set sebelum `payment.Verify()` → auto-posting jurnal memakai akun kas/bank yang benar.
- **Endpoint baru**:
  - `POST /api/v1/web/keuangan/payments/proof/presign` (auth-only) — presign upload bukti transfer (MinIO private bucket, TTL 15 menit, format jpg/png/webp/pdf).
  - `POST /api/v1/web/keuangan/payments` (auth-only) — santri submit pembayaran (validasi kepemilikan invoice, status `issued`/`partial`, `amount ≤ outstanding`, `proof_key` wajib).
  - `GET /api/v1/web/keuangan/admin/payments/:id/proof` (permission `verify_payment`) — presigned download URL bukti transfer.
- **`POST /admin/payments/:id/verify`** sekarang menerima body `{ "debit_account_id": "uuid" }`.
- **Jurnal otomatis**: tidak diubah — `VerifyPaymentUseCase` tetap posting jurnal; kini `debit_account_id` selalu terisi saat verify sehingga jurnal debit kas/bank benar.
- **Frontend**: modal verifikasi admin kini mewajibkan pilih akun kas/bank (`KeuanganAccountPicker` filter asset+cash_bank); halaman list/detail pembayaran admin punya aksi "Lihat Bukti Transfer"; halaman `/keuangan/tagihan/[id]/bayar.vue` + tombol "Bayar Tagihan" + `KeuanganPaymentProofUploader.vue`; riwayat santri menampilkan pesan status pending/verified/rejected.
- **Keamanan**: `proof_key` tidak di-trust sebagai path arbitrer — hanya dipakai untuk presign download; santri hanya bisa mem-preview bukti miliknya sendiri (jika perlu).
- **Rate limit**: pertimbangkan reuse middleware rate-limit yang ada untuk endpoint submit/presign.
