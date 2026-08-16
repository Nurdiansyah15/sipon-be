package query

import (
	"context"
	"strings"

	"sipon-be/internal/modules/identity/application"
	scoperepo "sipon-be/internal/modules/identity/domain/scope/repository"
	scopeservice "sipon-be/internal/modules/identity/domain/scope/service"
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
	scopeRepo         scoperepo.ScopeRepository
	getUserScopeSetUC *GetUserScopeSetUseCase
}

func NewGetUserScopeAccessUseCase(scopeRepo scoperepo.ScopeRepository, getUserScopeSetUC *GetUserScopeSetUseCase) *GetUserScopeAccessUseCase {
	return &GetUserScopeAccessUseCase{
		scopeRepo:         scopeRepo,
		getUserScopeSetUC: getUserScopeSetUC,
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

	return &UserScopeAccessResult{
		UserID:        userID,
		ScopeType:     scopeType,
		HasAccess:     decision.HasAccess,
		HasFullAccess: decision.HasFullAccess,
		AllowedCodes:  decision.AllowedCodes,
	}, nil
}

type CanAccessResourceUseCase struct {
	scopeRepo         scoperepo.ScopeRepository
	getUserScopeSetUC *GetUserScopeSetUseCase
}

func NewCanAccessResourceUseCase(scopeRepo scoperepo.ScopeRepository, getUserScopeSetUC *GetUserScopeSetUseCase) *CanAccessResourceUseCase {
	return &CanAccessResourceUseCase{
		scopeRepo:         scopeRepo,
		getUserScopeSetUC: getUserScopeSetUC,
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

	userSet, err := uc.getUserScopeSetUC.Execute(ctx, userID)
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
