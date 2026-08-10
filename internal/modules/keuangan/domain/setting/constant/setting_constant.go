package constant

import "sipon-be/internal/shared/kernel"

// Settings keys — hardcoded. Menambah setting baru cukup menambah key di sini
// (plus mapper di entity/DTO/use case) tanpa migration.
const (
	KeyDefaultPaymentDebitAccountID string = "default_payment_debit_account_id"
)

// SettingsRowID — ID tetap untuk single-row table keuangan_settings.
const SettingsRowID string = "00000000-0000-0000-0000-000000000001"

const (
	CodeSettingNotFound kernel.Code = "SETTING_NOT_FOUND"
	CodeSettingInvalid  kernel.Code = "SETTING_INVALID"
)
