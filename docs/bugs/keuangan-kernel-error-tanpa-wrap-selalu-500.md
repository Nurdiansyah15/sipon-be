# Bug: Error domain yang tidak dibungkus `WrapRepoErr`/`WrapConflictErr` selalu jadi HTTP 500

**Severity**: Tinggi — user-facing. Validasi bisnis yang seharusnya 400/409 tampil sebagai "Internal Server Error" generik. Ditemukan saat menyusun rencana penyempurnaan akuntansi karena rencana itu akan menambah beberapa error path baru yang rawan jatuh ke perangkap yang sama.

> **Status: Diperbaiki** — semua use-case keuangan kini memetakan error ke app error (`kernel.WrapMsg(application.ErrCode*, ...)`) secara inline per titik pemanggilan, sesuai pola identity module. Tidak ada lagi `kernel.New(domainCode)` telanjang, `WrapRepoErr`/`WrapConflictErr`, atau `WrapRepoErr(fmt.Errorf(...))` di `internal/modules/keuangan/application/`.

## Cara kerja pemetaan status HTTP (`internal/shared/httperror/http_error.go`)

```go
func statusFromCode(code string) int {
	switch code {
	case "ERR_BAD_REQUEST": return http.StatusBadRequest
	case "ERR_UNAUTHORIZED": return http.StatusUnauthorized
	case "ERR_FORBIDDEN": return http.StatusForbidden
	case "ERR_NOT_FOUND": return http.StatusNotFound
	case "ERR_CONFLICT": return http.StatusConflict
	case "ERR_GONE": return http.StatusGone
	case "ERR_TOO_MANY_REQUESTS": return http.StatusTooManyRequests
	case "ERR_UNPROCESSABLE_ENTITY": return http.StatusUnprocessableEntity
	default: return http.StatusInternalServerError  // <-- semua kode lain jatuh ke sini
	}
}
```

`statusFromCode` **hanya mengenali 8 kode generik** (`internal/modules/keuangan/application/errors.go`: `ErrCodeBadRequest`, `ErrCodeUnauthorized`, `ErrCodeForbidden`, `ErrCodeNotFound`, `ErrCodeConflict`, `ErrCodeUnprocessableEntity`, `ErrCodeInternal`, + `ERR_GONE`/`ERR_TOO_MANY_REQUESTS` dari modul lain). Kode domain spesifik (`PERIOD_OVERLAP`, `ACCOUNT_DUPLICATE`, `INVOICE_INVALID_STATUS`, dst) **tidak ada di daftar itu** — kalau sampai ke `httperror.Handle` tanpa diterjemahkan dulu lewat `application.WrapRepoErr`/`WrapConflictErr`, otomatis jadi 500.

## Lokasi yang sudah kena (dikonfirmasi lewat `grep -rn "return nil, kernel.New(\|WrapRepoErr(fmt.Errorf" internal/modules/keuangan/application`)

| File:Baris | Kode domain yang dikembalikan mentah | Seharusnya |
|---|---|---|
| `create_manual_journal.go:41` | `application.WrapRepoErr(fmt.Errorf("period is not open"), journalConst.CodeJournalPeriodClosed)` — `fmt.Errorf` bukan `*kernel.AppError`, jadi `errors.As` di dalam `WrapRepoErr` **tidak pernah match**, selalu jatuh ke `ErrCodeInternal` | 409 Conflict |
| `create_invoice_batch.go:83` | `kernel.New(invConst.CodeInvoiceInvalidStatus)` (skema tidak aktif) | 409 Conflict |
| `create_invoice_batch.go:91` | `kernel.New(bpConst.CodeBillingPeriodInvalidStatus)` | 409 Conflict |
| `create_invoice.go:63` | `kernel.New(feeConst.CodeFeeComponentNotFound)` (komponen tidak aktif) | 404 Not Found |
| `create_invoice.go:71` | `kernel.New(bpConst.CodeBillingPeriodInvalidStatus)` | 409 Conflict |
| `create_period.go:40` | `kernel.New(periodConst.CodePeriodOverlap)` | 409 Conflict |
| `create_fee_component.go:30` | `kernel.New(feeConst.CodeFeeComponentDuplicate)` | 409 Conflict |
| `create_fee_component.go:35` | `kernel.New(feeConst.CodeFeeComponentInvalidType)` | 400 Bad Request |
| `create_account.go:30` | `kernel.New(accConst.CodeAccountDuplicate)` | 409 Conflict |

## Gejala

Memanggil endpoint terkait dengan input yang secara bisnis salah (bukan bug, mis. membuat periode akuntansi dengan tanggal yang overlap periode lain, atau membuat komponen biaya dengan kode yang sudah dipakai) mengembalikan `500 Internal Server Error` alih-alih status 4xx yang informatif. Di frontend ini biasanya tampil sebagai toast generik "Terjadi kesalahan pada server" alih-alih pesan yang bisa ditindaklanjuti user, dan mencemari log error 500 dengan kejadian yang sebenarnya bukan bug.

## Akar Masalah

`kernel.New(domainCode)` menghasilkan `*kernel.AppError` dengan `Code` = kode domain aslinya (mis. `"PERIOD_OVERLAP"`), bukan salah satu dari 8 kode generik yang dikenali `statusFromCode`. Helper `application.WrapRepoErr`/`WrapConflictErr` ada justru untuk menerjemahkan kode domain (dari error yang datang dari layer lain, biasanya repository) menjadi salah satu dari 8 kode generik itu — tapi di lokasi-lokasi di atas, use-case mengembalikan `kernel.New(domainCode)` langsung tanpa lewat penerjemah itu.

## Dampak

Berpotensi mempengaruhi endpoint lain di luar daftar di atas juga (daftar ini hanya lingkup file yang tersentuh riset penyempurnaan akuntansi) — pola ini kemungkinan berulang di modul lain. Untuk dokumen ini, cukup dicatat lingkup `keuangan` dulu.

## Saran Perbaikan

**Untuk lokasi yang sudah ada** (opsional, di luar lingkup wajib rencana penyempurnaan akuntansi — boleh dikerjakan sekalian kalau menyentuh file yang sama):

```go
// Sebelum:
return nil, kernel.New(periodConst.CodePeriodOverlap)
// Sesudah (409 Conflict, bukan 500):
return nil, application.WrapConflictErr(kernel.New(periodConst.CodePeriodOverlap), periodConst.CodePeriodOverlap)
```

Untuk semantik "not found" (mis. `create_invoice.go:63`), pola yang sama tapi pakai `WrapRepoErr`:

```go
return nil, application.WrapRepoErr(kernel.New(feeConst.CodeFeeComponentNotFound), feeConst.CodeFeeComponentNotFound)
```

**Untuk kode BARU yang ditambahkan sebagai bagian dari rencana penyempurnaan akuntansi** (`docs/rules/akuntansi.md`, `docs/specs/keuangan-akuntansi-api.md`): **wajib** pakai pola di atas sejak awal — jangan `return nil, kernel.New(kodeBaru)` telanjang, dan jangan `WrapRepoErr`/`WrapConflictErr` dengan `fmt.Errorf(...)` sebagai argumen pertama (itu justru bug yang sama seperti `create_manual_journal.go:41`, karena `fmt.Errorf` bukan `*kernel.AppError` sehingga tidak pernah cocok). Argumen pertama ke `WrapRepoErr`/`WrapConflictErr` harus selalu `*kernel.AppError` asli (dari `kernel.New(...)`, `kernel.Wrap(...)`, atau error yang memang berasal dari domain/repo layer), dengan kode yang **sama persis** dengan argumen kedua.
