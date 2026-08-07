package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	roleconstant "sipon-be/internal/modules/identity/domain/role/constant"
	roleentity "sipon-be/internal/modules/identity/domain/role/entity"
	rolerepo "sipon-be/internal/modules/identity/domain/role/repository"
	rolevo "sipon-be/internal/modules/identity/domain/role/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

type AssignRoleScopeUseCase struct {
	roleRepo      rolerepo.RoleRepository
	roleScopeRepo rolerepo.RoleScopeRepository
}

func NewAssignRoleScopeUseCase(
	roleRepo rolerepo.RoleRepository,
	roleScopeRepo rolerepo.RoleScopeRepository,
) *AssignRoleScopeUseCase {
	return &AssignRoleScopeUseCase{
		roleRepo:      roleRepo,
		roleScopeRepo: roleScopeRepo,
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

	return &dto.ScopeItem{
		ID:         scope.ID,
		ScopeType:  string(scope.ScopeType),
		ScopeValue: scope.ScopeValue,
	}, nil
}

type DeleteRoleScopeUseCase struct {
	roleScopeRepo rolerepo.RoleScopeRepository
}

func NewDeleteRoleScopeUseCase(roleScopeRepo rolerepo.RoleScopeRepository) *DeleteRoleScopeUseCase {
	return &DeleteRoleScopeUseCase{roleScopeRepo: roleScopeRepo}
}

func (uc *DeleteRoleScopeUseCase) Execute(ctx context.Context, scopeID string) error {
	scopeID = strings.TrimSpace(scopeID)
	if scopeID == "" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID scope tidak boleh kosong", nil)
	}

	if _, err := uc.roleScopeRepo.FindByID(ctx, scopeID); err != nil {
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
	return nil
}
