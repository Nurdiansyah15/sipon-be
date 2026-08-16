package constant

import "sipon-be/internal/shared/kernel"

// ScopeType merepresentasikan klasifikasi scope master (mis. "gender").
// Bersifat terbuka — scope type baru (region, community, dst.) bisa
// ditambahkan ke tabel master tanpa mengubah kode.
type ScopeType string

const (
	ScopeTypeGender ScopeType = "gender"
)

// ScopeCodeGender adalah nilai scope untuk gender (laki-laki / perempuan).
const (
	ScopeCodeMale   = "male"
	ScopeCodeFemale = "female"
)

const (
	ErrCodeScopeNotFound     kernel.Code = "SCOPE_NOT_FOUND"
	ErrCodeScopeIDRequired   kernel.Code = "SCOPE_ID_REQUIRED"
	ErrCodeScopeTypeRequired kernel.Code = "SCOPE_TYPE_REQUIRED"
	ErrCodeScopeCodeRequired kernel.Code = "SCOPE_CODE_REQUIRED"
	ErrCodeScopeNameRequired kernel.Code = "SCOPE_NAME_REQUIRED"
	ErrCodeScopeInternal     kernel.Code = "SCOPE_INTERNAL_ERROR"
)
