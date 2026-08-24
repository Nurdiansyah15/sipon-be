package quota

import (
	"encoding/json"
	"testing"
)

func TestParseEmptyIsUnlimited(t *testing.T) {
	q, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse(nil) error: %v", err)
	}
	if _, ok := q.Limit("tahfidz_1"); ok {
		t.Fatal("limit kosong harus tidak dikonfigurasi (unlimited)")
	}
	if q.IsFull("tahfidz_1", 999) {
		t.Fatal("key tanpa konfigurasi tidak boleh dianggap penuh")
	}
	if _, ok := q.Remaining("tahfidz_1", 0); ok {
		t.Fatal("key tanpa konfigurasi tidak boleh punya remaining")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	if _, err := Parse(json.RawMessage(`{"tahfidz_1": "x"}`)); err == nil {
		t.Fatal("JSON tidak valid harus error")
	}
}

func TestParseNegativeLimit(t *testing.T) {
	if _, err := Parse(json.RawMessage(`{"tahfidz_1": -1}`)); err == nil {
		t.Fatal("limit negatif harus error")
	}
}

func TestParseLimitAndRemaining(t *testing.T) {
	q, err := Parse(json.RawMessage(`{"tahfidz_1": 20, "kitab_2": 5}`))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if got, ok := q.Limit("tahfidz_1"); !ok || got != 20 {
		t.Fatalf("Limit(tahfidz_1) = %d, %v; want 20, true", got, ok)
	}

	rem, ok := q.Remaining("tahfidz_1", 7)
	if !ok || rem == nil || *rem != 13 {
		t.Fatalf("Remaining(tahfidz_1,7) = %v, %v; want 13, true", rem, ok)
	}

	if rem, ok := q.Remaining("kitab_2", 10); !ok || rem == nil || *rem != 0 {
		t.Fatalf("Remaining negatif harus dipotong ke 0, dapat %v, %v", rem, ok)
	}

	if !q.IsFull("kitab_2", 5) {
		t.Fatal("used == limit harus dianggap penuh")
	}
	if q.IsFull("tahfidz_1", 19) {
		t.Fatal("used < limit tidak boleh dianggap penuh")
	}
}

func TestKeyForKnownCombinations(t *testing.T) {
	cases := []struct {
		code, gender, key string
		wantOK            bool
	}{
		{"TAHFIDZ", "1", "tahfidz_1", true},
		{"tahfidz", "2", "tahfidz_2", true},
		{"KITAB", "1", "kitab_1", true},
		{"KITAB", "2", "kitab_2", true},
		{"REGULER", "1", "", false},
		{"TAHFIDZ", "9", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		key, ok := KeyFor(c.code, c.gender)
		if ok != c.wantOK || key != c.key {
			t.Fatalf("KeyFor(%q, %q) = (%q, %v); want (%q, %v)", c.code, c.gender, key, ok, c.key, c.wantOK)
		}
	}
}

func TestValidateKeys(t *testing.T) {
	if err := ValidateKeys(json.RawMessage(`{"tahfidz_1": 20, "kitab_2": 5}`)); err != nil {
		t.Fatalf("key dikenal harus lolos: %v", err)
	}
	if err := ValidateKeys(json.RawMessage(`{"tahfidz_1": 20, "reguler_1": 5}`)); err == nil {
		t.Fatal("key asing harus ditolak")
	}
	if err := ValidateKeys(nil); err != nil {
		t.Fatalf("nil quota harus lolos: %v", err)
	}
}