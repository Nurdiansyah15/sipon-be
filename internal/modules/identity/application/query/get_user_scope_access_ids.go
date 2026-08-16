package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/identity/application"
	scoperepo "sipon-be/internal/modules/identity/domain/scope/repository"
	scopeservice "sipon-be/internal/modules/identity/domain/scope/service"
	"sipon-be/internal/shared/kernel"
)

// UserScopeAccessIDsResult sama dengan UserScopeAccessResult tetapi AllowedCodes
// diterjemahkan ke master scope ID (bukan kode). Dipakai pemanggil yang menyimpan
// scope_id pada resource (mis. kesantrian.surat) untuk auto-tagging & filter.
type UserScopeAccessIDsResult struct {
	UserID          string
	ScopeType       string
	HasAccess       bool
	HasFullAccess   bool
	AllowedScopeIDs []string
}

type GetUserScopeAccessIDsUseCase struct {
	scopeRepo         scoperepo.ScopeRepository
	getUserScopeSetUC *GetUserScopeSetUseCase
}

func NewGetUserScopeAccessIDsUseCase(scopeRepo scoperepo.ScopeRepository, getUserScopeSetUC *GetUserScopeSetUseCase) *GetUserScopeAccessIDsUseCase {
	return &GetUserScopeAccessIDsUseCase{
		scopeRepo:         scopeRepo,
		getUserScopeSetUC: getUserScopeSetUC,
	}
}

func (uc *GetUserScopeAccessIDsUseCase) Execute(ctx context.Context, userID, scopeType string) (*UserScopeAccessIDsResult, error) {
	userID = strings.TrimSpace(userID)
	scopeType = strings.ToLower(strings.TrimSpace(scopeType))
	if userID == "" || scopeType == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID user dan jenis scope wajib diisi", nil)
	}

	// Ambil scope master aktif + peta code→id.
	definedScopes, err := uc.scopeRepo.ListByType(ctx, scopeType, false)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membaca scope master", err)
	}
	definedCodes := make([]string, 0, len(definedScopes))
	codeToID := make(map[string]string, len(definedScopes))
	for _, s := range definedScopes {
		definedCodes = append(definedCodes, s.Code)
		codeToID[s.Code] = s.ID
	}

	userSet, err := uc.getUserScopeSetUC.Execute(ctx, userID)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membaca scope user", err)
	}

	userCodes := make([]string, 0, len(userSet.Values))
	for _, v := range userSet.Values {
		if v.ScopeType == scopeType {
			userCodes = append(userCodes, v.ScopeValue)
		}
	}

	decision := scopeservice.ResolveAccess(definedCodes, userCodes, userSet.HasSystemRole)

	result := &UserScopeAccessIDsResult{
		UserID:        userID,
		ScopeType:     scopeType,
		HasAccess:     decision.HasAccess,
		HasFullAccess: decision.HasFullAccess,
	}
	if decision.HasAccess && !decision.HasFullAccess {
		ids := make([]string, 0, len(decision.AllowedCodes))
		for _, code := range decision.AllowedCodes {
			if id, ok := codeToID[code]; ok {
				ids = append(ids, id)
			}
		}
		result.AllowedScopeIDs = ids
	}

	return result, nil
}
