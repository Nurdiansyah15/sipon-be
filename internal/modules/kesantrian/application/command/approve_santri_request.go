package command

import (
	"context"
	"errors"
	"log/slog"

	"sipon-be/internal/modules/kesantrian/application"
	"sipon-be/internal/modules/kesantrian/application/dto"
	ports "sipon-be/internal/modules/kesantrian/application/ports"
	requestconstant "sipon-be/internal/modules/kesantrian/domain/request/constant"
	requestrepo "sipon-be/internal/modules/kesantrian/domain/request/repository"
	santriconstant "sipon-be/internal/modules/kesantrian/domain/santri/constant"
	santrientity "sipon-be/internal/modules/kesantrian/domain/santri/entity"
	santrirepo "sipon-be/internal/modules/kesantrian/domain/santri/repository"
	santrivo "sipon-be/internal/modules/kesantrian/domain/santri/valueobject"
	"sipon-be/internal/shared/kernel"

	"github.com/google/uuid"
)

// ApproveSantriRequestUseCase's two writes (request-update + santri-insert)
// are BOTH intra-module, so they run inside one real Transactor.WithTx —
// unlike sipon-api, which left this multi-write flow non-atomic. The
// post-commit identity.Contract.AddNISLoginIdentity call is best-effort and
// cannot be part of that transaction (cross-module).
type ApproveSantriRequestUseCase struct {
	requestRepo requestrepo.SantriRequestRepository
	santriRepo  santrirepo.SantriRepository
	provisioner ports.AccountProvisioner
	transactor  ports.Transactor
}

func NewApproveSantriRequestUseCase(
	requestRepo requestrepo.SantriRequestRepository,
	santriRepo santrirepo.SantriRepository,
	provisioner ports.AccountProvisioner,
	transactor ports.Transactor,
) *ApproveSantriRequestUseCase {
	return &ApproveSantriRequestUseCase{
		requestRepo: requestRepo,
		santriRepo:  santriRepo,
		provisioner: provisioner,
		transactor:  transactor,
	}
}

func (uc *ApproveSantriRequestUseCase) Execute(ctx context.Context, reviewerID, requestID string, req dto.ApproveSantriRequestRequest) (*dto.MessageResponse, error) {
	nis, err := santrivo.NewNIS(req.NIS)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}

	if _, err := uc.santriRepo.FindByNIS(ctx, nis.String()); err == nil {
		return nil, kernel.New(application.ErrCodeConflict)
	} else {
		var ke *kernel.AppError
		if !errors.As(err, &ke) || ke.Code != santriconstant.CodeSantriNotFound {
			return nil, kernel.Wrap(application.ErrCodeInternal, err)
		}
	}

	sr, err := uc.requestRepo.FindByID(ctx, requestID)
	if err != nil {
		return nil, application.WrapRepoErr(err, requestconstant.CodeSantriRequestNotFound)
	}

	if err := sr.Approve(reviewerID, nis.String()); err != nil {
		var ke *kernel.AppError
		if errors.As(err, &ke) && ke.Code == requestconstant.CodeSantriRequestInvalidStatus {
			return nil, kernel.New(application.ErrCodeConflict)
		}
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	santri, err := santrientity.NewSantri(uuid.NewString(), sr.UserID)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	santri.SetNIS(nis)

	if err := uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.requestRepo.Update(txCtx, sr); err != nil {
			return err
		}
		return uc.santriRepo.Save(txCtx, santri)
	}); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	if err := uc.provisioner.AddNISLoginIdentity(ctx, sr.UserID, nis.String()); err != nil {
		slog.Warn("kesantrian: best-effort NIS login identity sync to identity failed", "user_id", sr.UserID, "error", err)
	}

	return &dto.MessageResponse{Message: "permintaan santri berhasil disetujui"}, nil
}
