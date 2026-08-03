package query

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	pentity "sipon-be/internal/modules/psb/domain/pendaftar/entity"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	srepo 	"sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type GetPendaftaranUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	settingRepo    srepo.PsbSettingRepository
}

func NewGetPendaftaranUseCase(pendaftarRepo prepo.PendaftarRepository, settingRepo srepo.PsbSettingRepository) *GetPendaftaranUseCase {
	return &GetPendaftaranUseCase{pendaftarRepo: pendaftarRepo, settingRepo: settingRepo}
}

func (uc *GetPendaftaranUseCase) Execute(ctx context.Context, userID string) (*dto.PendaftarResponse, error) {
	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, setting.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	return mapPendaftarToResponse(p), nil
}

func (uc *GetPendaftaranUseCase) ExecuteByID(ctx context.Context, id string) (*dto.PendaftarResponse, error) {
	p, err := uc.pendaftarRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	return mapPendaftarToResponse(p), nil
}

func mapPendaftarToResponse(p *pentity.Pendaftar) *dto.PendaftarResponse {
	return &dto.PendaftarResponse{
		ID: p.ID, UserID: p.UserID, PsbSettingID: p.PsbSettingID,
		Gender: p.Gender, Program: p.Program,
		Nickname: p.Nickname, Hobby: p.Hobby, Purpose: p.Purpose, MotivationEntry: p.MotivationEntry,
		POB: p.POB, DOB: p.DOB, Blood: p.Blood,
		Address: p.Address, SubDistrict: p.SubDistrict, District: p.District, Province: p.Province, PostalCode: p.PostalCode,
		PreviousPondokName: p.PreviousPondokName, PreviousPondokAddress: p.PreviousPondokAddress,
		PreviousPondokDiv: p.PreviousPondokDiv, PreviousPondokTime: p.PreviousPondokTime,
		NIK: p.NIK, NoKK: p.NoKK, NISN: p.NISN, NoKIP: p.NoKIP, NoKKS: p.NoKKS, NoPKH: p.NoPKH,
		Workplace: p.Workplace, Department: p.Department, HomeStatus: p.HomeStatus,
		Father: p.Father, FatherPN: p.FatherPN, FatherNIK: p.FatherNIK,
		FatherJob: p.FatherJob, FatherGraduate: p.FatherGraduate, FatherIncome: p.FatherIncome,
		Mother: p.Mother, MotherPN: p.MotherPN, MotherNIK: p.MotherNIK,
		MotherJob: p.MotherJob, MotherGraduate: p.MotherGraduate, MotherIncome: p.MotherIncome,
		GuardianRelationship: p.GuardianRelationship, Guardian: p.Guardian, GuardianPN: p.GuardianPN,
		GuardianNIK: p.GuardianNIK, GuardianJob: p.GuardianJob, GuardianGraduate: p.GuardianGraduate, GuardianIncome: p.GuardianIncome,
		Status: string(p.Status), AcceptedBy: p.AcceptedBy, AcceptedAt: p.AcceptedAt,
		SantriID: p.SantriID, NIS: p.NIS,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
