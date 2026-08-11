package command

import (
	"context"
	"errors"
	"log/slog"
	"strings"

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
	"sipon-be/internal/shared/kernel"
)

type PendaftarActionUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	fileUploader  ports.FileUploader
}

func NewPendaftarActionUseCase(
	pendaftarRepo prepo.PendaftarRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
	fileUploader ports.FileUploader,
) *PendaftarActionUseCase {
	return &PendaftarActionUseCase{
		pendaftarRepo: pendaftarRepo,
		dokumenRepo:   dokumenRepo,
		fileUploader:  fileUploader,
	}
}

func (uc *PendaftarActionUseCase) SubmitPendaftaran(ctx context.Context, userID, settingID string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, settingID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if err := validateProgramRequired(p); err != nil {
		return nil, err
	}

	if err := p.Submit(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == pconstant.CodePendaftarInvalidStatus {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "pendaftaran berhasil diajukan"}, nil
}

func (uc *PendaftarActionUseCase) SubmitDaftarUlang(ctx context.Context, userID, settingID string, dokumen []dto.FormulirDokumenItem) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, settingID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	for _, d := range dokumen {
		stagingKey := d.Key
		if !strings.HasPrefix(stagingKey, "pending/") {
			slog.Warn("psb: daftar ulang dokumen skip — key bukan staging prefix", "key", stagingKey)
			continue
		}

		if err := uc.fileUploader.ConfirmUpload(ctx, stagingKey); err != nil {
			slog.Warn("psb: daftar ulang dokumen confirm gagal", "key", stagingKey, "error", err)
			continue
		}

		finalKey := strings.TrimPrefix(stagingKey, "pending/")
		if err := uc.fileUploader.PromoteUpload(ctx, stagingKey, finalKey, ports.PrivacyPrivate); err != nil {
			slog.Warn("psb: daftar ulang dokumen promote gagal", "key", stagingKey, "error", err)
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

	if err := p.SubmitDaftarUlang(); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == pconstant.CodePendaftarInvalidStatus {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "daftar ulang berhasil diajukan"}, nil
}

// validateProgramRequired memastikan pendaftar memilih program sebelum
// mengajukan pendaftaran. ProgramID (referensi ke programs.id akademik)
// menjadi sumber kebenaran; field Program (string) tetap dipertahankan
// sebagai cache display.
func validateProgramRequired(p *pentity.Pendaftar) error {
	if p.ProgramID == nil || *p.ProgramID == "" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "program wajib dipilih sebelum mengajukan pendaftaran", nil)
	}
	return nil
}
