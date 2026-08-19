package command

import (
	"context"
	"encoding/json"
	"log/slog"

	"sipon-be/internal/modules/psb/application"
	"sipon-be/internal/modules/psb/application/dto"
	ports "sipon-be/internal/modules/psb/application/ports"
	dconstant "sipon-be/internal/modules/psb/domain/dokumen/constant"
	drepo "sipon-be/internal/modules/psb/domain/dokumen/repository"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	"sipon-be/internal/shared/kernel"
)

// notifyDokumenEvent best-effort mempublish event dokumen ke outbox. Recipient
// (user_id) diresolve lewat pendaftarRepo karena PendaftarDokumen hanya
// menyimpan pendaftar_id, bukan user_id secara langsung.
func notifyDokumenEvent(ctx context.Context, w ports.OutboxWriter, pendaftarRepo prepo.PendaftarRepository, routingKey, pendaftarID, dokumenKind string, notes *string) {
	if w == nil {
		return
	}
	p, err := pendaftarRepo.FindByID(ctx, pendaftarID)
	if err != nil {
		slog.Warn("psb: gagal resolve pendaftar untuk notifikasi dokumen", "pendaftar_id", pendaftarID, "error", err)
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"user_id":      p.UserID,
		"pendaftar_id": pendaftarID,
		"dokumen_kind": dokumenKind,
		"notes":        notes,
	})
	if err := w.Save(ctx, routingKey, payload); err != nil {
		slog.Warn("psb: gagal publish event dokumen", "routing_key", routingKey, "pendaftar_id", pendaftarID, "error", err)
	}
}

type DokumenVerifyUseCase struct {
	dokumenRepo   drepo.PendaftarDokumenRepository
	pendaftarRepo prepo.PendaftarRepository
	outboxWriter  ports.OutboxWriter
}

func NewDokumenVerifyUseCase(dokumenRepo drepo.PendaftarDokumenRepository, pendaftarRepo prepo.PendaftarRepository) *DokumenVerifyUseCase {
	return &DokumenVerifyUseCase{dokumenRepo: dokumenRepo, pendaftarRepo: pendaftarRepo}
}

// SetOutboxWriter memasang outbox writer untuk publikasi event verifikasi dokumen.
func (uc *DokumenVerifyUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *DokumenVerifyUseCase) Execute(ctx context.Context, verifierID, dokumenID string) (*dto.MessageResponse, error) {
	doc, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dconstant.CodeDokumenNotFound)
	}

	if err := doc.Verify(verifierID); err != nil {
		return nil, kernel.New(application.ErrCodeConflict)
	}

	if err := uc.dokumenRepo.Update(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	notifyDokumenEvent(ctx, uc.outboxWriter, uc.pendaftarRepo, RoutingDokumenVerified, doc.PendaftarID, string(doc.Kind), nil)

	return &dto.MessageResponse{Message: "dokumen diverifikasi"}, nil
}

type DokumenRejectUseCase struct {
	dokumenRepo   drepo.PendaftarDokumenRepository
	pendaftarRepo prepo.PendaftarRepository
	outboxWriter  ports.OutboxWriter
}

func NewDokumenRejectUseCase(dokumenRepo drepo.PendaftarDokumenRepository, pendaftarRepo prepo.PendaftarRepository) *DokumenRejectUseCase {
	return &DokumenRejectUseCase{dokumenRepo: dokumenRepo, pendaftarRepo: pendaftarRepo}
}

// SetOutboxWriter memasang outbox writer untuk publikasi event penolakan dokumen.
func (uc *DokumenRejectUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *DokumenRejectUseCase) Execute(ctx context.Context, verifierID, dokumenID string, notes *string) (*dto.MessageResponse, error) {
	doc, err := uc.dokumenRepo.FindByID(ctx, dokumenID)
	if err != nil {
		return nil, application.WrapRepoErr(err, dconstant.CodeDokumenNotFound)
	}

	doc.Reject(verifierID, notes)

	if err := uc.dokumenRepo.Update(ctx, doc); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	notifyDokumenEvent(ctx, uc.outboxWriter, uc.pendaftarRepo, RoutingDokumenRejected, doc.PendaftarID, string(doc.Kind), notes)

	return &dto.MessageResponse{Message: "dokumen ditolak"}, nil
}
