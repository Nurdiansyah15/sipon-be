# Gap Report — Endpoint Logic sipon-api vs sipon-be (Non-Santri)

> Cakupan: semua endpoint HTTP di luar domain santri.
> Sumber:[sipon-api `internal/interfaces/http/router/router.go`](../../sipon-api/internal/interfaces/http/router/router.go) & [sipon-be `internal/modules/identity/interfaces/http/router.go`](../internal/modules/identity/interfaces/http/router.go).
> Mode: harus sama persis. Status `BEDA` = ada perbedaan logic/guard/input/response/error. Status `BELUM ADA` = endpoint tidak terdaftar di sipon-be.

---

## Ringkasan Eksekutif

| Metrik | Nilai |
|---|---|
| Total endpoint non-santri di sipon-api | **46** |
| Total endpoint terdaftar di sipon-be | **45** |
| Endpoint BELUM ADA di sipon-be | **1** (`GET /role-permission/user-roles/:user_role_id`) |
| Endpoint BEDA guard | 11 route (5 read role-permission + 6 users group) |
| Endpoint BEDA input body | 6 (AssignUserRole, CreateUser, AvatarConfirm, UpdateProfile, CheckUsername, ChangeUsername) |
| Endpoint BEDA response field/strukture | hampir semua (message string, payload field) |
| Endpoint BEDA error status code | 8 (UpdateUserRole, Deactivate/ReactivateUserRole, CheckUsername, ForgotPassword, Delete* family, dll) |
| Handler kosong/stub hilang di sipon-be | 0 (semua 45 route punya implementasi di `handler.go`) |

### Temuan Kritis (prioritas perbaikan)

1. **1 endpoint belum diimplementasikan**: `GET /api/v1/web/role-permission/user-roles/:user_role_id` (GetUserRole). Route tidak terdaftar di router sipon-be. Lihat §3.
2. **Guard AND vs OR**: sipon-be menempatkan `RequirePermission("manage_users")` di **group** `/users` + `RequirePermission(...)` lagi di **route**. Karena middlewareGun Gin `c.Abort()` pada gagal, efektif user harus punya **kedua** permission (AND). Sipon-api hanya route-level tunggal (OR antar listed perms. Lihat §4.2.
3. **5 route read role-permission TANPA guard permission**: `ListRoles`, `GetRole`, `ListPermissions`, `ListUserRoles`, `ListScopes` di sipon-be hanya `JWTAuth + PrincipalLoad`, tanpa `RequirePermission`. Sipon-api memasang `readRoleGuard` (= `manage_roles` | `manage_role_permissions` | `assign_role`). Lihat §5.
4. **Rate limiter tidak dipasang**: `mb.RateLimitByIP()/ByUser()/ByAuth()` tersedia sebagai builder di sipon-be tetapi tidak dipasang di route mana pun. Sipon-api memasang rate-limit by-IP di public-auth dan by-user di protected. Lihat §6.
5. **Response message string semuanya beda**: setiap endpoint memakai kalimat message berbeda (mis. api `"login success"` vs be `"login success"` ada yang sama, tapi banyak beda: api `"change password success"` vs be `"Password changed successfully"`). Frontend yang hardcode match akan break.
6. **GetRole membuang permission list** (`handler.go:556`): `rolePermRepo.ListByRoleID` dipanggil lalu hasilnya `_ = permItems` di-drop. Response tidak menyertakan `permissions[]`. Sipon-api mengembalikan `permissions []string`. Lihat §5.2.
7. **CreateUser body beda**: sipon-api hanya `{username, fullname?, email, phone?}` (password auto-generated, return `generated_password` sekali). sipon-be mewajibkan `password` + `role_name`, response `data:null` (no generated_password). Lihat §4.3.
8. **AssignUserRole body beda**: api `role_id` + `scope_type` required enum `global|region|community` + `notes`. be `role_name` + `scope_type` opsional (default global) + tanpa `notes`. Lihat §5.10.
9. **AvatarConfirm input beda**: api pakai query `?key=`, be pakai JSON body `{key}`. Lihat §3.13.
10. **UpdateProfile behavior beda**: api hanya ijinkan ubah email/phone bila belum verified (di-guard di usecase); be langsung ubah tanpa OTP. Field juga beda: api semua opsional pointer; be `fullname` & `email` required. Lihat §3.9.
11. **CheckUsername error beda**: username kosong → api 422, be 400. Username invalid format → api 422, be `available:false` (no error). Be juga tidak exclude self (tidak kirim userID ke usecase). Lihat §3.10.
12. **ForgotPassword anti-enumeration beda**: api selalu 200 (anti-enumeration ketat). Be bisa 404 email-tidak-ditemukan (bocor enumeration). Lihat §1.6.
13. **UpdateUserRole/Deactivate/ReactivateUserRole not-found → 400 (bukan 404)** di be (mapping `ErrCodeBadRequest`). api 404. Lihat §5.11/5.12/5.13.
14. **ReactivateUserRole expired → 410 Gone** di be, tidak ada padanan di api. §5.13.
15. **Delete family (RolePermission/UserRole/RoleScope) di be langsung delete tanpa cek eksistensi** → error repo diteruskan sebagai 500 (no 404). api memvalidasi 404. §5.7/§5.14/§5.17.
16. **DeleteRoleScope di be pakai `:scope_id` saja, `:role_id` diabaikan** → deletion tidak scoped ke role (risiko silang-scope). api juga hanya pakai `:scope_id` (sama — bukan gap, namun dua-duanya berisiko). §5.17.
17. **Session revoke pada change-password/reset-password admin TIDAK dilakukan** di be. api me-revoke session aktif setelah change-password, dan me-revoke all session user target setelah admin reset-password/deactivate. §3.3 & §4.4 & §4.5.
18. **Error code string beda untuk forbidden**: api `ERR_FORBIDDEN`; be middleware `FORBIDDEN` (no principal) / `INSUFFICIENT_PERMISSION` (perm missing). Frontend yang match string akan break. §6.3.

---

## §1. Auth Publik — `/api/v1/web/auth` (tanpa JWT, rate-limit by-IP)

### 1.1 POST `/auth/register`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Handler | `AuthHandler.Register` (auth_handler.go) | `IdentityHandler.Register` (handler.go:172) | — |
| Guard | public + rate-limit auth by-IP | public (rate-limit **tidak dipasang**) | **BEDA** (rate-limit) |
| Input | `username`(req,3-30), `email`(req,email), `password`(req,min8), `phone`?, `fullname`?, `device_id`? | `fullname`(req), `email`(req,email), `phone`?, `username`(req), `password`(req,min8). Tidak ada `device_id` di DTO. | **BEDA** (device_id hilang, fullname jadi required) |
| Flow | hash → user+credential+identity → role "member" → OTP email async → access+refresh token | sama alurnya; role assignment via `AssignByRoleName("member")`; OTP via goroutine async | SAMA |
| Success | 201 `"register success"`, data `{user_id, token, refresh_token, user: UserMe}` | 201 `"register success"`, data `{user_id, username, email, phone, roles[], access_token, refresh_token, token_type:"Bearer", expires_in:900}` | **BEDA** (struktur field: api pakai `user: UserMe` + `token`/`refresh_token`; be pakai `access_token`/`refresh_token` + flatten + `token_type` + `expires_in`) |
| Errors | 400, 422, 409 conflict | 400, 422, 409, 500 | SAMA |

**Verdict: BEDA** — rate-limit hilang, input `device_id` hilang, `fullname` jadi required, struktur response `token`→`access_token`.

### 1.2 POST `/auth/login`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | public + rate-limit auth by-IP | public (no rate-limit) | **BEDA** |
| Input | `identifier`(req), `password`(req), `device_id`? | `identity`(req), `password`(req), `device_id`? | **BEDA** (field name `identifier` vs `identity`) |
| Flow | verify password → tandai login → access+refresh token | sama + `EnsureCanLogin()` (banned→403, locked→429, not-active→403, not-verified→403) | SAMA (be lebih eksplisit) |
| Success | 200 `"login success"`, data `{token, refresh_token, user: UserMe}` | 200 `"login success"`, data `{user_id, username, email, phone, roles[], permissions[], access_token, refresh_token, token_type, expires_in:900}` | **BEDA** (struktur) |
| Errors | 401, 403, 422, 500 | 401, 403, 429 (locked), 404 (user not found), 422, 500 | **BEDA** (be ekspos 404 & 429; api bungkam) |

**Verdict: BEDA** — field name `identifier`, struktur response, status code 404/429.

### 1.3 POST `/auth/request-otp`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | public + RL auth | public (no RL) | **BEDA** |
| Input | `identifier`(req) | `identity`(req) | **BEDA** (field name) |
| Flow | kirim OTP verifikasi email/phone sesuai identifier | sama | SAMA |
| Success | 200, data `{message}` | 200 `"OTP sent successfully"`, data `null` | **BEDA** (payload) |
| Errors | 404, 409 cooldown, 422, 500 | 404, 422, 500 (tidak ada 409 cooldown eksplisit) | **BEDA** (cooldown) |

**Verdict: BEDA** — field name, payload, cooldown handling.

### 1.4 POST `/auth/verify-otp`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | public + RL auth | public (no RL) | **BEDA** |
| Input | `identifier`(req), `otp`(req,len=6) | `identity`(req), `code`(req,len=6) | **BEDA** (field name `identifier`/`otp` vs `identity`/`code`) |
| Flow | validasi OTP → mark email/phone verified | sama + eksplisit mismatch→400, expired→410 Gone, used→409 | SAMA (be lebih granular) |
| Success | 200, data `{message}` | 200 `"Identity verified successfully"`, data `null` | **BEDA** (payload) |
| Errors | 404, 422, 500 | 404, 400 (mismatch), 410 (expired), 409 (used), 403 (identity not found), 422, 500 | **BEDA** (be ekspos 400/410/409) |

**Verdict: BEDA** — field name, payload, error code lebih granular di be.

### 1.5 POST `/auth/refresh-token`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | public + RL auth | public (no RL) | **BEDA** |
| Input | `refresh_token`(req) | `refresh_token`(req) | SAMA |
| Flow | parse refresh → cek revoke → terbitkan access+refresh baru | sama + cek `RevokedBefore` & `DeviceRevokedBefore` | SAMA |
| Success | 200 `"refresh token success"`, data `LoginResponse{token,refresh_token,user}` | 200 `"refresh token success"`, data `{access_token, refresh_token, token_type, expires_in:900}` | **BEDA** (be tidak return `user`; pakai `access_token`) |
| Errors | 401, 403, 422, 500 | 401, 500 | **BEDA** (be tidak ekspos 403/422) |

**Verdict: BEDA** — response tidak menyertakan `user`, field `token`→`access_token`.

### 1.6 POST `/auth/password/forgot`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | public + RL auth | public (no RL) | **BEDA** |
| Input | `email`(req,email) | `email`(req,email) | SAMA |
| Flow | **anti-enumeration: selalu 200** walau email tidak ada | `FindByIdentity(EMAIL)` → bila NotFound → **404**. Tidak ada anti-enumeration ketat. | **BEDA KRITIS** |
| Success | 200 `"forgot password success"`, data `{message}` | 200 `"Password reset OTP sent if email exists"`, data `null`; atau 404 bila email tidak ada | **BEDA** |
| Errors | 422, 500 (tidak ada 404) | 404 (email tidak ada), 422, 500 | **BEDA** (be bocor enumeration) |

**Verdict: BEDA KRITIS** — be tidak konsisten dengan anti-enumeration; mengembalikan 404 untuk email tidak terdaftar.

### 1.7 POST `/auth/password/reset`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | public + RL auth | public (no RL) | **BEDA** |
| Input | `email`(req,email), `token`(req), `password`(req,min8) | `email`(req,email), `code`(req,len=6), `password`(req,min8) | **BEDA** (field `token` vs `code`, be tambah `len=6`) |
| Flow | verifikasi token → hash → simpan → revoke session | verifikasi OTP → hash → `SetLocalPassword`; tidak revoke session | SAMA (beda side-effect revoke) |
| Success | 200 `"reset password success"`, data `{message}` | 200 `"Password reset successfully"`, data `null` | **BEDA** (message + payload) |
| Errors | 404 (token/email invalid/expired), 422, 500 | 404, 400 (mismatch / credential-not-local), 410 (expired), 409 (used), 422, 500 | **BEDA** (granularitas) |

**Verdict: BEDA** — field name, message, revoke session, error granularitas.

---

## §2. Session Bootstrap — `/api/v1` (JWT wajib, tanpa permission)

### 2.1 GET `/auth/session`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Handler | `AuthHandler.GetSession` | `IdentityHandler.GetSession` (handler.go:266) | — |
| Guard | `JWTAuth + PrincipalLoader` | `JWTAuth + PrincipalLoad` (route-level) | SAMA |
| Input | principal dari context | userID dari context | SAMA |
| Flow | baca user + roles + permissions + scopes dari principal | `FindByID` + loop role → permissions & scopes (error di-ignore `_, _`) | SAMA |
| Success | 200, data `SessionData{user, roles[], permissions[], scopes[]}` | 200 `"session retrieved"`, data `{user_id, username, fullname, email, phone, roles[], permissions[], scopes[]{scope_type,scope_id}}` | **BEDA** (struktur: api nested `user`, be flatten) |
| Errors | 401 (`"sesi tidak ditemukan"`), 500 | 404 (user not found), 401, 500 | **BEDA** (be ekspos 404) |

**Verdict: BEDA** — struktur response nested vs flatten, error 404.

### 2.2 POST `/auth/logout`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `JWTAuth + PrincipalLoader` | `JWTAuth` saja (tanpa PrincipalLoad) | **BEDA** (be tidak load principal; cukup parse token) |
| Input | `session_id` dari context | token dari header via `extractToken` | **BEDA** (cara identifikasi session) |
| Flow | `LogoutUseCase.Execute(sessionID)` → revoke session_id | `ParseAccessToken` → `RevokeSession(sessionID, TTL 30 hari)`. Punya `ExecuteRevokeAll`/`ExecuteRevokeDevice` di usecase tapi **tidak dipasang route**. | **BEDA** (be tidak expose logout-all/device) |
| Success | 200 `"Logout berhasil"`, data `null` | 200 `"Logged out successfully"`, data `null` | **BEDA** (message string) |
| Errors | 401 (`"sesi tidak ditemukan"`), 500 | 401 (token hilang/invalid), 500 | SAMA |

**Verdict: BEDA** — middleware PrincipalLoad, message string, revoke-all/device tidak ekspos.

---

## §3. Auth Protected — `/api/v1/web/auth/*` (JWTAuth + PrincipalLoad/Loader)

### 3.1 GET `/auth/me`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | userID dari context | userID dari context | SAMA |
| Flow | `MeUseCase.Execute` → user + roles | `MeUseCase.Execute` → user + roles, **permissions:nil** | **BEDA** (be tidak isi permissions) |
| Success | 200 `"me success"`, data `UserMe{id,username,email,is_email_verified,fullname,phone,is_phone_verified,status,created_at,has_password,avatar_url}` | 200 `"me retrieved"`, data `ProfileResponse{user_id,username,fullname,email,phone,avatar_url,roles[],permissions:null,status,created_at}` | **BEDA** (field: api pakai `is_email_verified`/`is_phone_verified`/`has_password`; be punya `roles[]`/`permissions`/`avatar_url` tanpa flag verified & has_password) |
| Errors | 401, 404, 500 | 401, 404, 500 | SAMA |

**Verdict: BEDA** — field response banyak beda, message string.

### 3.2 GET `/auth/profile`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | principal (untuk ambil roles/perms/scopes dari cache) | userID | SAMA |
| Flow | `GetProfileUseCase.Execute(userID, principal)` — join roles+perms+scopes dari principal/cache | `GetProfileUseCase.Execute` → resolve roles+permissions via repo (tidak pakai principal cache) | SAMA |
| Success | 200 `"OK"`, data `ProfileResponse{id,username,fullname,email,is_email_verified,phone,is_phone_verified,status,has_password,created_at,avatar_url,roles,permissions,scopes}` | 200 `"profile retrieved"`, data `ProfileResponse{user_id,username,fullname,email,phone,avatar_url,roles,permissions,status,created_at}` — **tidak ada `scopes`** | **BEDA** (be tidak punya `is_email_verified`/`is_phone_verified`/`has_password`/`scopes` dalam profile) |
| Errors | 401, 500 | 401, 404, 500 | **BEDA** (be ekspos 404) |

**Verdict: BEDA** — field response (verified flags, has_password, scopes), message string.

### 3.3 POST `/auth/change-password`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | `current_password`(req), `new_password`(req,min8) | `old_password`(req), `new_password`(req,min8) | **BEDA** (field `current_password` vs `old_password`) |
| Flow | verify old → hash new → simpan → **revoke session aktif (side effect)** | verify old → hash new → simpan; **tidak revoke session** | **BEDA KRIS KRITIS** (revoke session) |
| Success | 200 `"change password success"`, data `{message}` | 200 `"Password changed successfully"`, data `null` | **BEDA** (message+payload) |
| Errors | 401, 404, 422 (`PASSWORD_SAME_AS_CURRENT`, `INVALID_CURRENT_PASSWORD`), 500 | 401, 400 (credential-not-local), 404, 422, 500 | **BEDA** (kode error) |

**Verdict: BEDA KRITIS** — revoke session tidak dilakukan be, field name, error code.

### 3.4 POST `/auth/set-password`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | `new_password`(req,min8) | `password`(req,min8) | **BEDA** (field `new_password` vs `password`) |
| Flow | set password pertama; ditolak 409 bila sudah punya password | sama (409 bila credential tidak ada ATAU SecretHash sudah ada) | SAMA |
| Success | 200 `"set password success"`, data `{message}` | 200 `"Password set successfully"`, data `null` | **BEDA** (message+payload) |
| Errors | 401, 404, 409, 422, 500 | 409, 400 (credential-not-local), 404, 422, 500 | **BEDA** (be ekspos 400) |

**Verdict: BEDA** — field name, message, payload, error 400.

### 3.5 POST `/auth/change-email/request`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | `new_email`(req,email) | `new_value`(req) (email baru; tidak ada tag email format) | **BEDA** (field `new_email` vs `new_value`, validasi email format hilang) |
| Flow | `RequestChangeIdentity(Kind:"EMAIL", new_email)` → kirim OTP | `RequestChangeIdentity` (kind dideteksi otomatis dari `NewLoginIdentifier`) → kirim OTP | SAMA |
| Success | 200, message=`resp.Message`, data `{message}` | 200 `"OTP sent to new identity"`, data `null` | **BEDA** (payload) |
| Errors | 400, 401, 409, 422, 500 | 409, 404, 400 (new identity empty), 422, 500 | **BEDA** (granularitas) |

**Verdict: BEDA** — field name, validasi email format hilang, payload.

### 3.6 POST `/auth/change-email/confirm`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | `otp`(req,len=6) | `code`(req,len=6) | **BEDA** (field `otp` vs `code`) |
| Flow | verify OTP → update email + mark verified | sama + mismatch→400, expired→410, used→409; tx: soft-del identity lama, buat identity baru | SAMA (be lebih eksplisit tx) |
| Success | 200, data `ChangeIdentityResponse{message}` | 200 `"Identity changed successfully"`, data `null` | **BEDA** (payload) |
| Errors | 401, 404, 422, 500 | 404, 400, 410, 409, 422, 500 | **BEDA** (granularitas) |

**Verdict: BEDA** — field name, payload, error granularitas.

### 3.7 POST `/auth/change-phone/request`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Input | `new_phone`(req, string; tanpa regex) | `new_value`(req) | **BEDA** (field `new_phone` vs `new_value`) |
| Flow | `RequestChangeIdentity(Kind:"PHONE", new_phone)` | sama (kind auto-detected) | SAMA |
| Success | 200, data `ChangeIdentityResponse{message}` | 200 `"OTP sent to new identity"`, data `null` | **BEDA** |
| Errors | 401, 409, 422, 500 | 409, 404, 400, 422, 500 | **BEDA** |

**Verdict: BEDA** — field name, payload, error.

### 3.8 POST `/auth/change-phone/confirm`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Input | `otp`(req,len=6) | `code`(req,len=6) | **BEDA** |
| Flow | verify OTP → update phone + mark verified | sama (purpose PHONE) | SAMA |
| Success | 200, data `ChangeIdentityResponse` | 200 `"Identity changed successfully"`, data `null` | **BEDA** |
| Errors | 401, 404, 422, 500 | 404, 400, 410, 409, 422, 500 | **BEDA** |

**Verdict: BEDA** — field name, payload, error.

### 3.9 PUT `/auth/profile`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | `fullname`? (pointer), `email`? (pointer), `phone`? (pointer); **semua opsional**; aturan: email & phone hanya bisa diubah bila belum verified | `fullname`(req), `email`(req,email), `phone`? — **fullname & email WAJIB** | **BEDA** (required-ness; api opsional semua, be wajib fullname+email) |
| Flow | `UpdateProfile(userID, req)` — guard verified di usecase; ubah langsung email/phone tanpa OTP **jika belum verified** | `UpdateProfile` — langsung ubah email/phone tanpa OTP & tanpa cek verified; jika email berubah cek unik (409) | **BEDA** (be tidak cek verified) |
| Success | 200 `"update profile success"`, data `{message}` | 200 `"Profile updated successfully"`, data `null` | **BEDA** (message+payload) |
| Errors | 401, 404, 409, 422, 500 | 409, 404, 422, 500 | SAMA |

**Verdict: BEDA KRITIS** — required-ness input beda, validasi verified-flag hilang di be, message.

### 3.10 GET `/auth/check-username`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | JWTAuth + PrincipalLoader | JWTAuth + PrincipalLoad | SAMA |
| Input | query `?username=`; bila kosong → **422** `"parameter username wajib diisi"` | query `?username=`; bila kosong → **400** `ERR_BAD_REQUEST` | **BEDA** (status code 422 vs 400) |
| Flow | `CheckUsernameUseCase.Execute(userID, username)` — **exclude self** (kirim userID) | `CheckUsernameUseCase.Execute(username)` — **tidak kirim userID** (tidak exclude self); bila `NewUsername` invalid → return `{Available:false}` **tanpa error** | **BEDA KRITIS** (be tidak exclude self; invalid → false bukan 422) |
| Success | 200 `"check username success"`, data `{available:bool}` | 200 `"check username"`, data `{available:bool}` | SAMA (payload) |
| Errors | 400, 401, 422, 500 | 400, 500 (tidak ada 422) | **BEDA** |

**Verdict: BEDA KRITIS** — be tidak exclude user sendiri (akan return false untuk username milik sendiri), invalid format → false (silent), status code 422→400.

### 3.11 POST `/auth/change-username`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Input | `username`(req,min=3,max=30) | `username`(req) — **tidak ada tag min/max** | **BEDA** (validasi panjang hilang) |
| Flow | `ChangeUsernameUseCase.Execute(userID, req)` → ubah langsung tanpa OTP | sama (tanpa OTP) | SAMA |
| Success | 200 `"change username success"`, data `{message, username}` | 200 `"Username changed successfully"`, data `null` | **BEDA** (be tidak return `username`; message) |
| Errors | 401, 404, 409, 422, 500 | 409, 404, 422, 500 | SAMA |

**Verdict: BEDA** — validasi panjang username hilang, response tidak return `username`.

### 3.12 POST `/auth/profile/avatar/presign`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Input | `content_type`(req) | `content_type`(req) | SAMA |
| Flow | `AvatarPresignUseCase.Execute` → presigned URL MinIO | `RequestUpload(avatars/<uuid>, contentType, 10 menit, PrivacyPublic)` | SAMA |
| Success | 200 `"avatar presign success"`, data `{presign_url, key, expires_in}` | 200 `"avatar presign url generated"`, data `{presign_url, key}` | **BEDA** (be tidak return `expires_in`; message) |
| Errors | 400, 401, 422, 500 | 422, 500 | **BEDA** |

**Verdict: BEDA** — response tidak ada `expires_in`, message.

### 3.13 POST `/auth/profile/avatar/confirm`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Input | **query** `?key=`; bila kosong → 422 `"parameter key wajib diisi"` | **JSON body** `{key:req}` | **BEDA KRITIS** (sumber input: query vs body) |
| Flow | `AvatarConfirm(userID, key)` → verifikasi object → persist avatar_url | `AvatarConfirm(userID, key)` → `MarkDeleted` avatar lama → `ConfirmUpload` → set AvatarKey | SAMA (be lebih lengkap: mark old) |
| Success | 200 `"avatar confirm success"`, data `{avatar_url}` | 200 `"Avatar confirmed successfully"`, data `null` | **BEDA** (be tidak return `avatar_url`; message) |
| Errors | 400, 401, 404, 422, 500 | 404, 500 | **BEDA** |

**Verdict: BEDA KRITIS** — input source beda (query vs body), response tidak return `avatar_url`.

### 3.14 DELETE `/auth/profile/avatar`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Input | tidak ada | tidak ada | SAMA |
| Flow | `AvatarDelete(userID)` → hapus object + clear avatar_url | `AvatarDelete(userID)` → bila `AvatarKey` kosong → **return 200 no-op**; else delete + clear | **BEDA** (be no-op bila tidak ada avatar) |
| Success | 200 `"avatar delete success"`, data `null` | 200 `"Avatar deleted successfully"`, data `null` | **BEDA** (message) |
| Errors | 401, 404 (avatar tidak ada), 500 | 404 (user not found), 500 (tidak ada 404 khusus "avatar tidak ada") | **BEDA** (be tidak ekspos 404 avatar-not-exist) |

**Verdict: BEDA** — no-op bila avatar kosong (api raising 404), message.

---

## §4. User Management — `/api/v1/web/users/*`

### Catatan guard grup vs route (BERLAKU UNTUK §4.2–§4.6)

**sipon-api**: `users := protectedWeb.Group("/users")` lalu setiap route dapat **1** `RequirePermission` tunggal:
- ListUsers/GetUser/CreateUser → `manage_users`
- reset-password → `reset_user_password`
- deactivate/reactivate → `deactivate_user`

**sipon-be**: `users.Use(mb.PrincipalLoad(), mb.RequirePermission("manage_users"))` di **group** + `RequirePermission(...)` lagi di **route**. Karena `RequirePermission` memakai `c.Abort()` saat gagal, efeknya user **WAJIB punya `manage_users`** (lewat group) **DAN** permission route (`reset_user_password`/`deactivate_user`). Ini berbeda dari semantik OR tunggal di api.

### 4.1 GET `/users`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Handler | `UserManagementHandler.ListUsers` | `IdentityHandler.ListUsers` (handler.go:472) | — |
| Guard | `RequirePermission("manage_users")` (route) | `RequirePermission("manage_users")` (group) | SAMA (efektif) |
| Input | query `dto.ListUsersQuery{status, role_id, search, pagination}` | `dto.ListUsersRequest{status, role_id, search, PaginationParams}` | SAMA |
| Flow | `ListUsers.Execute` → list + roles per user (roles dikosongkan anti-N+1 di list, di-get user terisi) | `listUsersUC.Execute` → list + resolve roles per user via `FindActiveByUserID`+`FindByID` | **BEDA** (api kosongkan roles di list, be isi roles) |
| Success | 200 `"users fetched"`, `SuccessWithMeta{data:[UserManagementResponse], meta:{current_page, per_page, total, total_pages}}` | 200 `"users retrieved"`, data `{users:[UserItem{id,username,fullname,email,phone,status,roles[],created_at,updated_at,last_login_at}], meta:{page,limit,total_items,total_pages}}` | **BEDA** (struktur: api `{data, meta}` di top-level; be `{users, meta}` nested. Field meta: `current_page`/`per_page`/`total` vs `page`/`limit`/`total_items`) |
| Errors | 401, 403, 500 | 401, 403, 422, 500 | **BEDA** |

**Verdict: BEDA** — roles terisi vs kosong, struktur meta field, message.

### 4.2 GET `/users/:user_id`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_users` (route) | `manage_users` (group) | SAMA |
| Input | `:user_id` | `:user_id` | SAMA |
| Flow | `GetUser.Execute(userID)` → detail + roles aktif | `getUserUC.Execute` → `FindByID` + `FindActiveByUserID` → roles | SAMA |
| Success | 200 `"user fetched"`, data `UserManagementResponse` dengan `Roles []UserRoleSummary` terisi | 200 `"user retrieved"`, data `UserItem{id,...,roles[]}` | **BEDA** (struktur `UserItem` vs `UserManagementResponse`; message) |
| Errors | 401, 403, 404, 500 | 401, 403, 404, 500 | SAMA |

**Verdict: BEDA** — struktur response & message.

### 4.3 POST `/users` — CreateUser

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_users` (route) | `manage_users` (group) | SAMA |
| Input | `{username(req), fullname?, email(req), phone?}` — **tidak ada password, tidak ada role_name** (auto-generated password + default role via usecase) | `{fullname(req), email(req,email), phone?, username(req), password(req,min=8), role_name(req)}` — **wajib password & role_name** | **BEDA KRITIS** (be wajib password+role_name; api auto-generated) |
| Flow | `CreateUser.Execute(req)` → buat user + credential lokal tanpa password input → hash generated → persist → return generated_password sekali | `CreateUser.Execute(req, createdBy)` → validasi VO → cek unik → hash → tx: Save user → `AssignByRoleName(role_name)`; **tidak generate token**; createdBy dipakai sebagai AssignedBy | SAMA alur, BEDA input |
| Success | 201 `"user created"`, data `CreateUserResponse{UserManagementResponse, generated_password}` | 201 `"User created successfully"`, data `null` (no generated_password) | **BEDA KRITIS** (be tidak return generated_password) |
| Errors | 401, 403, 409, 422, 500 | 409, 404 (role not found), 422, 403, 401, 500 | **BEDA** |

**Verdict: BEDA KRITIS** — input wajib password + role_name, response tidak ada generated_password, message.

### 4.4 POST `/users/:user_id/reset-password`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `reset_user_password` (route tunggal) | group `manage_users` AND route `reset_user_password` | **BEDA** (AND vs single) |
| Input | `:user_id` tanpa body (auto-generate new password) | `:user_id` + JSON `{new_password(req,min=8)}` — **body wajib** | **BEDA** (be minta password baru di body) |
| Flow | `ResetUserPassword.Execute(userID)` → generate password baru → hash → persist → **revoke all session user target** → return password sekali | `ResetUserPasswordUseCase.Execute` → `FindByID` → hash new → `SetLocalPassword` → Update; **tidak revoke session** | **BEDA KRITIS** (revoke session hilang; be minta input password) |
| Success | 200 `"user password reset"`, data `{generated_password}` | 200 `"User password reset successfully"`, data `null` | **BEDA** (be tidak return generated_password) |
| Errors | 401, 403, 404, 500 | 403, 404, 422, 500 | **BEDA** (be ekspos 422) |

**Verdict: BEDA KRITIS** — guard AND, input body, revoke session, response generated_password.

### 4.5 POST `/users/:user_id/deactivate`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `deactivate_user` (route) | group `manage_users` AND `deactivate_user` | **BEDA** (AND) |
| Input | `:user_id` tanpa body | `:user_id` | SAMA |
| Flow | `DeactivateUser.Execute` → set status BANNED → **revoke all session user** | `DeactivateUser.Execute` → `user.Deactivate()` (sudah banned→409) → Update; **tidak revoke session** | **BEDA KRITIS** (revoke session) |
| Success | 200 `"user deactivated"`, data response (status baru) | 200 `"User deactivated successfully"`, data `null` | **BEDA** (payload) |
| Errors | 401, 403, 404, 409 (sudah deactivated), 500 | 403, 404, 409 (sudah banned `ERR_USER_ALREADY_BANNED`), 500 | SAMA |

**Verdict: BEDA KRITIS** — guard AND, revoke session, payload.

### 4.6 POST `/users/:user_id/reactivate`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `deactivate_user` (route) | group `manage_users` AND `deactivate_user` | **BEDA** (AND) |
| Input | `:user_id` | `:user_id` | SAMA |
| Flow | `ReactivateUser.Execute` → kembalikan status aktif | `ReactivateUser.Execute` → `user.Reactivate()` (status bukan banned→409 `ERR_USER_NOT_ACTIVE`) → reset FailedAttempts + LockedUntil → Update | SAMA (be reset attempts) |
| Success | 200 `"user reactivated"` | 200 `"User reactivated successfully"`, data `null` | **BEDA** (payload+message) |
| Errors | 401, 403, 404, 409 (bukan deactivated), 500 | 403, 404, 409, 500 | SAMA |

**Verdict: BEDA** — guard AND, payload, message.

---

## §5. Role & Permission — `/api/v1/web/role-permission/*`

### Catatan guard read (BERLAKU UNTUK §5.1–§5.3, §5.8, §5.15)

**sipon-api** memasang `readRoleGuard = RequirePermission("manage_roles" | "manage_role_permissions" | "assign_role")` di:
- GET `/roles`, GET `/roles/:role_id`, GET `/permission-keys`
- GET `/user-roles`, GET `/user-roles/:user_role_id`
- GET `/roles/:role_id/scopes`

`userRoleReadGuard = RequirePermission("assign_role" | "manage_users")` untuk GET `/user-roles` & GET `/user-roles/:id`.

**sipon-be** TIDAK memasang `RequirePermission` apa pun di kelima route read ini — hanya `JWTAuth + PrincipalLoad`. Artinya **siapa pun yang login** bisa list roles, list permissions, list user-roles, list scopes. Ini **gap guard kritis**.

### 5.1 GET `/role-permission/roles`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `readRoleGuard` | **tidak ada permission** | **BEDA KRITIS** |
| Input | query `{role_type, scope_type, assignable, pagination}` | `{role_type, scope_type, assignable*bool, pagination}` | SAMA |
| Flow | `ListRoles.Execute` | `listRolesUC.Execute` → `roleListRepo.List` | SAMA |
| Success | 200 `"roles fetched"`, `SuccessWithMeta{data, meta}` | 200 `"roles retrieved"`, data `{roles:[RoleItem], meta}` | **BEDA** (struktur: api `{data,meta}` top-level; be `{roles,meta}` nested) |
| Errors | 400, 401, 403, 500 | 401, 422, 500 (tidak ada 403 karena no guard) | **BEDA** |

**Verdict: BEDA KRITIS** — guard hilang, struktur response.

### 5.2 GET `/role-permission/roles/:role_id`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `readRoleGuard` | **tidak ada permission** | **BEDA KRITIS** |
| Input | `:role_id` | `:role_id` | SAMA |
| Flow | `GetRole.Execute(role_id)` → role + **permissions[]** (system: `PermissionsForRole`; custom: tabel) | `getRoleUC.Execute` → `FindByID` + `ListByRoleID` tapi **`_ = permItems`** → permissions di-drop | **BEDA KRITIS** (be membuang permissions) |
| Success | 200 `"role fetched"`, data `RoleResponse{id,name,display_name,description,role_type,scope_type,assignable,created_at,updated_at,permissions[]string}` | 200 `"role retrieved"`, data `RoleItem{id,name,display_name,description,role_type,scope_type,assignable,created_at,updated_at}` — **tidak ada `permissions[]`** | **BEDA KRITIS** |
| Errors | 401, 403, 404 (`DOMAIN_ROLE_NOT_FOUND`), 500 | 401, 404, 500 (no 403) | **BEDA** |

**Verdict: BEDA KRITIS** — guard hilang, permissions[] tidak dikembalikan (bug drop `permItems`).

### 5.3 GET `/role-permission/permission-keys`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `readRoleGuard` | **tidak ada permission** | **BEDA KRITIS** |
| Input | tidak ada | tidak ada | SAMA |
| Flow | `ListPermissionKeys.Execute` → `constant.AllPermissionDefinitions()` (dari konstanta) | `listPermissionsUC.Execute` → `domain.AllPermissionDefinitions` (statis 7 key) | SAMA |
| Success | 200 `"permission keys fetched"`, data `[{key, display_name, description}]` (array) | 200 `"permissions retrieved"`, data `{permissions:[{key, display_name, description}]}` (object berisi array) | **BEDA** (api array langsung; be object dengan key `permissions`) |
| Errors | 401, 403 | 401, 500 (no 403) | **BEDA** |

**Verdict: BEDA KRITIS** — guard hilang, struktur response (array vs object).

### 5.4 POST `/role-permission/roles`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_roles` | `manage_roles` (route) | SAMA |
| Input | `{name(req), display_name(req), description?, role_type(req,oneof=system|custom), scope_type(req,oneof=global|region|community), assignable}` | `{name(req), display_name(req), description?, scope_type(req), assignable}` — **tidak ada `role_type` di DTO** (hardcode custom di usecase) | **BEDA** (be hilangkan input `role_type` & validasi enum scope_type) |
| Flow | `CreateRole.Execute(req)` → persist custom/system sesuai input | `CreateRoleUseCase` → `RoleType=custom` (hardcode); `domain.NewRole` validasi scope | SAMA |
| Success | 201 `"role created"`, data `RoleResponse` | 201 `"role created"`, data `RoleItem` | **BEDA** (struktur+message) |
| Errors | 400, 401, 403, 409 (`DOMAIN_ROLE_DUPLICATE_NAME`), 422 (enum), 500 | 403, 422 (`INVALID_SCOPE_TYPE`), 500 — **tidak ada 409 duplicate name** | **BEDA** (be tidak cek duplikat nama) |

**Verdict: BEDA** — input `role_type` & enum `scope_type` hilang, tidak ada cek duplikat nama 409.

### 5.5 PUT `/role-permission/roles/:role_id`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_roles` | `manage_roles` | SAMA |
| Input | `{display_name?, description?, assignable*bool?}` (pointer, partial update) | `{display_name?, description?, assignable*bool?}` | SAMA |
| Flow | `UpdateRole.Execute(role_id, req)` | `updateRoleUC` → `FindByID` → `role.EnsureCustom()` (system→400/403) → apply → Update | SAMA (be eksplisit EnsureCustom) |
| Success | 200 `"role updated"`, data `RoleResponse` | 200 `"role updated"`, data `RoleItem` | **BEDA** (struktur) |
| Errors | 401, 403, 404, 422, 500 | 403 (fallback system), 400 (system), 404, 500 | **BEDA** |

**Verdict: BEDA** — struktur response, error system-role (400 vs 403).

### 5.6 POST `/role-permission/roles/:role_id/permissions`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_role_permissions` | `manage_role_permissions` | SAMA |
| Input | `{permission_key(req), notes?}` + `:role_id` | `{permission_key(req), notes?}` + `:role_id`; `assignedBy := GetUserID(c)` | SAMA |
| Flow | `AssignRolePermission.Execute(actorUserID, role_id, req)` → validasi `IsValidPermissionKey`, role harus custom | `assignRolePermissionUC.Execute(roleID, assignedBy, req)` → `FindByID` → `EnsureCustom` → `NewRolePermission` (invalid key→422) → Save | SAMA |
| Success | 201 `"permission assigned to role"`, data response role-permission | 200 `"Permission assigned to role successfully"`, data `null` | **BEDA** (status 201 vs 200; payload) |
| Errors | 400, 401, 403, 404, 409 (`DOMAIN_ROLE_PERMISSION_REQUIRES_CUSTOM_ROLE`, `DOMAIN_ROLE_PERMISSION_DUPLICATE`), 422, 500 | 403, 400 (system), 404, 422 (`INVALID_PERMISSION_KEY`), 500 — **tidak ada 409 duplicate eksplisit** | **BEDA** |

**Verdict: BEDA** — status code 201→200, payload, 409 duplicate.

### 5.7 DELETE `/role-permission/roles/:role_id/permissions/:permission_key`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_role_permissions` | `manage_role_permissions` | SAMA |
| Input | `:role_id`, `:permission_key` | `:role_id`, `:permission_key` | SAMA |
| Flow | `RevokeRolePermission.Execute(role_id, permission_key)` — ditolak 409 untuk system role | `deleteRolePermissionUC.Execute(roleID, permissionKey)` → **langsung `rolePermRepo.Delete`** tanpa validasi role/permission eksis | **BEDA** (be tidak cek eksistensi & system role) |
| Success | 200 `"permission revoked from role"`, data response | 200 `"Permission removed from role successfully"`, data `null` | **BEDA** (payload) |
| Errors | 401, 403, 404 (`DOMAIN_ROLE_NOT_FOUND`/`DOMAIN_ROLE_PERMISSION_NOT_FOUND`), 409 (system), 500 | 403, 500 — **tidak ada 404 & 409** | **BEDA KRITIS** (be tidak ekspos 404/409) |

**Verdict: BEDA KRITIS** — be langsung delete tanpa validasi system role & eksistensi.

### 5.8 GET `/role-permission/user-roles`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `userRoleReadGuard = assign_role|manage_users` | **tidak ada permission** | **BEDA KRITIS** |
| Input | query `{user_id, role_id, scope_type, scope_id, is_active, pagination}` | `{user_id, role_id, scope_type, is_active*bool, pagination}` — **tidak ada `scope_id`** | **BEDA** (be tidak ada filter `scope_id`) |
| Flow | `ListUserRoles.Execute` → list + resolve role/user/permissions per row | `listUserRolesUC.Execute` → resolve role/user/permissions; **skip baris jika FindByID gagal** (`continue`) | SAMA (be silent-skip) |
| Success | 200 `"user roles fetched"`, `SuccessWithMeta{data:[UserRoleResponse{...}], meta}` | 200 `"user roles retrieved"`, data `{user_roles:[UserRoleItem{...}], meta}` | **BEDA** (struktur; api `{data,meta}`; be `{user_roles,meta}`) |
| Errors | 400, 401, 403, 500 | 401, 422, 500 (no 403) | **BEDA** |

**Verdict: BEDA KRITIS** — guard hilang, filter `scope_id` hilang, struktur response.

### 5.9 GET `/role-permission/user-roles/:user_role_id` — **BELUM ADA**

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Route | router.go:155 `userRoleReadGuard` | **TIDAK TERDAFTAR** | **BELUM ADA** |
| Flow | `GetUserRole.Execute(user_role_id)` → detail 1 user-role | — | — |
| Success | 200 `"user role fetched"`, data `UserRoleResponse` | — | — |
| Errors | 401, 403, 404, 500 | — | — |

**Verdict: BELUM ADA** — klien harus filter via `ListUserRoles?...` di be. Tidak ada padanan detail-by-id.

### 5.10 POST `/role-permission/user-roles` — AssignUserRole

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `assign_role` | `assign_role` | SAMA |
| Input | `{user_id(req), role_id(req), scope_type(req,oneof=global|region|community), scope_id?, expired_at?, notes?}` | `{user_id(req), role_name(req), scope_type?(default global), scope_id?, expired_at?}` — **tidak ada `notes`** | **BEDA KRITIS** (api `role_id`, be `role_name`; api scope_type required+enum, be opsional; api ada `notes`) |
| Flow | `AssignUserRole.Execute(actorUserID, req)` → validasi role assignable, scope mismatch (409), duplicate (409) | `assignUserRoleUC.Execute(assignedBy, req)` → `FindByID(user)` → `AssignByRoleName` (role not found→404, not assignable→403, scope mismatch→400, sudah ada→409) → `NewUserRole` → Save | SAMA |
| Success | 201 `"user role assigned"`, data `UserRoleResponse` | 201 `"Role assigned to user successfully"`, data `null` | **BEDA** (payload) |
| Errors | 400, 401, 403, 404, 409 (`DOMAIN_USER_ROLE_DUPLICATE`/`DOMAIN_ROLE_USER_ASSIGNMENT_DUPLICATE`/`DOMAIN_ROLE_NOT_ASSIGNABLE`), 422 (scope), 500 | 403, 400 (scope mismatch), 404, 409, 422, 500 | SAMA (mapping beda) |

**Verdict: BEDA KRITIS** — input `role_id`→`role_name`, `scope_type` required→optional, `notes` hilang, payload.

### 5.11 PUT `/role-permission/user-roles/:user_role_id`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `assign_role` | `assign_role` | SAMA |
| Input | `{expired_at?}` | `{expired_at?}` | SAMA |
| Flow | `UpdateUserRole.Execute(id, req)` — not found → **404** | `updateUserRoleUC.Execute` — `FindByID` not found → **400 `ERR_BAD_REQUEST`** | **BEDA** (not found 404 vs 400) |
| Success | 200 `"user role updated"`, data `UserRoleResponse` | 200 `"User role updated successfully"`, data `null` | **BEDA** (payload) |
| Errors | 401, 403, 404, 422, 500 | 403, 400, 500 | **BEDA** |

**Verdict: BEDA** — not-found 404→400, payload.

### 5.12 POST `/role-permission/user-roles/:user_role_id/deactivate`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `assign_role` | `assign_role` | SAMA |
| Input | `:user_role_id` | `:user_role_id` | SAMA |
| Flow | `DeactivateUserRole.Execute` → not found → 404 | `deactivateUserRoleUC` → not found → **400**; `userRole.Deactivate()` (sudah inactive → 400 `USER_ROLE_NOT_ACTIVE`) | **BEDA** (404 vs 400) |
| Success | 200 `"user role deactivated"`, data `UserRoleResponse` | 200 `"User role deactivated"`, data `null` | **BEDA** (payload) |
| Errors | 401, 403, 404, 500 | 403, 400, 500 | **BEDA** |

**Verdict: BEDA** — not-found 404→400, sudah-inactive handling, payload.

### 5.13 POST `/role-permission/user-roles/:user_role_id/reactivate`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `assign_role` | `assign_role` | SAMA |
| Input | `:user_role_id` | `:user_role_id` | SAMA |
| Flow | `ReactivateUserRole.Execute` — not found → 404 | `reactivateUserRoleUC` — not found → **400**; masih aktif → 400; **expired → 410 `ERR_GONE`** (`USER_ROLE_EXPIRED`) | **BEDA** (404→400 + 410 baru) |
| Success | 200 `"user role reactivated"` | 200 `"User role reactivated successfully"`, data `null` | **BEDA** (payload) |
| Errors | 401, 403, 404, 500 | 403, 400, 410, 500 | **BEDA** |

**Verdict: BEDA** — not-found 404→400, expired→410 (baru), payload.

### 5.14 DELETE `/role-permission/user-roles/:user_role_id`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `assign_role` | `assign_role` | SAMA |
| Input | `:user_role_id` | `:user_role_id` | SAMA |
| Flow | `DeleteUserRole.Execute` (hanya return error) | `deleteUserRoleUC` → **langsung `userRoleRepo.Delete`** tanpa validasi eksistensi | **BEDA** (be tidak cek eksistensi) |
| Success | 200 `"User role deleted successfully"`, data `{message:"User role deleted successfully"}` | 200 `"User role deleted successfully"`, data `null` | **BEDA** (payload: api return message string, be null) |
| Errors | 401, 403, 404, 500 | 403, 500 — **tidak ada 404** | **BEDA** |

**Verdict: BEDA** — be tidak ekspos 404, payload beda.

### 5.15 GET `/role-permission/roles/:role_id/scopes`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `readRoleGuard` | **tidak ada permission** | **BEDA KRITIS** |
| Input | `:role_id` | `:role_id` | SAMA |
| Flow | `ListRoleScopes.Execute(role_id)` | `listScopesUC.Execute` → `roleScopeRepo.FindByRoleID` | SAMA |
| Success | 200 `"role scopes fetched"`, data `[{id, scope_type, scope_value}]` (array) | 200 `"scopes retrieved"`, data `{scopes:[{id, scope_type, scope_value}]}` (object) | **BEDA** (array vs object) |
| Errors | 401, 403, 404 (role), 500 | 401, 500 — **tidak ada 403 & 404** | **BEDA** |

**Verdict: BEDA KRITIS** — guard hilang, struktur array vs object, 404 role hilang.

### 5.16 POST `/role-permission/roles/:role_id/scopes`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_role_permissions` | `manage_role_permissions` | SAMA |
| Input | `{scope_type(req,oneof=gender), scope_value(req,oneof=male|female)}` | `{scope_type(req), scope_value(req)}` — **tanpa enum oneof** | **BEDA** (be hilangkan validasi enum) |
| Flow | `AssignRoleScope.Execute(role_id, req)` — hanya custom role | `assignRoleScopeUC` → `FindByID` → `EnsureCustom` → `NewRoleScope` (scope invalid→422) → Save | SAMA |
| Success | 201 `"role scope assigned"`, data `RoleScopeResponse{id,scope_type,scope_value}` | 201 `"Scope assigned to role successfully"`, data `null` | **BEDA** (payload) |
| Errors | 400, 401, 403, 404, 422 (enum), 500 | 403, 400 (system), 404, 422 (`INVALID_SCOPE_TYPE`), 500 | SAMA |

**Verdict: BEDA** — validasi enum hilang di DTO, payload.

### 5.17 DELETE `/role-permission/roles/:role_id/scopes/:scope_id`

| Aspek | sipon-api | sipon-be | Status |
|---|---|---|---|
| Guard | `manage_role_permissions` | `manage_role_permissions` | SAMA |
| Input | `:role_id`, `:scope_id` — handler **hanya pakai `:scope_id`** | `:role_id`, `:scope_id` — handler **hanya pakai `:scope_id`** | SAMA (keduanya ignore `:role_id` — **anomali sama**) |
| Flow | `RemoveRoleScope.Execute(scope_id)` → 404 bila scope tidak ada | `deleteRoleScopeUC.Execute(scopeID)` → **langsung `roleScopeRepo.Delete`** tanpa validasi | **BEDA** (be tidak cek eksistensi) |
| Success | 200 `"role scope removed"`, data `null` | 200 `"Scope removed from role successfully"`, data `null` | **BEDA** (message) |
| Errors | 401, 403, 404, 500 | 403, 500 — **tidak ada 404** | **BEDA** |

**Verdict: BEDA** — be tidak ekspos 404, message. Anomali `:role_id` diabaikan ada di kedua sisi (risiko silang-scope deletion — perlu fix di kedua codebase).

---

## §6. Perbedaan Cross-Cutting

### 6.1 Rate Limiter

| Aspek | sipon-api | sipon-be |
|---|---|---|
| Global by-IP | Dipasang di root `r.Use` bila `rlCfg.Enabled` | MiddlewareBuilder punya `RateLimitByIP()` **tidak dipanggil** di route mana pun |
| Auth by-IP | Dipasang ulang di group `/auth` public | **tidak dipasang** |
| User by-userID | Dipasang di `protectedWeb` | **tidak dipasang** |

**Gap**: be menyediakan builder rate-limit tetapi tidak memasang di router identity. Public-auth rentan brute-force.

### 6.2 Middleware Global

| Middleware | sipon-api | sipon-be |
|---|---|---|
| RequestID | dipasang (root) | (perlu cek root wiring — di luar modul identity) |
| CORS | dipasang (AllowOrigins `["*"]`) | (perlu cek root wiring) |
| RequestLogger | dipasang | (perlu cek root wiring) |
| httperror.Middleware | dipasang | (perlu cek root wiring) |
| Swagger dev | dipasang bila `appEnv=="development"` | tidak ada di modul identity |

### 6.3 Error Code String

| Kasus | sipon-api | sipon-be |
|---|---|---|
| Forbidden (no principal) | `ERR_FORBIDDEN` | `FORBIDDEN` |
| Insufficient permission | `ERR_FORBIDDEN` | `INSUFFICIENT_PERMISSION` |
| Bad request | `ERR_BAD_REQUEST` | `ERR_BAD_REQUEST` |
| NotFound | `ERR_NOT_FOUND` | `ERR_NOT_FOUND` |
| Conflict | `ERR_CONFLICT` | `ERR_CONFLICT` |
| Unprocessable | `ERR_UNPROCESSABLE_ENTITY` | `ERR_UNPROCESSABLE_ENTITY` |
| Gone | `ERR_GONE` | `ERR_GONE` |
| TooManyRequests | `ERR_TOO_MANY_REQUESTS` | `ERR_TOO_MANY_REQUESTS` |

**Gap**: kode forbidden pecah (`ERR_FORBIDDEN` vs `FORBIDDEN`/`INSUFFICIENT_PERMISSION`). Frontend yang match string wajib sesuaikan.

### 6.4 Response Envelope

| Sisi | Sukses | Error |
|---|---|---|
| sipon-api | `{status:"success", status_code, message, data, meta}` (helper `OK/Created/SuccessWithMeta`) | `{status:"error", status_code, error_code, errors}` |
| sipon-be | `{status:"success", status_code, message, data, meta}` (helper `OK/Created`) | `{status:"error", status_code, error_code, errors}` |

Envelope SAMA. Yang beda adalah **isi `data`** (struktur nested vs flatten) dan **`message` string** per endpoint.

### 6.5 Domain Error Code Mapping

- sipon-api mendefinisikan kode domain spesifik (`DOMAIN_ROLE_NOT_FOUND`, `DOMAIN_USER_ROLE_DUPLICATE`, `DOMAIN_ROLE_PERMISSION_REQUIRES_CUSTOM_ROLE`, dll) di `role_constant.go` → dipetakan ke HTTP via `apperror`.
- sipon-be pakai `kernel.Code` generik (`ERR_NOT_FOUND`, `ERR_CONFLICT`, `ERR_BAD_REQUEST`, `ERR_GONE`, `ERR_UNPROCESSABLE_ENTITY`) — kode domain spesifik (`DOMAIN_*`) **tidak dipakai**. Frontend yang match kode domain akan kehilangan granularitas.

### 6.6 Session Revocation Side-Effect

| Trigger | sipon-api | sipon-be |
|---|---|---|
| `change-password` (self) | revoke session aktif → re-login | **tidak revoke** |
| `password/reset` (admin reset) | revoke all session user target | **tidak revoke** |
| `deactivate` user | revoke all session user target | **tidak revoke** |

**Gap kritis**: setelah password change/admin-reset/deactivate, session lama di be **masih valid**. Risiko keamanan.

### 6.7 Message String (perbedaan kalimat)

Setiap endpoint memakai kalimat `message` berbeda antara dua sisi. Daftar singkat perbedaan signifikan:

| Endpoint | api | be |
|---|---|---|
| login | `"login success"` | `"login success"` (sama) |
| refresh-token | `"refresh token success"` | `"refresh token success"` (sama) |
| logout | `"Logout berhasil"` | `"Logged out successfully"` |
| me | `"me success"` | `"me retrieved"` |
| profile | `"OK"` | `"profile retrieved"` |
| change-password | `"change password success"` | `"Password changed successfully"` |
| set-password | `"set password success"` | `"Password set successfully"` |
| update-profile | `"update profile success"` | `"Profile updated successfully"` |
| avatar-confirm | `"avatar confirm success"` | `"Avatar confirmed successfully"` |
| avatar-delete | `"avatar delete success"` | `"Avatar deleted successfully"` |
| list-users | `"users fetched"` | `"users retrieved"` |
| get-user | `"user fetched"` | `"user retrieved"` |
| create-user | `"user created"` | `"User created successfully"` |
| reset-user-password | `"user password reset"` | `"User password reset successfully"` |
| deactivate-user | `"user deactivated"` | `"User deactivated successfully"` |
| reactivate-user | `"user reactivated"` | `"User reactivated successfully"` |
| list-roles | `"roles fetched"` | `"roles retrieved"` |
| get-role | `"role fetched"` | `"role retrieved"` |
| list-permissions | `"permission keys fetched"` | `"permissions retrieved"` |
| assign-role-permission | `"permission assigned to role"` | `"Permission assigned to role successfully"` |
| list-user-roles | `"user roles fetched"` | `"user roles retrieved"` |
| assign-user-role | `"user role assigned"` | `"Role assigned to user successfully"` |

Pattern: api pakai lowercase + verb-noun, be pakai Title Case + Verb-noun-`successfully`. Konsisten frontend wajib baca dari field `message` secara soft (display), tidak boleh match string.

---

## §7. Daftar Endpoint yang BELUM Diimplementasikan di sipon-be (Non-Santri)

Hanya **1 endpoint** belum ada route di sipon-be:

| # | Method | Path | Handler api | Status be |
|---|---|---|---|---|
| 1 | GET | `/api/v1/web/role-permission/user-roles/:user_role_id` | `RolePermissionHandler.GetUserRole` (router.go:155) | **BELUM ADA** — tidak ada route di `router.go` sipon-be |

**Catatan**: tidak ada handler stub/hilang. Semua 45 route yang terdaftar di router sipon-be punya definisi `func (h *IdentityHandler) X` di `handler.go`. Yang "belum" di sisi be hanya berupa route yang tidak didaftarkan (1 endpoint di atas) plus sejumlah behavior gap (§3-§5) di mana handler ada tetapi flow tidak setara (`GetRole` membuang permissions, `Logout` tidak ekspos revoke-all/device, `Delete*` family tanpa validasi, `ForgotPassword` tidak anti-enumeration, `CheckUsername` tidak exclude-self, `UpdateProfile` tanpa cek verified).

### Daftar handler dengan nama berbeda (referensi silang)

| api (handler method) | be (handler method) |
|---|---|
| `ListPermissionKeys` | `ListPermissions` |
| `RevokeRolePermission` | `DeleteRolePermission` |
| `RemoveRoleScope` | `DeleteRoleScope` |
| `ForgotPassword` / `ResetPassword` | `ForgotPassword` / `ResetPassword` (nama sama) |
| `RequestChangeEmail` / `ConfirmChangeEmail` | `RequestChangeIdentityEmail` / `ConfirmChangeIdentityEmail` |
| `RequestChangePhone` / `ConfirmChangePhone` | `RequestChangeIdentityPhone` / `ConfirmChangeIdentityPhone` |
| `ChangePasswordLocal` / `SetPasswordLocal` | `ChangePassword` / `SetPassword` |

---

## §8. Ringkas Perbaikan Prioritas

| Prioritas | Item | Lokasi be |
|---|---|---|
| P0 | Tambah route `GET /user-roles/:user_role_id` + handler | `router.go`, `handler.go` |
| P0 | Pasang `RequirePermission` read guard di 5 route read role-permission (roles, roles/:id, permission-keys, user-roles, scopes) | `router.go:113-128` |
| P0 | Revoke session pada change-password / admin reset-password / deactivate user | usecase `change_password.go`, `manage_user.go` |
| P0 | ForgotPassword anti-enumeration: selalu 200 walau email tidak ada | `command/forgot_password.go` |
| P0 | CheckUsername: kirim userID ke usecase (exclude self); invalid format → 422 (bukan `available:false`); username kosong → 422 (bukan 400) | `handler.go:407`, `query/check_username.go` |
| P1 | GetRole: return `permissions[]` (jangan drop `permItems`) | `query/get_role.go` |
| P1 | Rate-limit: pasang `RateLimitByIP` di public-auth + `RateLimitByUser` di protected | `router.go` |
| P1 | Guard users group: ubah ke route-level tunggal (OR) atau dokumentasikan AND | `router.go:99-108` |
| P1 | CreateUser:如意 generates password + return `generated_password` (atau bytebin API kontrak) | `application/dto/auth_dto.go:184`, `command/manage_user.go` |
| P1 | AssignUserRole: ganti `role_name` → `role_id`; `scope_type` required enum; tambah `notes` | `application/dto/auth_dto.go:290` |
| P1 | AvatarConfirm: input `?key=` query (bukan body) + return `avatar_url` | `application/dto/auth_dto.go:151`, `handler.go:449` |
| P2 | UpdateProfile: cek verified-flag sebelum ubah email/phone; fullname+email opsional | `command/update_profile.go` |
| P2 | Not-found error di Update/Deactivate/ReactivateUserRole → 404 (bukan 400) | usecase `manage_role.go` |
| P2 | Delete family (RolePermission/UserRole/RoleScope): validasi eksistensi → 404 bila tidak ada | `command/manage_permission.go`, `manage_role.go`, `manage_scope.go` |
| P2 | RevokeRolePermission: tolak 409 untuk system role | `command/manage_permission.go` |
| P2 | CreateRole: cek duplikat nama → 409 | `command/manage_role.go` |
| P2 | AssignRoleScope: enum `oneof=gender`/`oneof=male|female` di DTO | `application/dto/auth_dto.go` |
| P3 | Seragamkan struktur response `data` (nested vs flatten) & meta field (`page`/`limit` vs `current_page`/`per_page`) | semua DTO |
| P3 | Seragamkan error code forbidden (`ERR_FORBIDDEN` vs `FORBIDDEN`/`INSUFFICIENT_PERMISSION`) | `shared/middleware/auth.go` |
| P3 | Tambah field `expires_in` di AvatarPresign response | `command/avatar.go` |
| P3 | Logout: ekspos `revoke-all`/`revoke-device` (usecase sudah punya) | `router.go`, `handler.go` |

---

## Lampiran A — Sumber Router

- **sipon-api**: `internal/interfaces/http/router/router.go` (195 baris)
  - Auth public: baris 59-72
  - Session: baris 75-83
  - Auth protected: baris 86-192
  - Users group: baris 115-123
  - Role-permission group: baris 131-165
  - Santri (EXCLUDE): baris 168-191
- **sipon-be**: `internal/modules/identity/interfaces/http/router.go` (149 baris)
  - Auth public: baris 59-68
  - Session: baris 70-75
  - Auth protected: baris 77-97
  - Users group: baris 99-108
  - Role-permission group: baris 110-129

## Lampiran B — Konstanta Permission (sama kedua sisi)

- `manage_system_settings`
- `assign_role`
- `manage_users`
- `reset_user_password`
- `deactivate_user`
- `manage_roles`
- `manage_role_permissions`

## Lampiran C — Cek Kehadiran Handler vs Route di sipon-be

Semua 45 referensi `handler.X` di `router.go` sipon-be memiliki definisi `func (h *IdentityHandler) X(c *gin.Context)` di `handler.go`. **Tidak ada stub kosong / compile-error potensial dari sisi handler.** Yang "belum" hanya berupa:
- 1 route tidak terdaftar (GET `/user-roles/:user_role_id`), dan
- Behavior gap di handler yang ada (lihat §3-§5).