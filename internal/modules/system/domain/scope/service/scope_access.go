package service

// AccessDecision adalah hasil resolusi akses user terhadap satu scope type.
type AccessDecision struct {
	HasAccess     bool
	HasFullAccess bool
	// AllowedCodes berisi kode scope aktif yang boleh diakses. Kosong (nil)
	// ketika HasFullAccess true — pemanggil tidak perlu filter tambahan.
	AllowedCodes []string
}

// ResolveAccess menentukan akses user terhadap satu scope type berdasarkan
// scope yang dibawa user lewat role-nya.
//
// Aturan (sesuai requirement):
//  1. Role system (superuser) -> selalu full access.
//  2. User tanpa scope sama sekali -> tidak punya akses apa pun.
//  3. User memegang SELURUH kode scope terdefinisi (mis. male+female) ->
//     full access.
//  4. Selain itu -> akses hanya ke kode scope yang dimiliki.
//
// definedCodes: kode scope aktif dari master (mis. ["male","female"]).
// userCodes: kode scope efektif yang dibawa user lewat role-nya.
// hasSystemRole: true ketika user memegang role superuser (system).
func ResolveAccess(definedCodes, userCodes []string, hasSystemRole bool) AccessDecision {
	if hasSystemRole {
		return AccessDecision{HasAccess: true, HasFullAccess: true}
	}

	userSet := toSet(userCodes)
	if len(userSet) == 0 {
		return AccessDecision{HasAccess: false}
	}

	// Tidak ada master code aktif untuk type ini -> tidak ada data yang
	// dibatasi scope, user dengan scope apa pun dianggap penuh.
	if len(definedCodes) == 0 {
		return AccessDecision{HasAccess: true, HasFullAccess: true}
	}

	hasAll := true
	for _, c := range definedCodes {
		if _, ok := userSet[c]; !ok {
			hasAll = false
			break
		}
	}
	if hasAll {
		return AccessDecision{HasAccess: true, HasFullAccess: true}
	}

	allowed := make([]string, 0, len(definedCodes))
	for _, c := range definedCodes {
		if _, ok := userSet[c]; ok {
			allowed = append(allowed, c)
		}
	}
	if len(allowed) == 0 {
		return AccessDecision{HasAccess: false}
	}
	return AccessDecision{HasAccess: true, AllowedCodes: allowed}
}

// CanAccessResource mengecek apakah user boleh mengakses resource yang
// diklasifikasikan dengan resourceScopeCodes.
//
// - Resource tanpa scope (len==0) = publik -> boleh diakses siapa pun.
// - Role system -> selalu boleh.
// - User tanpa scope -> tidak boleh sama sekali.
// - Selain itu -> boleh jika setidaknya satu kode scope resource dimiliki user.
func CanAccessResource(definedCodes, userCodes []string, hasSystemRole bool, resourceScopeCodes []string) bool {
	if hasSystemRole {
		return true
	}
	if len(resourceScopeCodes) == 0 {
		return true
	}

	userSet := toSet(userCodes)
	if len(userSet) == 0 {
		return false
	}

	for _, c := range resourceScopeCodes {
		if _, ok := userSet[c]; ok {
			return true
		}
	}
	return false
}

func toSet(codes []string) map[string]struct{} {
	set := make(map[string]struct{}, len(codes))
	for _, c := range codes {
		if c != "" {
			set[c] = struct{}{}
		}
	}
	return set
}
