package command

import (
	"context"
	"path"
	"strings"
	"time"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	dokumenconstant "sipon-be/internal/modules/kesantrian/domain/dokumen/constant"
	dokumenentity "sipon-be/internal/modules/kesantrian/domain/dokumen/entity"
	dokumenrepo "sipon-be/internal/modules/kesantrian/domain/dokumen/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

var dokumenExtByContentType = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"application/pdf": ".pdf",
}

const dokumenPresignTTL = 10 * time.Minute

type DokumenPresignUseCase struct {
	fileUploader ports.FileUploader
}

func NewDokumenPresignUseCase(fileUploader ports.FileUploader) *DokumenPresignUseCase {
	return &DokumenPresignUseCase{fileUploader: fileUploader}
}

func (uc *DokumenPresignUseCase) Execute(ctx context.Context, req dto.DokumenPresignRequest) (*dto.DokumenPresignResponse, error) {
	ct := strings.ToLower(strings.TrimSpace(req.ContentType))
	if !dokumenconstant.AllowedContentTypes[ct] {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	kind := dokumenconstant.DokumenKind(req.Kind)
	if !dokumenconstant.ValidDokumenKinds[kind] {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	ext := dokumenExtByContentType[ct]
	objectName := path.Join("pending", "santri", "dokumen", string(kind), uuid.NewString()+ext)

	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, ct, dokumenPresignTTL, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		ExpiresIn:  int(dokumenPresignTTL.Seconds()),
	}, nil
}

type DokumenConfirmUseCase struct {
	santriRepo   santrirepo.SantriRepository
	dokumenRepo  dokumenrepo.SantriDokumenRepository
	fileUploader ports.FileUploader
	transactor   ports.Transactor
}

func NewDokumenConfirmUseCase(
	santriRepo santrirepo.SantriRepository,
	dokumenRepo dokumenrepo.SantriDokumenRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *DokumenConfirmUseCase {
	return &DokumenConfirmUseCase{
		santriRepo:   santriRepo,
		dokumenRepo:  dokumenRepo,
		fileUploader: fileUploader,
		transactor:   transactor,
	}
}

func (uc *DokumenConfirmUseCase) Execute(ctx context.Context, userID string, req dto.DokumenConfirmRequest) (*dto.DokumenConfirmResponse, error) {
	santri, err := uc.santriRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, application.WrapRepoErr(err, santriconstant.CodeSantriNotFound)
	}

	kind := dokumenconstant.DokumenKind(req.Kind)
	stagingKey := req.Key
	if !strings.HasPrefix(stagingKey, "pending/") {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	if err := uc.fileUploader.ConfirmUpload(ctx, stagingKey); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	finalKey := strings.TrimPrefix(stagingKey, "pending/")
	if err := uc.fileUploader.PromoteUpload(ctx, stagingKey, finalKey, ports.PrivacyPrivate); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	dokumen, err := dokumenentity.NewSantriDokumen(uuid.NewString(), santri.ID, kind, finalKey)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}
	dokumen.OriginalFilename = req.OriginalFilename
	dokumen.MimeType = req.MimeType
	dokumen.Size = req.Size

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.dokumenRepo.Save(txCtx, dokumen)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenConfirmResponse{
		ID:        dokumen.ID,
		Kind:      string(dokumen.Kind),
		Key:       dokumen.Key,
		Status:    string(dokumen.Status),
		CreatedAt: dokumen.CreatedAt,
	}, nil
}
