package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	ports "sipon-be/internal/modules/psb/application/ports"
	dconstant "sipon-be/internal/modules/psb/domain/dokumen/constant"
	dentity "sipon-be/internal/modules/psb/domain/dokumen/entity"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	pentity "sipon-be/internal/modules/psb/domain/pendaftar/entity"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	"sipon-be/internal/shared/kernel"
)

type UpsertFormulirUseCase struct {
	settingRepo   srepo.PsbSettingRepository
	pendaftarRepo prepo.PendaftarRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	fileUploader  ports.FileUploader
	kesantrian    ports.KesantrianProvisioner
}

func NewUpsertFormulirUseCase(
	settingRepo srepo.PsbSettingRepository,
	pendaftarRepo prepo.PendaftarRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
	fileUploader ports.FileUploader,
	kesantrian ports.KesantrianProvisioner,
) *UpsertFormulirUseCase {
	return &UpsertFormulirUseCase{
		settingRepo:   settingRepo,
		pendaftarRepo: pendaftarRepo,
		dokumenRepo:   dokumenRepo,
		fileUploader:  fileUploader,
		kesantrian:    kesantrian,
	}
}

func (uc *UpsertFormulirUseCase) Execute(ctx context.Context, userID string, req dto.UpsertFormulirRequest) (*dto.PendaftarResponse, error) {
	if err := uc.ensureNotSantri(ctx, userID); err != nil {
		return nil, err
	}

	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	isNew := false
	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, setting.ID)
	if err != nil {
		p = nil
	}

	if p == nil {
		isNew = true
		p, err = pentity.NewPendaftar(uuid.NewString(), userID, setting.ID, "1", req.Program)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		p.ProgramID = req.ProgramID
		noRegis, err := generateNoRegis(ctx, uc.pendaftarRepo)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
		p.NoRegis = &noRegis
	}

	err = p.UpsertFormulir(func(p *pentity.Pendaftar) {
		p.Gender = "1"
		p.Program = req.Program
		p.ProgramID = req.ProgramID
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

	if isNew {
		if err := uc.pendaftarRepo.Save(ctx, p); err != nil {
			return nil, application.WrapConflictErr(err, pconstant.CodePendaftarDuplicate)
		}
	} else {
		if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	for _, d := range req.Dokumen {
		stagingKey := d.Key
		if !strings.HasPrefix(stagingKey, "pending/") {
			slog.Warn("psb: formulir dokumen skip — key bukan staging prefix", "key", stagingKey)
			continue
		}

		if err := uc.fileUploader.ConfirmUpload(ctx, stagingKey); err != nil {
			slog.Warn("psb: formulir dokumen confirm gagal", "key", stagingKey, "error", err)
			continue
		}

		finalKey := strings.TrimPrefix(stagingKey, "pending/")
		if err := uc.fileUploader.PromoteUpload(ctx, stagingKey, finalKey, ports.PrivacyPrivate); err != nil {
			slog.Warn("psb: formulir dokumen promote gagal", "key", stagingKey, "error", err)
			continue
		}

		stage := dconstant.DokumenStage(d.Stage)
		kind := dconstant.DokumenKind(d.Kind)

		existing, _ := uc.dokumenRepo.FindByPendaftarIDAndStage(ctx, p.ID, stage)
		for _, ed := range existing {
			if ed.Kind == kind && ed.DeletedAt == nil {
				if err := uc.fileUploader.DeleteObject(ctx, ed.Key, ports.PrivacyPrivate); err != nil {
					slog.Warn("psb: best-effort hapus dokumen lama gagal", "key", ed.Key, "error", err)
				}
				ed.SoftDelete()
				if err := uc.dokumenRepo.Update(ctx, ed); err != nil {
					slog.Warn("psb: gagal update soft-delete dokumen lama", "id", ed.ID, "error", err)
				}
			}
		}

		docID := uuid.NewString()
		doc, err := dentity.NewPendaftarDokumen(docID, p.ID, stage, kind, finalKey)
		if err != nil {
			return nil, kernel.Wrap(application.ErrCodeBadRequest, err)
		}

		if err := uc.dokumenRepo.Save(ctx, doc); err != nil {
			_ = uc.fileUploader.DeleteObject(ctx, finalKey, ports.PrivacyPrivate)
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	return mapPendaftarToResponse(p), nil
}

func mapPendaftarToResponse(p *pentity.Pendaftar) *dto.PendaftarResponse {
	return &dto.PendaftarResponse{
		ID: p.ID, UserID: p.UserID, PsbSettingID: p.PsbSettingID,
		Gender: p.Gender, Program: p.Program, ProgramID: p.ProgramID,
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
		SantriID: p.SantriID, NIS: p.NIS, NoRegis: p.NoRegis,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func generateNoRegis(ctx context.Context, repo prepo.PendaftarRepository) (string, error) {
	latest, err := repo.FindLatestNoRegis(ctx)
	if err != nil {
		return "", err
	}

	year := time.Now().Format("06")
	prefix := "P1000" + year

	if latest == nil {
		return prefix + "000", nil
	}

	regYear := (*latest)[5:7]
	seqNum := 0
	if regYear <= year {
		seqStr := (*latest)[7:10]
		s, err := strconv.Atoi(seqStr)
		if err == nil {
			seqNum = s + 1
		}
	}

	return prefix + fmt.Sprintf("%03d", seqNum), nil
}

// ensureNotSantri menolak pengisian formulir jika user sudah terdaftar
// sebagai santri aktif (status SANTRI). Santri tidak boleh mendaftar lagi
// melalui PSB.
func (uc *UpsertFormulirUseCase) ensureNotSantri(ctx context.Context, userID string) error {
	if uc.kesantrian == nil {
		return nil
	}
	info, err := uc.kesantrian.GetSantriByUserID(ctx, userID)
	if err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == santriconstant.CodeSantriNotFound {
			return nil
		}
		return kernel.WrapMsg(application.ErrCodeInternal, "terjadi kesalahan internal", err)
	}
	if info != nil && info.Status == "SANTRI" {
		return kernel.WrapMsg(application.ErrCodeConflict, "Anda sudah terdaftar sebagai santri dan tidak dapat mengisi pendaftaran baru", nil)
	}
	return nil
}
