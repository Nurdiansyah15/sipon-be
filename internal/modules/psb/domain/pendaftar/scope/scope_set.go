// Package scope menyediakan abstraksi scope yang sudah diterapkan (applied)
// terhadap domain pendaftar PSB. Nilai gender yang dipakai di sini mengikuti
// kosakata domain pendaftar (kolom "gender": '1' laki-laki / '2' perempuan),
// bukan kode scope sistem ("male"/"female"). Repository pendaftar men-*translate*
// ScopeSet ini ke kolom skema miliknya.
package scope

import "strings"

// Gender pendaftar — nilai kolom "gender" pada entitas Pendaftar.
const (
	// GenderMale adalah nilai "gender" untuk pendaftar laki-laki.
	GenderMale = "1"
	// GenderFemale adalah nilai "gender" untuk pendaftar perempuan.
	GenderFemale = "2"
)

type scopeMode uint8

const (
	// modeNone adalah zero value — scope belum diterapkan, tanpa filter.
	modeNone scopeMode = iota
	// modeAll berarti user boleh mengakses seluruh pendaftar (tanpa filter).
	modeAll
	// modeRestricted berarti user hanya boleh mengakses gender tertentu.
	modeRestricted
)

// ScopeSet adalah himpunan pendaftar yang boleh diakses user berdasarkan scope.
type ScopeSet struct {
	mode           scopeMode
	allowedOptions []string
}

// Unrestricted menyatakan akses penuh ke seluruh pendaftar tanpa filter scope.
func Unrestricted() ScopeSet {
	return ScopeSet{mode: modeAll}
}

// Restricted membatasi akses ke daftar gender pendaftar yang diberikan. Nilai
// kosong dan duplikat dibuang; daftar yang kosong berarti tidak ada satu pun
// pendaftar yang boleh diakses (IsDenied).
func Restricted(allowedOptions []string) ScopeSet {
	options := make([]string, 0, len(allowedOptions))
	seen := make(map[string]struct{}, len(allowedOptions))
	for _, o := range allowedOptions {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if _, ok := seen[o]; ok {
			continue
		}
		seen[o] = struct{}{}
		options = append(options, o)
	}
	return ScopeSet{mode: modeRestricted, allowedOptions: options}
}

// IsNone true ketika scope belum diterapkan — pemanggil tidak perlu menambah
// filter apa pun.
func (s ScopeSet) IsNone() bool { return s.mode == modeNone }

// IsAll true ketika user boleh mengakses seluruh pendaftar tanpa filter tambahan.
func (s ScopeSet) IsAll() bool { return s.mode == modeAll }

// IsRestricted true ketika akses dibatasi ke subset gender pendaftar.
func (s ScopeSet) IsRestricted() bool { return s.mode == modeRestricted }

// IsDenied true ketika tidak ada satu pun pendaftar yang boleh diakses.
func (s ScopeSet) IsDenied() bool {
	return s.mode == modeRestricted && len(s.allowedOptions) == 0
}

// AllowedOptions mengembalikan salinan daftar gender pendaftar yang boleh
// diakses. Hanya bermakna saat IsRestricted true.
func (s ScopeSet) AllowedOptions() []string {
	return append([]string(nil), s.allowedOptions...)
}