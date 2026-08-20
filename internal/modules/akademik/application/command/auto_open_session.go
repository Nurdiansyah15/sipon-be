package command

import (
	"context"
	"log/slog"

	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
)

// SessionOpener adalah boundary untuk membuka sesi. Dipenuhi oleh
// OpenSessionUseCase; interface ini memudahkan unit test AutoOpenSessionUseCase.
type SessionOpener interface {
	Execute(ctx context.Context, id string) (*dto.ActivitySessionResponse, error)
}

var _ SessionOpener = (*OpenSessionUseCase)(nil)

// AutoOpenSessionUseCase membuka sesi secara otomatis saat waktu mulai tiba.
// Jika sesi sudah dibuka manual oleh user (status bukan "scheduled"), use case
// ini skip tanpa error — job auto-open dianggap sudah tidak relevan.
type AutoOpenSessionUseCase struct {
	sessionRepo sesRepo.ActivitySessionRepository
	openSession SessionOpener
}

func NewAutoOpenSessionUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	openSession SessionOpener,
) *AutoOpenSessionUseCase {
	return &AutoOpenSessionUseCase{
		sessionRepo: sessionRepo,
		openSession: openSession,
	}
}

func (uc *AutoOpenSessionUseCase) Execute(ctx context.Context, sessionID string) error {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}

	// Skip jika sesi sudah tidak dalam status "scheduled" — bisa jadi sudah
	// dibuka manual oleh user atau sudah dibatalkan/diselesaikan.
	if session.Status != constant.ActivitySessionStatusScheduled {
		slog.Info("akademik: auto-open skip — sesi sudah tidak scheduled",
			"session_id", sessionID, "status", session.Status)
		return nil
	}

	if _, err := uc.openSession.Execute(ctx, sessionID); err != nil {
		return err
	}

	slog.Info("akademik: sesi berhasil dibuka otomatis", "session_id", sessionID)
	return nil
}
