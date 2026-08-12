package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"sipon-be/internal/modules/fingerprint/application"
	"sipon-be/internal/modules/fingerprint/application/dto"
	"sipon-be/internal/modules/fingerprint/domain/scanlog/entity"
	"sipon-be/internal/modules/fingerprint/domain/scanlog/repository"
	"sipon-be/internal/shared/kernel"
)

const defaultSandboxSN = "SANDBOX-DEVICE-01"

// SimulateScanUseCase (sandbox) menulis satu scan palsu dengan skema yang
// sama dengan scan dari mesin fingerprint asli, supaya alur get-scan-info
// bisa dikembangkan & dites tanpa hardware fisik.
type SimulateScanUseCase struct {
	repo repository.ScanLogRepository
}

func NewSimulateScanUseCase(repo repository.ScanLogRepository) *SimulateScanUseCase {
	return &SimulateScanUseCase{repo: repo}
}

func (uc *SimulateScanUseCase) Execute(ctx context.Context, in dto.SimulateScanRequest) (*dto.ScanLogResponse, error) {
	if in.PIN == "" {
		return nil, kernel.WrapMsg(application.ErrCodeBadRequest, "pin wajib diisi", nil)
	}

	sn := defaultSandboxSN
	if in.SN != nil {
		sn = *in.SN
	}
	scanDate := time.Now()
	if in.ScanDate != nil {
		scanDate = *in.ScanDate
	}
	verifyMode := 0
	if in.VerifyMode != nil {
		verifyMode = *in.VerifyMode
	}
	inOutMode := 0
	if in.InOutMode != nil {
		inOutMode = *in.InOutMode
	}
	deviceIP := ""
	if in.DeviceIP != nil {
		deviceIP = *in.DeviceIP
	}

	log, err := entity.NewScanLog(uuid.NewString(), sn, in.PIN, deviceIP, scanDate, verifyMode, inOutMode)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeUnprocessableEntity, err)
	}
	if err := uc.repo.Insert(ctx, log); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return toScanLogResponse(log), nil
}

func toScanLogResponse(l *entity.ScanLog) *dto.ScanLogResponse {
	return &dto.ScanLogResponse{
		ID:         l.ID,
		SN:         l.SN,
		ScanDate:   l.ScanDate,
		PIN:        l.PIN,
		VerifyMode: l.VerifyMode,
		InOutMode:  l.InOutMode,
		DeviceIP:   l.DeviceIP,
		CreatedAt:  l.CreatedAt,
	}
}
