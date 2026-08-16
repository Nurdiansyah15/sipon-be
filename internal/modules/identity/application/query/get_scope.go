package query

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	"sipon-be/internal/modules/identity/application/dto"
	scopeconstant "sipon-be/internal/modules/identity/domain/scope/constant"
	scoperepo "sipon-be/internal/modules/identity/domain/scope/repository"
	"sipon-be/internal/shared/kernel"
)

type GetScopeUseCase struct {
	scopeRepo scoperepo.ScopeRepository
}

func NewGetScopeUseCase(scopeRepo scoperepo.ScopeRepository) *GetScopeUseCase {
	return &GetScopeUseCase{scopeRepo: scopeRepo}
}

func (uc *GetScopeUseCase) Execute(ctx context.Context, id string) (*dto.MasterScopeItem, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID scope wajib diisi", nil)
	}

	scope, err := uc.scopeRepo.FindByID(ctx, id)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case scopeconstant.ErrCodeScopeNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal mencari scope", err)
	}

	return dto.ToMasterScopeItem(scope), nil
}
