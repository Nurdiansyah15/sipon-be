package persistence

import (
	"database/sql"
	"time"
)

// UserModel represents the users table row
type UserModel struct {
	ID                  string
	Username            string
	Fullname            sql.NullString
	Email               string
	Phone               sql.NullString
	AvatarKey           sql.NullString
	Status              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastLoginAt         sql.NullTime
	DeletedAt           sql.NullTime
	FailedLoginAttempts int
	LockedUntil         sql.NullTime
}

type CredentialModel struct {
	ID            string
	UserID        string
	Type          string
	SecretHash    sql.NullString
	LastChangedAt sql.NullTime
	IsPrimary     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	LastLoginAt   sql.NullTime
	DeletedAt     sql.NullTime
}

type LoginIdentityModel struct {
	ID           string
	UserID       string
	CredentialID string
	Kind         string
	Value        string
	Status       string
	IsPrimary    bool
	VerifiedAt   sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    sql.NullTime
}

type RoleModel struct {
	ID          string
	Name        string
	DisplayName string
	Description sql.NullString
	RoleType    string
	ScopeType   string
	Assignable  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserRoleModel struct {
	ID            string
	UserID        string
	RoleID        string
	ScopeType     string
	ScopeID       sql.NullString
	AssignedAt    time.Time
	AssignedBy    string
	ExpiredAt     sql.NullTime
	IsActive      bool
	Notes         sql.NullString
	DeactivatedAt sql.NullTime
}

type RolePermissionModel struct {
	ID            string
	RoleID        string
	PermissionKey string
	AssignedAt    time.Time
	AssignedBy    string
	Notes         sql.NullString
}

type RoleScopeModel struct {
	ID         string
	RoleID     string
	ScopeType  string
	ScopeValue string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type VerificationCodeModel struct {
	ID               string
	UserID           string
	Code             string
	Purpose          string
	ExpiresAt        time.Time
	UsedAt           sql.NullTime
	CreatedAt        time.Time
	NewIdentityValue sql.NullString
}
