# API Spec: Notification (FCM)

Base path: `/api/v1/web`. Endpoint ini merupakan route uji coba push notification via Firebase Cloud Messaging (FCM), dipasang di branch `feat/notif` dan belum ada di `main`.

Dokumen ini memuat **(A)** endpoint yang baru ditambahkan dan **(B)** perubahan konfigurasi serta aturan pengiriman notifikasi yang terkait. Informasi teknis dan wiring implementasi berada di [`internal/modules/notification/service.go`](../../internal/modules/notification/service.go) serta bootstrap di [`cmd/api/main.go`](../../cmd/api/main.go).

---

## A. Endpoint yang Baru Ditambahkan

### Notification — Test
| Method | Path | Deskripsi |
|---|---|---|
| POST | `/notifications/test` | Mengirim notifikasi uji coba ke topic/default topic atau device token tertentu |

### Aturan dasar
- Endpoint hanya dipasang ketika `notification.NewService(cfg.Firebase)` berhasil dibuat.
- Kalau Firebase tidak tersedia atau config invalid, route tidak dipasang dan server tetap berjalan dengan `warn` log.
- Endpoint ini berfungsi sebagai route debug/test, bukan endpoint produksi yang dibatasi permission.

---

## B. Request & Response

### B.1 `POST /notifications/test`

#### Request body
```json
{
  "topic": "sipon_test",
  "title": "Sipon Notification",
  "body": "Tes notifikasi dari Sipon",
  "data": {
    "type": "test",
    "route": "/dashboard"
  }
}
```

| Field | Type | Wajib | Deskripsi |
|---|---|---:|---|
| `topic` | string | Tidak | Jika kosong, pakai `FIREBASE_DEFAULT_TOPIC` |
| `title` | string | Tidak | Default: `Sipon Notification` |
| `body` | string | Tidak | Default: `Tes notifikasi dari Sipon` |
| `data` | object | Tidak | Payload tambahan untuk notifikasi, default `{ "type": "test", "route": "/dashboard" }` |

#### Success response
```json
{
  "success": true,
  "message": "notification sent",
  "topic": "sipon_test"
}
```

Status: `200 OK`

#### Error response: invalid payload
```json
{
  "success": false,
  "message": "invalid payload",
  "error": "...error detail..."
}
```

Status: `400 Bad Request`

#### Error response: send failed
```json
{
  "success": false,
  "message": "failed to send FCM notification",
  "error": "...error detail..."
}
```

Status: `503 Service Unavailable`

---

## C. Aturan Notifikasi

### C.1 Config Firebase
Konfigurasi dibaca dari environment dan dimapping di [`internal/shared/config/config.go`](../../internal/shared/config/config.go):

```env
FIREBASE_ENABLED=true
FIREBASE_PROJECT_ID=
FIREBASE_DEFAULT_TOPIC=sipon_test
FIREBASE_SERVICE_ACCOUNT_PATH=
FIREBASE_SERVICE_ACCOUNT_JSON=
```

| Field | Deskripsi |
|---|---|
| `FIREBASE_ENABLED` | Menyalakan atau mematikan integrasi FCM |
| `FIREBASE_PROJECT_ID` | Project ID Firebase |
| `FIREBASE_DEFAULT_TOPIC` | Topic default bila request tidak menyebut `topic` |
| `FIREBASE_SERVICE_ACCOUNT_PATH` | Path file service account JSON |
| `FIREBASE_SERVICE_ACCOUNT_JSON` | JSON service account langsung, jika pakai inline credentials |

### C.2 Prioritas pengiriman
Saat build message, service mengirim payload dengan:
- `Notification.Title`
- `Notification.Body`
- `Data` map custom
- Android priority: `high`
- APNS priority header: `10`
- sound default: `default`

### C.3 Behavior fallback
- `topic` kosong → pakai `cfg.Firebase.DefaultTopic`
- `title` kosong → `Sipon Notification`
- `body` kosong → `Anda menerima notifikasi baru dari Sipon`
- `data` kosong → map default `{"type":"test","route":"/dashboard"}`

---

## D. Struktur Service

Implementasi utama ada di [`internal/modules/notification/service.go`](../../internal/modules/notification/service.go):

```go
type SendRequest struct {
    Token string
    Topic string
    Title string
    Body  string
    Data  map[string]string
}
```

Fungsi utama:
- `NewService(cfg config.FirebaseConfig)` → membuat client Firebase Messaging
- `Send(ctx context.Context, req SendRequest)` → kirim pesan ke token atau topic
- `buildFCMMessage(...)` → membangun payload FCM sesuai format default

---

## E. Coverage & Validasi

Ada test sederhana untuk validasi payload di [`internal/modules/notification/service_test.go`](../../internal/modules/notification/service_test.go). Test yang sekarang ada menilai:
- token benar masuk ke payload
- title/body terpasang
- data map dikirim dengan benar
- Android priority bernilai `high`

---

## F. File Terkait

- [`cmd/api/main.go`](../../cmd/api/main.go)
- [`internal/shared/config/config.go`](../../internal/shared/config/config.go)
- [`internal/modules/notification/service.go`](../../internal/modules/notification/service.go)
- [`internal/modules/notification/service_test.go`](../../internal/modules/notification/service_test.go)
- [`.env.example`](../../.env.example)
- [`go.mod`](../../go.mod)

---

## G. Catatan Implementasi

Endpoint ini saat ini bersifat test-oriented. Belum ada layer auth/permission khusus untuk route `/notifications/test`, jadi cocok dipakai untuk debugging atau uji coba lokal sebelum diintegrasikan ke flow user-facing yang lebih formal di masa depan.
