package command

import (
	"context"
	"errors"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	progConst "sipon-be/internal/modules/akademik/domain/program/constant"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	setConst "sipon-be/internal/modules/akademik/domain/setting/constant"
	setEntity "sipon-be/internal/modules/akademik/domain/setting/entity"
	setRepo "sipon-be/internal/modules/akademik/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type UpdateAkademikSettingUseCase struct {
	settingRepo setRepo.AkademikSettingRepository
	programRepo progRepo.ProgramRepository
}

func NewUpdateAkademikSettingUseCase(settingRepo setRepo.AkademikSettingRepository, programRepo progRepo.ProgramRepository) *UpdateAkademikSettingUseCase {
	return &UpdateAkademikSettingUseCase{settingRepo: settingRepo, programRepo: programRepo}
}

func (uc *UpdateAkademikSettingUseCase) Execute(ctx context.Context, req dto.UpdateAkademikSettingRequest) (*dto.AkademikSettingResponse, error) {
	if req.DefaultProgramID != nil {
		if err := uc.validateProgram(ctx, *req.DefaultProgramID); err != nil {
			return nil, err
		}
	}

	setting, err := uc.settingRepo.Find(ctx)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case setConst.CodeSettingNotFound:
				return nil, kernel.WrapMsg(application.ErrCodeNotFound, ke.Message, ke)
			}
		}
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	if err := setting.SetDefaultProgramID(req.DefaultProgramID); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "gagal menyimpan settings", err)
	}
	if err := uc.settingRepo.Update(ctx, setting); err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}

	return uc.toResponse(ctx, setting)
}

func (uc *UpdateAkademikSettingUseCase) validateProgram(ctx context.Context, programID string) error {
	prog, err := uc.programRepo.FindByID(ctx, programID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) {
			switch ke.Code {
			case progConst.CodeProgramNotFound:
				return kernel.WrapMsg(application.ErrCodeNotFound, "Program tidak ditemukan", ke)
			}
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if prog.Status != progConst.ProgramStatusActive {
		return kernel.WrapMsg(application.ErrCodeBadRequest, "Program default harus berstatus aktif", nil)
	}
	return nil
}

func (uc *UpdateAkademikSettingUseCase) toResponse(ctx context.Context, setting *setEntity.AkademikSetting) (*dto.AkademikSettingResponse, error) {
	programID, err := setting.GetDefaultProgramID()
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "gagal membaca settings", err)
	}
	resp := &dto.AkademikSettingResponse{
		DefaultProgramID: programID,
	}
	if programID != nil {
		prog, err := uc.programRepo.FindByID(ctx, *programID)
		if err == nil {
			resp.DefaultProgram = MapProgramToResponse(prog)
		}
	}
	return resp, nil
}
