package ports

type EmailSender interface {
	SendOTP(toEmail, username, otp string) error
	SendPasswordResetOTP(toEmail, username, otp string) error
}

type SMSSender interface {
	SendOTP(toPhone, otp string) error
}
