package query

import (
	"context"
	"time"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	docConst "sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

const herregDownloadExpiry = 5 * time.Minute

type DownloadHerregistrasiDocumentUseCase struct {
	kesantrianReader ports.KesantrianReader
	registrationRepo regRepo.SantriRegistrationRepository
	documentRepo     docRepo.HerregistrasiDocumentRepository
	fileUploader     ports.FileUploader
}

func NewDownloadHerregistrasiDocumentUseCase(
	kesantrianReader ports.KesantrianReader,
	registrationRepo regRepo.SantriRegistrationRepository,
	documentRepo docRepo.HerregistrasiDocumentRepository,
	fileUploader ports.FileUploader,
) *DownloadHerregistrasiDocumentUseCase {
	return &DownloadHerregistrasiDocumentUseCase{
		kesantrianReader: kesantrianReader,
		registrationRepo: registrationRepo,
		documentRepo:     documentRepo,
		fileUploader:     fileUploader,
	}
}

func (uc *DownloadHerregistrasiDocumentUseCase) Execute(ctx context.Context, userID, documentID string) (*dto.HerregistrasiDocumentDownloadResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	doc, err := uc.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return nil, application.WrapRepoErr(err, docConst.CodeHerregistrasiDocumentNotFound)
	}

	reg, err := uc.registrationRepo.FindByID(ctx, doc.SantriRegistrationID)
	if err != nil {
		return nil, application.WrapRepoErr(err, "SANTRI_REGISTRATION_NOT_FOUND")
	}
	if reg.SantriID != info.SantriID {
		return nil, kernel.New(application.ErrCodeForbidden)
	}

	url, err := uc.fileUploader.GeneratePresignedDownloadURL(ctx, doc.Key, herregDownloadExpiry, ports.PrivacyPrivate)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return &dto.HerregistrasiDocumentDownloadResponse{
		DownloadURL: url,
		ExpiresIn:   int(herregDownloadExpiry.Seconds()),
	}, nil
}
