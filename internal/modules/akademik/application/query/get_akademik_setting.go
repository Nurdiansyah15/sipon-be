package query

import (
	"context"
	"errors"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/command"
	"sipon-be/internal/modules/akademik/application/dto"
	progRepo "sipon-be/internal/modules/akademik/domain/program/repository"
	setConst "sipon-be/internal/modules/akademik/domain/setting/constant"
	setEntity "sipon-be/internal/modules/akademik/domain/setting/entity"
	setRepo "sipon-be/internal/modules/akademik/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type GetAkademikSettingUseCase struct {
	settingRepo setRepo.AkademikSettingRepository
	programRepo progRepo.ProgramRepository
}

func NewGetAkademikSettingUseCase(settingRepo setRepo.AkademikSettingRepository, programRepo progRepo.ProgramRepository) *GetAkademikSettingUseCase {
	return &GetAkademikSettingUseCase{settingRepo: settingRepo, programRepo: programRepo}
}

func (uc *GetAkademikSettingUseCase) Execute(ctx context.Context) (*dto.AkademikSettingResponse, error) {
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
	return uc.toResponse(ctx, setting)
}

func (uc *GetAkademikSettingUseCase) toResponse(ctx context.Context, setting *setEntity.AkademikSetting) (*dto.AkademikSettingResponse, error) {
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
			resp.DefaultProgram = command.MapProgramToResponse(prog)
		}
	}
	return resp, nil
}
