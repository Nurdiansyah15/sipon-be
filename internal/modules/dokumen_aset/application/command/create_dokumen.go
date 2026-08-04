package command

import (
	"context"
	"path"
	"strings"
	"time"

	"sipon-be/internal/modules/dokumen_aset/application"
	"sipon-be/internal/modules/dokumen_aset/application/dto"
	ports "sipon-be/internal/modules/dokumen_aset/application/ports"
	constant "sipon-be/internal/modules/dokumen_aset/domain/dokumen/constant"
	entity "sipon-be/internal/modules/dokumen_aset/domain/dokumen/entity"
	repo "sipon-be/internal/modules/dokumen_aset/domain/dokumen/repository"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

var dokumenAsetExtByContentType = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"application/pdf": ".pdf",
	"application/msword":                                                     ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	"application/vnd.ms-excel":                                                ".xls",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       ".xlsx",
	"application/zip":                                                         ".zip",
}

const dokumenAsetPresignTTL = 10 * time.Minute

type CreateDokumenAsetPresignUseCase struct {
	fileUploader ports.FileUploader
}

func NewCreateDokumenAsetPresignUseCase(fileUploader ports.FileUploader) *CreateDokumenAsetPresignUseCase {
	return &CreateDokumenAsetPresignUseCase{fileUploader: fileUploader}
}

func (uc *CreateDokumenAsetPresignUseCase) Execute(ctx context.Context, req dto.DokumenAsetPresignRequest) (*dto.DokumenAsetPresignResponse, error) {
	ct := strings.ToLower(strings.TrimSpace(req.ContentType))
	if !constant.AllowedContentTypes[ct] {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	kategori := constant.Kategori(req.Kategori)
	if !constant.ValidKategoris[kategori] {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	privacy := ports.PrivacyPublic
	if !req.IsPublic {
		privacy = ports.PrivacyPrivate
	}

	ext := dokumenAsetExtByContentType[ct]
	objectName := path.Join("pending", "dokumen-aset", uuid.NewString()+ext)

	presignURL, key, _, err := uc.fileUploader.RequestUpload(ctx, objectName, ct, dokumenAsetPresignTTL, privacy)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenAsetPresignResponse{
		PresignURL: presignURL,
		Key:        key,
		ExpiresIn:  int(dokumenAsetPresignTTL.Seconds()),
	}, nil
}

type CreateDokumenAsetConfirmUseCase struct {
	dokumenRepo  repo.DokumenAsetRepository
	fileUploader ports.FileUploader
	transactor   ports.Transactor
}

func NewCreateDokumenAsetConfirmUseCase(
	dokumenRepo repo.DokumenAsetRepository,
	fileUploader ports.FileUploader,
	transactor ports.Transactor,
) *CreateDokumenAsetConfirmUseCase {
	return &CreateDokumenAsetConfirmUseCase{
		dokumenRepo:  dokumenRepo,
		fileUploader: fileUploader,
		transactor:   transactor,
	}
}

func (uc *CreateDokumenAsetConfirmUseCase) Execute(ctx context.Context, createdBy string, req dto.DokumenAsetConfirmRequest) (*dto.DokumenAsetConfirmResponse, error) {
	kategori := constant.Kategori(req.Kategori)
	if !constant.ValidKategoris[kategori] {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	stagingKey := req.Key
	if !strings.HasPrefix(stagingKey, "pending/") {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	privacy := ports.PrivacyPublic
	if !req.IsPublic {
		privacy = ports.PrivacyPrivate
	}

	if err := uc.fileUploader.ConfirmUpload(ctx, stagingKey); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	finalKey := strings.TrimPrefix(stagingKey, "pending/")
	if err := uc.fileUploader.PromoteUpload(ctx, stagingKey, finalKey, privacy); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	dokumen, err := entity.NewDokumenAset(
		uuid.NewString(),
		req.Judul,
		kategori,
		finalKey,
		req.OriginalFilename,
		req.MimeType,
		req.Size,
		req.IsPublic,
		createdBy,
	)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if req.Deskripsi != "" {
		dokumen.Deskripsi = &req.Deskripsi
	}

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return uc.dokumenRepo.Save(txCtx, dokumen)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.DokumenAsetConfirmResponse{
		ID:        dokumen.ID,
		Judul:     dokumen.Judul,
		Kategori:  string(dokumen.Kategori),
		Key:       dokumen.Key,
		CreatedAt: dokumen.CreatedAt,
	}, nil
}
