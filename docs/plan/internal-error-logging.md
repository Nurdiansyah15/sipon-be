# Fix: Error 500 Tidak Ke-log Detailnya di Server

## Context

Saat debug bug "error internal" di submit formulir PSB (root cause: mismatch
jumlah placeholder SQL di `INSERT INTO pendaftar`, lihat commit fix-nya di
`postgres_pendaftar_repo.go`), ketauan ada celah observability yang bikin
debugging itu jauh lebih lama dari seharusnya: **error asli di balik response
500 nggak pernah ke-log di mana pun di server**, cuma keliatan generic message
("internal server error" / kode error) di response client.

Alur yang sudah ada sekarang:

- `httperror.Handle` (`internal/shared/httperror/http_error.go:57-58`) SUDAH
  menyimpan error asli ke Gin context tiap kali status yang di-mapping >= 500:
  `c.Set("internal_err", err)`.
- Tapi **tidak ada satu pun kode di repo yang membaca kembali key
  `internal_err` itu** (sudah dicek lewat grep menyeluruh). Jadi nilainya
  ke-set, lalu dibuang begitu saja.
- `RequestLogger` (`internal/shared/middleware/request_logger.go`) memang
  sudah log satu baris per request (method, path, status, duration,
  request_id, client_ip) lewat `slog.LogAttrs`, tapi pesannya statis
  (`"request completed"`) dan tidak pernah menyertakan error-nya.

Akibatnya: satu-satunya cara nemuin root cause bug kemarin adalah baca kode
manual / nebak, bukan baca log server — padahal error Go asli (mis. pesan
driver Postgres) sebenarnya sudah ada di tangan (`internal_err`), tinggal
dibaca dan ditulis ke log.

Urutan middleware di `cmd/api/main.go` (`RequestID` → `CORS` →
`RequestLogger` → `ErrorHandler`/`httperror.Middleware`) sudah otomatis benar
untuk ini: karena Gin menjalankan kode setelah `c.Next()` dengan urutan LIFO,
`RequestLogger` masuk ke `ErrorHandler` dulu (yang men-set `internal_err`
lewat `Handle`), baru balik lagi ke kode setelah `c.Next()` milik
`RequestLogger`. Jadi pada saat `RequestLogger` mau nge-log, `internal_err`
sudah pasti ke-set duluan — tidak perlu ubah urutan middleware.

## Rencana Perbaikan

**File:** `internal/shared/middleware/request_logger.go`

Tambahkan pembacaan `internal_err` dari context tepat sebelum baris
`logger.LogAttrs(...)` yang sudah ada, hanya untuk kasus `status >= 500`:

```go
attrs := []slog.Attr{
    slog.String("method", method),
    slog.String("path", path),
    slog.Int("status", status),
    slog.Duration("duration", duration),
    slog.String("request_id", rid),
    slog.String("client_ip", c.ClientIP()),
}

if status >= 500 {
    if v, ok := c.Get("internal_err"); ok {
        if err, ok := v.(error); ok {
            attrs = append(attrs, slog.String("error", err.Error()))
        }
    }
}

switch {
case status >= 500:
    logger.LogAttrs(c.Request.Context(), slog.LevelError, "request completed", attrs...)
...
```

Poin penting:

- **Tidak menyentuh** `httperror.Handle`/`Middleware` sama sekali — mekanisme
  `c.Set("internal_err", err)` sudah benar, cuma perlu dibaca.
- **Tidak ada perubahan pada response HTTP ke client** — `respond.Error` tetap
  cuma pakai `httpErr.message`/`errorCode`, jadi error mentah (yang mungkin
  berisi detail internal seperti nama kolom/tabel) tetap tidak pernah bocor ke
  luar, hanya masuk ke log server.
- Pakai `err.Error()` (bukan `%+v` atau sejenisnya) supaya konsisten dengan
  gaya logging yang sudah dipakai di tempat lain di repo (mis.
  `slog.Warn(msg, "error", err, ...)` di `upsert_formulir.go`,
  `middleware/auth.go`).

## Verifikasi

1. `go build ./...` dan `go vet ./...` harus tetap bersih.
2. Restart dev server (`go run ./cmd/api`), lalu picu satu error 500 asli
   (mis. bug SQL yang sengaja direproduksi sementara, atau matikan koneksi DB
   sesaat) dan pastikan baris log untuk request itu sekarang punya field
   `error` berisi pesan error Go yang sebenarnya, berdampingan dengan
   `request_id` yang sama seperti yang dikembalikan ke client — supaya kalau
   user lapor bug lagi, tinggal cocokkan `request_id` dari response ke baris
   log server.
3. Pastikan body response ke client TIDAK berubah (masih cuma error code +
   message generic seperti sebelumnya) — cukup diff manual response sebelum
   vs sesudah untuk skenario error yang sama.

## Status

**Sudah diimplementasikan** di `internal/shared/middleware/request_logger.go:33-38`.

Error 500 sekarang otomatis ke-log dengan detail error Go asli di server, berdampingan dengan `request_id` yang sama seperti yang dikembalikan ke client. Contoh output log:

```
time=2026-08-05T06:18:57.362Z level=ERROR msg="request completed" method=POST path=/api/v1/web/keuangan/admin/components status=500 duration=13.594255ms request_id=1275211b-5525-48e7-aef5-84776ccae6f4 client_ip=172.22.0.1 error="code: ERR_INTERNAL, message: , err: code: FEE_COMPONENT_QUERY_FAILED, message: , err: exists by code: ERROR: invalid input syntax for type uuid: \"\" (SQLSTATE 22P02)"
```

Perbaikan status code duplicate/conflict (pendaftar/santri yang salah ke-mapping jadi 500 padahal harusnya 409) dikerjakan terpisah, langsung sebagai fix (lihat commit terkait), tidak lewat dokumen plan ini.
