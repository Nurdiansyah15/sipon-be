package command

import (
	"context"
	"fmt"
	"time"

	"sipon-be/internal/modules/kesantrian"
	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	ports "sipon-be/internal/modules/psb/application/ports"
	dconstant "sipon-be/internal/modules/psb/domain/dokumen/constant"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type GenerateNISUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	settingRepo   srepo.PsbSettingRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	kesantrian    ports.KesantrianProvisioner
}

func NewGenerateNISUseCase(
	pendaftarRepo prepo.PendaftarRepository,
	settingRepo srepo.PsbSettingRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
	kesantrian ports.KesantrianProvisioner,
) *GenerateNISUseCase {
	return &GenerateNISUseCase{
		pendaftarRepo: pendaftarRepo,
		settingRepo:   settingRepo,
		dokumenRepo:   dokumenRepo,
		kesantrian:    kesantrian,
	}
}

func (uc *GenerateNISUseCase) Execute(ctx context.Context, pendaftarID, adminID string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByID(ctx, pendaftarID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	docs, err := uc.dokumenRepo.FindByPendaftarIDAndStage(ctx, pendaftarID, dconstant.StageDaftarUlang)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	for _, d := range docs {
		if d.Status != dconstant.DokumenStatusVerified {
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity,
				fmt.Errorf("dokumen %s belum diverifikasi", d.Kind))
		}
	}

	entryYear := time.Now().Format("06")

	dok := make([]kesantrian.SantriDokumenInput, len(docs))
	for i, d := range docs {
		verifiedBy := ""
		if d.VerifiedBy != nil {
			verifiedBy = *d.VerifiedBy
		}
		dok[i] = kesantrian.SantriDokumenInput{
			Kind:             string(d.Kind),
			Key:              d.Key,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			Size:             d.Size,
			VerifiedBy:       &verifiedBy,
			VerifiedAt:       d.VerifiedAt,
		}
	}

	result, err := uc.kesantrian.CreateSantriFromPendaftaran(ctx, kesantrian.CreateSantriFromPendaftaranInput{
		UserID:    p.UserID,
		Gender:    p.Gender,
		EntryYear: entryYear,
		Nickname:              p.Nickname,
		Program:               p.Program,
		Hobby:                 p.Hobby,
		Purpose:               p.Purpose,
		MotivationEntry:       p.MotivationEntry,
		POB:                   p.POB,
		DOB:                   p.DOB,
		Blood:                 p.Blood,
		Address:               p.Address,
		SubDistrict:           p.SubDistrict,
		District:              p.District,
		Province:              p.Province,
		PostalCode:            p.PostalCode,
		PreviousPondokName:    p.PreviousPondokName,
		PreviousPondokAddress: p.PreviousPondokAddress,
		PreviousPondokDiv:     p.PreviousPondokDiv,
		PreviousPondokTime:    p.PreviousPondokTime,
		NIK:                   p.NIK,
		NoKK:                  p.NoKK,
		NISN:                  p.NISN,
		NoKIP:                 p.NoKIP,
		NoKKS:                 p.NoKKS,
		NoPKH:                 p.NoPKH,
		Workplace:             p.Workplace,
		Department:            p.Department,
		HomeStatus:            p.HomeStatus,
		Father:                p.Father,
		FatherPN:              p.FatherPN,
		FatherNIK:             p.FatherNIK,
		FatherJob:             p.FatherJob,
		FatherGraduate:        p.FatherGraduate,
		FatherIncome:          p.FatherIncome,
		Mother:                p.Mother,
		MotherPN:              p.MotherPN,
		MotherNIK:             p.MotherNIK,
		MotherJob:             p.MotherJob,
		MotherGraduate:        p.MotherGraduate,
		MotherIncome:          p.MotherIncome,
		GuardianRelationship:  p.GuardianRelationship,
		Guardian:              p.Guardian,
		GuardianPN:            p.GuardianPN,
		GuardianNIK:           p.GuardianNIK,
		GuardianJob:           p.GuardianJob,
		GuardianGraduate:      p.GuardianGraduate,
		GuardianIncome:        p.GuardianIncome,
		Dokumen:               dok,
	})
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := p.GenerateNIS(result.SantriID, result.NIS); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "NIS berhasil digenerate, santri telah dibuat"}, nil
}
