// Package quota mendefinisikan domain kuota pendaftaran PSB. Kuota disimpan
// sebagai JSON (kolom `quota` pada psb_settings) berbentuk map flat:
//
//	{ "tahfidz_1": 20, "tahfidz_2": 20, "kitab_1": 15, "kitab_2": 15 }
//
// Key dibentuk dari kombinasi kode program (dari master `programs`) dan gender
// pendaftar ('1'/'2'). Kosakata program & gender diimpor dari
// domain/pendaftar/constant — domain pendaftar pemilik rule constant-nya,
// domain quota memakainya agar keduanya konsisten memvalidasi pendaftar.
//
// Kombinasi yang tidak dikenal (program/gender di luar konstanta) diperlakukan
// sebagai tanpa kuota (unlimited). Key yang tidak terkonfigurasi di JSON quota
// juga berarti tanpa batas.
package quota

import (
	"encoding/json"
	"fmt"
	"strings"

	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
)

// Combination adalah satu kombinasi kode program + gender yang dikenal sebagai
// bucket kuota beserta key JSON-nya.
type Combination struct {
	ProgramCode string
	Gender      string
	Key         string
}

// Combinations adalah daftar bucket kuota yang berlaku saat ini. Sumber
// kebenaran kombinasi dijaga sebagai konstanta karena nama jurusan & gender
// volatile — diselaraskan manual dengan master `programs`.
var Combinations = []Combination{
	{ProgramCode: pconstant.ProgramCodeTahfidz, Gender: pconstant.GenderMale, Key: "tahfidz_1"},
	{ProgramCode: pconstant.ProgramCodeTahfidz, Gender: pconstant.GenderFemale, Key: "tahfidz_2"},
	{ProgramCode: pconstant.ProgramCodeKitab, Gender: pconstant.GenderMale, Key: "kitab_1"},
	{ProgramCode: pconstant.ProgramCodeKitab, Gender: pconstant.GenderFemale, Key: "kitab_2"},
}

// KeyFor membentuk key kuota dari kode program dan gender. Mengembalikan
// (key, true) bila kombinasinya dikenal; (key, false) bila tidak — pemanggil
// harus memperlakukan kombinasi tak dikenal sebagai tanpa kuota.
func KeyFor(programCode, gender string) (string, bool) {
	code := strings.ToUpper(strings.TrimSpace(programCode))
	gender = strings.TrimSpace(gender)
	for _, c := range Combinations {
		if c.ProgramCode == code && c.Gender == gender {
			return c.Key, true
		}
	}
	return "", false
}

// Quota adalah himpunan batas kuota (limit) per key.
type Quota struct {
	limits map[string]int
}

// Parse membangun Quota dari JSON raw kolom `quota`. Nil/empty JSON dianggap
// sebagai tanpa kuota (semua kombinasi unlimited). Angka negatif ditolak;
// format yang bukan map objek ditolak.
func Parse(raw json.RawMessage) (Quota, error) {
	q := Quota{limits: map[string]int{}}
	if len(raw) == 0 {
		return q, nil
	}

	var values map[string]int
	if err := json.Unmarshal(raw, &values); err != nil {
		return Quota{}, fmt.Errorf("quota: json tidak valid: %w", err)
	}
	for key, limit := range values {
		if limit < 0 {
			return Quota{}, fmt.Errorf("quota: key %q bernilai negatif", key)
		}
		q.limits[key] = limit
	}
	return q, nil
}

// Limit mengembalikan batas kuota untuk sebuah key. bool false berarti key
// tidak terkonfigurasi — tanpa batas (unlimited).
func (q Quota) Limit(key string) (int, bool) {
	limit, ok := q.limits[key]
	return limit, ok
}

// IsFull true bila kuota key sudah terpakai penuh (remaining <= 0). Key tanpa
// konfigurasi tidak pernah dianggap penuh.
func (q Quota) IsFull(key string, used int64) bool {
	limit, ok := q.limits[key]
	if !ok {
		return false
	}
	return used >= int64(limit)
}

// Remaining menghitung sisa kuota. Nil balik nil berarti unlimited (key tidak
// terkonfigurasi) sehingga tidak perlu membatasi. Sisa negatif dipotong ke 0.
func (q Quota) Remaining(key string, used int64) (*int, bool) {
	limit, ok := q.limits[key]
	if !ok {
		return nil, false
	}
	remaining := limit - int(used)
	if remaining < 0 {
		remaining = 0
	}
	return &remaining, true
}

// ValidateKeys memastikan seluruh key pada JSON quota termasuk kombinasi yang
// dikenal (bucket kuota). Key asing ditolak agar konfigurasi tetap konsisten
// dengan konstanta rule.
func ValidateKeys(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var values map[string]int
	if err := json.Unmarshal(raw, &values); err != nil {
		return fmt.Errorf("quota: json tidak valid: %w", err)
	}
	for key := range values {
		found := false
		for _, c := range Combinations {
			if c.Key == key {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("quota: key %q tidak dikenal", key)
		}
	}
	return nil
}