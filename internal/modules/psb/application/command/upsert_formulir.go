package command

import (
	"context"

	"github.com/google/uuid"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	pentity "sipon-be/internal/modules/psb/domain/pendaftar/entity"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type UpsertFormulirUseCase struct {
	settingRepo   srepo.PsbSettingRepository
	pendaftarRepo prepo.PendaftarRepository
}

func NewUpsertFormulirUseCase(settingRepo srepo.PsbSettingRepository, pendaftarRepo prepo.PendaftarRepository) *UpsertFormulirUseCase {
	return &UpsertFormulirUseCase{settingRepo: settingRepo, pendaftarRepo: pendaftarRepo}
}

func (uc *UpsertFormulirUseCase) Execute(ctx context.Context, userID string, req dto.UpsertFormulirRequest) (*dto.PendaftarResponse, error) {
	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, setting.ID)
	if err != nil {
		p = nil
	}

	if p == nil {
		p, err = pentity.NewPendaftar(uuid.NewString(), userID, setting.ID, "1", req.Program)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	err = p.UpsertFormulir(func(p *pentity.Pendaftar) {
		p.Gender = "1"
		p.Program = req.Program
		p.Nickname = req.Nickname
		p.Hobby = req.Hobby
		p.Purpose = req.Purpose
		p.MotivationEntry = req.MotivationEntry
		p.POB = req.POB
		p.DOB = req.DOB
		p.Blood = req.Blood
		p.Address = req.Address
		p.SubDistrict = req.SubDistrict
		p.District = req.District
		p.Province = req.Province
		p.PostalCode = req.PostalCode
		p.PreviousPondokName = req.PreviousPondokName
		p.PreviousPondokAddress = req.PreviousPondokAddress
		p.PreviousPondokDiv = req.PreviousPondokDiv
		p.PreviousPondokTime = req.PreviousPondokTime
		p.NIK = req.NIK
		p.NoKK = req.NoKK
		p.NISN = req.NISN
		p.NoKIP = req.NoKIP
		p.NoKKS = req.NoKKS
		p.NoPKH = req.NoPKH
		p.Workplace = req.Workplace
		p.Department = req.Department
		p.HomeStatus = req.HomeStatus
		p.Father = req.Father
		p.FatherPN = req.FatherPN
		p.FatherNIK = req.FatherNIK
		p.FatherJob = req.FatherJob
		p.FatherGraduate = req.FatherGraduate
		p.FatherIncome = req.FatherIncome
		p.Mother = req.Mother
		p.MotherPN = req.MotherPN
		p.MotherNIK = req.MotherNIK
		p.MotherJob = req.MotherJob
		p.MotherGraduate = req.MotherGraduate
		p.MotherIncome = req.MotherIncome
		p.GuardianRelationship = req.GuardianRelationship
		p.Guardian = req.Guardian
		p.GuardianPN = req.GuardianPN
		p.GuardianNIK = req.GuardianNIK
		p.GuardianJob = req.GuardianJob
		p.GuardianGraduate = req.GuardianGraduate
		p.GuardianIncome = req.GuardianIncome
	})
	if err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if p.CreatedAt.IsZero() {
		if err := uc.pendaftarRepo.Save(ctx, p); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	} else {
		if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
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
