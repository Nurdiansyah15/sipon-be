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
//
// Nilai scope (scope_value) TIDAK didefinisikan di sini — sumber kebenaran
// tunggal berada di master scope (tabel `scopes` milik module identity).
// Validasi bahwa scope_type + scope_value terdaftar di master dilakukan di
// application layer (AssignRoleScopeUseCase).
type RoleScopeType string

const (
	ErrCodeInvalidScopeType  kernel.Code = "INVALID_SCOPE_TYPE"
	ErrCodeInvalidScopeValue kernel.Code = "INVALID_SCOPE_VALUE"
)

// NormalizeScopeValue menormalkan scope_value (lowercase + trim) tanpa
// memvalidasi nilai spesifik — validasi terhadap master scope dilakukan oleh
// pemanggil.
func NormalizeScopeValue(scopeType RoleScopeType, rawValue string) (string, error) {
	if strings.TrimSpace(string(scopeType)) == "" {
		return "", kernel.WrapMsg(ErrCodeInvalidScopeType, "Jenis scope wajib diisi", nil)
	}
	v := strings.ToLower(strings.TrimSpace(rawValue))
	if v == "" {
		return "", kernel.WrapMsg(ErrCodeInvalidScopeValue, "Nilai scope wajib diisi", nil)
	}
	return v, nil
}
