package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/identity/application"
	scopeconstant "sipon-be/internal/modules/identity/domain/scope/constant"
	scoperepo "sipon-be/internal/modules/identity/domain/scope/repository"
	"sipon-be/internal/shared/kernel"
)

type DeleteScopeUseCase struct {
	scopeRepo scoperepo.ScopeRepository
}

func NewDeleteScopeUseCase(scopeRepo scoperepo.ScopeRepository) *DeleteScopeUseCase {
	return &DeleteScopeUseCase{scopeRepo: scopeRepo}
}

func (uc *DeleteScopeUseCase) Execute(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "ID scope wajib diisi", nil)
	}

	if _, err := uc.scopeRepo.FindByID(ctx, id); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case scopeconstant.ErrCodeScopeNotFound:
				return kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal mencari scope", err)
	}

	if err := uc.scopeRepo.Delete(ctx, id); err != nil {
		return kernel.WrapMsg(application.ErrCodeInternal, "gagal menghapus scope", err)
	}
	return nil
}
