package command

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	ports "sipon-be/internal/modules/psb/application/ports"
	dconstant "sipon-be/internal/modules/psb/domain/dokumen/constant"
	dentity "sipon-be/internal/modules/psb/domain/dokumen/entity"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

type DokumenPresignUseCase struct {
	fileUploader ports.FileUploader
}

func NewDokumenPresignUseCase(fileUploader ports.FileUploader) *DokumenPresignUseCase {
	return &DokumenPresignUseCase{fileUploader: fileUploader}
}

func (uc *DokumenPresignUseCase) Execute(ctx context.Context, req dto.DokumenPresignRequest) (*dto.DokumenPresignResponse, error) {
	if !dconstant.AllowedContentTypes[req.ContentType] {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	ext := filepath.Ext(req.Filename)
	objectName := fmt.Sprintf("psb/dokumen/%s/%s/%s%s", req.Stage, req.Kind, uuid.NewString(), ext)
	expiry := 15 * time.Minute

	presignURL, key, publicURL, err := uc.fileUploader.RequestUpload(ctx, objectName, req.ContentType, expiry, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		PublicURL:  publicURL,
	}, nil
}

type DokumenConfirmUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	settingRepo   srepo.PsbSettingRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	fileUploader  ports.FileUploader
	transactor    ports.Transactor
}

func NewDokumenConfirmUseCase(
	pendaftarRepo prepo.PendaftarRepository,
	settingRepo srepo.PsbSettingRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *DokumenConfirmUseCase {
	return &DokumenConfirmUseCase{
		pendaftarRepo: pendaftarRepo,
		settingRepo:   settingRepo,
		dokumenRepo:   dokumenRepo,
		fileUploader:  fileUploader,
		transactor:    transactor,
	}
}

func (uc *DokumenConfirmUseCase) Execute(ctx context.Context, userID string, req dto.DokumenConfirmRequest) (*dto.DokumenConfirmResponse, error) {
	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, setting.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	stage := dconstant.DokumenStage(req.Stage)
	kind := dconstant.DokumenKind(req.Kind)

	existing, _ := uc.dokumenRepo.FindByPendaftarIDAndStage(ctx, p.ID, stage)
	for _, d := range existing {
		if d.Kind == kind && d.DeletedAt == nil {
			_ = uc.fileUploader.DeleteObject(ctx, d.Key, ports.PrivacyPrivate)
			d.SoftDelete()
			_ = uc.dokumenRepo.Update(ctx, d)
		}
	}

	docID := uuid.NewString()
	doc, err := dentity.NewPendaftarDokumen(docID, p.ID, stage, kind, req.Key)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeBadRequest, err)
	}

	if err := uc.fileUploader.ConfirmUpload(ctx, req.Key); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.dokumenRepo.Save(ctx, doc); err != nil {
		_ = uc.fileUploader.DeleteObject(ctx, req.Key, ports.PrivacyPrivate)
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenConfirmResponse{ID: docID}, nil
}

type DokumenDeleteUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	fileUploader  ports.FileUploader
	settingRepo   srepo.PsbSettingRepository
	transactor    ports.Transactor
}

func NewDokumenDeleteUseCase(
	pendaftarRepo prepo.PendaftarRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
	fileUploader ports.FileUploader,
	settingRepo srepo.PsbSettingRepository,
	transactor ports.Transactor,
) *DokumenDeleteUseCase {
	return &DokumenDeleteUseCase{
		pendaftarRepo: pendaftarRepo,
		dokumenRepo:   dokumenRepo,
		fileUploader:  fileUploader,
		settingRepo:   settingRepo,
		transactor:    transactor,
	}
}

func (uc *DokumenDeleteUseCase) Execute(ctx context.Context, userID, dokumenID string) (*dto.MessageResponse, error) {
	doc, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dconstant.CodeDokumenNotFound)
	}

	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, setting.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if doc.PendaftarID != p.ID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	if err := uc.fileUploader.DeleteObject(ctx, doc.Key, ports.PrivacyPrivate); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	doc.SoftDelete()
	if err := uc.dokumenRepo.Update(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "dokumen berhasil dihapus"}, nil
}
