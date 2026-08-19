package command

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"sipon-be/internal/modules/akademik/application"
	"sipon-be/internal/modules/akademik/application/dto"
	"sipon-be/internal/modules/akademik/application/resolver"
	sesConst "sipon-be/internal/modules/akademik/domain/activity_session/constant"
	sesEntity "sipon-be/internal/modules/akademik/domain/activity_session/entity"
	sesRepo "sipon-be/internal/modules/akademik/domain/activity_session/repository"
	attConst "sipon-be/internal/modules/akademik/domain/attendance/constant"
	attEntity "sipon-be/internal/modules/akademik/domain/attendance/entity"
	attRepo "sipon-be/internal/modules/akademik/domain/attendance/repository"
	spRepo "sipon-be/internal/modules/akademik/domain/santri_program/repository"
	"sipon-be/internal/shared/kernel"
)

// CompleteSessionUseCase menyelesaikan sesi kegiatan. Saat sesi diselesaikan,
// sistem otomatis mencatat absensi "absent" (alpa) untuk semua santri pada
// program yang terkait dengan kegiatan sesi tersebut yang belum melakukan
// absensi. Sesi yang dibatalkan (cancelled) tidak menjalankan auto-absen.
type CompleteSessionUseCase struct {
	sessionRepo       sesRepo.ActivitySessionRepository
	attendanceRepo    attRepo.AttendanceRepository
	santriProgramRepo spRepo.SantriProgramRepository
	programResolver   *resolver.SessionProgramResolver
}

func NewCompleteSessionUseCase(
	sessionRepo sesRepo.ActivitySessionRepository,
	attendanceRepo attRepo.AttendanceRepository,
	santriProgramRepo spRepo.SantriProgramRepository,
	programResolver *resolver.SessionProgramResolver,
) *CompleteSessionUseCase {
	return &CompleteSessionUseCase{
		sessionRepo:       sessionRepo,
		attendanceRepo:    attendanceRepo,
		santriProgramRepo: santriProgramRepo,
		programResolver:   programResolver,
	}
}

func (uc *CompleteSessionUseCase) Execute(ctx context.Context, id string) (*dto.ActivitySessionResponse, error) {
	session, err := uc.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, application.WrapRepoErr(err, sesConst.CodeActivitySessionNotFound)
	}
	if err := session.Complete(); err != nil {
		return nil, application.WrapBadRequestErr(err, sesConst.CodeActivitySessionInvalidStatus)
	}

	if err := uc.autoAbsent(ctx, session); err != nil {
		return nil, err
	}

	if err := uc.sessionRepo.Update(ctx, session); err != nil {
		return nil, kernel.Wrap(application.ErrCodeInternal, err)
	}
	return MapSessionToResponse(session), nil
}

// autoAbsent mencatat absensi "absent" untuk semua santri aktif pada program
// yang terkait dengan sesi ini yang belum memiliki record absensi. Record yang
// sudah ada (present/excused) tidak ditimpa.
func (uc *CompleteSessionUseCase) autoAbsent(ctx context.Context, session *sesEntity.ActivitySession) error {
	programIDs, err := uc.programResolver.Resolve(ctx, session.ID)
	if err != nil {
		slog.Warn("akademik: resolve session programs failed", "session_id", session.ID, "error", err)
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	eligible := make(map[string]struct{})
	for _, pid := range programIDs {
		santriIDs, err := uc.santriProgramRepo.ListActiveSantriIDsByProgramID(ctx, pid)
		if err != nil {
			return kernel.Wrap(application.ErrCodeInternal, err)
		}
		for _, sid := range santriIDs {
			eligible[sid] = struct{}{}
		}
	}

	recorded, err := uc.attendanceRepo.ListSantriIDsBySession(ctx, session.ID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	for _, sid := range recorded {
		delete(eligible, sid)
	}

	for santriID := range eligible {
		att, err := attEntity.NewAttendance(uuid.NewString(), session.ID, santriID, attConst.AttendanceStatusAbsent)
		if err != nil {
			return application.WrapBadRequestErr(err, attConst.CodeAttendanceInvalidStatus)
		}
		if err := uc.attendanceRepo.Save(ctx, att); err != nil {
			return application.WrapConflictErr(err, attConst.CodeAttendanceDuplicate)
		}
	}
	return nil
}
