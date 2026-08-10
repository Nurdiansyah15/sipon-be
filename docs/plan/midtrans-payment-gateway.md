# Plan: Payment Online via Midtrans Snap (Module Keuangan)

## Context

Modul `keuangan` sudah memiliki invoice/tagihan, pembayaran manual, jurnal akuntansi, dan report. Belum ada mekanisme pembayaran **online** — santri hanya bisa membayar lewat transfer/cash/check yang diverifikasi admin secara manual.

Tujuan: mengintegrasikan **Midtrans Snap** sehingga santri bisa membayar invoice secara online. Alur:

1. Tagihan (invoice) dibuat oleh admin (`issued`/`partial`).
2. Santri menekan "Bayar Online" → backend membuat transaksi Snap → muncul popup Snap (atau redirect ke halaman pembayaran).
3. Santri menyelesaikan pembayaran di Midtrans (semua metode yang diaktifkan di dashboard Midtrans).
4. Midtrans mengirim **webhook** ke backend → status invoice & payment berubah otomatis.

> **Status: SUDAH DIIMPLEMENTASI** (commit plan ini). Dokumen ini dokumentasi hidup.

## Keputusan Arsitektur

1. **Bounded context baru `paymentgateway`** di dalam modul `keuangan` — transaksi gateway dipisah dari entity `Payment` internal. Ini memungkinkan:
   - Status Midtrans (`pending`, `capture`, `settlement`, `deny`, dll.) tidak tercampur dengan status internal (`pending`/`verified`/`rejected`).
   - `Payment` (yang memicu jurnal akuntansi) baru dibuat **setelah** ada settlement dari webhook, lalu langsung diverifikasi otomatis.
2. **Env mode `sandbox`/`production`** — `MIDTRANS_ENV` menentukan base URL Snap & Core API. Base URL bisa di-override manual (`MIDTRANS_SNAP_BASE_URL`, `MIDTRANS_API_BASE_URL`).
3. **Webhook public + verifikasi signature** — endpoint webhook tidak memakai JWT; keamanan dijamin dengan verifikasi `signature_key` sesuai standar Midtrans: `SHA512(server_key + order_id + status_code + gross_amount)`. Invalid → `401`.
4. **Idempotent** — webhook bisa datang berulang (retry Midtrans). Status sukses tidak pernah di-regresi, status final (`deny`/`failure`/`expire`/`cancel`) tidak berubah. Transaksi yang sudah ter-link ke payment tidak akan membuat payment ganda.
5. **Semua metode pembayaran Snap** — item & gross amount dikirim, sisanya diserahkan ke dashboard Midtrans (tidak ada filter metode di backend).
6. **Akun settlement via config** — `MIDTRANS_SETTLEMENT_ACCOUNT_ID` (akun kas/bank di COA) dipakai sebagai akun debit jurnal `payment_verified`. Jika kosong, payment dibuat `pending` dan menunggu verifikasi manual admin (fallback aman, tidak error jurnal).
7. **Owner check** — santri hanya bisa membayar invoice miliknya sendiri (`invoice.user_id == user id` dari JWT).

## Aktor

| Aktor | Peran |
|---|---|
| **Santri** | Membayar invoice online via Snap, melihat status pembayaran online (polling setelah Snap ditutup). |
| **Sistem (webhook)** | Menerima notifikasi Midtrans, memperbarui status transaksi gateway, membuat + memverifikasi payment, mem-posting jurnal. |
| **Admin** | Tidak berubah — tetap bisa verifikasi/reject manual lewat endpoint existing (fallback bila akun settlement kosong). |

## Skema Data (migration `20260810140000_create_payment_gateway_transactions`)

### `payment_gateway_transactions`
`id UUID PK`, `transaction_id VARCHAR(50) UNIQUE` (order_id Midtrans, format `SIPON-YYYYMMDDHHMMSS-<8char>`), `invoice_id UUID FK invoices`, `payment_id UUID FK payments` (nullable, terisi setelah settlement), `amount NUMERIC(14,2)`, `status VARCHAR(30)` CHECK (`pending`|`pending_challenge`|`capture`|`settlement`|`deny`|`failure`|`expire`|`cancel`), `payment_method VARCHAR(50)`, `snap_token TEXT`, `redirect_url TEXT`, `raw_notification JSONB`, `metadata JSONB`, `expired_at TIMESTAMPTZ`, timestamps.

Index: unique `transaction_id`, index `invoice_id`, `status`, `payment_id`, dan partial index transaksi aktif per invoice.

### Pemetaan status Midtrans → internal

| Midtrans `transaction_status` (+ `fraud_status`) | Internal | Dampak |
|---|---|---|
| `capture` + `accept` / no fraud | `settlement` | Sukses → buat + verifikasi payment |
| `capture` + `challenge` | `pending_challenge` | Menunggu approve/deny |
| `settlement` | `settlement` | Sukses → buat + verifikasi payment |
| `pending` | `pending` | Menunggu pembayaran |
| `deny` / `cancel` / `expire` / `failure` | `deny` / `cancel` / `expire` / `failure` | Gagal → tidak buat payment |
| tidak dikenal | `pending` | Fallback aman |

## Business Rules

1. Hanya invoice `issued`/`partial` yang bisa dibayar online. Invoice `paid`/`expired`/`cancelled`/`draft` → `409`.
2. Hanya pemilik invoice yang bisa memulai pembayaran online (`invoice.user_id != user id` → `403`).
3. **Idempotensi create**: bila sudah ada transaksi gateway aktif (`pending`/`pending_challenge`) untuk invoice itu, kembalikan snap token yang sudah ada (jangan buat baru). Bila sudah `settlement`/`capture` → `409` "sudah dibayar".
4. Nominal transaksi = `invoice.RemainingAmount()` saat itu.
5. Webhook sukses (`settlement`/`capture`) membuat `Payment` baru:
   - `method = transfer`, `reference_number = transaction_id`, `created_by = systemActor` (`00000000-0000-0000-0000-000000000000`, UUID sentinel untuk aksi sistem).
   - Bila `MIDTRANS_SETTLEMENT_ACCOUNT_ID` terisi → payment langsung `verify` → `invoice.AddPayment` → `autoPosting.PostPaymentVerified` (semua dalam satu DB transaction).
   - Bila kosong → payment tetap `pending` (verifikasi manual admin via endpoint existing `/payments/:id/verify`).
6. Webhook gagal/non-sukses → hanya memperbarui status transaksi gateway; tidak membuat payment.
7. `Payment` tidak dibuat dua kali untuk transaksi yang sama (guard `gatewayTx.PaymentID != nil`).
8. Webhook untuk `order_id` yang tidak dikenal → respons sukses tanpa aksi (mencegah retry tak berujung dari Midtrans).

## Konfigurasi (`.env`)

```bash
MIDTRANS_ENV=sandbox              # sandbox | production
MIDTRANS_SERVER_KEY=SB-Mid-server-xxxx
MIDTRANS_CLIENT_KEY=SB-Mid-client-xxxx
MIDTRANS_SNAP_BASE_URL=           # kosong = otomatis dari MIDTRANS_ENV
MIDTRANS_API_BASE_URL=            # kosong = otomatis dari MIDTRANS_ENV
MIDTRANS_SNAP_EXPIRY_MINUTES=1440 # 24 jam default
MIDTRANS_SETTLEMENT_ACCOUNT_ID=   # ID akun kas/bank COA utk jurnal otomatis
```

Base URL otomatis:
| Environment | Snap | Core API |
|---|---|---|
| `sandbox` | `https://app.sandbox.midtrans.com/snap/v1` | `https://api.sandbox.midtrans.com/v2` |
| `production` | `https://app.midtrans.com/snap/v1` | `https://api.midtrans.com/v2` |

> **Catatan production**: wajib mengganti `MIDTRANS_SERVER_KEY`/`MIDTRANS_CLIENT_KEY` dengan key production dari dashboard Midtrans, dan mendaftarkan URL webhook `https://<domain>/api/v1/web/keuangan/webhooks/midtrans` di dashboard Midtrans.

## API Endpoints

**Self-service santri (JWT):**
- `POST /api/v1/web/keuangan/payments/midtrans` — body `{ "invoice_id": "<uuid>" }` → `{ transaction_id, invoice_id, amount, snap_token, redirect_url, status, expires_at }`
- `GET /api/v1/web/keuangan/invoices/:id/payment-status` — status transaksi gateway terbaru untuk invoice tsb (dipakai frontend polling setelah Snap ditutup).

**Webhook Midtrans (public, tanpa JWT):**
- `POST /api/v1/web/keuangan/webhooks/midtrans` — body notifikasi Midtrans; validasi signature; `200` sukses, `401` signature invalid.

## Struktur Module

```
internal/modules/keuangan/
├── domain/paymentgateway/
│   ├── constant/payment_gateway_constant.go      # status + error codes
│   ├── entity/payment_gateway_transaction.go     # entity + ApplyNotification (idempotent)
│   ├── valueobject/transaction_id.go             # order_id format SIPON-<ts>-<8>
│   └── repository/interfaces.go                  # Save/Update/FindBy* 
├── application/
│   ├── command/create_midtrans_payment.go        # buat transaksi Snap
│   ├── command/process_midtrans_webhook.go       # verifikasi + settle payment
│   ├── query/get_payment_gateway_status.go       # status utk polling frontend
│   ├── dto/payment_gateway_dto.go
│   └── ports/midtrans_gateway.go                 # outbound port Snap API
├── infrastructure/
│   ├── external/midtrans_gateway.go              # HTTP client + VerifySignature
│   └── persistence/postgres_payment_gateway_repo.go
└── interfaces/http/handler.go & router.go        # endpoints + webhook
```

### Ketergantungan (dependency injection)

```
NewModule(db, cfg, kesantrian, jwtAuth, principalLoad)
  ├─ paymentGatewayRepo := persistence.NewPostgresPaymentGatewayRepository(db)
  ├─ midtransGateway    := external.NewMidtransGateway(cfg.Midtrans)
  ├─ createMidtransPaymentUC  (paymentGatewayRepo, invoiceRepo, feeRepo, midtransGateway, expiry)
  ├─ processMidtransWebhookUC (paymentGatewayRepo, paymentRepo, invoiceRepo, feeRepo,
  │                             midtransGateway, transactor, autoPostingService, settlementAccountID)
  └─ getPaymentGatewayStatusUC (paymentGatewayRepo, invoiceRepo)
```

## Alur End-to-End

```
1. Admin buat invoice (issued)
2. Santri: POST /payments/midtrans
   └─ cek owner + status invoice -> existing transaksi aktif? -> call Midtrans Snap API
      -> simpan payment_gateway_transactions (pending, simpan snap_token)
      -> return snap_token + redirect_url
3. Frontend buka Snap popup / redirect (pakai client key + snap_token)
4. Santri pilih metode & bayar di Midtrans
5. Midtrans kirim webhook POST /webhooks/midtrans
   └─ verify signature (SHA512 server_key+order_id+status_code+gross_amount)
      -> cari transaction by order_id -> ApplyNotification (idempotent)
      -> settlement? buat Payment -> verify -> invoice.AddPayment -> post jurnal
6. Frontend polling GET /invoices/:id/payment-status -> tampilkan hasil
```

## Verifikasi

1. `go build ./...`, `go vet ./...`, `go test ./internal/modules/keuangan/...` lolos.
2. Unit test baru:
   - `domain/paymentgateway/entity/payment_gateway_transaction_test.go` — factory, idempotensi (tidak regresi setelah sukses, status final sticky), invalid status, link payment sekali, mark rejected.
   - `application/command/process_midtrans_webhook_test.go` — pemetaan status Midtrans.
   - `infrastructure/external/midtrans_gateway_test.go` — verifikasi signature (valid, tampered, empty key).
3. Migration `up`/`down` bersih (nama `20260810140000_create_payment_gateway_transactions`).
4. Smoke test di sandbox: buat snap token → simulasikan notifikasi settlement via webhook (signature valid) → payment `verified`, invoice `paid`, jurnal `payment_verified` muncul → notifikasi duplikat tidak membuat payment ganda.
