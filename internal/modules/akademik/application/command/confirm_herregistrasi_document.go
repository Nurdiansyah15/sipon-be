package command

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	periodRepo "sipon-be/internal/modules/akademik/domain/academic_period/repository"
	reqRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document_requirement/repository"
	docEntity "sipon-be/internal/modules/akademik/domain/herregistrasi_document/entity"
	docRepo "sipon-be/internal/modules/akademik/domain/herregistrasi_document/repository"
	regConst "sipon-be/internal/modules/akademik/domain/santri_registration/constant"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	"sipon-be/internal/shared/kernel"
)

type ConfirmHerregistrasiDocumentUseCase struct {
	kesantrianReader ports.KesantrianReader
	periodRepo       periodRepo.AcademicPeriodRepository
	registrationRepo regRepo.SantriRegistrationRepository
	requirementRepo  reqRepo.HerregistrasiDocumentRequirementRepository
	documentRepo     docRepo.HerregistrasiDocumentRepository
	fileUploader     ports.FileUploader
}

func NewConfirmHerregistrasiDocumentUseCase(
	kesantrianReader ports.KesantrianReader,
	periodRepo periodRepo.AcademicPeriodRepository,
	registrationRepo regRepo.SantriRegistrationRepository,
	requirementRepo reqRepo.HerregistrasiDocumentRequirementRepository,
	documentRepo docRepo.HerregistrasiDocumentRepository,
	fileUploader ports.FileUploader,
) *ConfirmHerregistrasiDocumentUseCase {
	return &ConfirmHerregistrasiDocumentUseCase{
		kesantrianReader: kesantrianReader,
		periodRepo:       periodRepo,
		registrationRepo: registrationRepo,
		requirementRepo:  requirementRepo,
		documentRepo:     documentRepo,
		fileUploader:     fileUploader,
	}
}

func (uc *ConfirmHerregistrasiDocumentUseCase) Execute(ctx context.Context, userID string, req dto.HerregistrasiDocumentConfirmRequest) (*dto.HerregistrasiDocumentResponse, error) {
	info, err := application.ResolveSantriByUserID(ctx, uc.kesantrianReader, userID)
	if err != nil {
		return nil, err
	}

	period, err := uc.periodRepo.FindOpen(ctx)
	if err != nil {
		return nil, application.WrapRepoErr(err, application.PeriodNotFoundCode)
	}

	reg, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, info.SantriID, period.ID)
	if err != nil {
		return nil, application.WrapRepoErr(err, regConst.CodeSantriRegistrationNotFound)
	}
	// Upload dokumen hanya diperbolehkan saat herreg masih draft, pending,
	// atau revisi.
	if reg.Status != regConst.SantriRegistrationStatusDraft &&
		reg.Status != regConst.SantriRegistrationStatusPending &&
		reg.Status != regConst.SantriRegistrationStatusRevision {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	// Validasi kind terdaftar di blueprint periode ini.
	valid := false
	if requirements, err := uc.requirementRepo.FindByAcademicPeriod(ctx, period.ID); err == nil {
		for _, requirement := range requirements {
			if requirement.Kind == req.Kind {
				valid = true
				break
			}
		}
	}
	if !valid {
		return nil, kernel.New(application.ErrCodeUnprocessableEntity)
	}

	if !strings.HasPrefix(req.Key, "pending/") {
		return nil, kernel.New(application.ErrCodeBadRequest)
	}

	if err := uc.fileUploader.ConfirmUpload(ctx, req.Key); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	finalKey := strings.TrimPrefix(req.Key, "pending/")
	if err := uc.fileUploader.PromoteUpload(ctx, req.Key, finalKey, ports.PrivacyPrivate); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	// Soft-delete dokumen lama untuk (registration, kind) yang sama.
	if existing, err := uc.documentRepo.FindByRegistrationAndKind(ctx, reg.ID, req.Kind); err == nil {
		existing.SoftDelete()
		if err := uc.documentRepo.Update(ctx, existing); err != nil {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	doc, err := docEntity.NewHerregistrasiDocument(uuid.NewString(), reg.ID, req.Kind, finalKey)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	doc.SetMetadata(req.OriginalFilename, req.MimeType, req.Size)
	if err := uc.documentRepo.Save(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return MapHerregistrasiDocumentToResponse(doc, ""), nil
}
