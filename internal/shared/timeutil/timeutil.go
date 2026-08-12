// Package timeutil menyediakan helper waktu berbasis platform timezone.
//
// Platform timezone adalah timezone yang berlaku untuk seluruh pengguna
// (tidak per-user), dikonfigurasi lewat env APP_TIMEZONE (default
// Asia/Jakarta). Inisialisasi dilakukan sekali saat aplikasi start lewat
// Init().
//
// Konvensi:
//   - TIMESTAMPTZ di database tetap disimpan dalam UTC.
//   - Kolom DATE / TIME bersifat timezone-naive: DATE dan TIME merepresentasikan
//     wall-clock dalam platform timezone.
//   - Output timestamp ke UI selalu dikonversi ke platform timezone agar
//     frontend tidak perlu melakukan konversi.
package timeutil

import (
	"time"
)

var platformLoc *time.Location

// Init memuat lokasi platform timezone. Wajib dipanggil sekali saat
// aplikasi start, sebelum helper lain dipakai.
func Init(timezone string) error {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return err
	}
	platformLoc = loc
	return nil
}

// Loc mengembalikan lokasi platform timezone. Menghasilkan UTC bila Init
// belum dipanggil (fail-safe agar tidak panic).
func Loc() *time.Location {
	if platformLoc == nil {
		return time.UTC
	}
	return platformLoc
}

// Now mengembalikan waktu sekarang dalam platform timezone.
func Now() time.Time {
	return time.Now().In(Loc())
}

// ToPlatform mengonversi time.Time apa pun ke platform timezone.
func ToPlatform(t time.Time) time.Time {
	return t.In(Loc())
}

// ToPlatformPtr mengonversi *time.Time ke platform timezone, mempertahankan
// nil.
func ToPlatformPtr(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	v := t.In(Loc())
	return &v
}

// ParseDate mem-parse string tanggal "2006-01-02" dalam platform timezone.
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, Loc())
}

// ParseDatePtr mem-parse string tanggal nullable dalam platform timezone.
// Nilai nil / string kosong mengembalikan nil tanpa error.
func ParseDatePtr(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := ParseDate(*s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FormatDate mem-format time.Time menjadi "2006-01-02" dalam platform
// timezone.
func FormatDate(t time.Time) string {
	return t.In(Loc()).Format("2006-01-02")
}

// FormatDateTime mem-format time.Time menjadi RFC3339 dalam platform
// timezone, misal "2026-08-12T08:00:00+07:00".
func FormatDateTime(t time.Time) string {
	return t.In(Loc()).Format(time.RFC3339)
}

// DateOnly memotong komponen waktu, mempertahankan tanggal dalam platform
// timezone (pukul 00:00:00 platform timezone).
func DateOnly(t time.Time) time.Time {
	inLoc := t.In(Loc())
	return time.Date(inLoc.Year(), inLoc.Month(), inLoc.Day(), 0, 0, 0, 0, Loc())
}
