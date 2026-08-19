package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	ports "sipon-be/internal/modules/psb/application/ports"
	dconstant "sipon-be/internal/modules/psb/domain/dokumen/constant"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	pconstant "sipon-be/internal/modules/psb/domain/pendaftar/constant"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	rconstant "sipon-be/internal/modules/psb/domain/review/constant"
	rentity "sipon-be/internal/modules/psb/domain/review/entity"
	rrepo "sipon-be/internal/modules/psb/domain/review/repository"
	"sipon-be/internal/shared/kernel"
)

type AdminReviewUseCase struct {
	pendaftarRepo prepo.PendaftarRepository
	reviewRepo    rrepo.PendaftarReviewRepository
	dokumenRepo   drepo.PendaftarDokumenRepository
	outboxWriter  ports.OutboxWriter
}

func NewAdminReviewUseCase(
	pendaftarRepo prepo.PendaftarRepository,
	reviewRepo rrepo.PendaftarReviewRepository,
	dokumenRepo drepo.PendaftarDokumenRepository,
) *AdminReviewUseCase {
	return &AdminReviewUseCase{
		pendaftarRepo: pendaftarRepo,
		reviewRepo:    reviewRepo,
		dokumenRepo:   dokumenRepo,
	}
}

// SetOutboxWriter memasang outbox writer untuk publikasi event review admin.
func (uc *AdminReviewUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *AdminReviewUseCase) publishEvent(ctx context.Context, routingKey, userID, pendaftarID string, notes *string) {
	if uc.outboxWriter == nil {
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"user_id":      userID,
		"pendaftar_id": pendaftarID,
		"notes":        notes,
	})
	if err := uc.outboxWriter.Save(ctx, routingKey, payload); err != nil {
		slog.Warn("psb: gagal publish event review", "routing_key", routingKey, "pendaftar_id", pendaftarID, "error", err)
	}
}

func (uc *AdminReviewUseCase) RequestRevision(ctx context.Context, pendaftarID, adminID string, notes *string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByID(ctx, pendaftarID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if err := p.RequestRevision(); err != nil {
		return nil, wrapInvalidStatus(err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	review, _ := rentity.NewPendaftarReview(uuid.NewString(), pendaftarID, rconstant.StagePendaftaran, rconstant.ActionPerluRevisi, adminID, notes)
	_ = uc.reviewRepo.Save(ctx, review)

	uc.publishEvent(ctx, RoutingRevisionRequested, p.UserID, pendaftarID, notes)

	return &dto.MessageResponse{Message: "permintaan revisi dikirim"}, nil
}

func (uc *AdminReviewUseCase) Reject(ctx context.Context, pendaftarID, adminID string, notes *string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByID(ctx, pendaftarID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if err := p.Reject(); err != nil {
		return nil, wrapInvalidStatus(err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	review, _ := rentity.NewPendaftarReview(uuid.NewString(), pendaftarID, rconstant.StagePendaftaran, rconstant.ActionDitolak, adminID, notes)
	_ = uc.reviewRepo.Save(ctx, review)

	uc.publishEvent(ctx, RoutingPendaftaranRejected, p.UserID, pendaftarID, notes)

	return &dto.MessageResponse{Message: "pendaftaran ditolak"}, nil
}

func (uc *AdminReviewUseCase) Accept(ctx context.Context, pendaftarID, adminID string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByID(ctx, pendaftarID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	docs, err := uc.dokumenRepo.FindByPendaftarIDAndStage(ctx, pendaftarID, dconstant.StagePendaftaran)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	for _, d := range docs {
		if d.Status != dconstant.DokumenStatusVerified {
			return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity,
				fmt.Errorf("dokumen %s belum diverifikasi", d.Kind))
		}
	}

	if err := p.Accept(adminID); err != nil {
		return nil, wrapInvalidStatus(err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	review, _ := rentity.NewPendaftarReview(uuid.NewString(), pendaftarID, rconstant.StagePendaftaran, rconstant.ActionDiterima, adminID, nil)
	_ = uc.reviewRepo.Save(ctx, review)

	uc.publishEvent(ctx, RoutingPendaftaranAccepted, p.UserID, pendaftarID, nil)

	return &dto.MessageResponse{Message: "pendaftaran diterima"}, nil
}

func (uc *AdminReviewUseCase) MarkNotReregistered(ctx context.Context, pendaftarID, adminID string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByID(ctx, pendaftarID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if err := p.MarkNotReregistered(); err != nil {
		return nil, wrapInvalidStatus(err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	return &dto.MessageResponse{Message: "pendaftar ditandai mengundurkan diri"}, nil
}

func (uc *AdminReviewUseCase) RequestRevisionDaftarUlang(ctx context.Context, pendaftarID, adminID string, notes *string) (*dto.MessageResponse, error) {
	p, err := uc.pendaftarRepo.FindByID(ctx, pendaftarID)
	if err != nil {
		return nil, application.WrapRepoErr(err, pconstant.CodePendaftarNotFound)
	}

	if err := p.RequestRevisionDaftarUlang(); err != nil {
		return nil, wrapInvalidStatus(err)
	}

	if err := uc.pendaftarRepo.Update(ctx, p); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	review, _ := rentity.NewPendaftarReview(uuid.NewString(), pendaftarID, rconstant.StageDaftarUlang, rconstant.ActionPerluRevisi, adminID, notes)
	_ = uc.reviewRepo.Save(ctx, review)

	uc.publishEvent(ctx, RoutingRevisionRequestedDaftarUlang, p.UserID, pendaftarID, notes)

	return &dto.MessageResponse{Message: "permintaan revisi daftar ulang dikirim"}, nil
}

func wrapInvalidStatus(err error) error {
	var ke *kernel.AppError
	if errors.As(err, &ke) && ke.Code == pconstant.CodePendaftarInvalidStatus {
		return kernel.New(application.ErrCodeConflict)
	}
	return kernel.Wrap(application.ErrCodeInternal, err)
}
