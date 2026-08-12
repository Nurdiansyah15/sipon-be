# Plan: Perluasan Herregistrasi — Status Revisi & Dokumen

## Context

Herregistrasi saat ini hanya punya 3 status: `pending → completed / cancelled`.
Tidak ada mekanisme revisi dan tidak ada fitur upload dokumen.

Kebutuhan baru:
1. **Status revisi** — admin bisa minta revisi, santri bisa upload/update dokumen, admin
   mengecek ulang sebelum menerima.
2. **Dokumen herregistrasi** — santri upload dokumen (sesuai kebutuhan periode).
3. **Blueprint dokumen per periode** — admin mengatur jenis dokumen apa yang wajib/opsional
   per periode akademik. `is_required` bisa dimatikan/dinyalakan kapan saja.
4. **Admin hanya bisa terima** jika dokumen-dokumen yang wajib sudah terverifikasi.

---

## Domain Model Baru

### 1. Herregistrasi Status (perubahan)

```
SantriRegistrationStatus:
  pending     → diajukan santri, menunggu review admin
  revision    → admin minta revisi (ada notes), santri bisa update dokumen
  completed   → diterima admin (semua dokumen wajib sudah verified)
  cancelled   → ditolak admin
```

Transisi:
```
pending   → completed   (accept)
pending   → revision    (request revision)
pending   → cancelled   (reject)
revision  → completed   (accept after revision)
revision  → revision    (re-revision, notes baru)
revision  → cancelled   (reject after revision)
```

Santri bisa upload/update dokumen selama status `pending` atau `revision`.

Field baru di `santri_registrations`:
- `revision_notes TEXT` — catatan admin saat minta revisi (nullable, diisi saat transisi ke revision).

### 2. HerregistrasiDocumentRequirement (blueprint per periode)

Menentukan jenis dokumen apa saja yang bisa di-upload pada suatu periode akademik,
dan mana yang wajib.

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| academic_period_id | UUID FK → academic_periods | |
| kind | VARCHAR(50) | Kode dokumen, misal: "surat_pernyataan", "kk", "aktakelahiran" |
| label | VARCHAR(200) | Nama tampilan, misal: "Surat Pernyataan", "Kartu Keluarga" |
| is_required | BOOLEAN DEFAULT true | Wajib atau opsional |
| description | TEXT nullable | Keterangan tambahan (opsional) |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ nullable | |

Unique: `(academic_period_id, kind) WHERE deleted_at IS NULL`

**Perilaku**:
- Admin bisa menambah/menghapus requirement kapan saja.
- Admin bisa toggle `is_required` on/off.
- Saat `is_required = false`, dokumen bersifat opsional (santri bisa upload atau tidak).
- Saat `is_required = true`, santri WAJIB upload dan admin WAJIB verifikasi sebelum
  menerima herreg.

### 3. HerregistrasiDocument (dokumen yang di-upload santri)

| Field | Tipe | Keterangan |
|-------|------|-----------|
| id | UUID PK | |
| santri_registration_id | UUID FK → santri_registrations | |
| kind | VARCHAR(50) | Harus sesuai requirement kind |
| key | TEXT NOT NULL | MinIO object key |
| original_filename | VARCHAR(500) | |
| mime_type | VARCHAR(200) | |
| size | BIGINT | |
| status | VARCHAR(20) DEFAULT 'pending' | pending / verified / rejected |
| notes | TEXT nullable | Catatan admin saat reject |
| verified_by | UUID nullable | user_id verifier |
| verified_at | TIMESTAMPTZ nullable | |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |
| deleted_at | TIMESTAMPTZ nullable | |

Unique: `(santri_registration_id, kind) WHERE deleted_at IS NULL`
→ Satu dokumen per kind per herregistrasi. Upload baru = soft-delete lama + insert baru.

Constants:
```
HerregistrasiDocumentStatus: pending, verified, rejected
```

---

## Alur Data

```
Admin mengatur blueprint dokumen per periode:
  herregistrasi_document_requirements (kind, label, is_required)
      │
      ▼
Santri ajukan herreg → santri_registrations (status: pending)
      │
      ▼
Santri upload dokumen (sesuai blueprint):
  1. POST /my/herregistrasi/dokumen/presign → presign_url, key
  2. Upload langsung ke MinIO
  3. POST /my/herregistrasi/dokumen/confirm → simpan row herregistrasi_documents
      │
      ▼
Admin review:
  • Cek dokumen satu per satu → verify / reject (dengan notes)
  • Jika semua required docs verified → Complete herreg (status: completed)
  • Jika ada masalah → Request revision (status: revision, isi notes)
      │
      ▼
Jika revision:
  • Santri lihat notes dari admin
  • Santri upload ulang dokumen yang perlu diperbaiki (replace)
  • Admin review ulang
      │
      ▼
Admin terima → status: completed
```

---

## Endpoint HTTP

### Santri-facing (prefix `/my/herregistrasi`, tanpa permission)

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/my/herregistrasi` | Status herregistrasi santri (detail, termasuk dokumen) |
| POST | `/my/herregistrasi` | Ajukan herregistrasi (existing, ditambah response dokumen requirements) |
| POST | `/my/herregistrasi/dokumen/presign` | Request presign URL untuk upload dokumen |
| POST | `/my/herregistrasi/dokumen/confirm` | Konfirmasi upload setelah file sampai di MinIO |
| DELETE | `/my/herregistrasi/dokumen/:id` | Hapus dokumen (soft-delete, hanya jika status masih pending/rejected) |
| GET | `/my/herregistrasi/dokumen/:id/download` | Presigned download URL untuk dokumen |

### Admin-facing (prefix `/registrations`, butuh `manage_akademik`)

| Method | Path | Deskripsi |
|--------|------|-----------|
| POST | `/registrations/:id/complete` | Terima herreg (existing, ditambah validasi: semua required docs harus verified) |
| POST | `/registrations/:id/revision` | **Baru** — Request revisi dari admin |
| GET | `/registrations/:id/dokumen` | List dokumen herregistrasi |
| POST | `/registrations/:id/dokumen/:dokumenId/verify` | Verifikasi dokumen |
| POST | `/registrations/:id/dokumen/:dokumenId/reject` | Tolak dokumen (dengan notes) |

### Blueprint Dokumen (admin-facing, butuh `manage_akademik`)

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/periods/:periodId/dokumen-requirements` | List requirement dokumen per periode |
| POST | `/periods/:periodId/dokumen-requirements` | Tambah requirement baru |
| PUT | `/periods/:periodId/dokumen-requirements/:id` | Update requirement (label, is_required, description) |
| DELETE | `/periods/:periodId/dokumen-requirements/:id` | Hapus requirement |

---

## DTO Baru

```go
// --- Blueprint ---

type HerregistrasiDocumentRequirementResponse struct {
    ID               string `json:"id"`
    AcademicPeriodID string `json:"academic_period_id"`
    Kind             string `json:"kind"`
    Label            string `json:"label"`
    IsRequired       bool   `json:"is_required"`
    Description      *string `json:"description,omitempty"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

type CreateHerregistrasiDocumentRequirementRequest struct {
    Kind        string  `json:"kind" binding:"required"`
    Label       string  `json:"label" binding:"required"`
    IsRequired  *bool   `json:"is_required"`         // default true
    Description *string `json:"description,omitempty"`
}

type UpdateHerregistrasiDocumentRequirementRequest struct {
    Label       *string `json:"label,omitempty"`
    IsRequired  *bool   `json:"is_required,omitempty"`
    Description *string `json:"description,omitempty"`
}

// --- Dokumen ---

type HerregistrasiDocumentResponse struct {
    ID               string     `json:"id"`
    SantriRegistrationID string `json:"santri_registration_id"`
    Kind             string     `json:"kind"`
    KindLabel        string     `json:"kind_label,omitempty"`
    Key              string     `json:"key"`
    OriginalFilename *string    `json:"original_filename,omitempty"`
    MimeType         *string    `json:"mime_type,omitempty"`
    Size             *int64     `json:"size,omitempty"`
    Status           string     `json:"status"`
    Notes            *string    `json:"notes,omitempty"`
    VerifiedBy       *string    `json:"verified_by,omitempty"`
    VerifiedAt       *time.Time `json:"verified_at,omitempty"`
    CreatedAt        time.Time  `json:"created_at"`
    UpdatedAt        time.Time  `json:"updated_at"`
}

type HerregistrasiDocumentPresignRequest struct {
    Kind        string `json:"kind" binding:"required"`
    ContentType string `json:"content_type" binding:"required"`
    Filename    string `json:"filename" binding:"required"`
}

type HerregistrasiDocumentPresignResponse struct {
    PresignURL string `json:"presign_url"`
    Key        string `json:"key"`
    ExpiresIn  int    `json:"expires_in"`
}

type HerregistrasiDocumentConfirmRequest struct {
    Key              string `json:"key" binding:"required"`
    Kind             string `json:"kind" binding:"required"`
    OriginalFilename string `json:"original_filename"`
    MimeType         string `json:"mime_type"`
    Size             int64  `json:"size"`
}

type DokumenVerifyRequest struct {
    // empty body — verifier identity comes from JWT
}

type DokumenRejectRequest struct {
    Notes string `json:"notes" binding:"required"`
}

type RevisionRequest struct {
    Notes string `json:"notes" binding:"required"`
}
```

---

## Repository Changes

### SantriRegistrationRepository (tambah method)
```go
// UpdateStatus mengupdate status dan optionally revision_notes.
// Digunakan untuk transisi ke revision (dengan notes), complete, cancel.
```

### HerregistrasiDocumentRequirementRepository (baru)
```go
type HerregistrasiDocumentRequirementRepository interface {
    Save(ctx, req) error
    Update(ctx, req) error
    FindByID(ctx, id) (*entity, error)
    FindByAcademicPeriod(ctx, periodID) ([]*entity, error)
    Delete(ctx, id) error
}
```

### HerregistrasiDocumentRepository (baru)
```go
type HerregistrasiDocumentRepository interface {
    Save(ctx, doc) error
    Update(ctx, doc) error
    FindByID(ctx, id) (*entity, error)
    FindByRegistration(ctx, registrationID) ([]*entity, error)
    FindByRegistrationAndKind(ctx, registrationID, kind) (*entity, error)
    Delete(ctx, id) error
}
```

---

## Queries & Commands Baru

### Commands
| Use Case | Deskripsi |
|----------|-----------|
| `CreateHerregistrasiDocumentRequirementUseCase` | Tambah blueprint doc |
| `UpdateHerregistrasiDocumentRequirementUseCase` | Update blueprint doc (toggle is_required, label, dll) |
| `DeleteHerregistrasiDocumentRequirementUseCase` | Hapus blueprint doc |
| `PresignHerregistrasiDocumentUseCase` | Request presign URL |
| `ConfirmHerregistrasiDocumentUseCase` | Konfirmasi upload, soft-delete doc lama untuk kind yang sama, simpan baru |
| `DeleteHerregistrasiDocumentUseCase` | Santri hapus dokumen sendiri |
| `RequestRevisionUseCase` | Admin minta revisi (set status=revision + notes) |
| `VerifyHerregistrasiDocumentUseCase` | Admin verifikasi dokumen |
| `RejectHerregistrasiDocumentUseCase` | Admin tolak dokumen (dengan notes) |
| `CompleteRegistrationUseCase` (update) | Tambah validasi: semua required docs harus verified |

### Queries
| Use Case | Deskripsi |
|----------|-----------|
| `ListHerregistrasiDocumentRequirementsUseCase` | List blueprint docs per periode |
| `ListHerregistrasiDocumentsUseCase` | List dokumen untuk suatu herregistrasi |
| `GetHerregistrasiDetailUseCase` | Detail herregistrasi + status dokumen (untuk santri & admin) |

---

## Logic Penting

### CompleteRegistrationUseCase (update existing)
```go
func (uc *CompleteRegistrationUseCase) Execute(ctx, regID string) (*dto.Response, error) {
    reg := findRegistration(regID)
    if reg.Status != pending && reg.Status != revision {
        return error("invalid status")
    }

    // Check: semua required docs harus verified
    requirements := requirementRepo.FindByAcademicPeriod(ctx, reg.AcademicPeriodID)
    docs := docRepo.FindByRegistration(ctx, reg.ID)
    docMap := map docs by kind

    for _, req := range requirements {
        if req.IsRequired {
            doc, exists := docMap[req.Kind]
            if !exists {
                return error("dokumen wajib belum di-upload: " + req.Label)
            }
            if doc.Status != "verified" {
                return error("dokumen wajib belum terverifikasi: " + req.Label)
            }
        }
    }

    reg.Complete()
    regRepo.Update(ctx, reg)
    return mapToResponse(reg), nil
}
```

### ConfirmHerregistrasiDocumentUseCase
```go
func (uc *ConfirmHerregistrasiDocumentUseCase) Execute(ctx, req ConfirmRequest) (*dto.Response, error) {
    // 1. Resolve santri dari JWT
    // 2. Find herreg aktif santri (pending atau revision)
    // 3. Validate kind terdaftar di blueprint periode
    // 4. Validate key starts with "pending/"
    // 5. ConfirmUpload(key) → promote
    // 6. Soft-delete existing doc untuk (registration_id, kind) jika ada
    // 7. Save new doc (status=pending)
    // 8. Return doc response
}
```

### RequestRevisionUseCase
```go
func (uc *RequestRevisionUseCase) Execute(ctx, regID, notes string) (*dto.Response, error) {
    reg := findRegistration(regID)
    if reg.Status != pending && reg.Status != revision {
        return error("invalid status")
    }
    reg.RequestRevision(notes)
    regRepo.Update(ctx, reg)
    return mapToResponse(reg), nil
}
```

---

## FileUploader Port

Mengikuti pola PSB dan kesantrian, akademik butuh FileUploader sendiri:

**File**: `internal/modules/akademik/application/ports/file_uploader.go`
```go
type FileUploader interface {
    RequestUpload(ctx context.Context, objectName, contentType, expiry string, privacy PrivacyRule) (presignURL, key, publicURL string, err error)
    ConfirmUpload(ctx context.Context, key string) error
    DeleteObject(ctx context.Context, key string, privacy PrivacyRule) error
    PromoteUpload(ctx context.Context, stagingKey, finalKey string, privacy PrivacyRule) error
    EnsurePendingUploadLifecycle(ctx context.Context, expireDays int) error
    GeneratePresignedDownloadURL(ctx context.Context, key string, expiry int, privacy PrivacyRule) (string, error)
    PublicURL(key string) string
}

type PrivacyRule string
const (
    PrivacyPublic  PrivacyRule = "PUBLIC"
    PrivacyPrivate PrivacyRule = "PRIVATE"
)
```

**Adapter**: `internal/modules/akademik/infrastructure/minio_uploader/` — reuse logic dari
module lain (psb, kesantrian) yang sudah ada adapter MinIO.

Object key pattern: `pending/akademik/herregistrasi/{kind}/{uuid}{ext}`

---

## Migration

File: `migrations/YYYYMMDDHHMMSS_herregistrasi_revision_documents.up.sql`

```sql
-- 1. Tambah kolom revision_notes di santri_registrations
ALTER TABLE santri_registrations ADD COLUMN revision_notes TEXT;

-- 2. Update CHECK constraint untuk status baru
ALTER TABLE santri_registrations DROP CONSTRAINT IF EXISTS santri_registrations_status_check;
ALTER TABLE santri_registrations ADD CONSTRAINT santri_registrations_status_check
    CHECK (status IN ('pending', 'revision', 'completed', 'cancelled'));

-- 3. Tabel blueprint dokumen per periode
CREATE TABLE herregistrasi_document_requirements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    academic_period_id UUID NOT NULL REFERENCES academic_periods(id),
    kind VARCHAR(50) NOT NULL,
    label VARCHAR(200) NOT NULL,
    is_required BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (academic_period_id, kind)
);
CREATE INDEX idx_herreg_doc_req_period ON herregistrasi_document_requirements(academic_period_id) WHERE deleted_at IS NULL;

-- 4. Tabel dokumen herregistrasi
CREATE TABLE herregistrasi_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    santri_registration_id UUID NOT NULL REFERENCES santri_registrations(id),
    kind VARCHAR(50) NOT NULL,
    key TEXT NOT NULL,
    original_filename VARCHAR(500),
    mime_type VARCHAR(200),
    size BIGINT DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'verified', 'rejected')),
    notes TEXT,
    verified_by UUID,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (santri_registration_id, kind)
);
CREATE INDEX idx_herreg_doc_registration ON herregistrasi_documents(santri_registration_id) WHERE deleted_at IS NULL;
```

---

## Struktur Module

```
internal/modules/akademik/
  domain/
    santri_registration/
      entity/santri_registration.go     ← tambah RequestRevision(notes) method
      constant/santri_registration_constant.go  ← tambah status "revision"
    herregistrasi_document_requirement/  ← BARU
      entity/entity.go
      constant/constant.go
      repository/interfaces.go
    herregistrasi_document/              ← BARU
      entity/entity.go
      constant/constant.go
      repository/interfaces.go
  application/
    command/
      create_herregistrasi_document_requirement.go
      update_herregistrasi_document_requirement.go
      delete_herregistrasi_document_requirement.go
      presign_herregistrasi_document.go
      confirm_herregistrasi_document.go
      delete_herregistrasi_document.go
      request_revision.go
      verify_herregistrasi_document.go
      reject_herregistrasi_document.go
      complete_registration.go           ← UPDATE: tambah validasi dokumen
    query/
      list_herregistrasi_document_requirements.go
      list_herregistrasi_documents.go
      get_herregistrasi_detail.go
    dto/
      herregistrasi_document_dto.go
    ports/
      file_uploader.go                   ← BARU
  infrastructure/
    persistence/
      postgres_herregistrasi_document_requirement_repo.go
      postgres_herregistrasi_document_repo.go
      postgres_santri_registration_repo.go  ← UPDATE: tambah UpdateStatus/RevisionNotes
    minio_uploader/
      uploader.go                        ← BARU: adapter FileUploader untuk akademik
  interfaces/http/
    handler.go                           ← UPDATE: tambah handler baru
    router.go                            ← UPDATE: tambah routes baru
```

---

## Allowed Content Types

Mengikuti pola PSB/kesantrian:
```
image/jpeg, image/png, application/pdf
```

Maksimal ukuran: 10 MB per dokumen (configured via MinIO).

---

## Fase Pengerjaan

### Fase 1 — Migration & Entity
- [ ] Migration SQL (alter santri_registrations + 2 tabel baru)
- [ ] Entity: `herregistrasi_document_requirement`
- [ ] Entity: `herregistrasi_document`
- [ ] Update entity `santri_registration` (tambah `RevisionNotes`, `RequestRevision` method)
- [ ] Update constants (tambah `revision`)

### Fase 2 — Repository & Persistence
- [ ] Interface + Postgres repo: `HerregistrasiDocumentRequirementRepository`
- [ ] Interface + Postgres repo: `HerregistrasiDocumentRepository`
- [ ] Update `SantriRegistrationRepository` (support `revision_notes`)
- [ ] FileUploader port + MinIO adapter

### Fase 3 — Commands & Queries
- [ ] Blueprint: create, update, delete requirement
- [ ] Dokumen: presign, confirm, delete
- [ ] Admin: verify document, reject document, request revision
- [ ] Update `CompleteRegistrationUseCase` (validasi dokumen)
- [ ] Queries: list requirements, list documents, get herregistrasi detail

### Fase 4 — Handler & Router
- [ ] Admin routes untuk blueprint (`/periods/:id/dokumen-requirements`)
- [ ] Admin routes untuk review dokumen (`/registrations/:id/dokumen/...`)
- [ ] Admin route untuk request revision (`/registrations/:id/revision`)
- [ ] Santri routes untuk upload dokumen (`/my/herregistrasi/dokumen/...`)
- [ ] Wire di module.go

### Fase 5 — Frontend (santri portal)
- [ ] Update dashboard `/akademik` — tampilkan dokumen requirements + status upload
- [ ] Halaman upload dokumen herregistrasi
- [ ] Update status badge untuk `revision`

### Fase 6 — Frontend (admin)
- [ ] Halaman blueprint dokumen per periode
- [ ] Halaman review herregistrasi + verifikasi/reject dokumen
- [ ] Tombol request revision dengan notes
- [ ] Validasi: tombol "Terima" disabled jika required docs belum verified

---

## Verifikasi

1. `go build ./...` lolos di tiap fase.
2. Migration `up`/`down` bersih.
3. Smoke test:
   - Admin buat requirement: "Surat Pernyataan" (required), "KK" (required), "Akta" (optional)
   - Santri ajukan herreg → status pending
   - Santri upload 3 dokumen (2 required + 1 optional)
   - Admin verifikasi "Surat Pernyataan", reject "KK" (notes: "foto buram")
   - Admin request revision (notes: "KK perlu difoto ulang")
   - Status herreg: revision
   - Santri upload ulang KK
   - Admin verifikasi KK → semua required verified → admin complete → completed
4. Test toggle is_required: admin ubah "Akta" dari optional ke required → santri
   sekarang wajib upload.
5. Test constraint: upload dokumen dengan kind yang tidak ada di blueprint → ditolak.
6. Test upload duplikat: upload kind yang sama → soft-delete lama + insert baru.
7. Test complete tanpa required docs verified → error.
