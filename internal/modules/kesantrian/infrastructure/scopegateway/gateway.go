// Package scopegateway mengadaptasi identity.Contract ke ports.ScopeReader milik
// kesantrian, sekaligus menerjemahkan kode scope sistem ("male"/"female") ke
// nilai gender domain santri ('1'/'2'). Lihat
// docs/architecture/module-boundaries.md.
package scopegateway

import (
	"context"

	"sipon-be/internal/modules/identity"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
)

// scopeTypeGender adalah scope type master milik module identity yang
// mengklasifikasikan data santri berdasarkan gender. Scope type & nilai scope
// (kode "male"/"female") didefinisikan di master scope identity sebagai sumber
// kebenaran tunggal; paket ini hanyalah anti-corruption layer yang
// menerjemahkannya ke kosakata domain santri ('1'/'2').
const scopeTypeGender = "gender"

type Gateway struct {
	contract identity.Contract
}

func New(contract identity.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) GetSantriScopeSet(ctx context.Context, userID string) (santriscope.ScopeSet, error) {
	access, err := g.contract.GetUserScopeAccess(ctx, userID, scopeTypeGender)
	if err != nil {
		return santriscope.Restricted(nil), err
	}
	if !access.HasAccess {
		return santriscope.Restricted(nil), nil
	}
	if access.HasFullAccess {
		return santriscope.Unrestricted(), nil
	}

	options := make([]string, 0, len(access.AllowedCodes))
	for _, code := range access.AllowedCodes {
		switch code {
		case "male":
			options = append(options, santriscope.GenderMale)
		case "female":
			options = append(options, santriscope.GenderFemale)
		}
	}
	return santriscope.Restricted(options), nil
}

// GetSuratScopeAccess mengembalikan akses scope user untuk resource surat.
// AllowedScopeIDs berisi master scope ID (dari identity), karena tabel surat
// menyimpan scope_id sebagai referensi master scope.
func (g *Gateway) GetSuratScopeAccess(ctx context.Context, userID string) (*ports.SuratScopeAccess, error) {
	access, err := g.contract.GetUserScopeAccessIDs(ctx, userID, scopeTypeGender)
	if err != nil {
		return &ports.SuratScopeAccess{}, err
	}
	return &ports.SuratScopeAccess{
		HasAccess:       access.HasAccess,
		HasFullAccess:   access.HasFullAccess,
		AllowedScopeIDs: access.AllowedScopeIDs,
	}, nil
}
