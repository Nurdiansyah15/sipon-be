package service

import (
	"reflect"
	"testing"
)

func TestResolveAccess(t *testing.T) {
	defined := []string{"male", "female"}

	tests := []struct {
		name          string
		definedCodes  []string
		userCodes     []string
		hasSystemRole bool
		wantAccess    bool
		wantFull      bool
		wantAllowed   []string
	}{
		{
			name:          "system role selalu full access",
			definedCodes:  defined,
			userCodes:     nil,
			hasSystemRole: true,
			wantAccess:    true,
			wantFull:      true,
		},
		{
			name:         "tanpa scope sama sekali tidak bisa akses",
			definedCodes: defined,
			userCodes:    nil,
			wantAccess:   false,
			wantFull:     false,
		},
		{
			name:         "punya male dan female = full access",
			definedCodes: defined,
			userCodes:    []string{"male", "female"},
			wantAccess:   true,
			wantFull:     true,
		},
		{
			name:         "punya male saja akses terbatas ke male",
			definedCodes: defined,
			userCodes:    []string{"male"},
			wantAccess:   true,
			wantFull:     false,
			wantAllowed:  []string{"male"},
		},
		{
			name:         "punya female saja akses terbatas ke female",
			definedCodes: defined,
			userCodes:    []string{"female"},
			wantAccess:   true,
			wantFull:     false,
			wantAllowed:  []string{"female"},
		},
		{
			name:         "tidak ada master code terdefinisi = tidak ada pembatas",
			definedCodes: nil,
			userCodes:    []string{"male"},
			wantAccess:   true,
			wantFull:     true,
		},
		{
			name:         "user bawa code tak dikenal tetap akses terbatas sesuai master",
			definedCodes: defined,
			userCodes:    []string{"other"},
			wantAccess:   false,
			wantFull:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveAccess(tt.definedCodes, tt.userCodes, tt.hasSystemRole)
			if got.HasAccess != tt.wantAccess {
				t.Errorf("HasAccess = %v, want %v", got.HasAccess, tt.wantAccess)
			}
			if got.HasFullAccess != tt.wantFull {
				t.Errorf("HasFullAccess = %v, want %v", got.HasFullAccess, tt.wantFull)
			}
			if tt.wantAllowed == nil {
				if got.AllowedCodes != nil {
					t.Errorf("AllowedCodes = %v, want nil", got.AllowedCodes)
				}
			} else if !reflect.DeepEqual(got.AllowedCodes, tt.wantAllowed) {
				t.Errorf("AllowedCodes = %v, want %v", got.AllowedCodes, tt.wantAllowed)
			}
		})
	}
}

func TestCanAccessResource(t *testing.T) {
	defined := []string{"male", "female"}

	tests := []struct {
		name          string
		userCodes     []string
		hasSystemRole bool
		resourceCodes []string
		want          bool
	}{
		{
			name:          "system role boleh akses resource ber-scope apa pun",
			userCodes:     nil,
			hasSystemRole: true,
			resourceCodes: []string{"male"},
			want:          true,
		},
		{
			name:          "resource publik boleh diakses user tanpa scope",
			userCodes:     nil,
			hasSystemRole: false,
			resourceCodes: nil,
			want:          true,
		},
		{
			name:          "user tanpa scope tidak boleh akses resource ber-scope",
			userCodes:     nil,
			hasSystemRole: false,
			resourceCodes: []string{"male"},
			want:          false,
		},
		{
			name:          "scope cocok = boleh",
			userCodes:     []string{"male"},
			hasSystemRole: false,
			resourceCodes: []string{"male"},
			want:          true,
		},
		{
			name:          "scope tidak cocok = tidak boleh",
			userCodes:     []string{"male"},
			hasSystemRole: false,
			resourceCodes: []string{"female"},
			want:          false,
		},
		{
			name:          "salah satu scope resource cocok = boleh",
			userCodes:     []string{"male"},
			hasSystemRole: false,
			resourceCodes: []string{"male", "female"},
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanAccessResource(defined, tt.userCodes, tt.hasSystemRole, tt.resourceCodes)
			if got != tt.want {
				t.Errorf("CanAccessResource = %v, want %v", got, tt.want)
			}
		})
	}
}
