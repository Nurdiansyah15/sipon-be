package command

import (
	"context"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/ports"
	docConst "sipon-be/internal/modules/akademik/domain/herregistrasi_document/constant"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type DeleteHerregistrasiDocumentUseCase struct {
	kesantrianReader ports.KesantrianReader
	registrationRepo regRepo.SantriRegistrationRepository
	documentRepo     docRepo.HerregistrasiDocumentRepository
	fileUploader     ports.FileUploader
}

func NewDeleteHerregistrasiDocumentUseCase(
	kesantrianReader ports.KesantrianReader,
	registrationRepo regRepo.SantriRegistrationRepository,
	documentRepo docRepo.HerregistrasiDocumentRepository,
	fileUploader ports.FileUploader,
) *DeleteHerregistrasiDocumentUseCase {
	return &DeleteHerregistrasiDocumentUseCase{
		kesantrianReader: kesantrianReader,
		registrationRepo: registrationRepo,
		documentRepo:     documentRepo,
		fileUploader:     fileUploader,
	}
}

func (uc *DeleteHerregistrasiDocumentUseCase) Execute(ctx context.Context, userID, documentID string) error {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return err
	}

	doc, err := uc.documentRepo.FindByID(ctx, documentID)
	if err != nil {
		return application.WrapRepoErr(err, docConst.CodeHerregistrasiDocumentNotFound)
	}

	reg, err := uc.registrationRepo.FindByID(ctx, doc.SantriRegistrationID)
	if err != nil {
		return application.WrapRepoErr(err, "SANTRI_REGISTRATION_NOT_FOUND")
	}
	if reg.SantriID != info.SantriID {
		return kernel.New(application.ErrCodeForbidden)
	}
	// Hanya boleh dihapus jika belum diverifikasi.
	if doc.Status == docConst.HerregistrasiDocumentStatusVerified {
		return kernel.New(application.ErrCodeConflict)
	}

	if err := uc.fileUploader.DeleteObject(ctx, doc.Key, ports.PrivacyPrivate); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	doc.SoftDelete()
	if err := uc.documentRepo.Update(ctx, doc); err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	return nil
}
