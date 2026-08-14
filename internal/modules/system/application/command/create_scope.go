package command

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"sipon-be/internal/modules/system/application"
	"sipon-be/internal/modules/system/application/dto"
	scopesentity "sipon-be/internal/modules/system/domain/scope/entity"
	scoperepo "sipon-be/internal/modules/system/domain/scope/repository"
	scopesvo "sipon-be/internal/modules/system/domain/scope/valueobject"
	"sipon-be/internal/shared/kernel"
)

type CreateScopeUseCase struct {
	scopeRepo scoperepo.ScopeRepository
}

func NewCreateScopeUseCase(scopeRepo scoperepo.ScopeRepository) *CreateScopeUseCase {
	return &CreateScopeUseCase{scopeRepo: scopeRepo}
}

func (uc *CreateScopeUseCase) Execute(ctx context.Context, req dto.CreateScopeRequest) (*dto.ScopeItem, error) {
	scopeType, err := scopesvo.NormalizeScopeType(req.ScopeType)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, err.Error(), err)
	}

	code := strings.ToLower(strings.TrimSpace(req.Code))
	if code == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "Kode scope wajib diisi", nil)
	}

	existing, err := uc.scopeRepo.FindByTypeAndCode(ctx, string(scopeType), code)
	if err == nil && existing != nil {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "Kode scope sudah ada untuk jenis scope tersebut", nil)
	}

	scope, err := scopesentity.NewScope(uuid.NewString(), scopeType, code, req.Name, req.Description)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, err.Error(), err)
	}

	if err := uc.scopeRepo.Save(ctx, scope); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal menyimpan scope", err)
	}

	return dto.ToScopeItem(scope), nil
}
