// Package scopegateway mengadaptasi system.Contract ke ports.ScopeReader milik
// kesantrian, sekaligus menerjemahkan kode scope sistem ("male"/"female") ke
// nilai gender domain santri ('1'/'2'). Lihat
// docs/architecture/module-boundaries.md.
package scopegateway

import (
	"context"

	santriscope "sipon-be/internal/modules/kesantrian/domain/santri/scope"
	"sipon-be/internal/modules/system"
)

// scopeTypeGender adalah scope type master milik module system yang
// mengklasifikasikan data santri berdasarkan gender. Nilai ini harus sinkron
// dengan system/domain/scope/constant.ScopeTypeGender — kesantrian tidak
// boleh meng-import konstanta itu langsung (batas modul).
const scopeTypeGender = "gender"

type Gateway struct {
	contract system.Contract
}

func New(contract system.Contract) *Gateway {
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
