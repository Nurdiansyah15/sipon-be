package dto

import "time"

type PaginationParams struct {
	Page  int `form:"page" json:"page" binding:"min=1"`
	Limit int `form:"limit" json:"limit" binding:"min=1,max=100"`
}

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type RegisterRequest struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterResponse struct {
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
	Email        string   `json:"email"`
	Phone        *string  `json:"phone"`
	Roles        []string `json:"roles"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
}

type LoginRequest struct {
	Identity string `json:"identity" binding:"required"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id"`
}

type LoginResponse struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Phone       *string  `json:"phone"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	AccessToken string   `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int64    `json:"expires_in"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RequestOTPRequest struct {
	Identity string `json:"identity" binding:"required"`
}

type VerifyOTPRequest struct {
	Identity string `json:"identity" binding:"required"`
	Code     string `json:"code" binding:"required,len=6"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type SetPasswordRequest struct {
	Password string `json:"password" binding:"required,min=8"`
}

type ChangeIdentityRequest struct {
	NewValue string `json:"new_value" binding:"required"`
}

type ChangeIdentityConfirmRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

type UpdateProfileRequest struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
}

type ChangeUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

type SessionResponse struct {
	UserID      string          `json:"user_id"`
	Username    string          `json:"username"`
	Fullname    *string         `json:"fullname"`
	Email       string          `json:"email"`
	Phone       *string         `json:"phone"`
	Roles       []string        `json:"roles"`
	Permissions []string        `json:"permissions"`
	Scopes      []ScopeResponse `json:"scopes"`
}

type ScopeResponse struct {
	ScopeType string  `json:"scope_type"`
	ScopeID   *string `json:"scope_id"`
}

type ProfileResponse struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Fullname    *string   `json:"fullname"`
	Email       string    `json:"email"`
	Phone       *string   `json:"phone"`
	AvatarURL   *string   `json:"avatar_url"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type AvatarPresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
}

type AvatarPresignResponse struct {
	PresignURL string `json:"presign_url"`
	Key        string `json:"key"`
}

type AvatarConfirmRequest struct {
	Key string `json:"key" binding:"required"`
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

type ListUsersResponse struct {
	Users []UserItem `json:"users"`
	Meta  Meta       `json:"meta"`
}

type UserItem struct {
	ID          string     `json:"id"`
	Username    string     `json:"username"`
	Fullname    *string    `json:"fullname"`
	Email       string     `json:"email"`
	Phone       *string    `json:"phone"`
	Status      string     `json:"status"`
	Roles       []string   `json:"roles"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

type CreateUserRequest struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	RoleName string `json:"role_name" binding:"required"`
}

type ResetUserPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ListRolesRequest struct {
	RoleType   string `form:"role_type"`
	ScopeType  string `form:"scope_type"`
	Assignable *bool  `form:"assignable"`
	PaginationParams
}

type ListRolesResponse struct {
	Roles []RoleItem `json:"roles"`
	Meta  Meta       `json:"meta"`
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
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Description string `json:"description"`
	ScopeType   string `json:"scope_type" binding:"required"`
	Assignable  bool   `json:"assignable"`
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
	IsActive  *bool  `form:"is_active"`
	PaginationParams
}

type ListUserRolesResponse struct {
	UserRoles []UserRoleItem `json:"user_roles"`
	Meta      Meta           `json:"meta"`
}

type UserRoleItem struct {
	ID            string      `json:"id"`
	UserID        string      `json:"user_id"`
	User          UserSummary `json:"user"`
	RoleID        string      `json:"role_id"`
	Role          RoleSummary `json:"role"`
	ScopeType     string      `json:"scope_type"`
	ScopeID       *string     `json:"scope_id"`
	AssignedAt    time.Time   `json:"assigned_at"`
	AssignedBy    *string     `json:"assigned_by"`
	ExpiredAt     *time.Time  `json:"expired_at"`
	IsActive      bool        `json:"is_active"`
	DeactivatedAt *time.Time  `json:"deactivated_at"`
	Permissions   []string    `json:"permissions"`
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
	RoleName  string     `json:"role_name" binding:"required"`
	ScopeType string     `json:"scope_type"`
	ScopeID   *string    `json:"scope_id"`
	ExpiredAt *time.Time `json:"expired_at"`
}

type UpdateUserRoleRequest struct {
	ExpiredAt *time.Time `json:"expired_at"`
}

type ListScopesResponse struct {
	Scopes []ScopeItem `json:"scopes"`
}

type ScopeItem struct {
	ID         string `json:"id"`
	ScopeType  string `json:"scope_type"`
	ScopeValue string `json:"scope_value"`
}

type AssignScopeRequest struct {
	ScopeType  string `json:"scope_type" binding:"required"`
	ScopeValue string `json:"scope_value" binding:"required"`
}
