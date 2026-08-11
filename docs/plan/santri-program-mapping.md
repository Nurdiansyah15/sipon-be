# Plan: Pemetaan Santri ke Program (Akademik)

## Context

Saat ini santri memiliki field `Program` di module kesantrian (plain string), dan pendaftar PSB juga memiliki field `Program` (optional). Namun belum ada pemetaan formal antara santri dengan program di module akademik.

**Kebutuhan:**
1. Santri harus ter mapping ke program yang dia ikuti di sistem akademik
2. Santri bisa punya banyak program (historis), tapi saat ini hanya boleh **1 aktif**
3. PSB: santri memilih program sendiri saat registrasi → menjadi record di akademik
4. Admin generate santri manual → pakai default program dari pengaturan akademik
5. Pengaturan akademik (seperti keuangan settings) untuk menyimpan default program

**Masalah saat ini:**
- Field `Program` di PSB `Pendaftar` masih optional (`*string`) — perlu dibuat required saat submit
- Field `Program` di Kesantrian `Santri` adalah plain string — perlu direferensikan ke `programs.id` di akademik
- Belum ada table/entity untuk mapping santri ke program di akademik
- Belum ada pengaturan akademik untuk default program

## Keputusan Desain

| No | Pertanyaan | Keputusan |
|----|-----------|-----------|
| 1 | Bagaimana menyimpan mapping santri-program? | **Table baru `santri_programs`** di akademik (FK ke programs & santri via santri_id plain UUID) |
| 2 | Bagaimana enforce "hanya 1 aktif"? | **Unique partial index** `WHERE is_active = true` di database + validasi di domain |
| 3 | Bagaimana default program untuk admin create santri? | **Akademik settings** (single-row JSONB) dengan key `default_program_id` |
| 4 | Bagaimana PSB flow? | PSB pilih program → CreateSantriFromPendaftaran → after create, PSB (atau kesantrian) create record `santri_programs` via kontrak akademik |
| 5 | Apakah field Program di Santri entity dihapus? | **Tidak** — tetap sebagai denormalized cache, tapi sumber kebenaran adalah `santri_programs` |
| 6 | Apakah Program di PSB jadi required? | **Ya** — validasi saat submit pendaftaran, harus pilih program |

## Temuan Penting dari Kode Aktual

1. **Program entity sudah ada** di `akademik/domain/program/entity/program.go` dengan `Code` (TAHFIDZ, KITAB) dan `Name`.
2. **SantriRegistration** sudah ada di akademik — mapping santri ke `academic_period`, tapi TIDAK ke program.
3. **Keuangan settings pattern** sudah established: single-row JSONB table, hardcoded keys di constant, Find/Update repository.
4. **PSB → Kesantrian flow**: `generate_nis.go` memanggil `kesantrian.CreateSantriFromPendaftaran` yang menyimpan `Program` sebagai plain string.
5. **Program di PSB** adalah free-text string, tidak divalidasi terhadap `programs` table di akademik.

## Database — Migration

### File: `migrations/YYYYMMDDHHMMSS_create_akademik_settings.up.sql`

```sql
-- Akademik settings: single-row table untuk konfigurasi default akademik
-- Pattern sama dengan keuangan_settings

CREATE TABLE akademik_settings (
    id UUID PRIMARY KEY,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO akademik_settings (id, settings)
VALUES ('00000000-0000-0000-0000-000000000002', '{}'::jsonb);
```

### File: `migrations/YYYYMMDDHHMMSS_create_santri_programs.up.sql`

```sql
-- Pemetaan santri ke program (akademik)
-- Santri bisa punya banyak program historis, tapi hanya 1 yang aktif

CREATE TABLE santri_programs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id   UUID NOT NULL,  -- no FK, cross-module (kesantrian)
    program_id  UUID NOT NULL REFERENCES programs(id),
    is_active   BOOLEAN NOT NULL DEFAULT true,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Hanya 1 program aktif per santri
CREATE UNIQUE INDEX idx_santri_programs_active_unique
    ON santri_programs(santri_id)
    WHERE is_active = true AND deleted_at IS NULL;

CREATE INDEX idx_santri_programs_santri
    ON santri_programs(santri_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_santri_programs_program
    ON santri_programs(program_id)
    WHERE deleted_at IS NULL;
```

### Schema JSONB untuk akademik_settings

```json
{
  "default_program_id": "uuid-of-program"
}
```

## Backend — `sipon-be`

### A. Akademik Settings

#### A.1 Constant — `akademik/domain/setting/constant/setting_constant.go` (BARU)

```go
package constant

import "sipon-be/internal/shared/kernel"

const (
    KeyDefaultProgramID string = "default_program_id"
)

const SettingsRowID string = "00000000-0000-0000-0000-000000000002"

const (
    CodeSettingNotFound kernel.Code = "AKADEMIK_SETTING_NOT_FOUND"
    CodeSettingInvalid  kernel.Code = "AKADEMIK_SETTING_INVALID"
)
```

#### A.2 Entity — `akademik/domain/setting/entity/setting.go` (BARU)

```go
type AkademikSetting struct {
    ID        string
    Settings  json.RawMessage
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (s *AkademikSetting) GetDefaultProgramID() (*string, error)
func (s *AkademikSetting) SetDefaultProgramID(programID *string) error
```

#### A.3 Repository — `akademik/domain/setting/repository/interfaces.go` (BARU)

```go
type AkademikSettingRepository interface {
    Find(ctx context.Context) (*entity.AkademikSetting, error)
    Update(ctx context.Context, setting *entity.AkademikSetting) error
}
```

#### A.4 Postgres Repo — `akademik/infrastructure/persistence/postgres_akademik_setting_repo.go` (BARU)

- Pattern sama dengan `postgres_keuangan_setting_repo.go`

#### A.5 DTO — `akademik/application/dto/setting_dto.go` (BARU)

```go
type AkademikSettingResponse struct {
    DefaultProgramID   *string           `json:"default_program_id,omitempty"`
    DefaultProgram     *ProgramResponse  `json:"default_program,omitempty"`
}

type UpdateAkademikSettingRequest struct {
    DefaultProgramID *string `json:"default_program_id"`
}
```

#### A.6 Command — `akademik/application/command/update_akademik_setting.go` (BARU)

`UpdateAkademikSettingUseCase`:
1. Jika `DefaultProgramID` tidak nil → validasi program exists & active
2. `settingRepo.Find` → `SetDefaultProgramID` → `settingRepo.Update`

#### A.7 Query — `akademik/application/query/get_akademik_setting.go` (BARU)

`GetAkademikSettingUseCase`: Find → map ke response (include nested program brief)

#### A.8 HTTP Handler & Router (UPDATE)

```go
// router.go
admin.GET("/settings", middleware.RequirePermission("manage_akademik"), h.GetAkademikSetting)
admin.PUT("/settings", middleware.RequirePermission("manage_akademik"), h.UpdateAkademikSetting)
```

### B. Santri Program Entity

#### B.1 Constant — `akademik/domain/santri_program/constant/santri_program_constant.go` (BARU)

```go
const (
    CodeSantriProgramNotFound      kernel.Code = "SANTRI_PROGRAM_NOT_FOUND"
    CodeSantriProgramDuplicate     kernel.Code = "SANTRI_PROGRAM_DUPLICATE"
    CodeSantriProgramAlreadyActive kernel.Code = "SANTRI_PROGRAM_ALREADY_ACTIVE"
)
```

#### B.2 Entity — `akademik/domain/santri_program/entity/santri_program.go` (BARU)

```go
type SantriProgram struct {
    ID        string
    SantriID  string
    ProgramID string
    IsActive  bool
    StartedAt time.Time
    EndedAt   *time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}

func NewSantriProgram(id, santriID, programID string) (*SantriProgram, error)
func (sp *SantriProgram) Deactivate() error  // set is_active=false, ended_at=now
```

#### B.3 Repository — `akademik/domain/santri_program/repository/interfaces.go` (BARU)

```go
type SantriProgramRepository interface {
    Save(ctx context.Context, sp *entity.SantriProgram) error
    FindBySantriID(ctx context.Context, santriID string) (*entity.SantriProgram, error) // yang aktif
    FindActiveBySantriAndProgram(ctx context.Context, santriID, programID string) (*entity.SantriProgram, error)
    DeactivateBySantriID(ctx context.Context, santriID string) error // deactivate yg aktif
    Update(ctx context.Context, sp *entity.SantriProgram) error
}
```

#### B.4 Postgres Repo — `akademik/infrastructure/persistence/postgres_santri_program_repo.go` (BARU)

### C. Contract & Use Cases

#### C.1 Akademik Contract — `akademik/contract.go` (BARU)

```go
type Contract interface {
    GetDefaultProgramID(ctx context.Context) (*string, error)
    AssignSantriProgram(ctx context.Context, santriID, programID string) error
    GetSantriProgram(ctx context.Context, santriID string) (*SantriProgramInfo, error)
}

type SantriProgramInfo struct {
    SantriID  string
    ProgramID string
    ProgramCode string
    ProgramName string
}
```

#### C.2 Command — `akademik/application/command/assign_santri_program.go` (BARU)

`AssignSantriProgramUseCase`:
1. Validasi program exists & active
2. Deactivate existing active program untuk santri ini (jika ada)
3. Create new SantriProgram dengan is_active=true
4. Return info

#### C.3 Command — `akademik/application/command/create_santri_from_psb.go` (BARU)

`CreateSantriFromPsbUseCase` — dipanggil setelah santri dibuat di kesantrian:
1. Terima santriID + programID
2. Assign santri ke program (pakai AssignSantriProgramUseCase internal)

### D. Update Module Kesantrian

#### D.1 Contract Update — `kesantrian/contract.go` (UPDATE)

Tambah method atau parameter untuk menerima programID (UUID) bukan hanya plain string:

```go
type CreateSantriFromPendaftaranInput struct {
    // ... existing fields
    ProgramID *string  // tambah: reference ke programs.id di akademik
}
```

#### D.2 Update `create_santri_from_pendaftaran.go`

Setelah create santri, publish event atau call kontrak akademik untuk assign program.

### E. Update Module PSB

#### E.1 Validasi Program Required — `psb/application/command/upsert_formulir.go` (UPDATE)

Saat submit pendaftaran, validasi `Program` tidak boleh nil:

```go
func (uc *PendaftarActionUseCase) SubmitPendaftaran(ctx context.Context, userID, settingID string) (*dto.MessageResponse, error) {
    // ... existing code
    if p.Program == nil || *p.Program == "" {
        return nil, kernel.New(application.ErrCodeUnprocessableEntity) // atau error khusus
    }
    // ...
}
```

Atau lebih baik: validasi saat UpsertFormulir bahwa Program harus diisi.

#### E.2 Update `generate_nis.go` (UPDATE)

Setelah CreateSantriFromPendaftaran berhasil, call akademik contract untuk assign program:

```go
// Setelah result diterima
if p.Program != nil {
    // Resolve programID dari code/name yang dipilih pendaftar
    // Atau: simpan programID langsung di Pendaftar (migrate field)
    err = uc.akademik.AssignSantriProgram(ctx, result.SantriID, programID)
}
```

**Pertimbangan**: Field `Program` di `Pendaftar` saat ini plain string. Ada 2 opsi:
1. **Opsi A**: Tambah field `ProgramID *string` (UUID FK ke programs) di `pendaftars` table
2. **Opsi B**: Resolve program by code/name saat generate NIS

Rekomendasi: **Opsi A** — lebih clean, langsung reference ke program yang valid.

### F. Update Admin Create Santri

#### F.1 `kesantrian/application/command/create_santri.go` (UPDATE)

Saat admin create santri manual, jika tidak ada programID di request, ambil dari akademik settings:

```go
func (uc *CreateSantriUseCase) Execute(ctx context.Context, req dto.CreateSantriRequest) (*dto.CreateSantriResponse, error) {
    // ... existing code

    // Assign program
    programID := req.ProgramID
    if programID == nil {
        // Ambil default dari akademik settings
        defaultProgramID, _ := uc.akademikContract.GetDefaultProgramID(ctx)
        programID = defaultProgramID
    }

    if programID != nil {
        // Save santri dulu, lalu assign program via akademik contract
        uc.akademikContract.AssignSantriProgram(ctx, santri.ID, *programID)
    }

    // ...
}
```

### G. Module Wiring

#### G.1 `akademik/module.go` (UPDATE)

- Init `akademikSettingRepo`, `santriProgramRepo`
- Wire use cases baru
- Expose contract methods

#### G.2 `kesantrian/module.go` (UPDATE)

- Terima `akademik.Contract` di constructor
- Pass ke `CreateSantriUseCase` dan `CreateSantriFromPendaftaranUseCase`

#### G.3 `psb/module.go` (UPDATE)

- Terima `akademik.Contract` di constructor
- Pass ke `GenerateNISUseCase`

#### G.4 `cmd/api/main.go` (UPDATE)

- Init akademik module lebih dulu (karena jadi dependency)
- Pass `akademik.Contract` ke kesantrian dan PSB

## Frontend — `sipon-ui`

### A. Akademik Settings Page

#### A.1 Types — `shared/types/Akademik.ts` (UPDATE)

```ts
export interface AkademikSettingResponse {
  default_program_id?: string | null
  default_program?: ProgramBrief | null
}

export interface UpdateAkademikSettingRequest {
  default_program_id?: string | null
}
```

#### A.2 Store — `app/stores/akademik.ts` (UPDATE)

- `fetchAkademikSettings()` → `GET /api/v1/web/akademik/admin/settings`
- `updateAkademikSettings(payload)` → `PUT /api/v1/web/akademik/admin/settings`

#### A.3 Halaman — `app/pages/admin/akademik/settings/index.vue` (BARU)

- `definePageMeta({ layout: 'akademik' })`
- Dropdown/picker untuk pilih default program
- Tombol "Simpan Pengaturan"

#### A.4 Navigasi — `app/components/AppAkademikNavbar.vue` (UPDATE)

- Tambah item "Pengaturan" → `/admin/akademik/settings`

### B. PSB Form — Program Required

#### B.1 `app/pages/psb/pendaftaran/index.vue` (UPDATE)

- Field program jadi required (bintang merah / validasi)
- Fetch list program dari `/api/v1/web/akademik/programs` (dropdown)
- Kirim `program_id` (bukan plain text) saat upsert formulir

### C. Admin Create Santri — Default Program

#### C.1 `app/pages/admin/santri/create.vue` (UPDATE)

- Field program opsional
- Jika kosong, gunakan default program dari settings (tampilkan hint)
- Atau: dropdown program dengan option "Gunakan Default"

## Fase Pengerjaan

### Phase 1: Database & Domain Layer
1. Migration `akademik_settings` table
2. Migration `santri_programs` table
3. Migration add `program_id` ke `pendaftars` (optional, untuk PSB)
4. Domain: akademik setting constant, entity, repo interface
5. Domain: santri_program constant, entity, repo interface
6. Persistence: postgres repos

### Phase 2: Application & Contract Layer
7. Akademik: DTO settings, use cases (get/update)
8. Akademik: santri_program use cases (assign, get)
9. Akademik: contract definition & implementation
10. Update kontrak kesantrian (terima akademik contract)

### Phase 3: HTTP & Wiring
11. Akademik: handler & router untuk settings
12. Update module wiring (main.go, kesantrian/module.go, psb/module.go)

### Phase 4: Update Existing Flows
13. Update PSB: validasi program required, simpan program_id
14. Update PSB generate_nis: assign santri program setelah create
15. Update kesantrian create_santri: pakai default program dari settings
16. Update kesantrian create_santri_from_pendaftaran: terima program_id

### Phase 5: Frontend
17. Akademik settings page
18. PSB form: program required + dropdown dari API
19. Admin create santri: default program hint

## Catatan Implementasi (deviasi dari plan)

1. **Circular dependency** antara akademik (butuh `kesantrian.Contract`) dan kesantrian
   (butuh `akademik.Contract`) diselesaikan dengan **late-binding setter**:
   - `kesantrian/application/ports/akademik_provisioner.go` mendefinisikan port
     `AkademikProvisioner` (`GetDefaultProgramID`, `AssignSantriProgram`).
   - `kesantrian.Module.SetAkademikProvisioner(...)` dipanggil di `main.go` SETELAH
     module akademik terbentuk (akademik butuh kesantrian di konstruktor).
2. **Seeder program** baru `internal/seeders/program_seeder.go` (TAHFIDZ, KITAB)
   — `CreateProgramUseCase.ExecuteSeed` ternyata tidak terpakai.
3. **Public endpoint** `GET /api/v1/web/akademik/programs/active` ditambahkan karena
   pendaftar PSB (member) tidak punya permission `manage_akademik` untuk memilih program.
4. **`program_id`** diterima di 3 titik pembuatan santri:
   - Admin create santri (`POST /santri/admin`) — optional, fallback ke default.
   - Approve santri request (`POST /santri/admin/requests/:id/approve`) — optional,
     fallback ke default (user request non-PSB).
   - PSB generate NIS (via `CreateSantriFromPendaftaran.ProgramID`) — wajib terisi.

## Verifikasi

1. `go build ./...` dan `go test ./...` lolos
2. Migration berjalan tanpa error
3. **Akademik Settings**: GET/PUT `/admin/settings` berfungsi, validasi program exists
4. **PSB Flow**: Submit pendaftaran tanpa program → ditolak. Dengan program → santri dibuat + santri_program record terbentuk
5. **Admin Create Santri**: Tanpa program → pakai default dari settings. Dengan program → pakai yang dipilih
6. **Santri Program**: Seorang santri hanya bisa punya 1 program aktif (unique index)
7. **Frontend**: Typecheck lolos, UI berfungsi

## Catatan Tambahan

- **Backward compatibility**: Santri existing yang sudah punya `Program` string bisa di-migrate ke `santri_programs` via seeder/manual (resolve by code)
- **Program di Santri entity**: Tetap dipertahankan sebagai cache, tapi bukan source of truth
- **Extensibility**: Tambah setting baru di akademik cukup tambah constant + mapper, tanpa migration
- **Eventual consistency**: Jika call ke akademik contract gagal saat create santri, log warning (best-effort) atau retry — tergantung kebutuhan
