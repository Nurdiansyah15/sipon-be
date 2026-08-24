// Package scopegateway mengadaptasi identity.Contract ke ports.ScopeReader milik
// psb, sekaligus menerjemahkan kode scope sistem ("male"/"female") ke nilai
// gender domain pendaftar ('1'/'2'). Lihat
// docs/architecture/module-boundaries.md.
package scopegateway

import (
	"context"

	"sipon-be/internal/modules/identity"
	pendaftarscope "sipon-be/internal/modules/psb/domain/pendaftar/scope"
)

// scopeTypeGender adalah scope type master milik module identity yang
// mengklasifikasikan data pendaftar berdasarkan gender. Scope type & nilai scope
// (kode "male"/"female") didefinisikan di master scope identity sebagai sumber
// kebenaran tunggal; paket ini hanyalah anti-corruption layer yang
// menerjemahkannya ke kosakata domain pendaftar ('1'/'2').
const scopeTypeGender = "gender"

type Gateway struct {
	contract identity.Contract
}

func New(contract identity.Contract) *Gateway {
	return &Gateway{contract: contract}
}

func (g *Gateway) GetPendaftarScopeSet(ctx context.Context, userID string) (pendaftarscope.ScopeSet, error) {
	access, err := g.contract.GetUserScopeAccess(ctx, userID, scopeTypeGender)
	if err != nil {
		return pendaftarscope.Restricted(nil), err
	}
	if !access.HasAccess {
		return pendaftarscope.Restricted(nil), nil
	}
	if access.HasFullAccess {
		return pendaftarscope.Unrestricted(), nil
	}

	options := make([]string, 0, len(access.AllowedCodes))
	for _, code := range access.AllowedCodes {
		switch code {
		case "male":
			options = append(options, pendaftarscope.GenderMale)
		case "female":
			options = append(options, pendaftarscope.GenderFemale)
		}
	}
	return pendaftarscope.Restricted(options), nil
}