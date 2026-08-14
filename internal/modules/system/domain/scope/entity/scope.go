package entity

import (
	"strings"
	"time"

	scopeconstant "sipon-be/internal/modules/system/domain/scope/constant"
	scopesvo "sipon-be/internal/modules/system/domain/scope/valueobject"
	"sipon-be/internal/shared/kernel"
)

// Scope adalah master data klasifikasi scope (mis. gender: male/female).
// Entity ini berdiri sendiri per scope_type — resource mana pun di modul lain
// yang butuh klasifikasi gender cukup menyimpan scope_type + code dari sini.
type Scope struct {
	ID          string
	ScopeType   scopesvo.ScopeType
	Code        string
	Name        string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewScope(id string, scopeType scopesvo.ScopeType, code, name string, description *string) (*Scope, error) {
	if id == "" {
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeIDRequired, "ID scope wajib diisi", nil)
	}
	code = normalizeCode(code)
	if code == "" {
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeCodeRequired, "Kode scope wajib diisi", nil)
	}
	if strings.TrimSpace(name) == "" {
		return nil, kernel.WrapMsg(scopeconstant.ErrCodeScopeNameRequired, "Nama scope wajib diisi", nil)
	}

	now := time.Now()
	return &Scope{
		ID:          id,
		ScopeType:   scopeType,
		Code:        code,
		Name:        strings.TrimSpace(name),
		Description: description,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Scope) Touch() {
	s.UpdatedAt = time.Now()
}

func (s *Scope) UpdateDetails(name string, description *string, isActive *bool) {
	if strings.TrimSpace(name) != "" {
		s.Name = strings.TrimSpace(name)
	}
	if description != nil {
		s.Description = description
	}
	if isActive != nil {
		s.IsActive = *isActive
	}
	s.Touch()
}

func normalizeCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}
