package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/system/application"
	"sipon-be/internal/modules/system/application/ports"
	scoperepo "sipon-be/internal/modules/system/domain/scope/repository"
	scopeservice "sipon-be/internal/modules/system/domain/scope/service"
	"sipon-be/internal/shared/kernel"
)

type UserScopeAccessResult struct {
	UserID        string
	ScopeType     string
	HasAccess     bool
	HasFullAccess bool
	AllowedCodes  []string
}

type GetUserScopeAccessUseCase struct {
	scopeRepo      scoperepo.ScopeRepository
	identityReader ports.IdentityReader
}

func NewGetUserScopeAccessUseCase(scopeRepo scoperepo.ScopeRepository, identityReader ports.IdentityReader) *GetUserScopeAccessUseCase {
	return &GetUserScopeAccessUseCase{
		scopeRepo:      scopeRepo,
		identityReader: identityReader,
	}
}

func (uc *GetUserScopeAccessUseCase) Execute(ctx context.Context, userID, scopeType string) (*UserScopeAccessResult, error) {
	userID = strings.TrimSpace(userID)
	scopeType = strings.ToLower(strings.TrimSpace(scopeType))
	if userID == "" || scopeType == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID user dan jenis scope wajib diisi", nil)
	}

	definedCodes, err := uc.scopeRepo.ListActiveCodesByType(ctx, scopeType)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal membaca kode scope master", err)
	}

	userSet, err := uc.identityReader.GetUserScopeSet(ctx, userID)
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

	return &UserScopeAccessResult{
		UserID:        userID,
		ScopeType:     scopeType,
		HasAccess:     decision.HasAccess,
		HasFullAccess: decision.HasFullAccess,
		AllowedCodes:  decision.AllowedCodes,
	}, nil
}

type CanAccessResourceUseCase struct {
	scopeRepo      scoperepo.ScopeRepository
	identityReader ports.IdentityReader
}

func NewCanAccessResourceUseCase(scopeRepo scoperepo.ScopeRepository, identityReader ports.IdentityReader) *CanAccessResourceUseCase {
	return &CanAccessResourceUseCase{
		scopeRepo:      scopeRepo,
		identityReader: identityReader,
	}
}

func (uc *CanAccessResourceUseCase) Execute(ctx context.Context, userID, scopeType string, resourceScopeCodes []string) (bool, error) {
	userID = strings.TrimSpace(userID)
	scopeType = strings.ToLower(strings.TrimSpace(scopeType))
	if userID == "" || scopeType == "" {
		return false, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID user dan jenis scope wajib diisi", nil)
	}

	definedCodes, err := uc.scopeRepo.ListActiveCodesByType(ctx, scopeType)
	if err != nil {
		return false, kernel.WrapMsg(application.ErrCodeInternal, "gagal membaca kode scope master", err)
	}

	userSet, err := uc.identityReader.GetUserScopeSet(ctx, userID)
	if err != nil {
		return false, kernel.WrapMsg(application.ErrCodeInternal, "gagal membaca scope user", err)
	}

	userCodes := make([]string, 0, len(userSet.Values))
	for _, v := range userSet.Values {
		if v.ScopeType == scopeType {
			userCodes = append(userCodes, v.ScopeValue)
		}
	}

	return scopeservice.CanAccessResource(definedCodes, userCodes, userSet.HasSystemRole, resourceScopeCodes), nil
}
