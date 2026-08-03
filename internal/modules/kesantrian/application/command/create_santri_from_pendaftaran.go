package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/kesantrian/application"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	dokumenconstant "sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	dokumenentity "sipon-be/internal/modules/kesantrian/domain/dokumen/entity"
	dokumenrepo "sipon-be/internal/modules/kesantrian/domain/dokumen/repository"
	santrientity "sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	santrivo "sipon-be/internal/modules/kesantrian/domain/santri/valueobject"
	"sipon-be/internal/shared/kernel"
)

type PendaftaranDokumenInput struct {
	Kind             string
	Key              string
	OriginalFilename *string
	MimeType         *string
	Size             *int64
	VerifiedBy       *string
	VerifiedAt       *time.Time
}

type CreateSantriFromPendaftaranCmd struct {
	UserID    string
	Gender    string
	EntryYear string

	Nickname        *string
	Program         *string
	Hobby           *string
	Purpose         *string
	MotivationEntry *string
	POB             *string
	DOB             *time.Time
	Blood           *string

	Address     *string
	SubDistrict *string
	District    *string
	Province    *string
	PostalCode  *string

	PreviousPondokName    *string
	PreviousPondokAddress *string
	PreviousPondokDiv     *string
	PreviousPondokTime    *string

	NIK   *string
	NoKK  *string
	NISN  *string
	NoKIP *string
	NoKKS *string
	NoPKH *string

	Workplace  *string
	Department *string

	HomeStatus *string

	Father         *string
	FatherPN       *string
	FatherNIK      *string
	FatherJob      *string
	FatherGraduate *string
	FatherIncome   *string

	Mother         *string
	MotherPN       *string
	MotherNIK      *string
	MotherJob      *string
	MotherGraduate *string
	MotherIncome   *string

	GuardianRelationship *string
	Guardian             *string
	GuardianPN           *string
	GuardianNIK          *string
	GuardianJob          *string
	GuardianGraduate     *string
	GuardianIncome       *string

	Dokumen []PendaftaranDokumenInput
}

type CreateSantriFromPendaftaranResult struct {
	SantriID string
	NIS      string
}

type CreateSantriFromPendaftaranUseCase struct {
	santriRepo  santrirepo.SantriRepository
	dokumenRepo dokumenrepo.SantriDokumenRepository
	provisioner ports.AccountProvisioner
	transactor  ports.Transactor
}

func NewCreateSantriFromPendaftaranUseCase(
	santriRepo santrirepo.SantriRepository,
	dokumenRepo dokumenrepo.SantriDokumenRepository,
	provisioner ports.AccountProvisioner,
	transactor ports.Transactor,
) *CreateSantriFromPendaftaranUseCase {
	return &CreateSantriFromPendaftaranUseCase{
		santriRepo:  santriRepo,
		dokumenRepo: dokumenRepo,
		provisioner: provisioner,
		transactor:  transactor,
	}
}

func (uc *CreateSantriFromPendaftaranUseCase) Execute(ctx context.Context, cmd CreateSantriFromPendaftaranCmd) (*CreateSantriFromPendaftaranResult, error) {
	prefix := "1000" + cmd.Gender + cmd.EntryYear

	var santriID, nis string

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		seq, err := uc.santriRepo.FindMaxSequence(txCtx, prefix)
		if err != nil {
			return err
		}
		seq++

		nis = fmt.Sprintf("%s%03d", prefix, seq)

		nisVO, err := santrivo.NewNIS(nis)
		if err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}

		santriID = uuid.NewString()
		santri, err := santrientity.NewSantri(santriID, cmd.UserID)
		if err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
		santri.SetNIS(nisVO)

		applySantriFromPendaftaranCmd(santri, cmd)

		if err := uc.santriRepo.Save(txCtx, santri); err != nil {
			return err
		}

		for _, d := range cmd.Dokumen {
			docID := uuid.NewString()
			doc, err := dokumenentity.NewSantriDokumen(docID, santriID, dokumenconstant.DokumenKind(d.Kind), d.Key)
			if err != nil {
				return err
			}

			if d.VerifiedBy != nil && d.VerifiedAt != nil {
				doc.Status = dokumenconstant.DokumenStatusVerified
				doc.VerifiedBy = d.VerifiedBy
				doc.VerifiedAt = d.VerifiedAt
			}

			doc.OriginalFilename = d.OriginalFilename
			doc.MimeType = d.MimeType
			doc.Size = d.Size

			if err := uc.dokumenRepo.Save(txCtx, doc); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.provisioner.AddNISLoginIdentity(ctx, cmd.UserID, nis); err != nil {
		slog.Warn("kesantrian: best-effort NIS login identity sync to identity failed", "user_id", cmd.UserID, "error", err)
	}

	return &CreateSantriFromPendaftaranResult{
		SantriID: santriID,
		NIS:      nis,
	}, nil
}

func applySantriFromPendaftaranCmd(s *santrientity.Santri, cmd CreateSantriFromPendaftaranCmd) {
	s.Nickname = cmd.Nickname
	s.Program = cmd.Program
	s.Hobby = cmd.Hobby
	s.Purpose = cmd.Purpose
	s.MotivationEntry = cmd.MotivationEntry
	s.POB = cmd.POB
	s.DOB = cmd.DOB
	s.Blood = cmd.Blood

	s.Address = cmd.Address
	s.SubDistrict = cmd.SubDistrict
	s.District = cmd.District
	s.Province = cmd.Province
	s.PostalCode = cmd.PostalCode

	s.PreviousPondokName = cmd.PreviousPondokName
	s.PreviousPondokAddress = cmd.PreviousPondokAddress
	s.PreviousPondokDiv = cmd.PreviousPondokDiv
	s.PreviousPondokTime = cmd.PreviousPondokTime

	s.NIK = cmd.NIK
	s.NoKK = cmd.NoKK
	s.NISN = cmd.NISN
	s.NoKIP = cmd.NoKIP
	s.NoKKS = cmd.NoKKS
	s.NoPKH = cmd.NoPKH

	s.Workplace = cmd.Workplace
	s.Department = cmd.Department

	s.HomeStatus = cmd.HomeStatus

	s.Father = cmd.Father
	s.FatherPN = cmd.FatherPN
	s.FatherNIK = cmd.FatherNIK
	s.FatherJob = cmd.FatherJob
	s.FatherGraduate = cmd.FatherGraduate
	s.FatherIncome = cmd.FatherIncome

	s.Mother = cmd.Mother
	s.MotherPN = cmd.MotherPN
	s.MotherNIK = cmd.MotherNIK
	s.MotherJob = cmd.MotherJob
	s.MotherGraduate = cmd.MotherGraduate
	s.MotherIncome = cmd.MotherIncome

	s.GuardianRelationship = cmd.GuardianRelationship
	s.Guardian = cmd.Guardian
	s.GuardianPN = cmd.GuardianPN
	s.GuardianNIK = cmd.GuardianNIK
	s.GuardianJob = cmd.GuardianJob
	s.GuardianGraduate = cmd.GuardianGraduate
	s.GuardianIncome = cmd.GuardianIncome
}
