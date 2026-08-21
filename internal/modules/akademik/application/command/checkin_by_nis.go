package command

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/ports"
	"sipon-be/internal/modules/akademik/application/resolver"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	attConst "sipon-be/internal/modules/akademik/domain/attendance/constant"
	attEntity "sipon-be/internal/modules/akademik/domain/attendance/entity"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	regRepo "sipon-be/internal/modules/akademik/domain/santri_registration/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

// CheckinByNISUseCase mencatat kehadiran santri via NIS dari halaman presensi.
type CheckinByNISUseCase struct {
	sessionRepo       sesRepo.ActivitySessionRepository
	kesantrianReader  ports.KesantrianReader
	periodResolver    *resolver.SessionPeriodResolver
	registrationRepo  regRepo.SantriRegistrationRepository
	attendanceRepo    attRepo.AttendanceRepository
	santriProgramRepo spRepo.SantriProgramRepository
	programResolver   *resolver.SessionProgramResolver
	outboxWriter      ports.OutboxWriter
}

func NewCheckinByNISUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	kesantrianReader ports.KesantrianReader,
	periodResolver *resolver.SessionPeriodResolver,
	registrationRepo regRepo.SantriRegistrationRepository,
	attendanceRepo attRepo.AttendanceRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	programResolver *resolver.SessionProgramResolver,
) *CheckinByNISUseCase {
	return &CheckinByNISUseCase{
		sessionRepo:       sessionRepo,
		kesantrianReader:  kesantrianReader,
		periodResolver:    periodResolver,
		registrationRepo:  registrationRepo,
		attendanceRepo:    attendanceRepo,
		santriProgramRepo: santriProgramRepo,
		programResolver:   programResolver,
	}
}

// SetOutboxWriter memasang outbox writer untuk publikasi event notifikasi
// kehadiran yang tercatat (check-in manual NIS maupun sinkronisasi fingerprint).
func (uc *CheckinByNISUseCase) SetOutboxWriter(w ports.OutboxWriter) {
	uc.outboxWriter = w
}

// Execute adalah alias check-in manual via input NIS.
func (uc *CheckinByNISUseCase) Execute(ctx context.Context, sessionID, nis string) (*dto.CheckinResponse, error) {
	return uc.ExecuteWithSource(ctx, sessionID, nis, AttendanceSourceNIS)
}

// ExecuteWithSource mencatat kehadiran santri via NIS dengan menandai sumber
// pencatatan (manual NIS / sinkronisasi fingerprint) untuk keperluan notifikasi.
func (uc *CheckinByNISUseCase) ExecuteWithSource(ctx context.Context, sessionID, nis, source string) (*dto.CheckinResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, application.WrapRepoErr(err, sesConst.CodeActivitySessionNotFound)
	}
	if session.Status != "open" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "sesi tidak terbuka untuk presensi", nil)
	}

	info, err := uc.kesantrianReader.GetSantriByNIS(ctx, nis)
	if err != nil {
		return nil, kernel.WrapMsg(application.ErrCodeNotFound, "NIS tidak ditemukan", nil)
	}
	if info == nil || info.Status != "SANTRI" {
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri tidak aktif", nil)
	}

	academicPeriodID, err := uc.periodResolver.Resolve(ctx, sessionID)
	if err != nil {
		slog.Warn("akademik: resolve session academic period failed", "session_id", sessionID, "error", err)
		return nil, kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "periode akademik sesi tidak ditemukan", nil)
	}

	if err := uc.ensureHerreg(ctx, info.SantriID, academicPeriodID); err != nil {
		return nil, err
	}

	if err := uc.ensureProgramMembership(ctx, sessionID, info.SantriID); err != nil {
		return nil, err
	}

	existing, _ := uc.attendanceRepo.FindBySessionAndSantri(ctx, sessionID, info.SantriID)
	if existing != nil {
		return nil, kernel.WrapMsg(application.ErrCodeConflict, "NIS sudah tercatat hadir", nil)
	}

	attendance, err := attEntity.NewAttendance(uuid.NewString(), sessionID, info.SantriID, attConst.AttendanceStatusPresent)
	if err != nil {
		return nil, application.WrapBadRequestErr(err, attConst.CodeAttendanceInvalidStatus)
	}
	if err := uc.attendanceRepo.Save(ctx, attendance); err != nil {
		return nil, application.WrapConflictErr(err, attConst.CodeAttendanceDuplicate)
	}

	uc.publishAttendanceRecorded(ctx, info, attendance.ID, sessionID, source)

	resp := MapAttendanceToResponse(attendance)
	resp.SantriNIS = info.NIS
	resp.SantriName = info.Fullname

	name := ""
	if info.Fullname != nil {
		name = *info.Fullname
	}
	return &dto.CheckinResponse{
		Attendance: *resp,
		Message:    fmt.Sprintf("Selamat, %s! Kehadiran tercatat.", name),
	}, nil
}

// publishAttendanceRecorded menulis event notifikasi kehadiran tercatat ke
// outbox. Event dikonsumsi modul notification untuk mengirim notifikasi
// in-app/push kepada user santri yang bersangkutan.
func (uc *CheckinByNISUseCase) publishAttendanceRecorded(ctx context.Context, info *ports.SantriBasicInfo, attendanceID, sessionID, source string) {
	if uc.outboxWriter == nil || info == nil || info.UserID == "" {
		return
	}

	name := ""
	if info.Fullname != nil {
		name = *info.Fullname
	}
	nis := ""
	if info.NIS != nil {
		nis = *info.NIS
	}

	payload, _ := json.Marshal(attendanceRecordedPayload{
		UserID:       info.UserID,
		AttendanceID: attendanceID,
		SantriID:     info.SantriID,
		NIS:          nis,
		Name:         name,
		SessionID:    sessionID,
		Source:       source,
	})
	if err := uc.outboxWriter.Save(ctx, RoutingAttendanceRecorded, payload); err != nil {
		slog.Warn("akademik: gagal publish event kehadiran tercatat", "session_id", sessionID, "santri_id", info.SantriID, "error", err)
	}
}

func (uc *CheckinByNISUseCase) ensureHerreg(ctx context.Context, santriID, academicPeriodID string) error {
	reg, err := uc.registrationRepo.FindBySantriAndPeriod(ctx, santriID, academicPeriodID)
	if err != nil || reg == nil || reg.Status != "completed" {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri belum herregistrasi pada periode ini", nil)
	}
	return nil
}

// ensureProgramMembership memastikan santri berada pada program yang terkait
// dengan kegiatan sesi ini. Hanya santri pada program yang berkaitan yang boleh
// melakukan absensi.
func (uc *CheckinByNISUseCase) ensureProgramMembership(ctx context.Context, sessionID, santriID string) error {
	programIDs, err := uc.programResolver.Resolve(ctx, sessionID)
	if err != nil {
		slog.Warn("akademik: resolve session programs failed", "session_id", sessionID, "error", err)
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	sp, err := uc.santriProgramRepo.FindActiveBySantriID(ctx, santriID)
	if err != nil || sp == nil {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri tidak terdaftar di program manapun", nil)
	}

	found := false
	for _, pid := range programIDs {
		if sp.ProgramID == pid {
			found = true
			break
		}
	}
	if !found {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "santri tidak terdaftar di program kegiatan ini", nil)
	}
	return nil
}
