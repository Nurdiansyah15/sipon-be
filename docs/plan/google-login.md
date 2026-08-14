# Plan: Google Login & Linked Accounts

## Context

Sistem SIPON perlu menambahkan fitur login menggunakan Google OAuth untuk kemudahan akses user. Implementasi akan mengikuti pola yang sama dengan project k-forum-api/k-forum-backoffice untuk konsistensi arsitektur.

### Referensi Implementasi
- `k-forum-api/internal/infrastructure/external/googleoauth/verifier.go` - Google OAuth verifier
- `k-forum-api/internal/app/usecase/auth/google_login.go` - Google login usecase
- `k-forum-api/internal/app/usecase/usersettings/` - Linked accounts usecases

## Lingkup

### Yang Dikerjakan
1. Google OAuth verification
2. Google login flow (link existing account / create new account)
3. Linked accounts management (get status, link Google, unlink Google)
4. Set password untuk akun Google-only

### Yang Tidak Dikerjakan
- Apple Sign In (bisa ditambahkan nanti dengan pola sama)
- Link/unlink Google dari UI (hanya API endpoint)
- Multiple Google accounts per user

## Keputusan Arsitektur

1. **Credential Type baru `GOOGLE`** - Ditambahkan di `user_constant.go` terpisah dari `LOCAL`.
2. **Login Identifier Kind baru `GOOGLE`** - Untuk menyimpan Google Subject ID.
3. **Auto-link pattern** - Jika user sudah punya email yang sama, Google identity akan otomatis di-link ke akun tersebut.
4. **Unlink guard** - Unlink Google hanya boleh jika user sudah punya password lokal (mencegah user mengunci diri sendiri).
5. **Email verified check** - Google login hanya diterima jika email dari Google sudah verified.

## Domain Model Changes

### User Constant Additions

```go
// di internal/modules/identity/domain/user/constant/user_constant.go

type CredentialType string

const (
    CredentialTypeLocal  CredentialType = "LOCAL"
    CredentialTypeGoogle CredentialType = "GOOGLE"  // BARU
)

type LoginIdentifierKind string

const (
    LoginIdentifierKindEmail    LoginIdentifierKind = "EMAIL"
    LoginIdentifierKindPhone    LoginIdentifierKind = "PHONE"
    LoginIdentifierKindUsername LoginIdentifierKind = "USERNAME"
    LoginIdentifierKindNIS      LoginIdentifierKind = "NIS"
    LoginIdentifierKindGoogle   LoginIdentifierKind = "GOOGLE"  // BARU
)

// Error codes baru
const (
    ErrCodeGoogleNotLinked              kernel.Code = "GOOGLE_NOT_LINKED"
    ErrCodeGoogleAlreadyLinked          kernel.Code = "GOOGLE_ALREADY_LINKED"
    ErrCodeGoogleSubAlreadyUsed         kernel.Code = "GOOGLE_SUB_ALREADY_USED"
    ErrCodeGoogleUnlinkRequiresPassword kernel.Code = "GOOGLE_UNLINK_REQUIRES_PASSWORD"
    ErrCodeGoogleEmailNotVerified       kernel.Code = "GOOGLE_EMAIL_NOT_VERIFIED"
    ErrCodeGoogleTokenInvalid           kernel.Code = "GOOGLE_TOKEN_INVALID"
)
```

### User Entity Additions

```go
// di internal/modules/identity/domain/user/entity/user.go

// UnlinkGoogleCredential melepas credential + login identity GOOGLE dari user.
// Ditolak kalau user belum punya password lokal (mencegah lockout).
func (u *User) UnlinkGoogleCredential() error {
    if !u.HasLocalPassword() {
        return kernel.WrapMsg(constant.ErrCodeGoogleUnlinkRequiresPassword,
            "Wajib memiliki password lokal sebelum unlink Google", nil)
    }
    google := u.FindCredential(constant.CredentialTypeGoogle)
    if google == nil || google.DeletedAt != nil {
        return kernel.WrapMsg(constant.ErrCodeGoogleNotLinked,
            "Akun Google tidak tertaut", nil)
    }

    now := time.Now()
    google.DeletedAt = &now
    google.UpdatedAt = now
    for _, identity := range google.LoginIdentities {
        if identity.DeletedAt == nil {
            identity.DeletedAt = &now
            identity.UpdatedAt = now
        }
    }
    u.UpdatedAt = now
    return nil
}

// LinkGoogleCredential menambahkan Google credential ke user.
// Ditolak kalau user sudah punya Google credential aktif.
func (u *User) LinkGoogleCredential(credentialID, googleSub string) error {
    existing := u.FindCredential(constant.CredentialTypeGoogle)
    if existing != nil && existing.DeletedAt == nil {
        return kernel.WrapMsg(constant.ErrCodeGoogleAlreadyLinked,
            "Akun Google sudah tertaut", nil)
    }

    now := time.Now()
    googleCred := &Credential{
        ID:        credentialID,
        UserID:    u.ID,
        Type:      constant.CredentialTypeGoogle,
        IsPrimary: false,
        UpdatedAt: now,
    }

    googleIdentity, err := NewLoginIdentity(
        uuid.NewString(),
        u.ID,
        googleCred.ID,
        constant.LoginIdentifierKindGoogle,
        googleSub,
        false,
        &now,
    )
    if err != nil {
        return err
    }
    googleIdentity.Status = constant.LoginIdentityStatusVerified
    googleCred.AddLoginIdentity(googleIdentity)
    u.AddCredential(googleCred)
    u.UpdatedAt = now
    return nil
}
```

### Credential Entity Additions

```go
// di internal/modules/identity/domain/user/entity/credential.go

func NewGoogleCredential(id, userID string, isPrimary bool) *Credential {
    now := time.Now()
    return &Credential{
        ID:        id,
        UserID:    userID,
        Type:      constant.CredentialTypeGoogle,
        IsPrimary: isPrimary,
        UpdatedAt: now,
    }
}
```

## Infrastruktur

### Google OAuth Verifier

**File baru**: `internal/modules/identity/infrastructure/external/googleoauth/verifier.go`

```go
package googleoauth

// Verifier memvalidasi Google ID Token dengan memanggil endpoint tokeninfo Google.
// Response berisi: sub (Google Subject ID), email, name, picture, email_verified.
// Validasi: aud harus match dengan allowed client IDs, email harus verified.

type GoogleIdentityInfo struct {
    Subject string // Google unique user ID (sub claim)
    Email   string
    Name    string
    Picture string
}

type Verifier struct {
    client *http.Client
}

func NewVerifier() *Verifier

// VerifyIDToken memvalidasi ID token dan mengembalikan identitas user.
// allowedClientIDs untuk mendukung multiple client IDs (web, mobile, dev, prod).
func (v *Verifier) VerifyIDToken(ctx context.Context, idToken string, allowedClientIDs []string) (*GoogleIdentityInfo, error)
```

### Port Interface

**File baru**: `internal/modules/identity/application/ports/google_oauth.go`

```go
package ports

type GoogleIdentityInfo struct {
    Subject string
    Email   string
    Name    string
    Picture string
}

type GoogleOAuthVerifier interface {
    VerifyIDToken(ctx context.Context, idToken string, allowedClientIDs []string) (*GoogleIdentityInfo, error)
}
```

## Application Layer

### Google Login UseCase

**File baru**: `internal/modules/identity/application/command/google_login.go`

```go
package command

type GoogleLoginUseCase struct {
    userRepo              userrepo.UserRepository
    tokenGen              ports.TokenGenerator
    sessionRepo           ports.SessionRepository
    googleVerifier        ports.GoogleOAuthVerifier
    allowedGoogleClientID []string
    transactor            ports.Transactor
}

func NewGoogleLoginUseCase(
    userRepo userrepo.UserRepository,
    tokenGen ports.TokenGenerator,
    sessionRepo ports.SessionRepository,
    googleVerifier ports.GoogleOAuthVerifier,
    allowedGoogleClientIDs []string,
    transactor ports.Transactor,
) *GoogleLoginUseCase

// Execute melakukan login dengan Google ID token.
// Flow:
// 1. Verifikasi ID token dengan Google tokeninfo endpoint
// 2. Cari user berdasarkan Google identity (sub)
// 3. Jika tidak ditemukan, cari berdasarkan email
// 4. Jika email ditemukan, link Google identity ke user tersebut
// 5. Jika tidak ada email match, buat user baru dengan Google + Local credential
// 6. Generate token pair dan return
func (uc *GoogleLoginUseCase) Execute(ctx context.Context, req dto.GoogleLoginRequest, ipAddress string) (*dto.LoginResponse, error)
```

### Flow Detail

```
┌─────────────────────────────────────────────────────────────────┐
│                     Google Login Flow                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Client                                                         │
│    │                                                            │
│    ▼                                                            │
│  POST /auth/login/google { id_token: "...", device_id: "..." } │
│    │                                                            │
│    ▼                                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Verify ID Token with Google tokeninfo endpoint           │  │
│  │ - Check audience (client_id)                             │  │
│  │ - Check email_verified = true                            │  │
│  │ - Extract: sub, email, name, picture                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│    │                                                            │
│    ▼                                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ FindByIdentity(GOOGLE, sub)                              │  │
│  └──────────────────────────────────────────────────────────┘  │
│    │                                                            │
│    ├── Found ──────────────────────────────────────────────► Login │
│    │                                                            │
│    └── Not Found                                                │
│         │                                                       │
│         ▼                                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ FindByIdentity(EMAIL, email)                             │  │
│  └──────────────────────────────────────────────────────────┘  │
│    │                                                            │
│    ├── Found + Verified                                         │
│    │    │                                                       │
│    │    ▼                                                       │
│    │  Link Google identity ke existing user                     │
│    │    │                                                       │
│    │    └───────────────────────────────────────────────────► Login │
│    │                                                            │
│    ├── Found + Not Verified                                     │
│    │    │                                                       │
│    │    └───────────────────────────────────────────────────► Error │
│    │                                                            │
│    └── Not Found                                                │
│         │                                                       │
│         ▼                                                       │
│  Create new user:                                               │
│  - User entity                                                  │
│  - LOCAL credential (without password)                          │
│    - EMAIL identity (verified)                                  │
│    - USERNAME identity (verified, generated)                    │
│  - GOOGLE credential                                            │
│    - GOOGLE identity (verified, sub)                            │
│  - Default role assignment                                      │
│    │                                                            │
│    └───────────────────────────────────────────────────────► Login │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Linked Accounts UseCases

**File baru**: `internal/modules/identity/application/query/get_linked_accounts.go`

```go
package query

type GetLinkedAccountsUseCase struct {
    userRepo userrepo.UserRepository
}

func NewGetLinkedAccountsUseCase(userRepo userrepo.UserRepository) *GetLinkedAccountsUseCase

// Execute mengembalikan status linked accounts user.
// can_unlink = true jika user punya password lokal.
func (uc *GetLinkedAccountsUseCase) Execute(ctx context.Context, userID string) (*dto.LinkedAccountsResponse, error)
```

**File baru**: `internal/modules/identity/application/command/unlink_google.go`

```go
package command

type UnlinkGoogleUseCase struct {
    userRepo userrepo.UserRepository
}

func NewUnlinkGoogleUseCase(userRepo userrepo.UserRepository) *UnlinkGoogleUseCase

// Execute melepas Google credential dari user.
// Ditolak jika user belum punya password lokal.
func (uc *UnlinkGoogleUseCase) Execute(ctx context.Context, userID string) error
```

**File baru**: `internal/modules/identity/application/command/link_google.go`

```go
package command

type LinkGoogleUseCase struct {
    userRepo              userrepo.UserRepository
    googleVerifier        ports.GoogleOAuthVerifier
    allowedGoogleClientID []string
}

func NewLinkGoogleUseCase(
    userRepo userrepo.UserRepository,
    googleVerifier ports.GoogleOAuthVerifier,
    allowedGoogleClientIDs []string,
) *LinkGoogleUseCase

// Execute menambahkan Google credential ke user yang sedang login.
// Flow:
// 1. Verifikasi ID token dengan Google
// 2. Check apakah Google sub sudah dipakai user lain (reject jika ya)
// 3. Check apakah user sudah punya Google credential (reject jika ya)
// 4. Link Google credential ke user
func (uc *LinkGoogleUseCase) Execute(ctx context.Context, userID string, req dto.LinkGoogleRequest) error
```

### Link Google Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                      Link Google Flow                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  User sudah login dengan akun lokal                             │
│    │                                                            │
│    ▼                                                            │
│  POST /auth/linked-accounts/google { id_token: "..." }         │
│    │                                                            │
│    ▼                                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Verify ID Token with Google tokeninfo endpoint           │  │
│  │ - Check audience (client_id)                             │  │
│  │ - Check email_verified = true                            │  │
│  │ - Extract: sub, email, name, picture                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│    │                                                            │
│    ▼                                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ FindByIdentity(GOOGLE, sub) - global check               │  │
│  │ Apakah Google sub ini sudah dipakai user lain?           │  │
│  └──────────────────────────────────────────────────────────┘  │
│    │                                                            │
│    ├── Found (sudah dipakai user lain)                          │
│    │    │                                                       │
│    │    └───────────────────────────────────────────────────► Error │
│    │         "Google account sudah terhubung ke pengguna lain"  │
│    │                                                            │
│    └── Not Found                                                │
│         │                                                       │
│         ▼                                                       │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ FindByID(userID) - load current user                     │  │
│  └──────────────────────────────────────────────────────────┘  │
│    │                                                            │
│    ▼                                                            │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ user.FindCredential(GOOGLE)                              │  │
│  │ Apakah user ini sudah punya Google credential?           │  │
│  └──────────────────────────────────────────────────────────┘  │
│    │                                                            │
│    ├── Found & Active                                           │
│    │    │                                                       │
│    │    └───────────────────────────────────────────────────► Error │
│    │         "Akun Google sudah tertaut"                        │
│    │                                                            │
│    └── Not Found / Soft-deleted                                 │
│         │                                                       │
│         ▼                                                       │
│  user.LinkGoogleCredential(credentialID, googleSub)             │
│    │                                                            │
│    ▼                                                            │
│  userRepo.Update(user)                                          │
│    │                                                            │
│    └───────────────────────────────────────────────────────► Success │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## DTO

### Auth DTO Additions

**File**: `internal/modules/identity/application/dto/auth_dto.go`

```go
// Tambahkan di akhir file

type GoogleLoginRequest struct {
    IDToken  string `json:"id_token" binding:"required"`
    DeviceID string `json:"device_id,omitempty"`
}

func (r *GoogleLoginRequest) ResolveIDToken() string {
    return strings.TrimSpace(r.IDToken)
}

type LinkGoogleRequest struct {
    IDToken string `json:"id_token" binding:"required"`
}

func (r *LinkGoogleRequest) ResolveIDToken() string {
    return strings.TrimSpace(r.IDToken)
}

type GoogleLinkedAccount struct {
    Linked    bool    `json:"linked"`
    Email     *string `json:"email"`
    CanUnlink bool    `json:"can_unlink"`
}

type LinkedAccountsResponse struct {
    Google GoogleLinkedAccount `json:"google"`
}
```

## API Endpoints

### Auth Endpoints

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| POST | `/api/v1/web/auth/login/google` | `GoogleLogin` | Public | Login dengan Google ID token |

### Linked Accounts Endpoints

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| GET | `/api/v1/web/auth/linked-accounts` | `GetLinkedAccounts` | JWT | Get status linked accounts |
| POST | `/api/v1/web/auth/linked-accounts/google` | `LinkGoogle` | JWT | Link akun Google ke user |
| DELETE | `/api/v1/web/auth/linked-accounts/google` | `UnlinkGoogle` | JWT | Unlink akun Google |

### Router Additions

```go
// di internal/modules/identity/interfaces/http/router.go

// Public auth routes
auth := router.Group("/api/v1/web/auth")
{
    // ... existing routes
    auth.POST("/login/google", handler.GoogleLogin)  // BARU
}

// Protected auth routes
authGroup := web.Group("/auth")
{
    // ... existing routes
    authGroup.GET("/linked-accounts", handler.GetLinkedAccounts)       // BARU
    authGroup.POST("/linked-accounts/google", handler.LinkGoogle)      // BARU
    authGroup.DELETE("/linked-accounts/google", handler.UnlinkGoogle)  // BARU
}
```

## Configuration

### Environment Variables

**File**: `.env.example`

```bash
# Google OAuth
# Client IDs untuk Google Sign-In (comma-separated untuk multiple environments)
# Contoh: "web-client-id.apps.googleusercontent.com,mobile-client-id.apps.googleusercontent.com"
GOOGLE_CLIENT_IDS=
```

### Config Structure

```go
// di internal/shared/config/config.go

type Config struct {
    // ... existing fields
    Google GoogleConfig
}

type GoogleConfig struct {
    ClientIDs []string // Allowed Google OAuth client IDs
}

// Load dari env:
// GOOGLE_CLIENT_IDS -> split by comma -> trim spaces
```

## Database

Tidak ada perubahan schema database yang diperlukan. Struktur `credentials` dan `login_identities` sudah mendukung tipe baru:

- `credentials.type` = 'GOOGLE' (baru, tidak perlu alter enum karena VARCHAR)
- `login_identities.kind` = 'GOOGLE' (baru, tidak perlu alter enum karena VARCHAR)

## Handler Implementation

### Auth Handler Additions

**File**: `internal/modules/identity/interfaces/http/handler.go`

```go
type IdentityHandler struct {
    // ... existing usecases
    
    googleLogin       *command.GoogleLoginUseCase
    getLinkedAccounts *query.GetLinkedAccountsUseCase
    linkGoogle        *command.LinkGoogleUseCase
    unlinkGoogle      *command.UnlinkGoogleUseCase
}

func NewIdentityHandler(
    // ... existing params
    googleLogin *command.GoogleLoginUseCase,
    getLinkedAccounts *query.GetLinkedAccountsUseCase,
    linkGoogle *command.LinkGoogleUseCase,
    unlinkGoogle *command.UnlinkGoogleUseCase,
) *IdentityHandler {
    // ...
}

// GoogleLogin godoc
// @Summary Login with Google
// @Description Login menggunakan Google ID token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.GoogleLoginRequest true "Google login payload"
// @Success 200 {object} dto.LoginResponse
// @Router /api/v1/web/auth/login/google [post]
func (h *IdentityHandler) GoogleLogin(c *gin.Context) {
    var req dto.GoogleLoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // handle error
    }
    
    ipAddress := c.ClientIP()
    resp, err := h.googleLogin.Execute(c.Request.Context(), req, ipAddress)
    if err != nil {
        // handle error
    }
    
    respond.OK(c, "Login Google berhasil", resp)
}

// GetLinkedAccounts godoc
// @Summary Get linked accounts status
// @Description Menampilkan status akun yang tertaut (Google)
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.LinkedAccountsResponse
// @Router /api/v1/web/auth/linked-accounts [get]
func (h *IdentityHandler) GetLinkedAccounts(c *gin.Context) {
    userID := middleware.GetUserID(c)
    
    resp, err := h.getLinkedAccounts.Execute(c.Request.Context(), userID)
    if err != nil {
        // handle error
    }
    
    respond.OK(c, "Linked accounts berhasil diambil", resp)
}

// LinkGoogle godoc
// @Summary Link Google account
// @Description Menautkan akun Google ke user yang sedang login
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.LinkGoogleRequest true "Google link payload"
// @Success 200 {object} dto.MessageResponse
// @Router /api/v1/web/auth/linked-accounts/google [post]
func (h *IdentityHandler) LinkGoogle(c *gin.Context) {
    var req dto.LinkGoogleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // handle error
    }
    
    userID := middleware.GetUserID(c)
    if err := h.linkGoogle.Execute(c.Request.Context(), userID, req); err != nil {
        // handle error
    }
    
    respond.OK(c, "Akun Google berhasil ditautkan", nil)
}

// UnlinkGoogle godoc
// @Summary Unlink Google account
// @Description Melepas tautan akun Google dari user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MessageResponse
// @Router /api/v1/web/auth/linked-accounts/google [delete]
func (h *IdentityHandler) UnlinkGoogle(c *gin.Context) {
    userID := middleware.GetUserID(c)
    
    if err := h.unlinkGoogle.Execute(c.Request.Context(), userID); err != nil {
        // handle error
    }
    
    respond.OK(c, "Akun Google berhasil dilepas", nil)
}
```

## Module Wiring

**File**: `internal/modules/identity/module.go`

```go
func (m *Module) wireDependencies() {
    // ... existing wiring
    
    // Google OAuth
    googleVerifier := googleoauth.NewVerifier()
    
    googleLoginUC := command.NewGoogleLoginUseCase(
        m.userRepo,
        m.tokenGen,
        m.sessionRepo,
        googleVerifier,
        m.cfg.Google.ClientIDs,
        m.transactor,
    )
    
    getLinkedAccountsUC := query.NewGetLinkedAccountsUseCase(m.userRepo)
    linkGoogleUC := command.NewLinkGoogleUseCase(m.userRepo, googleVerifier, m.cfg.Google.ClientIDs)
    unlinkGoogleUC := command.NewUnlinkGoogleUseCase(m.userRepo)
    
    m.handler = http.NewIdentityHandler(
        // ... existing params
        googleLoginUC,
        getLinkedAccountsUC,
        linkGoogleUC,
        unlinkGoogleUC,
    )
}
```

## Testing Strategy

### Unit Tests
1. `GoogleLoginUseCase` - Test all branches:
   - Valid token, existing Google user
   - Valid token, existing email user (auto-link)
   - Valid token, new user (create)
   - Invalid token
   - Unverified email from Google

2. `UnlinkGoogleUseCase` - Test:
    - Success case (has local password)
    - Rejected (no local password)
    - Not linked

3. `LinkGoogleUseCase` - Test:
    - Success case (valid token, no existing Google)
    - Rejected (Google sub already used by another user)
    - Rejected (user already has Google credential)
    - Invalid token
    - Unverified email

4. `User.UnlinkGoogleCredential()` - Test domain logic:
   - Success unlink
   - Rejected without local password
   - Already unlinked

### Integration Tests
- Full flow: register with Google -> check linked accounts -> unlink -> verify cannot login with Google
- Link flow: register local -> link Google -> verify can login with Google -> unlink -> verify cannot login with Google

## Migration Plan

1. **Phase 1**: Backend implementation
   - Add constants, entity methods
   - Implement Google OAuth verifier
   - Implement usecases
   - Add API endpoints
   - Unit & integration tests

2. **Phase 2**: Frontend integration (sipon-fe)
   - Google Sign-In button di login page
   - Linked accounts section di profile settings
   - Set password flow untuk Google-only users

3. **Phase 3**: Configuration & deployment
   - Setup Google Cloud Console OAuth credentials
   - Add client IDs to environment
   - Test di staging environment

## Error Handling

| Scenario | Error Code | HTTP Status | Message |
|----------|------------|-------------|---------|
| ID token kosong | `TOKEN_REQUIRED` | 422 | "ID token wajib diisi" |
| ID token invalid | `GOOGLE_TOKEN_INVALID` | 401 | "Token Google tidak valid" |
| Email belum verified di Google | `GOOGLE_EMAIL_NOT_VERIFIED` | 403 | "Email Google belum terverifikasi" |
| Audience tidak match | `GOOGLE_TOKEN_INVALID` | 401 | "Token tidak ditujukan untuk aplikasi ini" |
| Google sub sudah dipakai user lain | `GOOGLE_SUB_ALREADY_USED` | 409 | "Akun Google sudah terhubung ke pengguna lain" |
| User sudah punya Google credential | `GOOGLE_ALREADY_LINKED` | 409 | "Akun Google sudah tertaut" |
| Unlink tanpa password lokal | `GOOGLE_UNLINK_REQUIRES_PASSWORD` | 409 | "Wajib memiliki password lokal sebelum unlink Google" |
| Google tidak tertaut | `GOOGLE_NOT_LINKED` | 404 | "Akun Google tidak tertaut" |

## Notes

1. **Username Generation**: Saat membuat user baru dari Google login, generate username unik dari nama/email (ikuti pola k-forum-api: `generateUniqueUsername`).

2. **Avatar/Profile Picture**: Simpan URL gambar dari Google sebagai avatar awal user. Gunakan prefix `ext:` untuk external URL (ikuti pola yang sudah ada).

3. **Default Role**: User baru dari Google login dapat role default (member/user). Pastikan ada role default yang sudah di-seed.

4. **Device ID**: Optional, untuk tracking session per device.

5. **IP Address**: Dicatat untuk audit trail dan security.

## Checklist Implementasi

- [ ] Tambah constant `CredentialTypeGoogle` dan `LoginIdentifierKindGoogle`
- [ ] Tambah error codes untuk Google-related errors
- [ ] Implementasi `User.UnlinkGoogleCredential()`
- [ ] Implementasi `User.LinkGoogleCredential()`
- [ ] Implementasi `NewGoogleCredential()`
- [ ] Buat `GoogleOAuthVerifier` port dan implementation
- [ ] Buat `GoogleLoginUseCase` dengan semua flow
- [ ] Buat `GetLinkedAccountsUseCase`
- [ ] Buat `LinkGoogleUseCase`
- [ ] Buat `UnlinkGoogleUseCase`
- [ ] Tambah DTO: `GoogleLoginRequest`, `LinkGoogleRequest`, `LinkedAccountsResponse`
- [ ] Tambah handler methods: `GoogleLogin`, `GetLinkedAccounts`, `LinkGoogle`, `UnlinkGoogle`
- [ ] Register routes di router
- [ ] Wire dependencies di module.go
- [ ] Tambah `GOOGLE_CLIENT_IDS` di config dan .env.example
- [ ] Unit tests untuk semua usecase
- [ ] Integration test untuk full flow
- [ ] Update API documentation (swagger)
