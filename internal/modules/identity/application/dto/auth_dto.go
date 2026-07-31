package dto

import "time"

type PaginationParams struct {
	Page  int `form:"page" json:"page" binding:"min=1"`
	Limit int `form:"limit" json:"limit" binding:"min=1,max=100"`
}

type Meta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	TotalPages  int `json:"total_pages"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type RegisterRequest struct {
	Username string  `json:"username" binding:"required,min=3,max=30"`
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	Phone    *string `json:"phone,omitempty"`
	Fullname *string `json:"fullname,omitempty"`
	DeviceID string  `json:"device_id,omitempty"`
}

type UserMe struct {
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	IsEmailVerified bool      `json:"is_email_verified"`
	Fullname        *string   `json:"fullname"`
	Phone           *string   `json:"phone,omitempty"`
	IsPhoneVerified bool      `json:"is_phone_verified"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	HasPassword     bool      `json:"has_password"`
	AvatarURL       *string   `json:"avatar_url,omitempty"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         UserMe `json:"user"`
}

type RegisterResponse struct {
	UserID string `json:"user_id"`
	LoginResponse
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
	DeviceID   string `json:"device_id,omitempty"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RequestOTPRequest struct {
	Identifier string `json:"identifier" binding:"required"`
}

type VerifyOTPRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	OTP        string `json:"otp" binding:"required,len=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type SetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type RequestChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type ConfirmChangeEmailRequest struct {
	OTP string `json:"otp" binding:"required,len=6"`
}

type RequestChangePhoneRequest struct {
	NewPhone string `json:"new_phone" binding:"required"`
}

type ConfirmChangePhoneRequest struct {
	OTP string `json:"otp" binding:"required,len=6"`
}

type ChangeIdentityResponse struct {
	Message string `json:"message"`
}

type UpdateProfileRequest struct {
	Fullname *string `json:"fullname,omitempty"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
}

type ChangeUsernameRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
}

type ChangeUsernameResponse struct {
	Message  string `json:"message"`
	Username string `json:"username"`
}

// SessionRole, SessionPermission and SessionUserScope back both GetSession and
// Profile — sipon-api resolves roles/permissions/scopes as rich objects (not
// bare string arrays) on these two endpoints specifically.
type SessionUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type SessionRole struct {
	Name      string  `json:"name"`
	RoleType  string  `json:"role_type"`
	ScopeType string  `json:"scope_type"`
	ScopeID   *string `json:"scope_id"`
}

type SessionPermission struct {
	Key   string `json:"key"`
	Scope string `json:"scope"`
}

type SessionUserScope struct {
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
}

type SessionResponse struct {
	User        SessionUser         `json:"user"`
	Roles       []SessionRole       `json:"roles"`
	Permissions []SessionPermission `json:"permissions"`
	Scopes      []SessionUserScope  `json:"scopes"`
}

type ProfileResponse struct {
	ID              string              `json:"id"`
	Username        string              `json:"username"`
	Fullname        *string             `json:"fullname"`
	Email           string              `json:"email"`
	IsEmailVerified bool                `json:"is_email_verified"`
	Phone           *string             `json:"phone,omitempty"`
	IsPhoneVerified bool                `json:"is_phone_verified"`
	Status          string              `json:"status"`
	HasPassword     bool                `json:"has_password"`
	CreatedAt       time.Time           `json:"created_at"`
	AvatarURL       *string             `json:"avatar_url,omitempty"`
	Roles           []SessionRole       `json:"roles"`
	Permissions     []SessionPermission `json:"permissions"`
	Scopes          []SessionUserScope  `json:"scopes"`
}

type AvatarPresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
}

type AvatarPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
	ExpiresIn  int    `json:"expires_in"`
}

type AvatarConfirmResponse struct {
	AvatarURL string `json:"avatar_url"`
}

type CheckUsernameResponse struct {
	Available bool `json:"available"`
}

type ListUsersRequest struct {
	Status string `form:"status"`
	RoleID string `form:"role_id"`
	Search string `form:"search"`
	PaginationParams
}

type UserRoleSummaryResponse struct {
	ID        string  `json:"id"`
	RoleID    string  `json:"role_id"`
	RoleName  string  `json:"role_name"`
	ScopeType string  `json:"scope_type"`
	ScopeID   *string `json:"scope_id"`
	IsActive  bool    `json:"is_active"`
}

type UserManagementResponse struct {
	ID          string                    `json:"id"`
	Username    string                    `json:"username"`
	Fullname    *string                   `json:"fullname,omitempty"`
	Email       string                    `json:"email"`
	Phone       *string                   `json:"phone,omitempty"`
	Status      string                    `json:"status"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	LastLoginAt *time.Time                `json:"last_login_at,omitempty"`
	Roles       []UserRoleSummaryResponse `json:"roles,omitempty"`
}

type CreateUserRequest struct {
	Fullname *string `json:"fullname,omitempty"`
	Email    string  `json:"email" binding:"required,email"`
	Phone    *string `json:"phone,omitempty"`
	Username string  `json:"username" binding:"required"`
}

type CreateUserResponse struct {
	UserManagementResponse
	GeneratedPassword string `json:"generated_password"`
}

type ResetUserPasswordResponse struct {
	GeneratedPassword string `json:"generated_password"`
}

type ListRolesRequest struct {
	RoleType   string `form:"role_type"`
	ScopeType  string `form:"scope_type"`
	Assignable *bool  `form:"assignable"`
	PaginationParams
}

type RoleItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description *string   `json:"description"`
	RoleType    string    `json:"role_type"`
	ScopeType   string    `json:"scope_type"`
	Assignable  bool      `json:"assignable"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Permissions []string  `json:"permissions,omitempty"`
}

type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required"`
	DisplayName string  `json:"display_name" binding:"required"`
	Description *string `json:"description,omitempty"`
	RoleType    string  `json:"role_type" binding:"required,oneof=system custom"`
	ScopeType   string  `json:"scope_type" binding:"required,oneof=global region community"`
	Assignable  bool    `json:"assignable"`
}

type UpdateRoleRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Assignable  *bool  `json:"assignable"`
}

type PermissionItem struct {
	Key         string `json:"key"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

type AssignRolePermissionRequest struct {
	PermissionKey string `json:"permission_key" binding:"required"`
	Notes         string `json:"notes"`
}

type ListUserRolesRequest struct {
	UserID    string `form:"user_id"`
	RoleID    string `form:"role_id"`
	ScopeType string `form:"scope_type"`
	ScopeID   string `form:"scope_id"`
	IsActive  *bool  `form:"is_active"`
	PaginationParams
}

type UserRoleItem struct {
	ID            string      `json:"id"`
	UserID        string      `json:"user_id"`
	User          UserSummary `json:"user"`
	RoleID        string      `json:"role_id"`
	Role          RoleSummary `json:"role"`
	ScopeType     string      `json:"scope_type"`
	ScopeID       *string     `json:"scope_id,omitempty"`
	AssignedAt    time.Time   `json:"assigned_at"`
	AssignedBy    *string     `json:"assigned_by,omitempty"`
	ExpiredAt     *time.Time  `json:"expired_at,omitempty"`
	IsActive      bool        `json:"is_active"`
	DeactivatedAt *time.Time  `json:"deactivated_at,omitempty"`
	Permissions   []string    `json:"permissions,omitempty"`
}

type UserSummary struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

type RoleSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	RoleType    string `json:"role_type"`
	Assignable  bool   `json:"assignable"`
}

type AssignUserRoleRequest struct {
	UserID    string     `json:"user_id" binding:"required"`
	RoleID    string     `json:"role_id" binding:"required"`
	ScopeType string     `json:"scope_type" binding:"required,oneof=global region community"`
	ScopeID   *string    `json:"scope_id,omitempty"`
	ExpiredAt *time.Time `json:"expired_at,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

type UpdateUserRoleRequest struct {
	ExpiredAt *time.Time `json:"expired_at,omitempty"`
}

type ScopeItem struct {
	ID         string `json:"id"`
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
}

type AssignScopeRequest struct {
	ScopeType  string `json:"scope_type" binding:"required,oneof=gender"`
	ScopeValue string `json:"scope_value" binding:"required,oneof=male female"`
}
