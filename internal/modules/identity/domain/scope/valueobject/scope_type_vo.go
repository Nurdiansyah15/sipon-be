package valueobject

import (
	"strings"

	scopeconstant "sipon-be/internal/modules/identity/domain/scope/constant"
	"sipon-be/internal/shared/kernel"
)

type ScopeType string

const (
	ScopeTypeGender ScopeType = "gender"
)

// NormalizeScopeType menormalisasi dan memvalidasi scope type. Scope type
// sengaja bersifat terbuka (tidak dibatasi hanya "gender") supaya scope
// master bisa diperluas tanpa perubahan kode.
func NormalizeScopeType(raw string) (ScopeType, error) {
	t := strings.ToLower(strings.TrimSpace(raw))
	if t == "" {
		return "", kernel.WrapMsg(scopeconstant.ErrCodeScopeTypeRequired, "Jenis scope wajib diisi", nil)
	}
	return ScopeType(t), nil
}

func (t ScopeType) String() string {
	return string(t)
}
