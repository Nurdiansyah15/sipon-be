package command

import (
	"context"
	"errors"
	"time"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	"sipon-be/internal/shared/kernel"
)

// SyncAttendanceFromFingerprintUseCase menarik scan mesin fingerprint dalam
// rentang waktu sesi, lalu mencatatnya sebagai kehadiran lewat CheckinByNISUseCase
// yang sudah ada — sehingga semua validasi eligibility (sesi open, santri aktif,
// herreg completed, program membership, cek duplikat) dipakai persis sama dengan
// check-in manual by NIS, tanpa duplikasi logic.
type SyncAttendanceFromFingerprintUseCase struct {
	sessionRepo       sesRepo.ActivitySessionRepository
	fingerprintReader ports.FingerprintReader
	checkin           *CheckinByNISUseCase
}

func NewSyncAttendanceFromFingerprintUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	fingerprintReader ports.FingerprintReader,
	checkin *CheckinByNISUseCase,
) *SyncAttendanceFromFingerprintUseCase {
	return &SyncAttendanceFromFingerprintUseCase{
		sessionRepo:       sessionRepo,
		fingerprintReader: fingerprintReader,
		checkin:           checkin,
	}
}

func (uc *SyncAttendanceFromFingerprintUseCase) Execute(ctx context.Context, sessionID string) (*dto.SyncFingerprintResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, application.WrapRepoErr(err, sesConst.CodeActivitySessionNotFound)
	}
	if session.Status != "open" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "sesi tidak terbuka", nil)
	}

	to := time.Now()
	if session.EndsAt.Before(to) {
		to = session.EndsAt // jangan ambil scan setelah sesi selesai
	}

	scans, err := uc.fingerprintReader.ListDistinctPinInRange(ctx, session.StartsAt, to)
	if err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}

	resp := &dto.SyncFingerprintResponse{TotalScans: len(scans)}
	for _, scan := range scans {
		_, err := uc.checkin.Execute(ctx, sessionID, scan.PIN)
		switch {
		case err == nil:
			resp.Recorded++
		case isFingerprintConflict(err):
			resp.Skipped++
		default:
			resp.Errors = append(resp.Errors, dto.SyncFingerprintError{PIN: scan.PIN, Reason: err.Error()})
		}
	}
	return resp, nil
}

// isFingerprintConflict membedakan "NIS sudah tercatat" (skip, bukan error)
// dari error lain, pakai pola errors.As yang sudah dipakai di codebase.
func isFingerprintConflict(err error) bool {
	var ke *kernel.AppError
	return errors.As(err, &ke) && ke.Code == application.ErrCodeConflict
}
