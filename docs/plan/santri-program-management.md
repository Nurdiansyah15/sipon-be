# Plan: Manajemen Program Santri & Request Transfer Program

## Context

Santri aktif dikelompokkan ke dalam **program** melalui `santri_programs`. Saat ini,
penugasan program santri hanya terjadi secara implisit saat santri dibuat (PSB / admin create /
approve request). Tidak ada endpoint admin untuk mengubah program santri setelah santri dibuat,
dan tidak ada mekanisme bagi santri untuk mengajukan pindah program.

**Kebutuhan:**
1. **Admin manage santri program** — admin bisa langsung assign / ubah program santri
2. **Santri request transfer program** — santri bisa mengajukan pindah program (satu arah,
   bukan menambah program baru — karena saat ini hanya ada satu jenis program aktif per santri)
3. **Approval flow** — admin approve/reject request transfer santri

**Kendala:**
- `santri_programs` memiliki unique index `WHERE is_active=true AND deleted_at IS NULL`,
  sehingga hanya 1 program aktif per santri.
- Saat ini kode hanya mendukung satu jenis program (move only, bukan multi-program).

---

## Alur Sistem

### A. Admin Assign / Ubah Program

```
Admin memilih santri → pilih program → submit
    │
    ▼
Backend:
    • Validasi santri aktif & program aktif
    • Jika santri sudah punya program aktif (sama) → no-op (idempotent)
    • Jika berbeda → deactivate yang lama, buat yang baru
    • Return info program terbaru santri
```

### B. Santri Request Transfer Program

```
Santri memilih program tujuan → submit request (status: pending)
    │
    ▼
Admin review request:
    ├── Approve → deactivate program lama, buat program baru
    └── Reject  → set status rejected, simpan catatan admin
```

---

## Keputusan Desain

| No | Pertanyaan | Keputusan |
|----|-----------|-----------|
| 1 | Endpoint admin untuk assign program? | `PUT /admin/santri/:santriId/program` di module **akademik** |
| 2 | Endpoint admin untuk lihat program santri? | `GET /admin/santri/:santriId/program` di module **akademik** |
| 3 | Endpoint santri untuk list program-nya? | `GET /my/program` (sudah ada, tidak berubah) |
| 4 | Di mana entity request transfer? | Module **akademik** (karena efeknya di `santri_programs`) |
| 5 | Nama entity request? | `program_transfer_request` — table baru |
| 6 | Status request? | `pending` → `approved` / `rejected` (sama pola SantriRequest) |
| 7 | Apakah santri bisa punya request pending ganda? | **Tidak** — unique index `WHERE status='pending'` |
| 8 | Apakah santri bisa request program yang sama? | **Tidak** — validasi dari_program_id != to_program_id |
| 9 | Validasi santri saat request? | Harus `status=SANTRI` (cek via kesantrian port) |
| 10 | Siapa yang approve? | Admin dengan permission `manage_akademik` |
| 11 | Apa yang terjadi saat approve? | Deactivate program lama → create program baru (atomic via transactor) |

---

## Database

### Migration: `migrations/YYYYMMDDHHMMSS_create_program_transfer_requests.up.sql`

```sql
CREATE TABLE program_transfer_requests (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_id       UUID NOT NULL,
    from_program_id UUID NOT NULL REFERENCES programs(id),
    to_program_id   UUID NOT NULL REFERENCES programs(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    notes           TEXT,
    admin_notes     TEXT,
    reviewed_by     UUID,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Hanya 1 request pending per santri
CREATE UNIQUE INDEX idx_program_transfer_requests_pending_unique
    ON program_transfer_requests(santri_id)
    WHERE status = 'pending' AND deleted_at IS NULL;

CREATE INDEX idx_program_transfer_requests_santri
    ON program_transfer_requests(santri_id)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_program_transfer_requests_status
    ON program_transfer_requests(status)
    WHERE deleted_at IS NULL;
```

---

## Backend — Module Akademik

### A. Domain: Program Transfer Request

#### A.1 Constant — `akademik/domain/program_transfer_request/constant/` (BARU)

```go
type ProgramTransferRequestStatus string

const (
    StatusPending  ProgramTransferRequestStatus = "pending"
    StatusApproved ProgramTransferRequestStatus = "approved"
    StatusRejected ProgramTransferRequestStatus = "rejected"
)

const (
    CodeProgramTransferRequestNotFound          kernel.Code = "PROGRAM_TRANSFER_REQUEST_NOT_FOUND"
    CodeProgramTransferRequestInvalidStatus     kernel.Code = "PROGRAM_TRANSFER_REQUEST_INVALID_STATUS"
    CodeProgramTransferRequestDuplicate         kernel.Code = "PROGRAM_TRANSFER_REQUEST_DUPLICATE"
    CodeProgramTransferRequestSameProgram       kernel.Code = "PROGRAM_TRANSFER_REQUEST_SAME_PROGRAM"
    CodeProgramTransferRequestPersistenceFailed kernel.Code = "PROGRAM_TRANSFER_REQUEST_PERSISTENCE_FAILED"
    CodeProgramTransferRequestQueryFailed       kernel.Code = "PROGRAM_TRANSFER_REQUEST_QUERY_FAILED"
)
```

#### A.2 Entity — `akademik/domain/program_transfer_request/entity/` (BARU)

```go
type ProgramTransferRequest struct {
    ID            string
    SantriID      string
    FromProgramID string
    ToProgramID   string
    Status        constant.ProgramTransferRequestStatus
    Notes         *string
    AdminNotes    *string
    ReviewedBy    *string
    ReviewedAt    *time.Time
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

func NewProgramTransferRequest(id, santriID, fromProgramID, toProgramID string, notes *string) (*ProgramTransferRequest, error)
func (r *ProgramTransferRequest) Approve(adminID string) error
func (r *ProgramTransferRequest) Reject(adminID string, adminNotes *string) error
```

#### A.3 Repository — `akademik/domain/program_transfer_request/repository/interfaces.go` (BARU)

```go
type ProgramTransferRequestListQuery struct {
    SantriID *string
    Status   *string
    Page     int
    Limit    int
}

type ProgramTransferRequestListResult struct {
    Items []*entity.ProgramTransferRequest
    Total int64
}

type ProgramTransferRequestRepository interface {
    Save(ctx context.Context, req *entity.ProgramTransferRequest) error
    Update(ctx context.Context, req *entity.ProgramTransferRequest) error
    FindByID(ctx context.Context, id string) (*entity.ProgramTransferRequest, error)
    FindPendingBySantriID(ctx context.Context, santriID string) (*entity.ProgramTransferRequest, error)
    List(ctx context.Context, query ProgramTransferRequestListQuery) (*ProgramTransferRequestListResult, error)
}
```

#### A.4 Postgres Repo — `akademik/infrastructure/persistence/postgres_program_transfer_request_repo.go` (BARU)

### B. Commands

#### B.1 Admin Assign Program — `akademik/application/command/assign_santri_program_admin.go` (BARU)

Endpoint: `PUT /admin/santri/:santriId/program`

```go
type AssignSantriProgramAdminUseCase struct {
    sessionRepo       sesRepo.ActivitySessionRepository // (not needed here, just using existing)
    santriProgramRepo spRepo.SantriProgramRepository
    programRepo       progRepo.ProgramRepository
    transactor        ports.Transactor
}

type AssignProgramRequest struct {
    ProgramID string `json:"program_id" binding:"required"`
}

func (uc *AssignSantriProgramAdminUseCase) Execute(ctx context.Context, santriID, programID string) (*dto.SantriProgramAdminResponse, error) {
    // 1. Validasi program exists & active
    // 2. Cek program aktif santri saat ini
    // 3. Jika sama → idempotent, return current
    // 4. Dalam transactor: deactivate all + save baru
    // 5. Return info program baru
}
```

**Catatan**: Use case existing `AssignSantriProgramUseCase` (yang di-expose via contract) tidak
pakai transactor. Use case baru ini pakai transactor karena langsung dipanggil admin (perlu
atomicity). Existing contract-based use case tetap dipertahankan untuk backward compatibility.

#### B.2 Santri Request Transfer — `akademik/application/command/request_program_transfer.go` (BARU)

Endpoint: `POST /my/program-transfer-requests`

```go
type RequestProgramTransferUseCase struct {
    ptrRepo          ptrRepo.ProgramTransferRequestRepository
    santriProgramRepo spRepo.SantriProgramRepository
    programRepo       progRepo.ProgramRepository
    kesantrianReader  ports.KesantrianReader
    santriResolver    *application.SantriResolver
}

type RequestProgramTransferRequest struct {
    ToProgramID string  `json:"to_program_id" binding:"required"`
    Notes       *string `json:"notes"`
}

func (uc *RequestProgramTransferUseCase) Execute(ctx context.Context, userID string, req dto.RequestProgramTransferRequest) (*dto.ProgramTransferRequestResponse, error) {
    // 1. Resolve santri dari JWT (userID → santri)
    // 2. Validasi santri aktif (status=SANTRI)
    // 3. Validasi to_program exists & active
    // 4. Validasi dari != to
    // 5. Get current active program (from_program_id)
    // 6. Cek tidak ada pending request (FindPendingBySantriID)
    // 7. Create request (status=pending)
    // 8. Return response
}
```

#### B.3 Admin Approve Request — `akademik/application/command/approve_program_transfer.go` (BARU)

Endpoint: `POST /admin/program-transfer-requests/:id/approve`

```go
type ApproveProgramTransferUseCase struct {
    ptrRepo           ptrRepo.ProgramTransferRequestRepository
    santriProgramRepo spRepo.SantriProgramRepository
    programRepo       progRepo.ProgramRepository
    transactor        ports.Transactor
}

func (uc *ApproveProgramTransferUseCase) Execute(ctx context.Context, requestID, adminID string) (*dto.ProgramTransferRequestResponse, error) {
    // 1. Find request by ID
    // 2. request.Approve(adminID)
    // 3. Dalam transactor:
    //    a. Update request → status=approved, reviewed_at=now
    //    b. Deactivate all santri_programs untuk santri ini
    //    c. Create new santri_program (to_program_id)
    // 4. Return response
}
```

#### B.4 Admin Reject Request — `akademik/application/command/reject_program_transfer.go` (BARU)

Endpoint: `POST /admin/program-transfer-requests/:id/reject`

```go
type RejectProgramTransferUseCase struct {
    ptrRepo ptrRepo.ProgramTransferRequestRepository
}

type RejectRequest struct {
    AdminNotes *string `json:"admin_notes"`
}

func (uc *RejectProgramTransferUseCase) Execute(ctx context.Context, requestID, adminID string, req dto.RejectRequest) (*dto.ProgramTransferRequestResponse, error) {
    // 1. Find request by ID
    // 2. request.Reject(adminID, adminNotes)
    // 3. Update request → status=rejected, reviewed_at=now
    // 4. Return response
}
```

### C. Queries

#### C.1 Admin List Requests — `akademik/application/query/list_program_transfer_requests.go` (BARU)

Endpoint: `GET /admin/program-transfer-requests`

#### C.2 Admin Get Request Detail — `akademik/application/query/get_program_transfer_request.go` (BARU)

Endpoint: `GET /admin/program-transfer-requests/:id`

#### C.3 Santri List My Transfer Requests — `akademik/application/query/list_my_program_transfer_requests.go` (BARU)

Endpoint: `GET /my/program-transfer-requests`

### D. DTO — `akademik/application/dto/program_transfer_request_dto.go` (BARU)

```go
type ProgramTransferRequestResponse struct {
    ID            string  `json:"id"`
    SantriID      string  `json:"santri_id"`
    SantriName    *string `json:"santri_name,omitempty"`
    FromProgramID string  `json:"from_program_id"`
    FromProgram   *ProgramBrief `json:"from_program,omitempty"`
    ToProgramID   string  `json:"to_program_id"`
    ToProgram     *ProgramBrief `json:"to_program,omitempty"`
    Status        string  `json:"status"`
    Notes         *string `json:"notes,omitempty"`
    AdminNotes    *string `json:"admin_notes,omitempty"`
    ReviewedBy    *string `json:"reviewed_by,omitempty"`
    ReviewedAt    *string `json:"reviewed_at,omitempty"`
    CreatedAt     string  `json:"created_at"`
}

type SantriProgramAdminResponse struct {
    SantriID  string        `json:"santri_id"`
    ProgramID string        `json:"program_id"`
    Program   ProgramBrief  `json:"program"`
    IsActive  bool          `json:"is_active"`
}
```

### E. HTTP Router — UPDATE `akademik/interfaces/http/router.go`

```go
// Santri portal (JWT, tanpa permission)
akademik.POST("/my/program-transfer-requests", h.RequestProgramTransfer)
akademik.GET("/my/program-transfer-requests", h.ListMyProgramTransferRequests)

// Admin (JWT + permission manage_akademik)
admin := akademik.Group("/admin")
admin.Use(middleware.RequirePermission("manage_akademik"))
{
    // Santri program management
    admin.PUT("/santri/:santriId/program", h.AssignSantriProgramAdmin)
    admin.GET("/santri/:santriId/program", h.GetSantriProgramAdmin)

    // Program transfer requests
    admin.GET("/program-transfer-requests", h.ListProgramTransferRequests)
    admin.GET("/program-transfer-requests/:id", h.GetProgramTransferRequest)
    admin.POST("/program-transfer-requests/:id/approve", h.ApproveProgramTransferRequest)
    admin.POST("/program-transfer-requests/:id/reject", h.RejectProgramTransferRequest)
}
```

### F. Module Wiring — UPDATE `akademik/module.go`

- Init `programTransferRequestRepo`
- Create use cases: assign admin, request transfer, approve, reject, list, get, list-my
- Wire handler & router

---

## Struktur Module

```
internal/modules/akademik/
  domain/
    program_transfer_request/
      constant/
        program_transfer_request_constant.go     ← BARU
      entity/
        program_transfer_request.go              ← BARU
        program_transfer_request_test.go         ← BARU
      repository/
        interfaces.go                            ← BARU
  application/
    command/
      assign_santri_program_admin.go             ← BARU
      request_program_transfer.go                ← BARU
      approve_program_transfer.go                ← BARU
      reject_program_transfer.go                 ← BARU
    query/
      list_program_transfer_requests.go          ← BARU
      get_program_transfer_request.go            ← BARU
      list_my_program_transfer_requests.go       ← BARU
    dto/
      program_transfer_request_dto.go            ← BARU
  infrastructure/
    persistence/
      postgres_program_transfer_request_repo.go  ← BARU
  interfaces/http/
    handler.go                                   ← UPDATE: tambah handler baru
    router.go                                    ← UPDATE: tambah routes baru
  module.go                                      ← UPDATE: wiring baru
```

---

## Fase Pengerjaan

### Fase 1 — Database & Domain Layer
- [ ] Migration: `create_program_transfer_requests`
- [ ] Domain: constant, entity, repository interface
- [ ] Entity test (state machine: pending → approved/rejected)
- [ ] `go build ./...`

### Fase 2 — Persistence
- [ ] Postgres repo: `postgres_program_transfer_request_repo.go`
- [ ] `go build ./...`

### Fase 3 — Commands
- [ ] `AssignSantriProgramAdminUseCase` (admin assign/ubah program)
- [ ] `RequestProgramTransferUseCase` (santri ajukan transfer)
- [ ] `ApproveProgramTransferUseCase` (admin approve + efek ke santri_programs)
- [ ] `RejectProgramTransferUseCase` (admin reject)
- [ ] DTO: `program_transfer_request_dto.go`
- [ ] `go build ./...`

### Fase 4 — Queries
- [ ] `ListProgramTransferRequestsUseCase` (admin list)
- [ ] `GetProgramTransferRequestUseCase` (admin detail)
- [ ] `ListMyProgramTransferRequestsUseCase` (santri list miliknya)
- [ ] `go build ./...`

### Fase 5 — HTTP & Wiring
- [ ] Handler: semua handler baru
- [ ] Router: routes baru (admin + santri portal)
- [ ] Wiring di `module.go`
- [ ] `go build ./...`

### Fase 6 — Testing
- [ ] Unit test entity (state machine)
- [ ] Unit test approve: deactivate lama + create baru
- [ ] Unit test reject: status berubah, santri_programs tidak berubah
- [ ] Unit test request: tidak bisa request program sama, tidak bisa double pending
- [ ] Unit test admin assign: idempotent jika program sama
- [ ] `go test ./...`

### Fase 7 — Frontend: Types & Store
- [ ] `shared/types/Akademik.ts` — tambah type `SantriProgramAdminResponse`,
      `ProgramTransferRequest`, `RequestProgramTransferRequest`, dll.
- [ ] `app/stores/akademik.ts` — tambah actions:
      `fetchSantriProgram`, `assignSantriProgram`,
      `fetchProgramTransferRequests`, `approveProgramTransferRequest`,
      `rejectProgramTransferRequest`
- [ ] `app/stores/akademik-santri.ts` — tambah actions:
      `requestProgramTransfer`, `fetchMyProgramTransferRequests`
- [ ] Typecheck: `npx nuxi typecheck`

### Fase 8 — Frontend: Admin Santri Program Management
- [ ] `app/pages/admin/akademik/program/[id]/index.vue` atau
      `app/pages/admin/kesantrian/santri/[id].vue` — tambahkan section/panel
      "Program Santri" yang menampilkan program aktif santri
- [ ] Tombol "Ubah Program" → buka modal `AdminAssignProgramModal.vue`
- [ ] `app/components/admin/akademik/AssignProgramModal.vue` — modal dengan
      dropdown program aktif + konfirmasi ubah program (gaya `ConfirmActionModal`)
- [ ] Integrasikan dengan list santri — kolom baru "Program" di tabel santri
      admin kesantrian
- [ ] Typecheck

### Fase 9 — Frontend: Admin Review Program Transfer Requests
- [ ] `app/pages/admin/akademik/program-transfer-requests/index.vue` — list
      request transfer (filter status, pagination, tabel santri + program asal/tujuan)
- [ ] Row action: tombol "Setujui" (approve) dan "Tolak" (reject)
- [ ] `app/components/admin/akademik/ApproveProgramTransferModal.vue` — konfirmasi
      approve (tampilkan from → to, tombol konfirmasi)
- [ ] `app/components/admin/ConfirmActionModal.vue` (reuse) — reject dengan textarea
      admin_notes
- [ ] Navigasi di sidebar akademik: item "Permintaan Pindah Program"
- [ ] Typecheck

### Fase 10 — Frontend: Santri Portal — Request Transfer & History
- [ ] `app/pages/akademik/program.vue` — halaman program santri:
      - Kartu program aktif saat ini
      - Tombol "Ajukan Pindah Program" (jika tidak ada pending request)
      - History request transfer (tabel status: pending/approved/rejected)
- [ ] `app/components/akademik/RequestProgramTransferModal.vue` — modal form:
      dropdown program tujuan (exclude program saat ini), textarea notes opsional
- [ ] Badge/tag status request (pending=yellow, approved=green, rejected=red)
- [ ] Link navigasi di dashboard santri (`app/pages/akademik/index.vue`):
      kartu "Program" atau link "Lihat Program & Ajukan Pindah"
- [ ] Typecheck

### Fase 11 — Frontend: Integration Testing & Smoke Test
- [ ] Smoke test admin assign program:
      - Pilih santri → ubah program → program berubah
      - Ubah ke program yang sama → no-op, tidak error
- [ ] Smoke test santri request:
      - Ajukan pindah → request muncul di history (pending)
      - Tidak bisa ajukan 2x (button disabled / error)
      - Program tujuan sama dengan saat ini → ditolak
- [ ] Smoke test admin approve/reject:
      - Approve → santri pindah program, request jadi approved
      - Reject → santri program tetap, request jadi rejected
      - Request yang sudah diproses tidak bisa di-approve/reject lagi
- [ ] Typecheck final: `npx nuxi typecheck`

---

## Frontend (sipon-ui)

### Tech Stack
- **Nuxt 4** (Vue 3) + **Nuxt UI v4** + **Tailwind CSS v4**
- **Pinia** (state), **Zod** (form validation), **Lucide** icons
- API base: `NUXT_PUBLIC_API_BASE` → `http://localhost:8888`

### Komponen Baru

| Komponen | Lokasi | Deskripsi |
|----------|--------|-----------|
| `AssignProgramModal` | `app/components/admin/akademik/` | Admin assign/ubah program santri |
| `ApproveProgramTransferModal` | `app/components/admin/akademik/` | Admin konfirmasi approve transfer |
| `RequestProgramTransferModal` | `app/components/akademik/` | Santri form ajukan pindah program |

### Halaman Baru / Update

| Halaman | Deskripsi |
|---------|-----------|
| `admin/akademik/program-transfer-requests` | List & review request transfer (admin) |
| `akademik/program` | Program santri + ajukan pindah + history (santri) |
| `admin/kesantrian/santri/[id]` (update) | Tambah section program santri + tombol ubah |
| `akademik/index` (update) | Tambah link ke halaman program |

### Store Actions Baru

**`useAkademikStore`** (`app/stores/akademik.ts`):
```ts
fetchSantriProgram(santriId: string)                    // GET /admin/santri/:id/program
assignSantriProgram(santriId: string, programId: string) // PUT /admin/santri/:id/program
fetchProgramTransferRequests(query?)                      // GET /admin/program-transfer-requests
approveProgramTransferRequest(id: string)                 // POST /admin/program-transfer-requests/:id/approve
rejectProgramTransferRequest(id: string, adminNotes?: string) // POST /admin/program-transfer-requests/:id/reject
```

**`useAkademikSantriStore`** (`app/stores/akademik-santri.ts`):
```ts
requestProgramTransfer(toProgramId: string, notes?: string) // POST /my/program-transfer-requests
fetchMyProgramTransferRequests()                            // GET /my/program-transfer-requests
```

### Type Definitions Baru (`shared/types/Akademik.ts`)

```ts
export interface SantriProgramAdminResponse {
  santri_id: string
  program_id: string
  program: ProgramBrief
  is_active: boolean
}

export interface ProgramTransferRequest {
  id: string
  santri_id: string
  santri_name?: string | null
  from_program_id: string
  from_program?: ProgramBrief | null
  to_program_id: string
  to_program?: ProgramBrief | null
  status: 'pending' | 'approved' | 'rejected'
  notes?: string | null
  admin_notes?: string | null
  reviewed_by?: string | null
  reviewed_at?: string | null
  created_at: string
}

export interface ProgramBrief {
  id: string
  code: string
  name: string
}
```

### Navigasi

**Sidebar Admin Akademik** — tambah item:
- Label: "Permintaan Pindah Program"
- Icon: `i-lucide-arrow-right-left`
- Route: `/admin/akademik/program-transfer-requests`
- Permission: `manage_akademik`

**Sidebar Santri Portal Akademik** — tambah item:
- Label: "Program Saya"
- Icon: `i-lucide-graduation-cap`
- Route: `/akademik/program`

---

## Verifikasi

1. `go build ./...` dan `go test ./...` lolos di tiap fase.
2. **Admin assign program**:
   - `PUT /admin/santri/:id/program` → santri pindah program.
   - Request dengan program yang sama → no-op (idempotent).
   - Request program tidak aktif → ditolak.
   - Request untuk santri tidak ditemukan → 404.
3. **Santri request transfer**:
   - `POST /my/program-transfer-requests` → request dibuat (status=pending).
   - Tidak bisa request program yang sama → 422.
   - Sudah ada pending request → 409 conflict.
   - Santri tidak aktif → 422.
4. **Admin approve**:
   - `POST /admin/program-transfer-requests/:id/approve` → santri pindah program.
   - Program lama deactivated, program baru activated.
   - Request yang sudah approved/rejected tidak bisa di-approve lagi → 422.
5. **Admin reject**:
   - `POST /admin/program-transfer-requests/:id/reject` → status=rejected.
   - Santri program tidak berubah.
   - Request yang sudah diproses tidak bisa di-reject lagi → 422.
6. **List & detail**:
   - Admin bisa list semua request (filter status, pagination).
   - Santri hanya bisa list request miliknya sendiri.

---

## Catatan

- **Backward compatibility**: `AssignSantriProgramUseCase` existing (via contract) tetap
  dipertahankan. Use case baru `AssignSantriProgramAdminUseCase` adalah versi yang
  atomic (pakai transactor) untuk admin.
- **Cross-module**: Santri resolver (resolve santriID dari JWT userID) sudah tersedia
  di `application/santri_resolver.go`.
- **ProgramBrief**: DTO shared untuk nested program info, sudah ada pattern di
  `get_santri_program.go` (`SantriProgramInfo`).
- **Extensibility**: Jika nanti perlu multi-program, flow request bisa diextend menjadi
  "request tambah program" (create tanpa deactivate). Saat ini cukup "pindah" saja.
