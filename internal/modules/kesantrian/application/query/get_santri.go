package query

import (
	"context"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrientity "sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
)

type GetSantriUseCase struct {
	santriRepo   santrirepo.SantriRepository
	provisioner  ports.AccountProvisioner
	fileUploader ports.FileUploader
}

func NewGetSantriUseCase(santriRepo santrirepo.SantriRepository, provisioner ports.AccountProvisioner, fileUploader ports.FileUploader) *GetSantriUseCase {
	return &GetSantriUseCase{santriRepo: santriRepo, provisioner: provisioner, fileUploader: fileUploader}
}

func (uc *GetSantriUseCase) Execute(ctx context.Context, userID string) (*dto.GetSantriResponse, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	summary, err := uc.provisioner.GetUserSummary(ctx, userID)
	if err != nil {
		return nil, application.WrapRepoErr(err, application.ErrCodeNotFound)
	}

	var avatarURL *string
	if summary.AvatarKey != nil && *summary.AvatarKey != "" {
		u := uc.fileUploader.PublicURL(*summary.AvatarKey)
		avatarURL = &u
	}

	return mapSantriToResponse(santri, summary.Username, summary.Email, summary.Fullname, avatarURL), nil
}

func mapSantriToResponse(s *santrientity.Santri, username, email string, fullname, avatarURL *string) *dto.GetSantriResponse {
	resp := &dto.GetSantriResponse{
		ID:        s.ID,
		UserID:    s.UserID,
		Username:  username,
		Email:     email,
		Fullname:  fullname,
		AvatarURL: avatarURL,

		Nickname:        s.Nickname,
		Program:         s.Program,
		Option:          s.Option,
		Hobby:           s.Hobby,
		Purpose:         s.Purpose,
		MotivationEntry: s.MotivationEntry,
		POB:             s.POB,
		DOB:             s.DOB,
		Blood:           s.Blood,

		Address:     s.Address,
		SubDistrict: s.SubDistrict,
		District:    s.District,
		Province:    s.Province,
		PostalCode:  s.PostalCode,

		PreviousPondokName:    s.PreviousPondokName,
		PreviousPondokAddress: s.PreviousPondokAddress,
		PreviousPondokDiv:     s.PreviousPondokDiv,
		PreviousPondokTime:    s.PreviousPondokTime,

		NIK:   s.NIK,
		NoKK:  s.NoKK,
		NISN:  s.NISN,
		NoKIP: s.NoKIP,
		NoKKS: s.NoKKS,
		NoPKH: s.NoPKH,

		Workplace:  s.Workplace,
		Department: s.Department,

		HomeStatus: s.HomeStatus,

		Father:         s.Father,
		FatherPN:       s.FatherPN,
		FatherNIK:      s.FatherNIK,
		FatherJob:      s.FatherJob,
		FatherGraduate: s.FatherGraduate,
		FatherIncome:   s.FatherIncome,

		Mother:         s.Mother,
		MotherPN:       s.MotherPN,
		MotherNIK:      s.MotherNIK,
		MotherJob:      s.MotherJob,
		MotherGraduate: s.MotherGraduate,
		MotherIncome:   s.MotherIncome,

		GuardianRelationship: s.GuardianRelationship,
		Guardian:             s.Guardian,
		GuardianPN:           s.GuardianPN,
		GuardianNIK:          s.GuardianNIK,
		GuardianJob:          s.GuardianJob,
		GuardianGraduate:     s.GuardianGraduate,
		GuardianIncome:       s.GuardianIncome,

		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
	if s.NIS != nil {
		nis := s.NIS.String()
		resp.NIS = &nis
	}
	status := string(s.Status)
	resp.Status = &status
	return resp
}
