package valueobject

import (
	"regexp"
	"strings"

	"sipon-be/internal/modules/identity/domain/user/constant"
	"sipon-be/internal/shared/kernel"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

var digitRegex = regexp.MustCompile(`^\d{6}$`)

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{6,14}$`)

var uppercaseRegex = regexp.MustCompile(`[A-Z]`)
var digitRegexCompiled = regexp.MustCompile(`[0-9]`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ToLower(raw)
	if raw == "" {
		return Email{}, kernel.New(constant.ErrCodeEmailEmpty)
	}
	if !emailRegex.MatchString(raw) {
		return Email{}, kernel.New(constant.ErrCodeEmailInvalidFormat)
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
		return PhoneNumber{}, kernel.New(constant.ErrCodePhoneNumberEmpty)
	}
	normalized := NormalizePhoneNumber(raw)
	if !phoneRegex.MatchString(normalized) {
		return PhoneNumber{}, kernel.New(constant.ErrCodePhoneNumberInvalidFormat)
	}
	return PhoneNumber{value: normalized}, nil
}

// NormalizePhoneNumber menormalisasi nomor telepon Indonesia ke format +62.
func NormalizePhoneNumber(raw string) string {
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
		return HashedPassword{}, kernel.New(constant.ErrCodeHashedPasswordTooShort)
	}
	return HashedPassword{hash: hash}, nil
}

func (h HashedPassword) String() string {
	return h.hash
}

type PlainPassword struct {
	plain string
}

func NewPlainPassword(raw string) (PlainPassword, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return PlainPassword{}, kernel.New(constant.ErrCodePlainPasswordEmpty)
	}
	if len(raw) < 8 {
		return PlainPassword{}, kernel.New(constant.ErrCodePlainPasswordTooShort)
	}
	if !uppercaseRegex.MatchString(raw) {
		return PlainPassword{}, kernel.New(constant.ErrCodePlainPasswordNoUppercase)
	}
	if !digitRegexCompiled.MatchString(raw) {
		return PlainPassword{}, kernel.New(constant.ErrCodePlainPasswordNoDigit)
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
		return OTPCode{}, kernel.New(constant.ErrCodeOTPCodeEmpty)
	}
	if len(code) != 6 {
		return OTPCode{}, kernel.New(constant.ErrCodeOTPCodeInvalidLength)
	}
	if !digitRegex.MatchString(code) {
		return OTPCode{}, kernel.New(constant.ErrCodeOTPCodeNotDigit)
	}
	return OTPCode{code: code}, nil
}

func (o OTPCode) String() string {
	return o.code
}

func (o OTPCode) Match(input string) bool {
	return o.code == input
}

type Username struct {
	value string
}

func NewUsername(raw string) (Username, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return Username{}, kernel.New(constant.ErrCodeUsernameEmpty)
	}
	if len(raw) < 3 {
		return Username{}, kernel.New(constant.ErrCodeUsernameTooShort)
	}
	if len(raw) > 30 {
		return Username{}, kernel.New(constant.ErrCodeUsernameTooLong)
	}
	if !usernameRegex.MatchString(raw) {
		return Username{}, kernel.New(constant.ErrCodeUsernameInvalidChar)
	}
	return Username{value: raw}, nil
}

func (u Username) String() string {
	return u.value
}

type LoginIdentifier struct {
	Kind  constant.LoginIdentifierKind
	Value string
}

func NewLoginIdentifier(raw string) (LoginIdentifier, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ToLower(raw)
	if raw == "" {
		return LoginIdentifier{}, kernel.New(constant.ErrCodeLoginIdentifierEmpty)
	}

	if emailRegex.MatchString(raw) {
		return LoginIdentifier{Kind: constant.LoginIdentifierKindEmail, Value: raw}, nil
	}

	normalizedPhone := NormalizePhoneNumber(raw)
	if phoneRegex.MatchString(normalizedPhone) {
		return LoginIdentifier{Kind: constant.LoginIdentifierKindPhone, Value: normalizedPhone}, nil
	}

	if usernameRegex.MatchString(raw) && len(raw) >= 3 && len(raw) <= 30 {
		return LoginIdentifier{Kind: constant.LoginIdentifierKindUsername, Value: raw}, nil
	}

	return LoginIdentifier{}, kernel.New(constant.ErrCodeLoginIdentifierUnknownKind)
}

// NormalizeLoginIdentityValue menormalisasi & memvalidasi nilai suatu login
// identity sesuai kind-nya. Dipakai entity.LoginIdentity saat konstruksi.
func NormalizeLoginIdentityValue(kind constant.LoginIdentifierKind, rawValue string) (string, error) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return "", kernel.New(constant.ErrCodeInvalidLoginIdentityValue)
	}

	switch kind {
	case constant.LoginIdentifierKindEmail:
		rawValue = strings.ToLower(rawValue)
		if !emailRegex.MatchString(rawValue) {
			return "", kernel.New(constant.ErrCodeEmailInvalidFormat)
		}
		return rawValue, nil

	case constant.LoginIdentifierKindPhone:
		normalized := NormalizePhoneNumber(rawValue)
		if !phoneRegex.MatchString(normalized) {
			return "", kernel.New(constant.ErrCodePhoneNumberInvalidFormat)
		}
		return normalized, nil

	case constant.LoginIdentifierKindUsername:
		if !usernameRegex.MatchString(rawValue) {
			return "", kernel.New(constant.ErrCodeUsernameInvalidChar)
		}
		if len(rawValue) < 3 {
			return "", kernel.New(constant.ErrCodeUsernameTooShort)
		}
		if len(rawValue) > 30 {
			return "", kernel.New(constant.ErrCodeUsernameTooLong)
		}
		return rawValue, nil

	default:
		return "", kernel.New(constant.ErrCodeLoginIdentifierUnknownKind)
	}
}
