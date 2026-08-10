# Plan: Keuangan Settings (Default Payment Debit Account) di `sipon-be` & `sipon-ui`

## Context

Fitur pembayaran manual santri (upload bukti transfer) membutuhkan `debit_account_id` untuk setiap pembayaran:

1. **Database constraint**: `payments.debit_account_id` sudah **NOT NULL** (migration `20260809130000_refine_akuntansi_fase1.up.sql:10`).
2. **Validasi existing**: `CreateManualPaymentUseCase` (`create_manual_payment.go`) memvalidasi account harus `postable` dan `sub_type = cash_bank`.
3. **Masalah**: Pembayaran santri (self-service) tidak perlu memilih account debit — admin ingin memakai akun kas/bank default yang sudah dikonfigurasi.

**Solusi**: Buat **Keuangan Settings** untuk menyimpan konfigurasi default, dimulai dengan:
- `default_payment_debit_account_id` — account kas/bank default untuk pembayaran.

Settings disimpan sebagai **JSONB** (single-row table) dengan **key hardcoded di constant**. Menambah setting baru di masa depan cukup menambah key di constant + mapper — tanpa migration.

### Keputusan desain

| No | Pertanyaan | Keputusan |
|----|-----------|-----------|
| 1 | Bagaimana menyimpan settings? | **Single-row table dengan JSONB column** (`keuangan_settings`) |
| 2 | Bagaimana key didefinisikan? | **Hardcoded di Go constant** (`domain/setting/constant`) |
| 3 | Bagaimana value diisi? | **Admin pilih dari COA** di UI (`KeuanganAccountPicker`, filter asset + cash_bank) |
| 4 | Bagaimana jika key baru ditambahkan? | **Tambah constant + mapper** — tanpa migration schema |
| 5 | Apakah validasi diperlukan? | **Ya** — account harus `postable` + `sub_type = cash_bank` |

## Temuan penting dari kode aktual

1. **Pattern single-row JSONB sudah ada** di PSB: `psb_settings` pakai `json.RawMessage` di entity + helper `jsonBytes()` di repo.
2. **`payments.debit_account_id` NOT NULL** (migration `20260809130000`), dan penanganannya saat verifikasi sudah ada di `VerifyPaymentUseCase`.
3. **`KeuanganAccountPicker.vue`** sudah ada di UI dan mendukung prop `filter` (AccountType) + `subType` (AccountSubType) — cukup dipakai dengan `filter="asset"` dan `sub-type="cash_bank"`.
4. **Permission key**: `manage_keuangan` sudah ada — dipakai untuk endpoint settings (konsisten dengan komponen biaya, skema, dll).
5. **Single-row enforcement**: table `keuangan_settings` memakai **ID tetap** `00000000-0000-0000-0000-000000000001`; migration meng-insert 1 baris default `{}`.

## Database — Migration

**File: `migrations/20260811090000_create_keuangan_settings.up.sql`**

```sql
CREATE TABLE keuangan_settings (
    id UUID PRIMARY KEY,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO keuangan_settings (id, settings)
VALUES ('00000000-0000-0000-0000-000000000001', '{}'::jsonb);
```

**File: `migrations/20260811090000_create_keuangan_settings.down.sql`**

```sql
DROP TABLE IF EXISTS keuangan_settings;
```

### Schema JSONB

```json
{
  "default_payment_debit_account_id": "uuid-of-account"
}
```

## Backend — `sipon-be`

### A. Constant — `domain/setting/constant/setting_constant.go` (BARU)

```go
package constant

import "sipon-be/internal/shared/kernel"

// Settings keys — hardcoded. Tambah key baru di sini (plus mapper) tanpa migration.
const (
	KeyDefaultPaymentDebitAccountID string = "default_payment_debit_account_id"
)

// ID tetap untuk single-row table keuangan_settings.
const SettingsRowID string = "00000000-0000-0000-0000-000000000001"

const (
	CodeSettingNotFound kernel.Code = "SETTING_NOT_FOUND"
	CodeSettingInvalid  kernel.Code = "SETTING_INVALID"
)
```

### B. Entity — `domain/setting/entity/setting.go` (BARU)

- `KeuanganSetting{ ID, Settings json.RawMessage, CreatedAt, UpdatedAt }`.
- `GetDefaultPaymentDebitAccountID() (*string, error)` — parse key dari JSONB.
- `SetDefaultPaymentDebitAccountID(*string) error` — set/hapus key, preserve key lain.
- Test: `setting_test.go` (set/get/clear/preserve unknown keys).

### C. Repository Interface — `domain/setting/repository/interfaces.go` (BARU)

```go
type KeuanganSettingRepository interface {
	Find(ctx context.Context) (*entity.KeuanganSetting, error)
	Update(ctx context.Context, setting *entity.KeuanganSetting) error
}
```

### D. Postgres Repo — `infrastructure/persistence/postgres_keuangan_setting_repo.go` (BARU)

- `Find` → `SELECT ... WHERE id = SettingsRowID`, scan via `scanKeuanganSetting`.
- `Update` → `UPDATE keuangan_settings SET settings=$1, updated_at=NOW() WHERE id=$2`.
- Tambah helper `jsonBytes` di `infrastructure/persistence/helpers.go`.

### E. DTO — `application/dto/setting_dto.go` (BARU)

```go
type KeuanganSettingResponse struct {
	DefaultPaymentDebitAccountID *string               `json:"default_payment_debit_account_id,omitempty"`
	DefaultPaymentDebitAccount   *AccountBriefResponse `json:"default_payment_debit_account,omitempty"`
}

type UpdateKeuanganSettingRequest struct {
	DefaultPaymentDebitAccountID *string `json:"default_payment_debit_account_id"`
}
```

### F. Command — `application/command/update_keuangan_setting.go` (BARU)

`UpdateKeuanganSettingUseCase`:
1. Jika `DefaultPaymentDebitAccountID` tidak nil → validasi akun: `FindByID`, `EnsurePostable`, `SubType == SubTypeCashBank`.
2. `settingRepo.Find` → `SetDefaultPaymentDebitAccountID` → `settingRepo.Update`.
3. Return response dengan nested account brief.

### G. Query — `application/query/get_keuangan_setting.go` (BARU)

`GetKeuanganSettingUseCase`: `Find` → map ke response (termasuk nested account brief).

### H. HTTP Handler — `interfaces/http/handler.go` (UPDATE)

- Tambah field `updateSettingUC`, `getSettingUC` ke struct + constructor.
- Tambah `GetKeuanganSetting` dan `UpdateKeuanganSetting`.

### I. Router — `interfaces/http/router.go` (UPDATE)

```go
admin.GET("/settings", middleware.RequirePermission("manage_keuangan"), h.GetKeuanganSetting)
admin.PUT("/settings", middleware.RequirePermission("manage_keuangan"), h.UpdateKeuanganSetting)
```

### J. Module Wiring — `module.go` (UPDATE)

- `keuanganSettingRepo := persistence.NewPostgresKeuanganSettingRepository(db)`.
- Wire `updateSettingUC`, `getSettingUC`, dan pass ke handler.

## Frontend — `sipon-ui`

### A. Types — `shared/types/Keuangan.ts` (UPDATE)

```ts
export interface KeuanganSettingResponse {
  default_payment_debit_account_id?: string | null
  default_payment_debit_account?: AccountBrief | null
}

export interface UpdateKeuanganSettingRequest {
  default_payment_debit_account_id?: string | null
}
```

### B. Store — `app/stores/keuangan.ts` (UPDATE)

- State: `settings: KeuanganSettingResponse | null`.
- `fetchKeuanganSettings()` → `GET /api/v1/web/keuangan/admin/settings`.
- `updateKeuanganSettings(payload)` → `PUT /api/v1/web/keuangan/admin/settings`.

### C. Halaman — `app/pages/admin/keuangan/settings/index.vue` (BARU)

- `definePageMeta({ layout: 'keuangan' })`.
- Load settings on mount, tampilkan `KeuanganAccountPicker` (filter asset + cash_bank).
- Tombol "Simpan Pengaturan" → `updateKeuanganSettings` → toast sukses.
- Tampilkan akun terpilih (code + name) + skeleton loading.

### D. Navigasi — `app/components/AppKeuanganNavbar.vue` (UPDATE)

- Tambah item "Pengaturan" (`i-lucide-settings`) di section Master Data → `/admin/keuangan/settings`.

## Fase Pengerjaan (implementasi yang sudah selesai)

1. **Migration**: table `keuangan_settings` + row default.
2. **Backend domain & infrastructure**: constant, entity (+test), repo interface, postgres repo, helper `jsonBytes`.
3. **Backend application**: DTO, `UpdateKeuanganSettingUseCase`, `GetKeuanganSettingUseCase`.
4. **Backend HTTP**: handler + router (`GET/PUT /admin/settings`) + module wiring.
5. **Frontend types & store**: types + `fetchKeuanganSettings` / `updateKeuanganSettings`.
6. **Frontend UI**: halaman settings + item sidebar.

## Verifikasi

1. `go build ./...` dan `go test ./internal/modules/keuangan/...` lolos (termasuk `domain/setting/entity`).
2. `npx nuxi typecheck` di `sipon-ui` tanpa error baru.
3. **Validasi akun**: simpan akun non-kas/non-postable → ditolak dengan pesan jelas.
4. **Empty state**: settings default `{}` → response `default_payment_debit_account_id = null`.
5. **Update**: ganti akun default → response mencerminkan akun baru.
6. **Single-row**: hanya ada 1 baris di `keuangan_settings` (ID tetap).

## Catatan tambahan

- **Integrasi lanjutan**: plan `payment-manual-santri.md` **belum** memakai settings ini — sesuai keputusan, `debit_account_id` pembayaran santri diisi admin saat verifikasi (lihat catatan implementasi di `payment-manual-santri.md`). Settings `default_payment_debit_account_id` akan dipakai di masa depan bila diperlukan.
- **Ekstensibilitas**: tambah setting baru cukup (1) constant, (2) entity mapper, (3) DTO + use case — tanpa migration.
- **JSONB**: dipakai untuk fleksibilitas; tidak ada relasi FK ke `accounts` di JSONB (nilai divalidasi di use case).
