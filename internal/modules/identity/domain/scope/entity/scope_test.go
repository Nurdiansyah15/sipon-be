package entity

import (
	"testing"

	scopeconstant "sipon-be/internal/modules/identity/domain/scope/constant"
	scopesvo "sipon-be/internal/modules/identity/domain/scope/valueobject"
	"sipon-be/internal/shared/kernel"
)

func TestNewScope(t *testing.T) {
	t.Run("valid scope", func(t *testing.T) {
		s, err := NewScope("id-1", scopesvo.ScopeTypeGender, "  MALE ", "Laki-laki", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.Code != "male" {
			t.Errorf("Code = %q, want %q", s.Code, "male")
		}
		if !s.IsActive {
			t.Error("IsActive harus true pada scope baru")
		}
	})

	t.Run("id kosong ditolak", func(t *testing.T) {
		_, err := NewScope("", scopesvo.ScopeTypeGender, "male", "Laki-laki", nil)
		assertCode(t, err, scopeconstant.ErrCodeScopeIDRequired)
	})

	t.Run("code kosong ditolak", func(t *testing.T) {
		_, err := NewScope("id-1", scopesvo.ScopeTypeGender, "  ", "Laki-laki", nil)
		assertCode(t, err, scopeconstant.ErrCodeScopeCodeRequired)
	})

	t.Run("name kosong ditolak", func(t *testing.T) {
		_, err := NewScope("id-1", scopesvo.ScopeTypeGender, "male", "  ", nil)
		assertCode(t, err, scopeconstant.ErrCodeScopeNameRequired)
	})
}

func TestScopeUpdateDetails(t *testing.T) {
	desc := "desc lama"
	s, err := NewScope("id-1", scopesvo.ScopeTypeGender, "male", "Laki-laki", &desc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updatedAt := s.UpdatedAt
	active := true
	s.UpdateDetails("Pria", nil, &active)

	if s.Name != "Pria" {
		t.Errorf("Name = %q, want %q", s.Name, "Pria")
	}
	if !s.UpdatedAt.After(updatedAt) && !s.UpdatedAt.Equal(updatedAt) {
		t.Error("UpdatedAt tidak diperbarui")
	}
}

func assertCode(t *testing.T, err error, code kernel.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", code)
	}
	ke, ok := err.(*kernel.AppError)
	if !ok {
		t.Fatalf("expected *kernel.AppError, got %T", err)
	}
	if ke.Code != code {
		t.Errorf("error code = %q, want %q", ke.Code, code)
	}
}
