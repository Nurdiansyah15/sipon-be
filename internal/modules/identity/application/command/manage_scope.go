package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	"sipon-be/internal/modules/identity/application/ports"
	scopeconstant "sipon-be/internal/modules/identity/domain/scope/constant"
	scoperepo "sipon-be/internal/modules/identity/domain/scope/repository"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	rolevo "sipon-be/internal/modules/identity/domain/role/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AssignRoleScopeUseCase struct {
	roleRepo       rolerepo.RoleRepository
	roleScopeRepo  rolerepo.RoleScopeRepository
	userRoleRepo   rolerepo.UserRoleRepository
	principalCache ports.PrincipalCacheInvalidator
	scopeRepo      scoperepo.ScopeRepository
}

func NewAssignRoleScopeUseCase(
	roleRepo rolerepo.RoleRepository,
	roleScopeRepo rolerepo.RoleScopeRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	principalCache ports.PrincipalCacheInvalidator,
	scopeRepo scoperepo.ScopeRepository,
) *AssignRoleScopeUseCase {
	return &AssignRoleScopeUseCase{
		roleRepo:       roleRepo,
		roleScopeRepo:  roleScopeRepo,
		userRoleRepo:   userRoleRepo,
		principalCache: principalCache,
		scopeRepo:      scopeRepo,
	}
}

func (uc *AssignRoleScopeUseCase) Execute(ctx context.Context, roleID string, req dto.AssignScopeRequest) (*dto.ScopeItem, error) {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID role tidak boleh kosong", nil)
	}

	role, err := uc.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeRoleNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mencari role", err)
	}

	if role.IsSystem() {
		return nil, kernel.WrapMsg(application.ErrCodeForbidden, "Tidak dapat menambah scope pada role sistem", nil)
	}

	scopeType := rolevo.RoleScopeType(strings.TrimSpace(strings.ToLower(req.ScopeType)))
	scopeValue := strings.TrimSpace(strings.ToLower(req.ScopeValue))

	// Validasi scope terhadap master scope (sumber kebenaran tunggal). Nilai
	// role scope harus merujuk ke scope yang terdaftar & aktif di tabel
	// `scopes` — bukan nilai bebas/hardcoded.
	if err := uc.validateAgainstMaster(ctx, string(scopeType), scopeValue); err != nil {
		return nil, err
	}

	existingScopes, err := uc.roleScopeRepo.FindByRoleID(ctx, roleID)
	if err == nil {
		for _, es := range existingScopes {
			if es.ScopeType == scopeType && es.ScopeValue == scopeValue {
				return nil, kernel.WrapMsg(application.ErrCodeConflict, "Scope role sudah ada", nil)
			}
		}
	}

	scope, err := roleentity.NewRoleScope(uuid.NewString(), roleID, scopeType, scopeValue)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, ke.Message, ke)
		}
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "data tidak dapat diproses", err)
	}

	if err := uc.roleScopeRepo.Save(ctx, scope); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal menyimpan scope role", err)
	}

	invalidateRoleHolders(ctx, uc.principalCache, uc.userRoleRepo, role.Name)

	return &dto.ScopeItem{
		ID:         scope.ID,
		ScopeType:  string(scope.ScopeType),
		ScopeValue: scope.ScopeValue,
	}, nil
}

// validateAgainstMaster memastikan scope_type + scope_value terdaftar dan aktif
// di master scope (tabel `scopes`). Ini menjaga role_scopes selalu merujuk ke
// sumber kebenaran tunggal, bukan menyimpan nilai bebas.
func (uc *AssignRoleScopeUseCase) validateAgainstMaster(ctx context.Context, scopeType, scopeValue string) error {
	if uc.scopeRepo == nil {
		return nil
	}
	scope, err := uc.scopeRepo.FindByTypeAndCode(ctx, scopeType, scopeValue)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == scopeconstant.ErrCodeScopeNotFound {
			return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Scope tidak terdaftar di master scope", nil)
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal memvalidasi scope terhadap master", err)
	}
	if !scope.IsActive {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Scope tidak aktif di master scope", nil)
	}
	return nil
}

type DeleteRoleScopeUseCase struct {
	roleScopeRepo  rolerepo.RoleScopeRepository
	roleRepo       rolerepo.RoleRepository
	userRoleRepo   rolerepo.UserRoleRepository
	principalCache ports.PrincipalCacheInvalidator
}

func NewDeleteRoleScopeUseCase(
	roleScopeRepo rolerepo.RoleScopeRepository,
	roleRepo rolerepo.RoleRepository,
	userRoleRepo rolerepo.UserRoleRepository,
	principalCache ports.PrincipalCacheInvalidator,
) *DeleteRoleScopeUseCase {
	return &DeleteRoleScopeUseCase{
		roleScopeRepo:  roleScopeRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		principalCache: principalCache,
	}
}

func (uc *DeleteRoleScopeUseCase) Execute(ctx context.Context, scopeID string) error {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID scope tidak boleh kosong", nil)
	}

	scope, err := uc.roleScopeRepo.FindByID(ctx, scopeID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case roleconstant.ErrCodeRoleNotFound:
				return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal mencari scope role", err)
	}

	if err := uc.roleScopeRepo.Delete(ctx, scopeID); err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal menghapus scope role", err)
	}

	if role, err := uc.roleRepo.FindByID(ctx, scope.RoleID); err == nil {
		invalidateRoleHolders(ctx, uc.principalCache, uc.userRoleRepo, role.Name)
	}
	return nil
}
