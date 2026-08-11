package constant

import "sipon-be/internal/shared/kernel"

// Settings keys — hardcoded. Tambah key baru di sini (plus mapper di
// entity/DTO/use case) tanpa migration.
const (
	KeyDefaultProgramID string = "default_program_id"
)

// SettingsRowID — ID tetap untuk single-row table akademik_settings.
const SettingsRowID string = "00000000-0000-0000-0000-000000000002"

const (
	CodeSettingNotFound kernel.Code = "AKADEMIK_SETTING_NOT_FOUND"
	CodeSettingInvalid  kernel.Code = "AKADEMIK_SETTING_INVALID"
)
