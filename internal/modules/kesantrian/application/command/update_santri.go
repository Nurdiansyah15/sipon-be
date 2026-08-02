package command

import (
	"context"
	"log/slog"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrientity "sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type UpdateSantriUseCase struct {
	santriRepo  santrirepo.SantriRepository
	provisioner ports.AccountProvisioner
}

func NewUpdateSantriUseCase(santriRepo santrirepo.SantriRepository, provisioner ports.AccountProvisioner) *UpdateSantriUseCase {
	return &UpdateSantriUseCase{santriRepo: santriRepo, provisioner: provisioner}
}

func (uc *UpdateSantriUseCase) Execute(ctx context.Context, userID string, req dto.UpdateSantriRequest) (*dto.UpdateSantriResponse, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	applySantriUpdate(santri, req)
	santri.Update()

	if err := uc.santriRepo.Update(ctx, santri); err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	if req.Fullname != nil {
		if err := uc.provisioner.UpdateFullname(ctx, userID, *req.Fullname); err != nil {
			slog.Warn("kesantrian: best-effort fullname sync to identity failed", "user_id", userID, "error", err)
		}
	}

	return &dto.UpdateSantriResponse{Message: "profil santri berhasil diperbarui"}, nil
}

// applySantriUpdate only copies non-nil pointer fields (partial-update
// pattern), mirroring sipon-api's helpers.go.
func applySantriUpdate(s *santrientity.Santri, req dto.UpdateSantriRequest) {
	if req.Nickname != nil {
		s.Nickname = req.Nickname
	}
	if req.Program != nil {
		s.Program = req.Program
	}
	if req.Hobby != nil {
		s.Hobby = req.Hobby
	}
	if req.Purpose != nil {
		s.Purpose = req.Purpose
	}
	if req.MotivationEntry != nil {
		s.MotivationEntry = req.MotivationEntry
	}
	if req.POB != nil {
		s.POB = req.POB
	}
	if req.DOB != nil {
		s.DOB = req.DOB
	}
	if req.Blood != nil {
		s.Blood = req.Blood
	}
	if req.Address != nil {
		s.Address = req.Address
	}
	if req.SubDistrict != nil {
		s.SubDistrict = req.SubDistrict
	}
	if req.District != nil {
		s.District = req.District
	}
	if req.Province != nil {
		s.Province = req.Province
	}
	if req.PostalCode != nil {
		s.PostalCode = req.PostalCode
	}
	if req.PreviousPondokName != nil {
		s.PreviousPondokName = req.PreviousPondokName
	}
	if req.PreviousPondokAddress != nil {
		s.PreviousPondokAddress = req.PreviousPondokAddress
	}
	if req.PreviousPondokDiv != nil {
		s.PreviousPondokDiv = req.PreviousPondokDiv
	}
	if req.PreviousPondokTime != nil {
		s.PreviousPondokTime = req.PreviousPondokTime
	}
	if req.NIK != nil {
		s.NIK = req.NIK
	}
	if req.NoKK != nil {
		s.NoKK = req.NoKK
	}
	if req.NISN != nil {
		s.NISN = req.NISN
	}
	if req.NoKIP != nil {
		s.NoKIP = req.NoKIP
	}
	if req.NoKKS != nil {
		s.NoKKS = req.NoKKS
	}
	if req.NoPKH != nil {
		s.NoPKH = req.NoPKH
	}
	if req.Workplace != nil {
		s.Workplace = req.Workplace
	}
	if req.Department != nil {
		s.Department = req.Department
	}
	if req.HomeStatus != nil {
		s.HomeStatus = req.HomeStatus
	}
	if req.Father != nil {
		s.Father = req.Father
	}
	if req.FatherPN != nil {
		s.FatherPN = req.FatherPN
	}
	if req.FatherNIK != nil {
		s.FatherNIK = req.FatherNIK
	}
	if req.FatherJob != nil {
		s.FatherJob = req.FatherJob
	}
	if req.FatherGraduate != nil {
		s.FatherGraduate = req.FatherGraduate
	}
	if req.FatherIncome != nil {
		s.FatherIncome = req.FatherIncome
	}
	if req.Mother != nil {
		s.Mother = req.Mother
	}
	if req.MotherPN != nil {
		s.MotherPN = req.MotherPN
	}
	if req.MotherNIK != nil {
		s.MotherNIK = req.MotherNIK
	}
	if req.MotherJob != nil {
		s.MotherJob = req.MotherJob
	}
	if req.MotherGraduate != nil {
		s.MotherGraduate = req.MotherGraduate
	}
	if req.MotherIncome != nil {
		s.MotherIncome = req.MotherIncome
	}
	if req.GuardianRelationship != nil {
		s.GuardianRelationship = req.GuardianRelationship
	}
	if req.Guardian != nil {
		s.Guardian = req.Guardian
	}
	if req.GuardianPN != nil {
		s.GuardianPN = req.GuardianPN
	}
	if req.GuardianNIK != nil {
		s.GuardianNIK = req.GuardianNIK
	}
	if req.GuardianJob != nil {
		s.GuardianJob = req.GuardianJob
	}
	if req.GuardianGraduate != nil {
		s.GuardianGraduate = req.GuardianGraduate
	}
	if req.GuardianIncome != nil {
		s.GuardianIncome = req.GuardianIncome
	}
}
