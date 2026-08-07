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

// nisLoginPattern is identity's OWN copy of the NIS shape, used only to
// auto-detect a NIS at login time — it deliberately does NOT import
// kesantrian's NIS value object (identity must not depend on kesantrian).
// The authoritative NIS format validation still lives exactly once, in
// kesantrian's own domain/santri/valueobject/nis.go; this copy exists
// purely because a raw NIS digit string (e.g. "1000112345") also matches
// phoneRegex below, and without this check it would be misclassified as
// PHONE — for which no login identity exists — before ever reaching here.
var nisLoginPattern = regexp.MustCompile(`^1000[12][0-9]{5}$`)

type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.ToLower(raw)
	if raw == "" {
		return Email{}, kernel.WrapMsg(constant.ErrCodeEmailEmpty, "Email tidak boleh kosong", nil)
	}
	if !emailRegex.MatchString(raw) {
		return Email{}, kernel.WrapMsg(constant.ErrCodeEmailInvalidFormat, "Format email tidak valid", nil)
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
		return PhoneNumber{}, kernel.WrapMsg(constant.ErrCodePhoneNumberEmpty, "Nomor telepon tidak boleh kosong", nil)
	}
	normalized := NormalizePhoneNumber(raw)
	if !phoneRegex.MatchString(normalized) {
		return PhoneNumber{}, kernel.WrapMsg(constant.ErrCodePhoneNumberInvalidFormat, "Format nomor telepon tidak valid", nil)
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
		return HashedPassword{}, kernel.WrapMsg(constant.ErrCodeHashedPasswordTooShort, "Hashed password terlalu pendek", nil)
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
		return PlainPassword{}, kernel.WrapMsg(constant.ErrCodePlainPasswordEmpty, "Kata sandi tidak boleh kosong", nil)
	}
	if len(raw) < 8 {
		return PlainPassword{}, kernel.WrapMsg(constant.ErrCodePlainPasswordTooShort, "Kata sandi terlalu pendek (minimal 8 karakter)", nil)
	}
	if !uppercaseRegex.MatchString(raw) {
		return PlainPassword{}, kernel.WrapMsg(constant.ErrCodePlainPasswordNoUppercase, "Kata sandi harus mengandung huruf kapital", nil)
	}
	if !digitRegexCompiled.MatchString(raw) {
		return PlainPassword{}, kernel.WrapMsg(constant.ErrCodePlainPasswordNoDigit, "Kata sandi harus mengandung angka", nil)
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
		return OTPCode{}, kernel.WrapMsg(constant.ErrCodeOTPCodeEmpty, "Kode OTP tidak boleh kosong", nil)
	}
	if len(code) != 6 {
		return OTPCode{}, kernel.WrapMsg(constant.ErrCodeOTPCodeInvalidLength, "Panjang kode OTP harus 6 digit", nil)
	}
	if !digitRegex.MatchString(code) {
		return OTPCode{}, kernel.WrapMsg(constant.ErrCodeOTPCodeNotDigit, "Kode OTP harus berupa angka", nil)
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
		return Username{}, kernel.WrapMsg(constant.ErrCodeUsernameEmpty, "Username tidak boleh kosong", nil)
	}
	if len(raw) < 3 {
		return Username{}, kernel.WrapMsg(constant.ErrCodeUsernameTooShort, "Username terlalu pendek (minimal 3 karakter)", nil)
	}
	if len(raw) > 30 {
		return Username{}, kernel.WrapMsg(constant.ErrCodeUsernameTooLong, "Username terlalu panjang (maksimal 30 karakter)", nil)
	}
	if !usernameRegex.MatchString(raw) {
		return Username{}, kernel.WrapMsg(constant.ErrCodeUsernameInvalidChar, "Username hanya boleh mengandung huruf, angka, dan underscore", nil)
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
		return LoginIdentifier{}, kernel.WrapMsg(constant.ErrCodeLoginIdentifierEmpty, "Identitas login tidak boleh kosong", nil)
	}

	if emailRegex.MatchString(raw) {
		return LoginIdentifier{Kind: constant.LoginIdentifierKindEmail, Value: raw}, nil
	}

	if nisLoginPattern.MatchString(raw) {
		return LoginIdentifier{Kind: constant.LoginIdentifierKindNIS, Value: raw}, nil
	}

	normalizedPhone := NormalizePhoneNumber(raw)
	if phoneRegex.MatchString(normalizedPhone) {
		return LoginIdentifier{Kind: constant.LoginIdentifierKindPhone, Value: normalizedPhone}, nil
	}

	if usernameRegex.MatchString(raw) && len(raw) >= 3 && len(raw) <= 30 {
		return LoginIdentifier{Kind: constant.LoginIdentifierKindUsername, Value: raw}, nil
	}

	return LoginIdentifier{}, kernel.WrapMsg(constant.ErrCodeLoginIdentifierUnknownKind, "Jenis identitas login tidak dikenali", nil)
}

// NormalizeLoginIdentityValue menormalisasi & memvalidasi nilai suatu login
// identity sesuai kind-nya. Dipakai entity.LoginIdentity saat konstruksi.
func NormalizeLoginIdentityValue(kind constant.LoginIdentifierKind, rawValue string) (string, error) {
	rawValue = strings.TrimSpace(rawValue)
	if rawValue == "" {
		return "", kernel.WrapMsg(constant.ErrCodeInvalidLoginIdentityValue, "Nilai identitas login tidak valid", nil)
	}

	switch kind {
	case constant.LoginIdentifierKindEmail:
		rawValue = strings.ToLower(rawValue)
		if !emailRegex.MatchString(rawValue) {
			return "", kernel.WrapMsg(constant.ErrCodeEmailInvalidFormat, "Format email tidak valid", nil)
		}
		return rawValue, nil

	case constant.LoginIdentifierKindPhone:
		normalized := NormalizePhoneNumber(rawValue)
		if !phoneRegex.MatchString(normalized) {
			return "", kernel.WrapMsg(constant.ErrCodePhoneNumberInvalidFormat, "Format nomor telepon tidak valid", nil)
		}
		return normalized, nil

	case constant.LoginIdentifierKindUsername:
		if !usernameRegex.MatchString(rawValue) {
			return "", kernel.WrapMsg(constant.ErrCodeUsernameInvalidChar, "Username hanya boleh mengandung huruf, angka, dan underscore", nil)
		}
		if len(rawValue) < 3 {
			return "", kernel.WrapMsg(constant.ErrCodeUsernameTooShort, "Username terlalu pendek (minimal 3 karakter)", nil)
		}
		if len(rawValue) > 30 {
			return "", kernel.WrapMsg(constant.ErrCodeUsernameTooLong, "Username terlalu panjang (maksimal 30 karakter)", nil)
		}
		return rawValue, nil

	case constant.LoginIdentifierKindNIS:
		// Pass-through: NIS format is validated once by the caller (e.g.
		// kesantrian's own NIS value object) before it ever reaches
		// identity — identity trusts the value it's handed here, mirroring
		// sipon-api's own login_identity.go NIS branch.
		return rawValue, nil

	default:
		return "", kernel.WrapMsg(constant.ErrCodeLoginIdentifierUnknownKind, "Jenis identitas login tidak dikenali", nil)
	}
}
