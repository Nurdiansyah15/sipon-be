package valueobject

import (
	"testing"
)

func TestNormalizeScopeType(t *testing.T) {
	tests := []struct {
		raw     string
		want    ScopeType
		wantErr bool
	}{
		{raw: "gender", want: ScopeTypeGender},
		{raw: "  GENDER ", want: ScopeTypeGender},
		{raw: "region", want: "region"},
		{raw: "", wantErr: true},
		{raw: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := NormalizeScopeType(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("NormalizeScopeType(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
