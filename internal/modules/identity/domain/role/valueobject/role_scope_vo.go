package valueobject

import (
	"strings"

	"sipon-be/internal/shared/kernel"
)

// RoleScopeType adalah atribut pembatas tambahan pada sebuah role (mis. gender)
// — array of object yang berdiri sendiri per role, TIDAK terkait dengan
// constant.ScopeType (global/region/community) yang dipakai Role/UserRole
// untuk hierarki assignment. Menyamakan keduanya adalah bug: role_scopes
// menyimpan atribut arbitrer (saat ini baru "gender"), bukan level scope
// hierarkis.
type RoleScopeType string

const (
	RoleScopeTypeGender RoleScopeType = "gender"
)

const (
	RoleScopeValueMale   = "male"
	RoleScopeValueFemale = "female"
)

const (
	ErrCodeInvalidScopeType  kernel.Code = "INVALID_SCOPE_TYPE"
	ErrCodeInvalidScopeValue kernel.Code = "INVALID_SCOPE_VALUE"
)

// NewRoleScopeValue menormalisasi dan memvalidasi scope_value sesuai scope_type-nya.
// Saat ini hanya "gender" yang didukung; scope_type lain ditolak.
func NewRoleScopeValue(scopeType RoleScopeType, rawValue string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(rawValue))
	switch scopeType {
	case RoleScopeTypeGender:
		if v != RoleScopeValueMale && v != RoleScopeValueFemale {
			return "", kernel.WrapMsg(ErrCodeInvalidScopeValue, "Nilai scope tidak valid untuk gender", nil)
		}
		return v, nil
	default:
		return "", kernel.WrapMsg(ErrCodeInvalidScopeType, "Jenis scope tidak dikenali", nil)
	}
}
