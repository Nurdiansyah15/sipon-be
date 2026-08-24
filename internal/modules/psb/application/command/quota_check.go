package command

import (
	"context"

	"sipon-be/internal/modules/psb/application"
	prepo "sipon-be/internal/modules/psb/domain/pendaftar/repository"
	"sipon-be/internal/modules/psb/domain/quota"
	srepo "sipon-be/internal/modules/psb/domain/setting/repository"
	"sipon-be/internal/shared/kernel"
)

// ensureQuotaAvailable memeriksa apakah kuota kombinasi (program × gender)
// masih tersedia untuk pendaftar pada setting berjalan. Dipanggil saat submit
// dan saat admin accept — keduanya mencegah pendaftaran melebihi kuota.
//
// Kombinasi program/gender yang tidak dikenal, program tanpa kode, atau key
// kuota yang tidak dikonfigurasi diperlakukan sebagai tanpa batas (unlimited).
func ensureQuotaAvailable(ctx context.Context, settingRepo srepo.PsbSettingRepository, pendaftarRepo prepo.PendaftarRepository, psbSettingID, programID, gender string) error {
	if programID == "" || gender == "" {
		return nil
	}

	setting, err := settingRepo.FindByID(ctx, psbSettingID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	q, err := quota.Parse(setting.Quota)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	code, err := pendaftarRepo.FindProgramCode(ctx, programID)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}
	if code == nil {
		return nil
	}

	key, ok := quota.KeyFor(*code, gender)
	if !ok {
		return nil
	}

	used, err := pendaftarRepo.CountAcceptedByProgramAndGender(ctx, psbSettingID, programID, gender)
	if err != nil {
		return kernel.Wrap(application.ErrCodeInternal, err)
	}

	if q.IsFull(key, used) {
		return kernel.WrapMsg(application.ErrCodeUnprocessableEntity, "kuota program untuk kombinasi ini sudah penuh", nil)
	}
	return nil
}