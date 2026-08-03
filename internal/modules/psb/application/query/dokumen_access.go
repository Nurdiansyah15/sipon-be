package query

import (
	"context"
	"time"

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

const dokumenAccessTTL = 15 * time.Minute

type DokumenAccessUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	settingRepo   srepo.PsbSettingRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	fileUploader  ports.FileUploader
}

func NewDokumenAccessUseCase(
	pendaftarRepo prepo.PendaftarRepository,
	settingRepo srepo.PsbSettingRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
	fileUploader ports.FileUploader,
) *DokumenAccessUseCase {
	return &DokumenAccessUseCase{
		pendaftarRepo: pendaftarRepo,
		settingRepo:   settingRepo,
		dokumenRepo:   dokumenRepo,
		fileUploader:  fileUploader,
	}
}

// Execute generates a short-lived preview URL for a document owned by the
// requesting user's active pendaftaran.
func (uc *DokumenAccessUseCase) Execute(ctx context.Context, userID, dokumenID string) (*dto.DokumenAccessResponse, error) {
	setting, err := uc.settingRepo.FindActive(ctx)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeNotFound, err)
	}

	p, err := uc.pendaftarRepo.FindByUserIDAndSetting(ctx, userID, setting.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	doc, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dconstant.CodeDokumenNotFound)
	}

	if doc.PendaftarID != p.ID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	return uc.presign(ctx, doc.Key)
}

// ExecuteAdmin generates a preview URL for any document; callers must already
// be gated by an admin permission middleware.
func (uc *DokumenAccessUseCase) ExecuteAdmin(ctx context.Context, dokumenID string) (*dto.DokumenAccessResponse, error) {
	doc, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dconstant.CodeDokumenNotFound)
	}

	return uc.presign(ctx, doc.Key)
}

func (uc *DokumenAccessUseCase) presign(ctx context.Context, key string) (*dto.DokumenAccessResponse, error) {
	url, err := uc.fileUploader.GeneratePresignedDownloadURL(ctx, key, dokumenAccessTTL, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenAccessResponse{
		AccessURL: url,
		ExpiresIn: int(dokumenAccessTTL.Seconds()),
	}, nil
}
