# Bug: Cancel invoice tidak menolak invoice berstatus `partial`

**Severity**: Sedang — celah integritas data, belum tentu sering terpicu tapi konsekuensinya nyata kalau terjadi.

## Lokasi

`internal/modules/keuangan/domain/invoice/entity/invoice.go:92-99`

```go
func (i *Invoice) Cancel() error {
	if i.Status == constant.StatusPaid || i.Status == constant.StatusCancelled {
		return kernel.New(constant.CodeInvoiceInvalidStatus)
	}
	i.Status = constant.StatusCancelled
	i.UpdatedAt = time.Now()
	return nil
}
```

## Gejala

Invoice yang sudah menerima **sebagian** pembayaran terverifikasi (`status = partial`, `paid_amount > 0`) masih bisa dibatalkan lewat `POST /admin/invoices/:id/cancel` — hanya status `paid` (lunas penuh) yang ditolak.

## Akar Masalah

Pengecekan hanya membandingkan `Status`, bukan `PaidAmount`. Status `partial` lolos dari kedua kondisi penolakan.

## Dampak

Kalau invoice yang sudah `partial` dibatalkan, uang yang sudah masuk kas (sudah ada payment `verified` yang menunjuk ke invoice ini) jadi tidak jelas statusnya — invoice sumbernya sudah `cancelled` tapi kas & piutang yang sudah terlanjur tercatat (setelah bug auto-posting diperbaiki) tidak dikoreksi otomatis.

## Saran Perbaikan

Ubah kondisi jadi menolak setiap invoice dengan `PaidAmount > 0` (mencakup `paid` maupun `partial`):

```go
func (i *Invoice) Cancel() error {
	if i.PaidAmount > 0 || i.Status == constant.StatusCancelled {
		return kernel.New(constant.CodeInvoiceInvalidStatus)
	}
	...
}
```

Kalau memang perlu jalur untuk "invoice yang sudah ada pembayaran tapi harus dikoreksi", itu keputusan produk terpisah (retur/refund) yang sengaja belum dibuat (lihat `docs/plan/keuangan-module.md` poin 8: "Tidak ada refund — ditunda sampai ada kebutuhan bisnis yang jelas"). Jangan diam-diam ditambahkan lewat celah ini.
