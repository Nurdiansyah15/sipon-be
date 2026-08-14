package command

import (
	"context"
	"errors"
	"strings"

	"sipon-be/internal/modules/system/application"
	"sipon-be/internal/modules/system/application/dto"
	scopeconstant "sipon-be/internal/modules/system/domain/scope/constant"
	scoperepo "sipon-be/internal/modules/system/domain/scope/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateScopeUseCase struct {
	scopeRepo scoperepo.ScopeRepository
}

func NewUpdateScopeUseCase(scopeRepo scoperepo.ScopeRepository) *UpdateScopeUseCase {
	return &UpdateScopeUseCase{scopeRepo: scopeRepo}
}

func (uc *UpdateScopeUseCase) Execute(ctx context.Context, id string, req dto.UpdateScopeRequest) (*dto.ScopeItem, error) {
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

	scope.UpdateDetails(req.Name, req.Description, req.IsActive)
	if err := uc.scopeRepo.Update(ctx, scope); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "gagal memperbarui scope", err)
	}

	return dto.ToScopeItem(scope), nil
}
