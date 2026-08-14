package scope

import (
	"reflect"
	"testing"
)

func TestScopeSetZeroValueIsNone(t *testing.T) {
	var s ScopeSet
	if !s.IsNone() {
		t.Fatal("zero value harus IsNone")
	}
	if s.IsAll() || s.IsRestricted() || s.IsDenied() {
		t.Fatal("zero value tidak boleh menjadi mode lain")
	}
	if opts := s.AllowedOptions(); len(opts) != 0 {
		t.Fatalf("AllowedOptions zero value harus kosong, dapat %v", opts)
	}
}

func TestScopeSetUnrestricted(t *testing.T) {
	s := Unrestricted()
	if !s.IsAll() {
		t.Fatal("Unrestricted harus IsAll")
	}
	if s.IsNone() || s.IsRestricted() || s.IsDenied() {
		t.Fatal("Unrestricted tidak boleh menjadi mode lain")
	}
}

func TestScopeSetRestricted(t *testing.T) {
	s := Restricted([]string{"1", "2"})
	if !s.IsRestricted() {
		t.Fatal("Restricted harus IsRestricted")
	}
	if s.IsNone() || s.IsAll() || s.IsDenied() {
		t.Fatal("Restricted tidak boleh menjadi mode lain")
	}
	want := []string{"1", "2"}
	if got := s.AllowedOptions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("AllowedOptions = %v, want %v", got, want)
	}
}

func TestScopeSetRestrictedEmptyIsDenied(t *testing.T) {
	s := Restricted(nil)
	if !s.IsDenied() {
		t.Fatal("Restricted kosong harus IsDenied")
	}
	if s.IsNone() || s.IsAll() || !s.IsRestricted() {
		t.Fatal("Restricted kosong harus tetap IsRestricted")
	}
}

func TestScopeSetRestrictedDedupesAndDropsBlank(t *testing.T) {
	s := Restricted([]string{"1", " ", "1", "", "2", "2"})
	want := []string{"1", "2"}
	if got := s.AllowedOptions(); !reflect.DeepEqual(got, want) {
		t.Fatalf("AllowedOptions = %v, want %v", got, want)
	}
}

func TestScopeSetAllowedOptionsReturnsCopy(t *testing.T) {
	s := Restricted([]string{"1", "2"})
	got := s.AllowedOptions()
	got[0] = "9"
	if want := "1"; s.AllowedOptions()[0] != want {
		t.Fatalf("AllowedOptions harus mengembalikan salinan, isi asli berubah jadi %q", s.AllowedOptions()[0])
	}
}
