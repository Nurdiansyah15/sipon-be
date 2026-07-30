package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"sipon-be/internal/shared/kernel"
)

type CredentialType string

const (
	CredentialTypeLocal CredentialType = "LOCAL"
)

type LoginIdentifierKind string

const (
	LoginIdentifierKindEmail    LoginIdentifierKind = "EMAIL"
	LoginIdentifierKindPhone    LoginIdentifierKind = "PHONE"
	LoginIdentifierKindUsername LoginIdentifierKind = "USERNAME"
)

type LoginIdentityStatus string

const (
	LoginIdentityStatusVerified   LoginIdentityStatus = "VERIFIED"
	LoginIdentityStatusUnverified LoginIdentityStatus = "UNVERIFIED"
)

type UserStatus string

const (
	UserStatusActive UserStatus = "ACTIVE"
	UserStatusBanned UserStatus = "BANNED"
)

const (
	MaxLoginAttempts = 5
	LockoutDuration  = 15 * time.Minute
)

const (
	ErrCodeEmailInvalidFormat          kernel.Code = "EMAIL_INVALID_FORMAT"
	ErrCodeEmailEmpty                  kernel.Code = "EMAIL_EMPTY"
	ErrCodePhoneNumberInvalidFormat    kernel.Code = "PHONE_NUMBER_INVALID_FORMAT"
	ErrCodePhoneNumberEmpty            kernel.Code = "PHONE_NUMBER_EMPTY"
	ErrCodeHashedPasswordTooShort      kernel.Code = "HASHED_PASSWORD_TOO_SHORT"
	ErrCodePlainPasswordTooShort       kernel.Code = "PLAIN_PASSWORD_TOO_SHORT"
	ErrCodePlainPasswordNoUppercase    kernel.Code = "PLAIN_PASSWORD_NO_UPPERCASE"
	ErrCodePlainPasswordNoDigit        kernel.Code = "PLAIN_PASSWORD_NO_DIGIT"
	ErrCodePlainPasswordEmpty          kernel.Code = "PLAIN_PASSWORD_EMPTY"
	ErrCodeOTPCodeInvalidLength        kernel.Code = "OTP_CODE_INVALID_LENGTH"
	ErrCodeOTPCodeNotDigit             kernel.Code = "OTP_CODE_NOT_DIGIT"
	ErrCodeOTPCodeEmpty                kernel.Code = "OTP_CODE_EMPTY"
	ErrCodeUsernameTooShort            kernel.Code = "USERNAME_TOO_SHORT"
	ErrCodeUsernameTooLong             kernel.Code = "USERNAME_TOO_LONG"
	ErrCodeUsernameInvalidChar         kernel.Code = "USERNAME_INVALID_CHAR"
	ErrCodeUsernameEmpty               kernel.Code = "USERNAME_EMPTY"
	ErrCodeLoginIdentifierEmpty        kernel.Code = "LOGIN_IDENTIFIER_EMPTY"
	ErrCodeLoginIdentifierUnknownKind  kernel.Code = "LOGIN_IDENTIFIER_UNKNOWN_KIND"
	ErrCodeUserBanned                  kernel.Code = "USER_BANNED"
	ErrCodeUserLockedOut               kernel.Code = "USER_LOCKED_OUT"
	ErrCodeUserNotActive              kernel.Code = "USER_NOT_ACTIVE"
	ErrCodeUserAlreadyActive           kernel.Code = "USER_ALREADY_ACTIVE"
	ErrCodeUserAlreadyBanned           kernel.Code = "USER_ALREADY_BANNED"
	ErrCodeUserAlreadyDeleted          kernel.Code = "USER_ALREADY_DELETED"
	ErrCodeCredentialNotLocal          kernel.Code = "CREDENTIAL_NOT_LOCAL"
	ErrCodeIdentityNotVerified         kernel.Code = "IDENTITY_NOT_VERIFIED"
	ErrCodeNoPrimaryIdentity           kernel.Code = "NO_PRIMARY_IDENTITY"
	ErrCodeUsernameAlreadySet          kernel.Code = "USERNAME_ALREADY_SET"
	ErrCodeInvalidLoginIdentityValue   kernel.Code = "INVALID_LOGIN_IDENTITY_VALUE"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

var digitRegex = regexp.MustCompile(`^\d{6}$`)

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ToLower(raw)
	if raw == "" {
		return Email{}, kernel.New(ErrCodeEmailEmpty)
	}
	if !emailRegex.MatchString(raw) {
		return Email{}, kernel.New(ErrCodeEmailInvalidFormat)
	}
	return Email{value: raw}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) IsEmpty() bool {
	return e.value == ""
}

type PhoneNumber struct {
	value string
}

func NewPhoneNumber(raw string) (PhoneNumber, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return PhoneNumber{}, kernel.New(ErrCodePhoneNumberEmpty)
	}
	normalized := normalizePhoneNumber(raw)
	if !phoneRegex.MatchString(normalized) {
		return PhoneNumber{}, kernel.New(ErrCodePhoneNumberInvalidFormat)
	}
	return PhoneNumber{value: normalized}, nil
}

func normalizePhoneNumber(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, " ", "")
	raw = strings.ReplaceAll(raw, "-", "")
	if strings.HasPrefix(raw, "0") {
		raw = "+62" + raw[1:]
	}
	if strings.HasPrefix(raw, "62") && !strings.HasPrefix(raw, "+") {
		raw = "+" + raw
	}
	return raw
}

func (p PhoneNumber) String() string {
	return p.value
}

func (p PhoneNumber) IsEmpty() bool {
	return p.value == ""
}

type HashedPassword struct {
	hash string
}

func NewHashedPassword(hash string) (HashedPassword, error) {
	hash = strings.TrimSpace(hash)
	if len(hash) < 10 {
		return HashedPassword{}, kernel.New(ErrCodeHashedPasswordTooShort)
	}
	return HashedPassword{hash: hash}, nil
}

func (h HashedPassword) String() string {
	return h.hash
}

type PlainPassword struct {
	plain string
}

var uppercaseRegex = regexp.MustCompile(`[A-Z]`)
var digitRegexCompiled = regexp.MustCompile(`[0-9]`)

func NewPlainPassword(raw string) (PlainPassword, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return PlainPassword{}, kernel.New(ErrCodePlainPasswordEmpty)
	}
	if len(raw) < 8 {
		return PlainPassword{}, kernel.New(ErrCodePlainPasswordTooShort)
	}
	if !uppercaseRegex.MatchString(raw) {
		return PlainPassword{}, kernel.New(ErrCodePlainPasswordNoUppercase)
	}
	if !digitRegexCompiled.MatchString(raw) {
		return PlainPassword{}, kernel.New(ErrCodePlainPasswordNoDigit)
	}
	return PlainPassword{plain: raw}, nil
}

func (p PlainPassword) String() string {
	return p.plain
}

type OTPCode struct {
	code string
}

func NewOTPCode(code string) (OTPCode, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return OTPCode{}, kernel.New(ErrCodeOTPCodeEmpty)
	}
	if len(code) != 6 {
		return OTPCode{}, kernel.New(ErrCodeOTPCodeInvalidLength)
	}
	if !digitRegex.MatchString(code) {
		return OTPCode{}, kernel.New(ErrCodeOTPCodeNotDigit)
	}
	return OTPCode{code: code}, nil
}

func (o OTPCode) String() string {
	return o.code
}

type Username struct {
	value string
}

func NewUsername(raw string) (Username, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Username{}, kernel.New(ErrCodeUsernameEmpty)
	}
	if len(raw) < 3 {
		return Username{}, kernel.New(ErrCodeUsernameTooShort)
	}
	if len(raw) > 30 {
		return Username{}, kernel.New(ErrCodeUsernameTooLong)
	}
	if !usernameRegex.MatchString(raw) {
		return Username{}, kernel.New(ErrCodeUsernameInvalidChar)
	}
	return Username{value: raw}, nil
}

func (u Username) String() string {
	return u.value
}

type LoginIdentifier struct {
	Kind  LoginIdentifierKind
	Value string
}

func NewLoginIdentifier(raw string) (LoginIdentifier, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ToLower(raw)
	if raw == "" {
		return LoginIdentifier{}, kernel.New(ErrCodeLoginIdentifierEmpty)
	}

	if emailRegex.MatchString(raw) {
		return LoginIdentifier{Kind: LoginIdentifierKindEmail, Value: raw}, nil
	}

	normalizedPhone := normalizePhoneNumber(raw)
	if phoneRegex.MatchString(normalizedPhone) {
		return LoginIdentifier{Kind: LoginIdentifierKindPhone, Value: normalizedPhone}, nil
	}

	if usernameRegex.MatchString(raw) && len(raw) >= 3 && len(raw) <= 30 {
		return LoginIdentifier{Kind: LoginIdentifierKindUsername, Value: raw}, nil
	}

	return LoginIdentifier{}, kernel.New(ErrCodeLoginIdentifierUnknownKind)
}

type User struct {
	ID                  string
	Username            Username
	Fullname            *string
	Email               Email
	PhoneNumber         *PhoneNumber
	AvatarKey           *string
	Status              UserStatus
	Credentials         []*Credential
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastLoginAt         *time.Time
	DeletedAt           *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
}

func NewUser(id string, username Username, fullname *string, email Email, phoneNumber *PhoneNumber) (*User, error) {
	now := time.Now()

	user := &User{
		ID:          id,
		Username:    username,
		Fullname:    fullname,
		Email:       email,
		PhoneNumber: phoneNumber,
		Status:      UserStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return user, nil
}

func (u *User) AddCredential(c *Credential) {
	u.Credentials = append(u.Credentials, c)
	u.UpdatedAt = time.Now()
}

func (u *User) FindLoginIdentity(kind LoginIdentifierKind, value string) *LoginIdentity {
	for _, c := range u.Credentials {
		if li := c.FindLoginIdentity(kind, value); li != nil {
			return li
		}
	}
	return nil
}

func (u *User) FindLoginIdentityByKind(kind LoginIdentifierKind) *LoginIdentity {
	for _, c := range u.Credentials {
		if li := c.FindLoginIdentityByKind(kind); li != nil {
			return li
		}
	}
	return nil
}

func (u *User) FindCredential(typ CredentialType) *Credential {
	for _, c := range u.Credentials {
		if c.Type == typ {
			return c
		}
	}
	return nil
}

func (u *User) EnsureCanLogin() error {
	if u.Status == UserStatusBanned {
		return kernel.New(ErrCodeUserBanned)
	}
	if u.DeletedAt != nil {
		return kernel.New(ErrCodeUserNotActive)
	}
	return u.EnsureNotLockedOut()
}

func (u *User) Activate() error {
	if u.Status == UserStatusActive {
		return kernel.New(ErrCodeUserAlreadyActive)
	}
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Deactivate() error {
	if u.Status == UserStatusBanned {
		return kernel.New(ErrCodeUserAlreadyBanned)
	}
	u.Status = UserStatusBanned
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) Reactivate() error {
	if u.Status != UserStatusBanned {
		return kernel.New(ErrCodeUserNotActive)
	}
	u.Status = UserStatusActive
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
	return nil
}

func (u *User) MarkLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

func (u *User) IsLockedOut() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

func (u *User) EnsureNotLockedOut() error {
	if u.IsLockedOut() {
		return kernel.New(ErrCodeUserLockedOut)
	}
	return nil
}

func (u *User) IncrementFailedAttempts() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= MaxLoginAttempts {
		lockUntil := time.Now().Add(LockoutDuration)
		u.LockedUntil = &lockUntil
	}
	u.UpdatedAt = time.Now()
}

func (u *User) ResetFailedAttempts() {
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	u.UpdatedAt = time.Now()
}

func (u *User) SoftDelete() error {
	if u.DeletedAt != nil {
		return kernel.New(ErrCodeUserAlreadyDeleted)
	}
	now := time.Now()
	u.DeletedAt = &now
	u.UpdatedAt = now
	return nil
}

func (u *User) HasLocalPassword() bool {
	cred := u.FindCredential(CredentialTypeLocal)
	if cred == nil {
		return false
	}
	return cred.SecretHash != nil
}

func (u *User) SetLocalPassword(hashed HashedPassword) error {
	cred := u.FindCredential(CredentialTypeLocal)
	if cred == nil {
		return kernel.New(ErrCodeCredentialNotLocal)
	}
	cred.SecretHash = &hashed
	now := time.Now()
	cred.LastChangedAt = &now
	cred.UpdatedAt = now
	u.UpdatedAt = now
	return nil
}

func (u *User) ChangeUsername(newUsername Username) {
	u.Username = newUsername
	u.UpdatedAt = time.Now()
}

type Credential struct {
	ID              string
	UserID          string
	Type            CredentialType
	LoginIdentities []*LoginIdentity
	SecretHash      *HashedPassword
	LastChangedAt   *time.Time
	IsPrimary       bool
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
	DeletedAt       *time.Time
}

func NewLocalCredential(id, userID string, hashed HashedPassword, isPrimary bool) *Credential {
	now := time.Now()
	return &Credential{
		ID:        id,
		UserID:    userID,
		Type:      CredentialTypeLocal,
		SecretHash: &hashed,
		IsPrimary: isPrimary,
		UpdatedAt: now,
	}
}

func NewLocalCredentialWithoutPassword(id, userID string, isPrimary bool) *Credential {
	now := time.Now()
	return &Credential{
		ID:        id,
		UserID:    userID,
		Type:      CredentialTypeLocal,
		IsPrimary: isPrimary,
		UpdatedAt: now,
	}
}

func (c *Credential) AddLoginIdentity(identity *LoginIdentity) {
	c.LoginIdentities = append(c.LoginIdentities, identity)
	c.UpdatedAt = time.Now()
}

func (c *Credential) FindLoginIdentity(kind LoginIdentifierKind, value string) *LoginIdentity {
	for _, li := range c.LoginIdentities {
		if li.Kind == kind && li.Value == value && li.DeletedAt == nil {
			return li
		}
	}
	return nil
}

func (c *Credential) FindLoginIdentityByKind(kind LoginIdentifierKind) *LoginIdentity {
	for _, li := range c.LoginIdentities {
		if li.Kind == kind && li.DeletedAt == nil {
			return li
		}
	}
	return nil
}

type LoginIdentity struct {
	ID           string
	UserID       string
	CredentialID string
	Kind         LoginIdentifierKind
	Value        string
	Status       LoginIdentityStatus
	IsPrimary    bool
	VerifiedAt   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

func NewLoginIdentity(id, userID, credentialID string, kind LoginIdentifierKind, rawValue string, isPrimary bool, verifiedAt *time.Time) (*LoginIdentity, error) {
	normalizedValue, err := normalizeLoginIdentityValue(kind, rawValue)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &LoginIdentity{
		ID:           id,
		UserID:       userID,
		CredentialID: credentialID,
		Kind:         kind,
		Value:        normalizedValue,
		Status:       LoginIdentityStatusUnverified,
		IsPrimary:    isPrimary,
		VerifiedAt:   verifiedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (li *LoginIdentity) MarkVerified() {
	now := time.Now()
	li.Status = LoginIdentityStatusVerified
	li.VerifiedAt = &now
	li.UpdatedAt = now
}

func (li *LoginIdentity) IsVerified() bool {
	return li.Status == LoginIdentityStatusVerified
}

func (li *LoginIdentity) EnsureVerified() error {
	if !li.IsVerified() {
		return kernel.New(ErrCodeIdentityNotVerified)
	}
	return nil
}

func normalizeLoginIdentityValue(kind LoginIdentifierKind, rawValue string) (string, error) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return "", kernel.New(ErrCodeInvalidLoginIdentityValue)
	}

	switch kind {
	case LoginIdentifierKindEmail:
		rawValue = strings.ToLower(rawValue)
		if !emailRegex.MatchString(rawValue) {
			return "", kernel.New(ErrCodeEmailInvalidFormat)
		}
		return rawValue, nil

	case LoginIdentifierKindPhone:
		normalized := normalizePhoneNumber(rawValue)
		if !phoneRegex.MatchString(normalized) {
			return "", kernel.New(ErrCodePhoneNumberInvalidFormat)
		}
		return normalized, nil

	case LoginIdentifierKindUsername:
		if !usernameRegex.MatchString(rawValue) {
			return "", kernel.New(ErrCodeUsernameInvalidChar)
		}
		if len(rawValue) < 3 {
			return "", kernel.New(ErrCodeUsernameTooShort)
		}
		if len(rawValue) > 30 {
			return "", kernel.New(ErrCodeUsernameTooLong)
		}
		return rawValue, nil

	default:
		return "", fmt.Errorf("unknown login identity kind: %s", kind)
	}
}
