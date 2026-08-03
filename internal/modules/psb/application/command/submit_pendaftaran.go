package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	"sipon-be/internal/shared/kernel"
)

type PendaftarActionUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
}

func NewPendaftarActionUseCase(repo prepo.PendaftarRepository) *PendaftarActionUseCase {
	return &PendaftarActionUseCase{pendaftarRepo: repo}
}

func (uc *PendaftarActionUseCase) SubmitPendaftaran(ctx context.Context, userID, settingID string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, settingID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if err := p.Submit(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == pconstant.CodePendaftarInvalidStatus {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "pendaftaran berhasil diajukan"}, nil
}

func (uc *PendaftarActionUseCase) SubmitDaftarUlang(ctx context.Context, userID, settingID string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, settingID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if err := p.SubmitDaftarUlang(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == pconstant.CodePendaftarInvalidStatus {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "daftar ulang berhasil diajukan"}, nil
}
