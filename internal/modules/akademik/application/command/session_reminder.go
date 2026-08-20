package command

import (
	"context"
	"encoding/json"
	"log/slog"

	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/akademik/application/resolver"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	apRepo "sipon-be/internal/modules/akademik/domain/activity_period/repository"
	actRepo "sipon-be/internal/modules/akademik/domain/activity/repository"
	schRepo "sipon-be/internal/modules/akademik/domain/activity_schedule/repository"
)

// SessionReminderUseCase mengeksekusi reminder sesi: resolve periode akademik,
// kumpulkan semua user yang sudah herregistrasi (completed) pada periode tsb
// tanpa filter jurusan, lalu publish event notifikasi ke outbox.
type SessionReminderUseCase struct {
	sessionRepo     sesRepo.ActivitySessionRepository
	scheduleRepo    schRepo.ActivityScheduleRepository
	periodResolver  *resolver.SessionPeriodResolver
	registrationRepo regRepo.SantriRegistrationRepository
	activityPeriodRepo apRepo.ActivityPeriodRepository
	activityRepo    actRepo.ActivityRepository
	kesantrianReader ports.KesantrianReader
	outboxWriter    ports.OutboxWriter
	notifyRoutingKey string
}

func NewSessionReminderUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	scheduleRepo schRepo.ActivityScheduleRepository,
	periodResolver *resolver.SessionPeriodResolver,
	registrationRepo regRepo.SantriRegistrationRepository,
	activityPeriodRepo apRepo.ActivityPeriodRepository,
	activityRepo actRepo.ActivityRepository,
	kesantrianReader ports.KesantrianReader,
	outboxWriter ports.OutboxWriter,
	notifyRoutingKey string,
) *SessionReminderUseCase {
	return &SessionReminderUseCase{
		sessionRepo:      sessionRepo,
		scheduleRepo:     scheduleRepo,
		periodResolver:   periodResolver,
		registrationRepo: registrationRepo,
		activityPeriodRepo: activityPeriodRepo,
		activityRepo:     activityRepo,
		kesantrianReader: kesantrianReader,
		outboxWriter:     outboxWriter,
		notifyRoutingKey: notifyRoutingKey,
	}
}

func (uc *SessionReminderUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

func (uc *SessionReminderUseCase) Execute(ctx context.Context, sessionID string) error {
	if uc.outboxWriter == nil {
		slog.Warn("akademik: outbox writer nil, skip reminder", "session_id", sessionID)
		return nil
	}

	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return err
	}
	// Jangan reminder sesi yang sudah selesai/dibatalkan.
	if session.Status == sesConst.ActivitySessionStatusCompleted || session.Status == sesConst.ActivitySessionStatusCancelled {
		slog.Info("akademik: reminder skip — sesi sudah tidak aktif", "session_id", sessionID, "status", session.Status)
		return nil
	}

	academicPeriodID, err := uc.periodResolver.Resolve(ctx, sessionID)
	if err != nil {
		slog.Warn("akademik: reminder resolve period gagal", "session_id", sessionID, "error", err)
		return err
	}

	// Kumpulkan user yang sudah herregistrasi pada periode tsb (tanpa filter jurusan).
	registrations, err := uc.registrationRepo.ListCompletedByAcademicPeriod(ctx, academicPeriodID)
	if err != nil {
		slog.Warn("akademik: reminder list herreg gagal", "session_id", sessionID, "error", err)
		return err
	}

	userSet := make(map[string]struct{})
	for _, reg := range registrations {
		info, err := uc.kesantrianReader.GetSantriByID(ctx, reg.SantriID)
		if err != nil || info == nil || info.UserID == "" {
			slog.Warn("akademik: reminder enrich user gagal", "santri_id", reg.SantriID, "error", err)
			continue
		}
		userSet[info.UserID] = struct{}{}
	}
	if len(userSet) == 0 {
		slog.Info("akademik: reminder tidak ada user herreg", "session_id", sessionID)
		return nil
	}

	userIDs := make([]string, 0, len(userSet))
	for uid := range userSet {
		userIDs = append(userIDs, uid)
	}

	activityName := uc.resolveActivityName(ctx, sessionID)
	payload, _ := json.Marshal(map[string]interface{}{
		"session_id":     sessionID,
		"user_ids":       userIDs,
		"activity_name":  activityName,
		"starts_at":      session.StartsAt,
	})
	if err := uc.outboxWriter.Save(ctx, uc.notifyRoutingKey, payload); err != nil {
		slog.Warn("akademik: gagal publish event reminder", "session_id", sessionID, "error", err)
		return err
	}

	slog.Info("akademik: reminder event published", "session_id", sessionID, "recipients", len(userIDs))
	return nil
}

func (uc *SessionReminderUseCase) resolveActivityName(ctx context.Context, sessionID string) string {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return ""
	}
	schedule, err := uc.scheduleRepo.FindByID(ctx, session.ActivityScheduleID)
	if err != nil {
		return ""
	}
	period, err := uc.activityPeriodRepo.FindByID(ctx, schedule.ActivityPeriodID)
	if err != nil {
		return ""
	}
	activity, err := uc.activityRepo.FindByID(ctx, period.ActivityID)
	if err != nil {
		return ""
	}
	return activity.Name
}
